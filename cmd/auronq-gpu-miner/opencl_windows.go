//go:build windows

package main

import (
	"fmt"
	"runtime"
	"strings"
	"syscall"
	"unsafe"
)

type openclDLLBackend struct {
	dll         *syscall.LazyDLL
	runProc     *syscall.LazyProc
	closeProc   *syscall.LazyProc
	name        string
	recommended int
}

func openOpenCLBackend(path string, device int) (gpuBackend, error) {
	dll := syscall.NewLazyDLL(path)
	if err := dll.Load(); err != nil {
		return nil, fmt.Errorf("load OpenCL backend %s: %w", path, err)
	}
	initProc := dll.NewProc("aqm64_opencl_init")
	runProc := dll.NewProc("aqm64_opencl_run")
	closeProc := dll.NewProc("aqm64_opencl_shutdown")
	for name, proc := range map[string]*syscall.LazyProc{
		"aqm64_opencl_init": initProc,
		"aqm64_opencl_run": runProc,
		"aqm64_opencl_shutdown": closeProc,
	} {
		if err := proc.Find(); err != nil {
			return nil, fmt.Errorf("OpenCL backend missing %s: %w", name, err)
		}
	}

	nameBuf := make([]byte, 256)
	errBuf := make([]byte, 4096)
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
		return nil, fmt.Errorf("OpenCL init failed: %s", openCLCString(errBuf))
	}
	if recommended < 1 {
		recommended = 1
	}
	return &openclDLLBackend{
		dll: dll, runProc: runProc, closeProc: closeProc,
		name: openCLCString(nameBuf), recommended: int(recommended),
	}, nil
}

func (b *openclDLLBackend) Name() string { return b.name }
func (b *openclDLLBackend) RecommendedBatch() int { return b.recommended }

func (b *openclDLLBackend) Run(initial []uint64, count int) ([]uint64, error) {
	if count < 1 || len(initial) != count*256 {
		return nil, fmt.Errorf("invalid OpenCL batch: count=%d initial_words=%d", count, len(initial))
	}
	out := make([]uint64, count*128)
	errBuf := make([]byte, 4096)
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
		return nil, fmt.Errorf("OpenCL batch failed: %s", openCLCString(errBuf))
	}
	return out, nil
}

func (b *openclDLLBackend) Close() error {
	if b == nil || b.closeProc == nil {
		return nil
	}
	b.closeProc.Call()
	return nil
}

func openCLCString(b []byte) string {
	if i := strings.IndexByte(string(b), 0); i >= 0 {
		return string(b[:i])
	}
	return string(b)
}
