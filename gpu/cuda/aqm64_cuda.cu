#include "aqm64_cuda.h"

#include <cuda_runtime.h>
#include <algorithm>
#include <cstdio>
#include <cstring>

namespace {

using u64 = unsigned long long;

constexpr int kBlockWords = 128;
constexpr int kInitialWords = 256;
constexpr int kMemoryBlocks = 64 * 1024; // 64 MiB / 1 KiB
constexpr int kSyncPoints = 4;
constexpr int kSegments = kMemoryBlocks / kSyncPoints;
constexpr int kTimeCost = 2;
constexpr size_t kBytesPerCandidate =
    static_cast<size_t>(kMemoryBlocks) * kBlockWords * sizeof(u64);

int g_device = -1;
int g_capacity = 0;
u64* g_workspace = nullptr;
u64* g_initial = nullptr;
u64* g_final = nullptr;

void set_error(char* err, int cap, const char* text) {
    if (!err || cap <= 0) return;
    std::snprintf(err, static_cast<size_t>(cap), "%s", text ? text : "unknown CUDA error");
    err[cap - 1] = '\0';
}

void set_cuda_error(char* err, int cap, const char* where, cudaError_t code) {
    if (!err || cap <= 0) return;
    std::snprintf(err, static_cast<size_t>(cap), "%s: %s", where, cudaGetErrorString(code));
    err[cap - 1] = '\0';
}

void free_buffers() {
    if (g_workspace) cudaFree(g_workspace);
    if (g_initial) cudaFree(g_initial);
    if (g_final) cudaFree(g_final);
    g_workspace = nullptr;
    g_initial = nullptr;
    g_final = nullptr;
    g_capacity = 0;
}

int ensure_capacity(int count, char* err, int err_cap) {
    if (count <= g_capacity && g_workspace && g_initial && g_final) return 0;
    free_buffers();

    const size_t workspace_bytes = static_cast<size_t>(count) * kBytesPerCandidate;
    const size_t initial_bytes = static_cast<size_t>(count) * kInitialWords * sizeof(u64);
    const size_t final_bytes = static_cast<size_t>(count) * kBlockWords * sizeof(u64);

    cudaError_t rc = cudaMalloc(reinterpret_cast<void**>(&g_workspace), workspace_bytes);
    if (rc != cudaSuccess) {
        set_cuda_error(err, err_cap, "cudaMalloc workspace", rc);
        free_buffers();
        return 1;
    }
    rc = cudaMalloc(reinterpret_cast<void**>(&g_initial), initial_bytes);
    if (rc != cudaSuccess) {
        set_cuda_error(err, err_cap, "cudaMalloc initial", rc);
        free_buffers();
        return 1;
    }
    rc = cudaMalloc(reinterpret_cast<void**>(&g_final), final_bytes);
    if (rc != cudaSuccess) {
        set_cuda_error(err, err_cap, "cudaMalloc final", rc);
        free_buffers();
        return 1;
    }
    g_capacity = count;
    return 0;
}

__device__ __forceinline__ u64 rotr64(u64 x, int n) {
    return (x >> n) | (x << (64 - n));
}

__device__ __forceinline__ u64 blamka(u64 x, u64 y) {
    return x + y + 2ULL * static_cast<u64>(static_cast<unsigned int>(x)) *
                         static_cast<u64>(static_cast<unsigned int>(y));
}

__device__ __forceinline__ void gmix(u64& a, u64& b, u64& c, u64& d) {
    a = blamka(a, b);
    d = rotr64(d ^ a, 32);
    c = blamka(c, d);
    b = rotr64(b ^ c, 24);
    a = blamka(a, b);
    d = rotr64(d ^ a, 16);
    c = blamka(c, d);
    b = rotr64(b ^ c, 63);
}

__device__ __forceinline__ void round16(u64 v[16]) {
    gmix(v[0], v[4], v[8],  v[12]);
    gmix(v[1], v[5], v[9],  v[13]);
    gmix(v[2], v[6], v[10], v[14]);
    gmix(v[3], v[7], v[11], v[15]);
    gmix(v[0], v[5], v[10], v[15]);
    gmix(v[1], v[6], v[11], v[12]);
    gmix(v[2], v[7], v[8],  v[13]);
    gmix(v[3], v[4], v[9],  v[14]);
}

// Exact port of AuronQ internal/argon2pure processBlock/processBlockXOR.
// Every CUDA block handles one Argon2 candidate. 128 threads cooperate on
// each 1 KiB Argon2 block; eight threads execute the row/column permutations.
__device__ void compress_block(
    u64* out,
    const u64* in1,
    const u64* in2,
    bool xor_old,
    u64* r,
    u64* z)
{
    const int t = threadIdx.x;
    r[t] = in1[t] ^ in2[t];
    z[t] = r[t];
    __syncthreads();

    if (t < 8) {
        u64 v[16];
#pragma unroll
        for (int j = 0; j < 16; ++j) v[j] = z[t * 16 + j];
        round16(v);
#pragma unroll
        for (int j = 0; j < 16; ++j) z[t * 16 + j] = v[j];
    }
    __syncthreads();

    if (t < 8) {
        u64 v[16];
#pragma unroll
        for (int j = 0; j < 8; ++j) {
            v[2 * j]     = z[16 * j + 2 * t];
            v[2 * j + 1] = z[16 * j + 2 * t + 1];
        }
        round16(v);
#pragma unroll
        for (int j = 0; j < 8; ++j) {
            z[16 * j + 2 * t]     = v[2 * j];
            z[16 * j + 2 * t + 1] = v[2 * j + 1];
        }
    }
    __syncthreads();

    if (xor_old) {
        out[t] ^= r[t] ^ z[t];
    } else {
        out[t] = r[t] ^ z[t];
    }
    __syncthreads();
}

__device__ __forceinline__ unsigned int index_alpha_single_lane(
    u64 random,
    unsigned int pass,
    unsigned int slice,
    unsigned int index)
{
    u64 m;
    u64 s;
    if (pass == 0) {
        m = static_cast<u64>(slice) * kSegments + index;
        s = 0;
    } else {
        m = static_cast<u64>(3 * kSegments) + index;
        s = static_cast<u64>((slice + 1) % kSyncPoints) * kSegments;
    }
    // With one lane refLane == lane always, so AuronQ's generic indexAlpha
    // always subtracts one here.
    --m;

    u64 p = random & 0xffffffffULL;
    p = (p * p) >> 32;
    p = (p * m) >> 32;
    return static_cast<unsigned int>((s + m - (p + 1)) % kMemoryBlocks);
}

__global__ void argon2id_aqm64_kernel(
    u64* workspace,
    const u64* initial,
    int count,
    u64* final_out)
{
    const int candidate = blockIdx.x;
    if (candidate >= count || blockDim.x != kBlockWords) return;
    const int t = threadIdx.x;

    u64* B = workspace +
        static_cast<size_t>(candidate) * kMemoryBlocks * kBlockWords;
    const u64* init = initial + static_cast<size_t>(candidate) * kInitialWords;

    __shared__ u64 r[kBlockWords];
    __shared__ u64 z[kBlockWords];
    __shared__ u64 addresses[kBlockWords];
    __shared__ u64 input[kBlockWords];
    __shared__ u64 zero[kBlockWords];
    __shared__ u64 random_word;
    __shared__ unsigned int ref_offset;

    B[t] = init[t];
    B[kBlockWords + t] = init[kBlockWords + t];
    __syncthreads();

    for (unsigned int pass = 0; pass < kTimeCost; ++pass) {
        for (unsigned int slice = 0; slice < kSyncPoints; ++slice) {
            const bool independent = (pass == 0 && slice < kSyncPoints / 2);

            addresses[t] = 0;
            input[t] = 0;
            zero[t] = 0;
            __syncthreads();

            if (independent && t == 0) {
                input[0] = pass;
                input[1] = 0; // lane
                input[2] = slice;
                input[3] = kMemoryBlocks;
                input[4] = kTimeCost;
                input[5] = 2; // Argon2id
            }
            __syncthreads();

            unsigned int index = 0;
            if (pass == 0 && slice == 0) {
                index = 2;
                if (t == 0) ++input[6];
                __syncthreads();
                compress_block(addresses, input, zero, false, r, z);
                compress_block(addresses, addresses, zero, false, r, z);
            }

            unsigned int offset = slice * kSegments + index;
            while (index < kSegments) {
                unsigned int prev = offset - 1;
                if (index == 0 && slice == 0) prev += kMemoryBlocks;

                if (independent && (index % kBlockWords) == 0) {
                    if (t == 0) ++input[6];
                    __syncthreads();
                    compress_block(addresses, input, zero, false, r, z);
                    compress_block(addresses, addresses, zero, false, r, z);
                }

                if (t == 0) {
                    random_word = independent
                        ? addresses[index % kBlockWords]
                        : B[static_cast<size_t>(prev) * kBlockWords];
                    ref_offset = index_alpha_single_lane(random_word, pass, slice, index);
                }
                __syncthreads();

                compress_block(
                    B + static_cast<size_t>(offset) * kBlockWords,
                    B + static_cast<size_t>(prev) * kBlockWords,
                    B + static_cast<size_t>(ref_offset) * kBlockWords,
                    pass != 0,
                    r,
                    z);

                ++index;
                ++offset;
            }
        }
    }

    final_out[static_cast<size_t>(candidate) * kBlockWords + t] =
        B[static_cast<size_t>(kMemoryBlocks - 1) * kBlockWords + t];
}

} // namespace

