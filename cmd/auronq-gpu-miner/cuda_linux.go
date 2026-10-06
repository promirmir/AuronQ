//go:build linux && cgo

package main

/*
#cgo LDFLAGS: -ldl
#include <dlfcn.h>
#include <stdint.h>
#include <stdio.h>
#include <string.h>

typedef int (*aq_init_fn)(int, int*, char*, int, char*, int);
typedef int (*aq_run_fn)(const uint64_t*, int, uint64_t*, char*, int);
typedef void (*aq_shutdown_fn)(void);

static void* aq_handle = NULL;
static aq_init_fn aq_init_ptr = NULL;
static aq_run_fn aq_run_ptr = NULL;
static aq_shutdown_fn aq_shutdown_ptr = NULL;
static char aq_loader_err[512];

static const char* aq_cuda_last_error(void) {
	return aq_loader_err;
}

static int aq_cuda_load(const char* path, int device, int* recommended, char* name, int name_cap, char* err, int err_cap) {
	aq_loader_err[0] = '\0';
	if (aq_handle != NULL) {
		snprintf(aq_loader_err, sizeof(aq_loader_err), "CUDA backend already loaded");
		return -1;
	}

	dlerror();
	aq_handle = dlopen(path, RTLD_NOW | RTLD_LOCAL);
	if (aq_handle == NULL) {
		const char* e = dlerror();
		snprintf(aq_loader_err, sizeof(aq_loader_err), "dlopen(%s): %s", path, e ? e : "unknown error");
		return -1;
	}

	dlerror();
	aq_init_ptr = (aq_init_fn)dlsym(aq_handle, "aqm64_cuda_init");
	aq_run_ptr = (aq_run_fn)dlsym(aq_handle, "aqm64_cuda_run");
	aq_shutdown_ptr = (aq_shutdown_fn)dlsym(aq_handle, "aqm64_cuda_shutdown");
	const char* symerr = dlerror();
	if (symerr != NULL || aq_init_ptr == NULL || aq_run_ptr == NULL || aq_shutdown_ptr == NULL) {
		snprintf(aq_loader_err, sizeof(aq_loader_err), "CUDA backend symbols: %s", symerr ? symerr : "missing symbol");
		dlclose(aq_handle);
		aq_handle = NULL;
		aq_init_ptr = NULL;
		aq_run_ptr = NULL;
		aq_shutdown_ptr = NULL;
		return -1;
	}

	int rc = aq_init_ptr(device, recommended, name, name_cap, err, err_cap);
	if (rc != 0) {
		aq_shutdown_ptr();
		dlclose(aq_handle);
		aq_handle = NULL;
		aq_init_ptr = NULL;
		aq_run_ptr = NULL;
		aq_shutdown_ptr = NULL;
	}
	return rc;
}

static int aq_cuda_run_bridge(const uint64_t* initial, int count, uint64_t* out, char* err, int err_cap) {
	if (aq_handle == NULL || aq_run_ptr == NULL) {
		snprintf(aq_loader_err, sizeof(aq_loader_err), "CUDA backend is not loaded");
		return -1;
	}
	return aq_run_ptr(initial, count, out, err, err_cap);
}

static void aq_cuda_close_bridge(void) {
	if (aq_handle != NULL) {
		if (aq_shutdown_ptr != NULL) {
			aq_shutdown_ptr();
		}
		dlclose(aq_handle);
	}
	aq_handle = NULL;
	aq_init_ptr = NULL;
	aq_run_ptr = NULL;
	aq_shutdown_ptr = NULL;
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

type cudaSOBackend struct {
	name        string
	recommended int
	closed      bool
}

func openCUDABackend(path string, device int) (gpuBackend, error) {
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))

	nameBuf := make([]byte, 256)
	errBuf := make([]byte, 512)
	var recommended C.int

	rc := C.aq_cuda_load(
		cpath,
		C.int(device),
		&recommended,
		(*C.char)(unsafe.Pointer(&nameBuf[0])),
		C.int(len(nameBuf)),
		(*C.char)(unsafe.Pointer(&errBuf[0])),
		C.int(len(errBuf)),
	)
	if rc != 0 {
		msg := cStringLinux(errBuf)
		if msg == "" {
			msg = C.GoString(C.aq_cuda_last_error())
		}
		if msg == "" {
			msg = fmt.Sprintf("CUDA init failed with code %d", int(rc))
		}
		return nil, fmt.Errorf("%s", msg)
	}

	rec := int(recommended)
	if rec < 1 {
		rec = 1
	}
	return &cudaSOBackend{name: cStringLinux(nameBuf), recommended: rec}, nil
}

func (b *cudaSOBackend) Name() string { return b.name }
func (b *cudaSOBackend) RecommendedBatch() int { return b.recommended }

func (b *cudaSOBackend) Run(initial []uint64, count int) ([]uint64, error) {
	if b == nil || b.closed {
		return nil, fmt.Errorf("CUDA backend is closed")
	}
	if count < 1 || len(initial) != count*256 {
		return nil, fmt.Errorf("invalid CUDA batch: count=%d initial_words=%d", count, len(initial))
	}

	out := make([]uint64, count*128)
	errBuf := make([]byte, 512)
	rc := C.aq_cuda_run_bridge(
		(*C.uint64_t)(unsafe.Pointer(&initial[0])),
		C.int(count),
		(*C.uint64_t)(unsafe.Pointer(&out[0])),
		(*C.char)(unsafe.Pointer(&errBuf[0])),
		C.int(len(errBuf)),
	)
	if rc != 0 {
		msg := cStringLinux(errBuf)
		if msg == "" {
			msg = C.GoString(C.aq_cuda_last_error())
		}
		return nil, fmt.Errorf("CUDA batch failed: %s", msg)
	}
	return out, nil
}

func (b *cudaSOBackend) Close() error {
	if b == nil || b.closed {
		return nil
	}
	C.aq_cuda_close_bridge()
	b.closed = true
	return nil
}

func cStringLinux(b []byte) string {
	for i, v := range b {
		if v == 0 {
			return string(b[:i])
		}
	}
	return string(b)
}
