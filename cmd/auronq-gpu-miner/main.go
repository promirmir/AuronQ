package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	aq "auronq/internal/auronq"
	"auronq/internal/argon2pure"
)

const mainnetNetworkID = "44e62c2ace002a6660c14e252173c1aa303529c68e40c998e92da2b453f44f30b1e58c94d533587e2186004593fb856c433fcdb5418ed430ec8617e29529365c"

func defaultDLLPath() string {
	exe, err := os.Executable()
	if err != nil {
		return "auronq-aqm64-cuda.dll"
	}
	return filepath.Join(filepath.Dir(exe), "auronq-aqm64-cuda.dll")
}

func main() {
	nodeURL := flag.String("node", "http://127.0.0.1:18444", "AuronQ full-node URL")
	address := flag.String("address", "", "AURQ reward address")
	device := flag.Int("device", 0, "CUDA device index")
	devicesFlag := flag.String("devices", "", "comma-separated CUDA device indices or 'all'; overrides --device")
	multiChild := flag.Bool("multi-child", false, "internal multi-GPU child worker")
	batchFlag := flag.Int("batch", 0, "nonces per GPU batch (0 = automatic)")
	dllPath := flag.String("cuda-dll", defaultDLLPath(), "path to auronq-aqm64-cuda.dll")
	selfTest := flag.Bool("self-test", false, "compare one full AQM64 GPU result with the CPU reference")
	benchmark := flag.Bool("benchmark", false, "run an offline end-to-end AQM64 throughput benchmark")
	benchmarkSeconds := flag.Int("benchmark-seconds", 20, "approximate benchmark duration in seconds")
	autoTune := flag.Bool("auto-tune", false, "benchmark safe batch sizes and automatically select the fastest one before mining")
	autoTuneSeconds := flag.Int("auto-tune-seconds", 1, "autotune measurement time per batch size in seconds")
	thermalAuto := flag.Bool("thermal-auto", false, "automatically reduce/increase GPU batch to stay below a temperature target")
	thermalLimit := flag.Int("thermal-limit", 85, "hard GPU temperature limit in C; reaching it stops mining")
	thermalTarget := flag.Int("thermal-target", 0, "target GPU temperature in C (0 = thermal-limit minus 5 C)")
	noncePrefix := flag.Uint64("nonce-prefix", 0, "starting nonce prefix/base used to partition work between GPUs")
	flag.Parse()

	if runtime.GOOS != "windows" {
		fmt.Fprintln(os.Stderr, "AuronQ GPU Miner CUDA v0.1 currently targets Windows x64.")
		os.Exit(2)
	}

	if !*multiChild && *devicesFlag != "" {
		devices, err := resolveCUDADevices(*devicesFlag)
		if err != nil {
			fmt.Fprintln(os.Stderr, "CUDA devices:", err)
			os.Exit(2)
		}
		if len(devices) > 1 {
			err := runMultiGPU(devices, multiGPUOptions{
				Node:             *nodeURL,
				Address:          *address,
				Batch:            *batchFlag,
				DLLPath:          *dllPath,
				SelfTest:         *selfTest,
				Benchmark:        *benchmark,
				BenchmarkSeconds: *benchmarkSeconds,
				AutoTune:         *autoTune,
				AutoTuneSeconds:  *autoTuneSeconds,
				ThermalAuto:      *thermalAuto,
				ThermalLimit:     *thermalLimit,
				ThermalTarget:    *thermalTarget,
				NoncePrefix:      *noncePrefix,
			})
			if err != nil {
				fmt.Fprintln(os.Stderr, "MULTI-GPU FAILED:", err)
				os.Exit(1)
			}
			return
		}
		*device = devices[0]
	}

	backend, err := openCUDABackend(*dllPath, *device)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer backend.Close()

	batch := *batchFlag
	if batch <= 0 {
		batch = backend.RecommendedBatch()
	}
	if batch < 1 {
		batch = 1
	}
	if batch > 64 {
		batch = 64
	}

	fmt.Printf("AuronQ GPU Miner v0.3.3-alpha CUDA\n")
	fmt.Printf("GPU: %s\n", backend.Name())
	fmt.Printf("Batch: %d nonces\n", batch)

	if *selfTest {
		fmt.Println("Running full AQM64 GPU/CPU equivalence self-test...")
		if err := runSelfTest(backend); err != nil {
			fmt.Fprintln(os.Stderr, "SELF-TEST FAILED:", err)
			os.Exit(1)
		}
		fmt.Println("SELF-TEST OK")
		if !*benchmark && *address == "" {
			return
		}
	}

	if *autoTune {
		if *autoTuneSeconds < 1 {
			fmt.Fprintln(os.Stderr, "--auto-tune-seconds must be at least 1")
			os.Exit(2)
		}
		bestBatch, bestRate, err := runAutoTune(backend, time.Duration(*autoTuneSeconds)*time.Second)
		if err != nil {
			fmt.Fprintln(os.Stderr, "AUTOTUNE FAILED:", err)
			os.Exit(1)
		}
		batch = bestBatch
		fmt.Printf("AUTOTUNE OK best_batch=%d best=%.3f H/s\n", bestBatch, bestRate)
	}

	if *benchmark {
		if *benchmarkSeconds < 1 {
			fmt.Fprintln(os.Stderr, "--benchmark-seconds must be at least 1")
			os.Exit(2)
		}
		if err := runBenchmark(backend, batch, time.Duration(*benchmarkSeconds)*time.Second); err != nil {
			fmt.Fprintln(os.Stderr, "BENCHMARK FAILED:", err)
			os.Exit(1)
		}
		return
	}

	var thermal *thermalController
	if *thermalAuto {
		if *thermalLimit < 60 || *thermalLimit > 95 {
			fmt.Fprintln(os.Stderr, "--thermal-limit must be between 60 and 95 C")
			os.Exit(2)
		}
		target := *thermalTarget
		if target == 0 {
			target = *thermalLimit - 5
		}
		if target < 50 || target >= *thermalLimit {
			fmt.Fprintln(os.Stderr, "--thermal-target must be at least 50 C and below --thermal-limit")
			os.Exit(2)
		}
		thermal = newThermalController(*device, target, *thermalLimit, batch)
		fmt.Printf("THERMAL AUTO target=%dC limit=%dC batch_range=1..%d\n", target, *thermalLimit, batch)
	}

	if *address == "" {
		fmt.Fprintln(os.Stderr, "--address is required for mining")
		os.Exit(2)
	}
	netByte, _, _, err := aq.DecodeAddress(*address)
	if err != nil || netByte != aq.MainnetNetworkByte {
		fmt.Fprintln(os.Stderr, "--address must be a valid AuronQ Mainnet address")
		os.Exit(2)
	}

	client := aq.NewClient(*nodeURL)
	st, err := client.Status()
	if err != nil {
		fmt.Fprintln(os.Stderr, "node status:", err)
		os.Exit(1)
	}
	if st.NetworkID.String() != mainnetNetworkID {
		fmt.Fprintln(os.Stderr, "refusing to mine: node Network ID mismatch")
		os.Exit(1)
	}
	fmt.Printf("Node: %s height=%d peers=%d\n", *nodeURL, st.Height, st.Peers)

	if err := mineLoop(client, backend, *address, batch, *noncePrefix, thermal); err != nil {
		fmt.Fprintln(os.Stderr, "miner stopped:", err)
		os.Exit(1)
	}
}