AQ_EXPORT int aqm64_cuda_init(
    int device,
    int* recommended_batch,
    char* device_name,
    int device_name_cap,
    char* err,
    int err_cap)
{
    aqm64_cuda_shutdown();

    int count = 0;
    cudaError_t rc = cudaGetDeviceCount(&count);
    if (rc != cudaSuccess) {
        set_cuda_error(err, err_cap, "cudaGetDeviceCount", rc);
        return 1;
    }
    if (device < 0 || device >= count) {
        set_error(err, err_cap, "requested CUDA device does not exist");
        return 1;
    }
    rc = cudaSetDevice(device);
    if (rc != cudaSuccess) {
        set_cuda_error(err, err_cap, "cudaSetDevice", rc);
        return 1;
    }

    cudaDeviceProp prop{};
    rc = cudaGetDeviceProperties(&prop, device);
    if (rc != cudaSuccess) {
        set_cuda_error(err, err_cap, "cudaGetDeviceProperties", rc);
        return 1;
    }

    size_t free_bytes = 0;
    size_t total_bytes = 0;
    rc = cudaMemGetInfo(&free_bytes, &total_bytes);
    if (rc != cudaSuccess) {
        set_cuda_error(err, err_cap, "cudaMemGetInfo", rc);
        return 1;
    }

    // Keep substantial VRAM headroom for the desktop/display driver and CUDA.
    const size_t usable = free_bytes * 60 / 100;
    int memory_lanes = static_cast<int>(usable / kBytesPerCandidate);
    memory_lanes = std::max(1, memory_lanes);
    int occupancy_lanes = std::max(1, prop.multiProcessorCount * 2);
    int rec = std::min({64, memory_lanes, occupancy_lanes});

    if (recommended_batch) *recommended_batch = rec;
    if (device_name && device_name_cap > 0) {
        std::snprintf(device_name, static_cast<size_t>(device_name_cap),
            "%s (SM %d.%d, %d SM, %.1f GiB VRAM)",
            prop.name, prop.major, prop.minor, prop.multiProcessorCount,
            static_cast<double>(total_bytes) / (1024.0 * 1024.0 * 1024.0));
        device_name[device_name_cap - 1] = '\0';
    }
    if (err && err_cap > 0) err[0] = '\0';
    g_device = device;
    return 0;
}

