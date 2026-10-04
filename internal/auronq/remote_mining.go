package auronq

import (
	"context"
	"errors"
	"time"
)

var ErrMiningTemplateStale = errors.New("mining template is stale")

// MiningTemplateCurrent reports whether template still extends the node's
// currently selected canonical tip. Comparing both height and PrevHash catches
// ordinary forward progress as well as a same-height reorganization.
func MiningTemplateCurrent(template Block, status Status) bool {
	if template.Header.Height == 0 {
		return false
	}
	return status.Height == template.Header.Height-1 && status.Tip == template.Header.PrevHash
}

// MineRemoteTemplate mines a template while watching the connected node's
// canonical tip. If another miner advances the chain or a reorg changes the tip,
// the stale work is cancelled and ErrMiningTemplateStale is returned so callers
// can fetch a fresh template immediately.
func MineRemoteTemplate(
	ctx context.Context,
	client *Client,
	template Block,
	threads int,
	pollInterval time.Duration,
	progress func(uint64, time.Duration),
) (MineResult, error) {
	if client == nil {
		return MineResult{}, errors.New("nil mining client")
	}
	if pollInterval <= 0 {
		pollInterval = time.Second
	}

	mineCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	stale := make(chan struct{}, 1)
	watcherDone := make(chan struct{})

	markStale := func() {
		select {
		case stale <- struct{}{}:
		default:
		}
		cancel()
	}

	go func() {
		defer close(watcherDone)
		ticker := time.NewTicker(pollInterval)
		defer ticker.Stop()

		check := func() bool {
			status, err := client.StatusContext(mineCtx)
			if err != nil {
				// A temporary status failure must not destroy otherwise valid
				// work. Submit will still fail safely if the node is unavailable.
				return false
			}
			if !MiningTemplateCurrent(template, status) {
				markStale()
				return true
			}
			return false
		}

		if check() {
			return
		}

		for {
			select {
			case <-mineCtx.Done():
				return
			case <-ticker.C:
				if check() {
					return
				}
			}
		}
	}()

	result, err := MineParallel(mineCtx, template, threads, progress)
	cancel()
	<-watcherDone

	select {
	case <-stale:
		return MineResult{}, ErrMiningTemplateStale
	default:
	}

	if ctx.Err() != nil {
		return MineResult{}, ctx.Err()
	}
	return result, err
}
