package auronq

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func miningWatchTemplate(tip Hash) Block {
	return Block{Header: BlockHeader{
		PowAlgo:  PowAlgorithmAQM64,
		Height:   11,
		PrevHash: tip,
		Target:   Target{},
	}}
}

func TestMiningTemplateCurrentChecksHeightAndTip(t *testing.T) {
	var tip, other Hash
	tip[0] = 1
	other[0] = 2
	tpl := miningWatchTemplate(tip)

	if !MiningTemplateCurrent(tpl, Status{Height: 10, Tip: tip}) {
		t.Fatal("matching canonical tip marked stale")
	}
	if MiningTemplateCurrent(tpl, Status{Height: 11, Tip: tip}) {
		t.Fatal("advanced chain did not mark template stale")
	}
	if MiningTemplateCurrent(tpl, Status{Height: 10, Tip: other}) {
		t.Fatal("same-height reorg did not mark template stale")
	}
}

func TestMineRemoteTemplateCancelsWhenTipChanges(t *testing.T) {
	var tip, other Hash
	tip[0] = 1
	other[0] = 2
	tpl := miningWatchTemplate(tip)

	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/status" {
			http.NotFound(w, r)
			return
		}
		n := calls.Add(1)
		status := Status{Height: 10, Tip: tip}
		if n >= 2 {
			status.Tip = other
		}
		_ = json.NewEncoder(w).Encode(status)
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	started := time.Now()
	_, err := MineRemoteTemplate(ctx, NewClient(srv.URL), tpl, 1, 5*time.Millisecond, nil)
	if !errors.Is(err, ErrMiningTemplateStale) {
		t.Fatalf("error=%v want ErrMiningTemplateStale", err)
	}
	if calls.Load() < 2 {
		t.Fatalf("tip watcher made only %d status request(s)", calls.Load())
	}
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("stale work was not cancelled promptly: %v", elapsed)
	}
}

func TestMineRemoteTemplateKeepsCurrentTipUntilCallerCancels(t *testing.T) {
	var tip Hash
	tip[0] = 1
	tpl := miningWatchTemplate(tip)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/status" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(Status{Height: 10, Tip: tip})
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	_, err := MineRemoteTemplate(ctx, NewClient(srv.URL), tpl, 1, 5*time.Millisecond, nil)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("error=%v want context deadline exceeded", err)
	}
}