AQ_EXPORT int aqm64_cuda_run(
    const uint64_t* initial_blocks,
    int count,
    uint64_t* final_blocks,
    char* err,
    int err_cap)
{
    if (g_device < 0) {
        set_error(err, err_cap, "CUDA backend is not initialized");
        return 1;
    }
    if (!initial_blocks || !final_blocks || count < 1 || count > 64) {
        set_error(err, err_cap, "invalid AQM64 CUDA batch");
        return 1;
    }

    if (ensure_capacity(count, err, err_cap) != 0) return 1;

    const size_t initial_bytes =
        static_cast<size_t>(count) * kInitialWords * sizeof(u64);
    const size_t final_bytes =
        static_cast<size_t>(count) * kBlockWords * sizeof(u64);

    cudaError_t rc = cudaMemcpy(
        g_initial, initial_blocks, initial_bytes, cudaMemcpyHostToDevice);
    if (rc != cudaSuccess) {
        set_cuda_error(err, err_cap, "cudaMemcpy initial", rc);
        return 1;
    }

    argon2id_aqm64_kernel<<<count, kBlockWords>>>(
        g_workspace, g_initial, count, g_final);
    rc = cudaGetLastError();
    if (rc != cudaSuccess) {
        set_cuda_error(err, err_cap, "AQM64 kernel launch", rc);
        return 1;
    }
    rc = cudaDeviceSynchronize();
    if (rc != cudaSuccess) {
        set_cuda_error(err, err_cap, "AQM64 kernel execution", rc);
        return 1;
    }

    rc = cudaMemcpy(
        final_blocks, g_final, final_bytes, cudaMemcpyDeviceToHost);
    if (rc != cudaSuccess) {
        set_cuda_error(err, err_cap, "cudaMemcpy final", rc);
        return 1;
    }
    if (err && err_cap > 0) err[0] = '\0';
    return 0;
}

AQ_EXPORT void aqm64_cuda_shutdown() {
    free_buffers();
    g_device = -1;
}