func runSelfTest(backend gpuBackend) error {
	// First verify the raw Argon2id accelerator boundary.
	password := make([]byte, 64)
	salt := make([]byte, 32)
	for i := range password {
		password[i] = byte(i*7 + 3)
	}
	for i := range salt {
		salt[i] = byte(i*11 + 5)
	}
	initial, err := argon2pure.PrepareSingleLaneBlocks(
		password, salt, aq.AQM64TimeCost, aq.AQM64MemoryKiB, 64,
	)
	if err != nil {
		return err
	}
	gpuFinal, err := backend.Run(initial, 1)
	if err != nil {
		return err
	}
	gpuKey, err := argon2pure.ExtractSingleLaneKey(gpuFinal[:128], 64)
	if err != nil {
		return err
	}
	cpuKey := argon2pure.IDKey(
		password, salt, aq.AQM64TimeCost, aq.AQM64MemoryKiB, aq.AQM64Parallelism, 64,
	)
	if !bytes.Equal(gpuKey, cpuKey) {
		return fmt.Errorf("GPU Argon2id result does not match AuronQ CPU reference")
	}

	// Then verify the complete AQM64 pipeline on a deterministic synthetic
	// header: header serialization + SHAKE256 PRE + salt + GPU Argon2id +
	// SHAKE256 FINAL must match the canonical CPU PowHash byte-for-byte.
	header := aq.BlockHeader{
		Version:   aq.BlockVersion,
		PowAlgo:   aq.PowAlgorithmAQM64,
		Height:    12345,
		Timestamp: 1790951480,
		Target:    aq.PowLimit,
		Nonce:     0x0123456789abcdef,
	}
	for i := range header.PrevHash {
		header.PrevHash[i] = byte(i*3 + 1)
	}
	for i := range header.MerkleRoot {
		header.MerkleRoot[i] = byte(i*5 + 7)
	}
	p, err := prepareCandidate(header)
	if err != nil {
		return err
	}
	final, err := backend.Run(p.initial, 1)
	if err != nil {
		return err
	}
	gpuHash, err := finishCandidate(p.pre, final[:128])
	if err != nil {
		return err
	}
	cpuHash, err := aq.PowHash(header)
	if err != nil {
		return err
	}
	if gpuHash != cpuHash {
		return fmt.Errorf("full AQM64 GPU result does not match canonical CPU PowHash")
	}
	return nil
}

