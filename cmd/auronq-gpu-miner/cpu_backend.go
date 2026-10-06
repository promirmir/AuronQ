package main

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"

	aq "auronq/internal/auronq"
	"auronq/internal/argon2pure"
)

type cpuAQM64Backend struct {
	threads    int
	workspaces []*argon2pure.SingleLaneWorkspace
}

func safeCPUThreads(requested int) int {
	cpus := runtime.NumCPU()
	if cpus < 1 {
		cpus = 1
	}
	if requested > 0 {
		if requested > cpus {
			requested = cpus
		}
		if requested > 16 {
			requested = 16
		}
		if requested < 1 {
			requested = 1
		}
		return requested
	}

	// Conservative universal default: about one quarter of the logical CPUs,
	// never more than two simultaneous 64 MiB AQM64 scratchpads. Without a
	// trustworthy cross-vendor CPU temperature sensor, AUTO deliberately leaves
	// substantial thermal and interactive headroom on laptops and small PCs.
	n := cpus / 4
	if n < 1 {
		n = 1
	}
	if n > 2 {
		n = 2
	}
	return n
}

func openCPUBackend(threads int) (gpuBackend, error) {
	threads = safeCPUThreads(threads)
	workspaces := make([]*argon2pure.SingleLaneWorkspace, threads)
	for i := range workspaces {
		workspaces[i] = argon2pure.NewSingleLaneWorkspace(aq.AQM64MemoryKiB)
	}
	return &cpuAQM64Backend{threads: threads, workspaces: workspaces}, nil
}

func (b *cpuAQM64Backend) Name() string {
	if b == nil {
		return "CPU"
	}
	return fmt.Sprintf("CPU AQM64 (%d safe threads)", b.threads)
}

func (b *cpuAQM64Backend) RecommendedBatch() int {
	if b == nil || b.threads < 1 {
		return 1
	}
	return b.threads
}

func (b *cpuAQM64Backend) Run(initial []uint64, count int) ([]uint64, error) {
	if b == nil || b.threads < 1 || len(b.workspaces) != b.threads {
		return nil, fmt.Errorf("CPU backend is not initialized")
	}
	if count < 1 || len(initial) != count*256 {
		return nil, fmt.Errorf("invalid CPU AQM64 batch: count=%d initial_words=%d", count, len(initial))
	}

	out := make([]uint64, count*128)
	var next atomic.Int64
	var firstErr error
	var errMu sync.Mutex
	var wg sync.WaitGroup

	workers := b.threads
	if workers > count {
		workers = count
	}
	wg.Add(workers)
	for worker := 0; worker < workers; worker++ {
		ws := b.workspaces[worker]
		go func() {
			defer wg.Done()
			for {
				idx := int(next.Add(1) - 1)
				if idx >= count {
					return
				}
				finalBlock, err := ws.ProcessPreparedSingleLane(
					initial[idx*256:(idx+1)*256],
					aq.AQM64TimeCost,
				)
				if err != nil {
					errMu.Lock()
					if firstErr == nil {
						firstErr = err
					}
					errMu.Unlock()
					return
				}
				copy(out[idx*128:(idx+1)*128], finalBlock)
			}
		}()
	}
	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
	}
	return out, nil
}

func (b *cpuAQM64Backend) Close() error {
	if b != nil {
		b.workspaces = nil
	}
	return nil
}
