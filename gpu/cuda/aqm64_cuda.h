#pragma once
#include <stdint.h>

#ifdef _WIN32
#define AQ_EXPORT extern "C" __declspec(dllexport)
#else
#define AQ_EXPORT extern "C" __attribute__((visibility("default")))
#endif

// Initializes the selected CUDA device and returns a conservative batch size.
// Returns 0 on success. Error text is written to err on failure.
AQ_EXPORT int aqm64_cuda_init(
    int device,
    int* recommended_batch,
    char* device_name,
    int device_name_cap,
    char* err,
    int err_cap);

// Processes count independent AQM64 Argon2id lanes.
// initial_blocks: count * 256 uint64 words (block0 || block1 per candidate).
// final_blocks:   count * 128 uint64 words (last Argon2 memory block per candidate).
AQ_EXPORT int aqm64_cuda_run(
    const uint64_t* initial_blocks,
    int count,
    uint64_t* final_blocks,
    char* err,
    int err_cap);

AQ_EXPORT void aqm64_cuda_shutdown();
