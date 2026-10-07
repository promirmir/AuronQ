//go:build windows

package main

import (
	"fmt"
	"runtime"
	"strings"
	"syscall"
	"unsafe"
)

type cudaDLLBackend struct {
	dll         *syscall.LazyDLL
	runProc     *syscall.LazyProc
	closeProc   *syscall.LazyProc
	name        string
	recommended int
}

func openCUDABackend(path string, device int) (gpuBackend, error) {
	dll := syscall.NewLazyDLL(path)
	if err := dll.Load(); err != nil {
		return nil, fmt.Errorf("load CUDA backend %s: %w", path, err)
	}
	initProc := dll.NewProc("aqm64_cuda_init")
	runProc := dll.NewProc("aqm64_cuda_run")
	closeProc := dll.NewProc("aqm64_cuda_shutdown")
	for name, proc := range map[string]*syscall.LazyProc{
		"aqm64_cuda_init": initProc,
		"aqm64_cuda_run": runProc,
		"aqm64_cuda_shutdown": closeProc,
	} {
		if err := proc.Find(); err != nil {
			return nil, fmt.Errorf("CUDA backend missing %s: %w", name, err)
		}
	}

	nameBuf := make([]byte, 256)
	errBuf := make([]byte, 512)
	var recommended int32
	r1, _, _ := initProc.Call(
		uintptr(device),
		uintptr(unsafe.Pointer(&recommended)),
		uintptr(unsafe.Pointer(&nameBuf[0])),
		uintptr(len(nameBuf)),
		uintptr(unsafe.Pointer(&errBuf[0])),
		uintptr(len(errBuf)),
	)
	runtime.KeepAlive(nameBuf)
	runtime.KeepAlive(errBuf)
	if int32(r1) != 0 {
		return nil, fmt.Errorf("CUDA init failed: %s", cString(errBuf))
	}
	if recommended < 1 {
		recommended = 1
	}
	return &cudaDLLBackend{
		dll: dll, runProc: runProc, closeProc: closeProc,
		name: cString(nameBuf), recommended: int(recommended),
	}, nil
}

func (b *cudaDLLBackend) Name() string { return b.name }
func (b *cudaDLLBackend) RecommendedBatch() int { return b.recommended }

func (b *cudaDLLBackend) Run(initial []uint64, count int) ([]uint64, error) {
	if count < 1 || len(initial) != count*256 {
		return nil, fmt.Errorf("invalid CUDA batch: count=%d initial_words=%d", count, len(initial))
	}
	out := make([]uint64, count*128)
	errBuf := make([]byte, 512)
	r1, _, _ := b.runProc.Call(
		uintptr(unsafe.Pointer(&initial[0])),
		uintptr(count),
		uintptr(unsafe.Pointer(&out[0])),
		uintptr(unsafe.Pointer(&errBuf[0])),
		uintptr(len(errBuf)),
	)
	runtime.KeepAlive(initial)
	runtime.KeepAlive(out)
	runtime.KeepAlive(errBuf)
	if int32(r1) != 0 {
		return nil, fmt.Errorf("CUDA batch failed: %s", cString(errBuf))
	}
	return out, nil
}

func (b *cudaDLLBackend) Close() error {
	if b == nil || b.closeProc == nil {
		return nil
	}
	b.closeProc.Call()
	return nil
}

func cString(b []byte) string {
	if i := strings.IndexByte(string(b), 0); i >= 0 {
		return string(b[:i])
	}
	return string(b)
}