func runAutoTune(backend gpuBackend, perBatch time.Duration) (int, float64, error) {
	if perBatch <= 0 {
		perBatch = time.Second
	}
	candidates := []int{4, 8, 16, 24, 32, 40, 48, 56, 60, 64}
	recommended := backend.RecommendedBatch()
	if recommended >= 1 && recommended <= 64 {
		found := false
		for _, v := range candidates {
			if v == recommended {
				found = true
				break
			}
		}
		if !found {
			candidates = append(candidates, recommended)
			for i := len(candidates) - 1; i > 0 && candidates[i] < candidates[i-1]; i-- {
				candidates[i], candidates[i-1] = candidates[i-1], candidates[i]
			}
		}
	}

	var template aq.Block
	template.Header = aq.BlockHeader{
		Version:   aq.BlockVersion,
		PowAlgo:   aq.PowAlgorithmAQM64,
		Height:    54321,
		Timestamp: 1790951480,
		Target:    aq.PowLimit,
	}
	for i := range template.Header.PrevHash {
		template.Header.PrevHash[i] = byte(i*13 + 9)
	}
	for i := range template.Header.MerkleRoot {
		template.Header.MerkleRoot[i] = byte(i*17 + 11)
	}

	bestBatch := 0
	bestRate := 0.0
	fmt.Printf("AUTOTUNE start candidates=%v per_batch=%s\n", candidates, perBatch.Round(time.Second))
	for _, batch := range candidates {
		start := time.Now()
		deadline := start.Add(perBatch)
		var total uint64
		var nonce uint64
		valid := true

		for {
			prepared, initial, err := buildBatch(template, nonce, batch)
			if err != nil {
				return 0, 0, err
			}
			finals, err := backend.Run(initial, batch)
			if err != nil {
				fmt.Printf("AUTOTUNE batch=%d skipped: %v\n", batch, err)
				valid = false
				break
			}
			for i := 0; i < batch; i++ {
				if _, err := finishCandidate(prepared[i].pre, finals[i*128:(i+1)*128]); err != nil {
					return 0, 0, err
				}
			}
			total += uint64(batch)
			nonce += uint64(batch)
			if time.Now().After(deadline) && total > 0 {
				break
			}
		}

		if !valid || total == 0 {
			continue
		}
		elapsed := time.Since(start)
		rate := float64(total) / elapsed.Seconds()
		fmt.Printf("AUTOTUNE batch=%d rate=%.3f H/s\n", batch, rate)
		if bestBatch == 0 || rate > bestRate {
			bestBatch = batch
			bestRate = rate
		}
	}
	if bestBatch == 0 {
		return 0, 0, fmt.Errorf("no usable CUDA batch size found")
	}
	return bestBatch, bestRate, nil
}

