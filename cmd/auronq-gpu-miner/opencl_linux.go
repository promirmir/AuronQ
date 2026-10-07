//go:build linux && cgo

package main

/*
#cgo LDFLAGS: -ldl
#include <stdint.h>
#include <stdlib.h>
#include <stdio.h>
#include <dlfcn.h>

typedef int (*aq_opencl_init_fn)(int, int*, char*, int, char*, int);
typedef int (*aq_opencl_run_fn)(const uint64_t*, int, uint64_t*, char*, int);
typedef void (*aq_opencl_shutdown_fn)(void);

static void* aq_opencl_handle = NULL;
static aq_opencl_init_fn aq_opencl_init_ptr = NULL;
static aq_opencl_run_fn aq_opencl_run_ptr = NULL;
static aq_opencl_shutdown_fn aq_opencl_shutdown_ptr = NULL;
static char aq_opencl_loader_err[512];

static const char* aq_opencl_last_error(void) {
	return aq_opencl_loader_err;
}

static int aq_opencl_load(const char* path, int device, int* recommended, char* name, int name_cap, char* err, int err_cap) {
	aq_opencl_loader_err[0] = 0;
	aq_opencl_handle = dlopen(path, RTLD_NOW | RTLD_LOCAL);
	if (aq_opencl_handle == NULL) {
		snprintf(aq_opencl_loader_err, sizeof(aq_opencl_loader_err), "%s", dlerror());
		return -1;
	}
	dlerror();
	aq_opencl_init_ptr = (aq_opencl_init_fn)dlsym(aq_opencl_handle, "aqm64_opencl_init");
	aq_opencl_run_ptr = (aq_opencl_run_fn)dlsym(aq_opencl_handle, "aqm64_opencl_run");
	aq_opencl_shutdown_ptr = (aq_opencl_shutdown_fn)dlsym(aq_opencl_handle, "aqm64_opencl_shutdown");
	const char* symerr = dlerror();
	if (symerr != NULL || aq_opencl_init_ptr == NULL || aq_opencl_run_ptr == NULL || aq_opencl_shutdown_ptr == NULL) {
		snprintf(aq_opencl_loader_err, sizeof(aq_opencl_loader_err), "OpenCL backend symbols: %s", symerr ? symerr : "missing symbol");
		dlclose(aq_opencl_handle);
		aq_opencl_handle = NULL;
		aq_opencl_init_ptr = NULL;
		aq_opencl_run_ptr = NULL;
		aq_opencl_shutdown_ptr = NULL;
		return -1;
	}

	int rc = aq_opencl_init_ptr(device, recommended, name, name_cap, err, err_cap);
	if (rc != 0) {
		aq_opencl_shutdown_ptr();
		dlclose(aq_opencl_handle);
		aq_opencl_handle = NULL;
		aq_opencl_init_ptr = NULL;
		aq_opencl_run_ptr = NULL;
		aq_opencl_shutdown_ptr = NULL;
	}
	return rc;
}

static int aq_opencl_run_bridge(const uint64_t* initial, int count, uint64_t* out, char* err, int err_cap) {
	if (aq_opencl_handle == NULL || aq_opencl_run_ptr == NULL) {
		snprintf(aq_opencl_loader_err, sizeof(aq_opencl_loader_err), "OpenCL backend is not loaded");
		return -1;
	}
	return aq_opencl_run_ptr(initial, count, out, err, err_cap);
}

static void aq_opencl_close_bridge(void) {
	if (aq_opencl_handle != NULL) {
		if (aq_opencl_shutdown_ptr != NULL) aq_opencl_shutdown_ptr();
		dlclose(aq_opencl_handle);
	}
	aq_opencl_handle = NULL;
	aq_opencl_init_ptr = NULL;
	aq_opencl_run_ptr = NULL;
	aq_opencl_shutdown_ptr = NULL;
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

type openclSOBackend struct {
	name        string
	recommended int
	closed      bool
}

func openOpenCLBackend(path string, device int) (gpuBackend, error) {
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))

	nameBuf := make([]byte, 256)
	errBuf := make([]byte, 4096)
	var recommended C.int

	rc := C.aq_opencl_load(
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
			msg = C.GoString(C.aq_opencl_last_error())
		}
		if msg == "" {
			msg = fmt.Sprintf("OpenCL init failed with code %d", int(rc))
		}
		return nil, fmt.Errorf("%s", msg)
	}

	rec := int(recommended)
	if rec < 1 {
		rec = 1
	}
	return &openclSOBackend{name: cStringLinux(nameBuf), recommended: rec}, nil
}

func (b *openclSOBackend) Name() string { return b.name }
func (b *openclSOBackend) RecommendedBatch() int { return b.recommended }

func (b *openclSOBackend) Run(initial []uint64, count int) ([]uint64, error) {
	if b == nil || b.closed {
		return nil, fmt.Errorf("OpenCL backend is closed")
	}
	if count < 1 || len(initial) != count*256 {
		return nil, fmt.Errorf("invalid OpenCL batch: count=%d initial_words=%d", count, len(initial))
	}
	out := make([]uint64, count*128)
	errBuf := make([]byte, 4096)
	rc := C.aq_opencl_run_bridge(
		(*C.uint64_t)(unsafe.Pointer(&initial[0])),
		C.int(count),
		(*C.uint64_t)(unsafe.Pointer(&out[0])),
		(*C.char)(unsafe.Pointer(&errBuf[0])),
		C.int(len(errBuf)),
	)
	if rc != 0 {
		msg := cStringLinux(errBuf)
		if msg == "" {
			msg = C.GoString(C.aq_opencl_last_error())
		}
		return nil, fmt.Errorf("OpenCL batch failed: %s", msg)
	}
	return out, nil
}

func (b *openclSOBackend) Close() error {
	if b == nil || b.closed {
		return nil
	}
	C.aq_opencl_close_bridge()
	b.closed = true
	return nil
}
