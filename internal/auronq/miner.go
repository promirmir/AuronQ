package auronq

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type MineResult struct {
	Block    Block
	Hashes   uint64
	Duration time.Duration
}

func MineParallel(ctx context.Context, template Block, threads int, progress func(uint64, time.Duration)) (MineResult, error) {
	if threads <= 0 {
		threads = runtime.NumCPU()
	}
	if threads < 1 {
		threads = 1
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	start := time.Now()
	var hashes atomic.Uint64
	found := make(chan Block, 1)
	var wg sync.WaitGroup
	if template.Header.PowAlgo != PowAlgorithmAQM64 {
		return MineResult{}, errors.New("unsupported proof-of-work algorithm")
	}
	target := template.Header.Target.Big()
	for worker := 0; worker < threads; worker++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			hasher := newAQM64Hasher()
			b := template
			b.Header.Nonce = uint64(id)
			step := uint64(threads)
			local := uint64(0)
			for {
				select {
				case <-ctx.Done():
					hashes.Add(local)
					return
				default:
				}
				h := hasher.hash(b.Header)
				local++
				if h.Big().Cmp(target) <= 0 {
					hashes.Add(local)
					select {
					case found <- b:
						cancel()
					default:
					}
					return
				}
				b.Header.Nonce += step
				if local%16 == 0 {
					hashes.Add(16)
					local -= 16
				}
				if b.Header.Nonce < step {
					b.Header.Timestamp = time.Now().Unix()
					b.Header.MerkleRoot = MerkleRoot(b.Transactions)
				}
			}
		}(worker)
	}
	var tick <-chan time.Time
	var ticker *time.Ticker
	if progress != nil {
		ticker = time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		tick = ticker.C
	}
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	for {
		select {
		case b := <-found:
			<-done
			return MineResult{b, hashes.Load(), time.Since(start)}, nil
		case <-ctx.Done():
			<-done
			select {
			case b := <-found:
				return MineResult{b, hashes.Load(), time.Since(start)}, nil
			default:
				return MineResult{}, errors.New("mining cancelled")
			}
		case <-done:
			select {
			case b := <-found:
				return MineResult{b, hashes.Load(), time.Since(start)}, nil
			default:
				return MineResult{}, errors.New("mining stopped")
			}
		case <-tick:
			progress(hashes.Load(), time.Since(start))
		}
	}
}