func runBenchmark(backend gpuBackend, batch int, duration time.Duration) error {
	var template aq.Block
	template.Header = aq.BlockHeader{
		Version:   aq.BlockVersion,
		PowAlgo:   aq.PowAlgorithmAQM64,
		Height:    54321,
		Timestamp: 1790951480,
		Target:    aq.PowLimit,
	}
	for i := range template.Header.PrevHash {
		template.Header.PrevHash[i] = byte(i*13 + 9)
	}
	for i := range template.Header.MerkleRoot {
		template.Header.MerkleRoot[i] = byte(i*17 + 11)
	}

	fmt.Printf("Running offline AQM64 benchmark for about %s...\n", duration.Round(time.Second))
	start := time.Now()
	deadline := start.Add(duration)
	var total uint64
	var nonce uint64

	for {
		prepared, initial, err := buildBatch(template, nonce, batch)
		if err != nil {
			return err
		}
		finals, err := backend.Run(initial, batch)
		if err != nil {
			return err
		}
		for i := 0; i < batch; i++ {
			if _, err := finishCandidate(prepared[i].pre, finals[i*128:(i+1)*128]); err != nil {
				return err
			}
		}
		total += uint64(batch)
		nonce += uint64(batch)
		if time.Now().After(deadline) && total > 0 {
			break
		}
	}

	elapsed := time.Since(start)
	rate := float64(total) / elapsed.Seconds()
	fmt.Printf("BENCHMARK OK hashes=%d elapsed=%s avg=%.3f H/s batch=%d\n",
		total, elapsed.Round(time.Millisecond), rate, batch)
	return nil
}

func mineLoop(client *aq.Client, backend gpuBackend, address string, batch int, noncePrefix uint64, thermal *thermalController) error {
	var total uint64
	start := time.Now()
	lastReport := start
	var lastReportTotal uint64
	var nextNonce uint64

	for {
		template, err := client.Template(address)
		if err != nil {
			return fmt.Errorf("template: %w", err)
		}
		if template.Header.PowAlgo != aq.PowAlgorithmAQM64 {
			return fmt.Errorf("unsupported PoW algorithm %d", template.Header.PowAlgo)
		}
		nextNonce = noncePrefix
		fmt.Printf("Mining height %d target=%s nonce_base=%d\n", template.Header.Height, template.Header.Target.String(), noncePrefix)

		for {
			var thermalPause time.Duration
			if thermal != nil {
				newBatch, temp, pause, action, err := thermal.Adjust(batch)
				if err != nil {
					return err
				}
				if action != "" {
					fmt.Printf("THERMAL temp=%dC target=%dC limit=%dC batch=%d->%d pause=%s action=%s\n",
						temp, thermal.Target(), thermal.Limit(), batch, newBatch, pause, action)
				}
				batch = newBatch
				thermalPause = pause
			}

			st, err := client.Status()
			if err == nil && !aq.MiningTemplateCurrent(template, st) {
				fmt.Printf("Tip changed at height %d; refreshing template\n", st.Height)
				break
			}

			prepared, initial, err := buildBatch(template, nextNonce, batch)
			if err != nil {
				return err
			}
			finals, err := backend.Run(initial, batch)
			if err != nil {
				return err
			}

			total += uint64(batch)
			for i := 0; i < batch; i++ {
				hash, err := finishCandidate(prepared[i].pre, finals[i*128:(i+1)*128])
				if err != nil {
					return err
				}
				if !hashMeetsTarget(hash, template.Header.Target) {
					continue
				}

				found := template
				found.Header.Nonce = prepared[i].nonce
				resp, err := client.SubmitBlock(found)
				if err != nil {
					fmt.Printf("Candidate nonce=%d rejected: %v\n", prepared[i].nonce, err)
					break
				}
				elapsed := time.Since(start)
				rate := float64(total) / elapsed.Seconds()
				fmt.Printf("BLOCK FOUND height=%d hash=%s nonce=%d avg=%.2f H/s\n",
					resp.Height, resp.Hash.String(), prepared[i].nonce, rate)
				break
			}

			st, err = client.Status()
			if err == nil && !aq.MiningTemplateCurrent(template, st) {
				break
			}
			prevNonce := nextNonce
			nextNonce += uint64(batch)
			if nextNonce < prevNonce {
				break
			}

			if thermalPause > 0 {
				time.Sleep(thermalPause)
			}

			now := time.Now()
			if now.Sub(lastReport) >= time.Second {
				interval := now.Sub(lastReport)
				intervalHashes := total - lastReportTotal
				rate := 0.0
				if interval > 0 {
					rate = float64(intervalHashes) / interval.Seconds()
				}
				elapsed := now.Sub(start)
				avg := 0.0
				if elapsed > 0 {
					avg = float64(total) / elapsed.Seconds()
				}
				fmt.Printf("hashes=%d rate=%.2f H/s avg=%.2f H/s current_height=%d batch=%d\n",
					total, rate, avg, template.Header.Height, batch)
				lastReport = now
				lastReportTotal = total
			}
		}
	}
}
