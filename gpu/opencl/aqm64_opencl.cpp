#include "aqm64_opencl_kernel.h"

#include <algorithm>
#include <cstdint>
#include <cstdio>
#include <cstdlib>
#include <cstring>
#include <string>
#include <vector>

#ifdef _WIN32
#ifndef NOMINMAX
#define NOMINMAX
#endif
#include <windows.h>
#define AQ_EXPORT extern "C" __declspec(dllexport)
#else
#include <dlfcn.h>
#define AQ_EXPORT extern "C" __attribute__((visibility("default")))
#endif

using cl_int = int32_t;
using cl_uint = uint32_t;
using cl_ulong = uint64_t;
using cl_bitfield = uintptr_t;
using cl_device_type = cl_bitfield;
using cl_mem_flags = cl_bitfield;
using cl_command_queue_properties = cl_bitfield;
using cl_context_properties = intptr_t;
using cl_platform_id = struct _cl_platform_id*;
using cl_device_id = struct _cl_device_id*;
using cl_context = struct _cl_context*;
using cl_command_queue = struct _cl_command_queue*;
using cl_mem = struct _cl_mem*;
using cl_program = struct _cl_program*;
using cl_kernel = struct _cl_kernel*;
using cl_event = struct _cl_event*;

constexpr cl_int CL_SUCCESS = 0;
constexpr cl_int CL_DEVICE_NOT_FOUND = -1;
constexpr cl_device_type CL_DEVICE_TYPE_GPU = (1u << 2);
constexpr cl_device_type CL_DEVICE_TYPE_ALL = static_cast<cl_device_type>(~static_cast<cl_device_type>(0));
constexpr cl_mem_flags CL_MEM_READ_WRITE = (1u << 0);
constexpr cl_mem_flags CL_MEM_WRITE_ONLY = (1u << 1);
constexpr cl_mem_flags CL_MEM_READ_ONLY = (1u << 2);
constexpr cl_uint CL_TRUE = 1;
constexpr cl_uint CL_DEVICE_MAX_COMPUTE_UNITS = 0x1002;
constexpr cl_uint CL_DEVICE_MAX_WORK_GROUP_SIZE = 0x1004;
constexpr cl_uint CL_DEVICE_GLOBAL_MEM_SIZE = 0x101F;
constexpr cl_uint CL_DEVICE_LOCAL_MEM_SIZE = 0x1023;
constexpr cl_uint CL_DEVICE_NAME = 0x102B;
constexpr cl_uint CL_DEVICE_VENDOR = 0x102C;
constexpr cl_uint CL_PROGRAM_BUILD_LOG = 0x1183;

using PFN_clGetPlatformIDs = cl_int (*)(cl_uint, cl_platform_id*, cl_uint*);
using PFN_clGetDeviceIDs = cl_int (*)(cl_platform_id, cl_device_type, cl_uint, cl_device_id*, cl_uint*);
using PFN_clGetDeviceInfo = cl_int (*)(cl_device_id, cl_uint, size_t, void*, size_t*);
using PFN_clCreateContext = cl_context (*)(const cl_context_properties*, cl_uint, const cl_device_id*, void*, void*, cl_int*);
using PFN_clCreateCommandQueue = cl_command_queue (*)(cl_context, cl_device_id, cl_command_queue_properties, cl_int*);
using PFN_clCreateProgramWithSource = cl_program (*)(cl_context, cl_uint, const char**, const size_t*, cl_int*);
using PFN_clBuildProgram = cl_int (*)(cl_program, cl_uint, const cl_device_id*, const char*, void*, void*);
using PFN_clGetProgramBuildInfo = cl_int (*)(cl_program, cl_device_id, cl_uint, size_t, void*, size_t*);
using PFN_clCreateKernel = cl_kernel (*)(cl_program, const char*, cl_int*);
using PFN_clCreateBuffer = cl_mem (*)(cl_context, cl_mem_flags, size_t, void*, cl_int*);
using PFN_clSetKernelArg = cl_int (*)(cl_kernel, cl_uint, size_t, const void*);
using PFN_clEnqueueWriteBuffer = cl_int (*)(cl_command_queue, cl_mem, cl_uint, size_t, size_t, const void*, cl_uint, const cl_event*, cl_event*);
using PFN_clEnqueueNDRangeKernel = cl_int (*)(cl_command_queue, cl_kernel, cl_uint, const size_t*, const size_t*, const size_t*, cl_uint, const cl_event*, cl_event*);
using PFN_clFinish = cl_int (*)(cl_command_queue);
using PFN_clEnqueueReadBuffer = cl_int (*)(cl_command_queue, cl_mem, cl_uint, size_t, size_t, void*, cl_uint, const cl_event*, cl_event*);
using PFN_clReleaseMemObject = cl_int (*)(cl_mem);
using PFN_clReleaseKernel = cl_int (*)(cl_kernel);
using PFN_clReleaseProgram = cl_int (*)(cl_program);
using PFN_clReleaseCommandQueue = cl_int (*)(cl_command_queue);
using PFN_clReleaseContext = cl_int (*)(cl_context);

struct OpenCLAPI {
    PFN_clGetPlatformIDs clGetPlatformIDs{};
    PFN_clGetDeviceIDs clGetDeviceIDs{};
    PFN_clGetDeviceInfo clGetDeviceInfo{};
    PFN_clCreateContext clCreateContext{};
    PFN_clCreateCommandQueue clCreateCommandQueue{};
    PFN_clCreateProgramWithSource clCreateProgramWithSource{};
    PFN_clBuildProgram clBuildProgram{};
    PFN_clGetProgramBuildInfo clGetProgramBuildInfo{};
    PFN_clCreateKernel clCreateKernel{};
    PFN_clCreateBuffer clCreateBuffer{};
    PFN_clSetKernelArg clSetKernelArg{};
    PFN_clEnqueueWriteBuffer clEnqueueWriteBuffer{};
    PFN_clEnqueueNDRangeKernel clEnqueueNDRangeKernel{};
    PFN_clFinish clFinish{};
    PFN_clEnqueueReadBuffer clEnqueueReadBuffer{};
    PFN_clReleaseMemObject clReleaseMemObject{};
    PFN_clReleaseKernel clReleaseKernel{};
    PFN_clReleaseProgram clReleaseProgram{};
    PFN_clReleaseCommandQueue clReleaseCommandQueue{};
    PFN_clReleaseContext clReleaseContext{};
};

static OpenCLAPI g_api;
#ifdef _WIN32
static HMODULE g_lib = nullptr;
#else
static void* g_lib = nullptr;
#endif

static cl_context g_context = nullptr;
static cl_command_queue g_queue = nullptr;
static cl_program g_program = nullptr;
static cl_kernel g_kernel = nullptr;
static cl_mem g_workspace = nullptr;
static cl_mem g_initial = nullptr;
static cl_mem g_final = nullptr;
static cl_device_id g_device = nullptr;
static int g_capacity = 0;

constexpr int kBlockWords = 128;
constexpr int kInitialWords = 256;
constexpr int kMemoryBlocks = 64 * 1024;
constexpr size_t kBytesPerCandidate =
    static_cast<size_t>(kMemoryBlocks) * kBlockWords * sizeof(uint64_t);

static void set_error(char* err, int cap, const std::string& text) {
    if (!err || cap <= 0) return;
    std::snprintf(err, static_cast<size_t>(cap), "%s", text.c_str());
    err[cap - 1] = '\0';
}

#ifdef _WIN32
static void* load_symbol(const char* name) {
    return g_lib ? reinterpret_cast<void*>(GetProcAddress(g_lib, name)) : nullptr;
}
#else
static void* load_symbol(const char* name) {
    return g_lib ? dlsym(g_lib, name) : nullptr;
}
#endif

template <typename T>
static bool bind_symbol(T& out, const char* name) {
    out = reinterpret_cast<T>(load_symbol(name));
    return out != nullptr;
}

static bool load_opencl(std::string& why) {
    if (g_lib) return true;
#ifdef _WIN32
    g_lib = LoadLibraryA("OpenCL.dll");
#else
    g_lib = dlopen("libOpenCL.so.1", RTLD_NOW | RTLD_LOCAL);
    if (!g_lib) g_lib = dlopen("libOpenCL.so", RTLD_NOW | RTLD_LOCAL);
#endif
    if (!g_lib) {
        why = "OpenCL runtime/ICD loader is not installed";
        return false;
    }

#define AQ_BIND(x) if (!bind_symbol(g_api.x, #x)) { why = std::string("missing OpenCL symbol ") + #x; return false; }
    AQ_BIND(clGetPlatformIDs)
    AQ_BIND(clGetDeviceIDs)
    AQ_BIND(clGetDeviceInfo)
    AQ_BIND(clCreateContext)
    AQ_BIND(clCreateCommandQueue)
    AQ_BIND(clCreateProgramWithSource)
    AQ_BIND(clBuildProgram)
    AQ_BIND(clGetProgramBuildInfo)
    AQ_BIND(clCreateKernel)
    AQ_BIND(clCreateBuffer)
    AQ_BIND(clSetKernelArg)
    AQ_BIND(clEnqueueWriteBuffer)
    AQ_BIND(clEnqueueNDRangeKernel)
    AQ_BIND(clFinish)
    AQ_BIND(clEnqueueReadBuffer)
    AQ_BIND(clReleaseMemObject)
    AQ_BIND(clReleaseKernel)
    AQ_BIND(clReleaseProgram)
    AQ_BIND(clReleaseCommandQueue)
    AQ_BIND(clReleaseContext)
#undef AQ_BIND
    return true;
}

static void release_buffers() {
    if (g_workspace) g_api.clReleaseMemObject(g_workspace);
    if (g_initial) g_api.clReleaseMemObject(g_initial);
    if (g_final) g_api.clReleaseMemObject(g_final);
    g_workspace = nullptr;
    g_initial = nullptr;
    g_final = nullptr;
    g_capacity = 0;
}

AQ_EXPORT void aqm64_opencl_shutdown() {
    release_buffers();
    if (g_kernel) g_api.clReleaseKernel(g_kernel);
    if (g_program) g_api.clReleaseProgram(g_program);
    if (g_queue) g_api.clReleaseCommandQueue(g_queue);
    if (g_context) g_api.clReleaseContext(g_context);
    g_kernel = nullptr;
    g_program = nullptr;
    g_queue = nullptr;
    g_context = nullptr;
    g_device = nullptr;
}

static std::string device_string(cl_device_id d, cl_uint key) {
    size_t n = 0;
    if (g_api.clGetDeviceInfo(d, key, 0, nullptr, &n) != CL_SUCCESS || n == 0) return "";
    std::vector<char> b(n + 1, 0);
    if (g_api.clGetDeviceInfo(d, key, n, b.data(), nullptr) != CL_SUCCESS) return "";
    return std::string(b.data());
}

static std::vector<cl_device_id> enumerate_gpus(std::string& why) {
    std::vector<cl_device_id> out;
    cl_uint pc = 0;
    cl_int rc = g_api.clGetPlatformIDs(0, nullptr, &pc);
    if (rc != CL_SUCCESS || pc == 0) {
        why = "no OpenCL platforms";
        return out;
    }
    std::vector<cl_platform_id> platforms(pc);
    if (g_api.clGetPlatformIDs(pc, platforms.data(), nullptr) != CL_SUCCESS) {
        why = "clGetPlatformIDs failed";
        return out;
    }
    const char* test_all = std::getenv("AURONQ_OPENCL_TEST_ALL");
    const cl_device_type wanted_type =
        (test_all && std::strcmp(test_all, "1") == 0) ? CL_DEVICE_TYPE_ALL : CL_DEVICE_TYPE_GPU;

    for (auto p : platforms) {
        cl_uint dc = 0;
        rc = g_api.clGetDeviceIDs(p, wanted_type, 0, nullptr, &dc);
        if (rc == CL_DEVICE_NOT_FOUND || dc == 0) continue;
        if (rc != CL_SUCCESS) continue;
        std::vector<cl_device_id> ds(dc);
        if (g_api.clGetDeviceIDs(p, wanted_type, dc, ds.data(), nullptr) == CL_SUCCESS) {
            out.insert(out.end(), ds.begin(), ds.end());
        }
    }
    if (out.empty()) why = "no OpenCL GPU devices";
    return out;
}

static int ensure_capacity(int count, char* err, int err_cap) {
    if (count <= g_capacity && g_workspace && g_initial && g_final) return 0;
    release_buffers();

    cl_int rc = CL_SUCCESS;
    const size_t workspace_bytes = static_cast<size_t>(count) * kBytesPerCandidate;
    const size_t initial_bytes = static_cast<size_t>(count) * kInitialWords * sizeof(uint64_t);
    const size_t final_bytes = static_cast<size_t>(count) * kBlockWords * sizeof(uint64_t);

    g_workspace = g_api.clCreateBuffer(g_context, CL_MEM_READ_WRITE, workspace_bytes, nullptr, &rc);
    if (!g_workspace || rc != CL_SUCCESS) {
        set_error(err, err_cap, "OpenCL workspace allocation failed (rc=" + std::to_string(rc) + ")");
        release_buffers();
        return 1;
    }
    g_initial = g_api.clCreateBuffer(g_context, CL_MEM_READ_ONLY, initial_bytes, nullptr, &rc);
    if (!g_initial || rc != CL_SUCCESS) {
        set_error(err, err_cap, "OpenCL initial-buffer allocation failed (rc=" + std::to_string(rc) + ")");
        release_buffers();
        return 1;
    }
    g_final = g_api.clCreateBuffer(g_context, CL_MEM_WRITE_ONLY, final_bytes, nullptr, &rc);
    if (!g_final || rc != CL_SUCCESS) {
        set_error(err, err_cap, "OpenCL final-buffer allocation failed (rc=" + std::to_string(rc) + ")");
        release_buffers();
        return 1;
    }
    g_capacity = count;
    return 0;
}

AQ_EXPORT int aqm64_opencl_device_count(char* err, int err_cap) {
    std::string why;
    if (!load_opencl(why)) {
        set_error(err, err_cap, why);
        return -1;
    }
    auto ds = enumerate_gpus(why);
    if (ds.empty() && !why.empty()) set_error(err, err_cap, why);
    return static_cast<int>(ds.size());
}

AQ_EXPORT int aqm64_opencl_init(
    int device,
    int* recommended_batch,
    char* device_name,
    int device_name_cap,
    char* err,
    int err_cap)
{
    aqm64_opencl_shutdown();

    std::string why;
    if (!load_opencl(why)) {
        set_error(err, err_cap, why);
        return 1;
    }
    auto devices = enumerate_gpus(why);
    if (device < 0 || device >= static_cast<int>(devices.size())) {
        set_error(err, err_cap, why.empty() ? "requested OpenCL GPU does not exist" : why);
        return 1;
    }
    g_device = devices[static_cast<size_t>(device)];

    size_t max_wg = 0;
    cl_ulong global_mem = 0;
    cl_ulong local_mem = 0;
    cl_uint compute_units = 0;
    g_api.clGetDeviceInfo(g_device, CL_DEVICE_MAX_WORK_GROUP_SIZE, sizeof(max_wg), &max_wg, nullptr);
    g_api.clGetDeviceInfo(g_device, CL_DEVICE_GLOBAL_MEM_SIZE, sizeof(global_mem), &global_mem, nullptr);
    g_api.clGetDeviceInfo(g_device, CL_DEVICE_LOCAL_MEM_SIZE, sizeof(local_mem), &local_mem, nullptr);
    g_api.clGetDeviceInfo(g_device, CL_DEVICE_MAX_COMPUTE_UNITS, sizeof(compute_units), &compute_units, nullptr);
    if (max_wg < 128) {
        set_error(err, err_cap, "OpenCL GPU cannot provide the required 128-work-item group");
        return 1;
    }
    if (local_mem < 8 * 1024) {
        set_error(err, err_cap, "OpenCL GPU has insufficient local memory for AQM64");
        return 1;
    }

    cl_int rc = CL_SUCCESS;
    g_context = g_api.clCreateContext(nullptr, 1, &g_device, nullptr, nullptr, &rc);
    if (!g_context || rc != CL_SUCCESS) {
        set_error(err, err_cap, "clCreateContext failed (rc=" + std::to_string(rc) + ")");
        aqm64_opencl_shutdown();
        return 1;
    }
    g_queue = g_api.clCreateCommandQueue(g_context, g_device, 0, &rc);
    if (!g_queue || rc != CL_SUCCESS) {
        set_error(err, err_cap, "clCreateCommandQueue failed (rc=" + std::to_string(rc) + ")");
        aqm64_opencl_shutdown();
        return 1;
    }

    const char* src = kAQM64OpenCLSource;
    size_t src_len = std::strlen(src);
    g_program = g_api.clCreateProgramWithSource(g_context, 1, &src, &src_len, &rc);
    if (!g_program || rc != CL_SUCCESS) {
        set_error(err, err_cap, "clCreateProgramWithSource failed (rc=" + std::to_string(rc) + ")");
        aqm64_opencl_shutdown();
        return 1;
    }

    rc = g_api.clBuildProgram(g_program, 1, &g_device, "-cl-std=CL1.2", nullptr, nullptr);
    if (rc != CL_SUCCESS) {
        size_t n = 0;
        g_api.clGetProgramBuildInfo(g_program, g_device, CL_PROGRAM_BUILD_LOG, 0, nullptr, &n);
        std::vector<char> log(n + 1, 0);
        if (n) g_api.clGetProgramBuildInfo(g_program, g_device, CL_PROGRAM_BUILD_LOG, n, log.data(), nullptr);
        set_error(err, err_cap, "OpenCL AQM64 kernel build failed: " + std::string(log.data()));
        aqm64_opencl_shutdown();
        return 1;
    }

    g_kernel = g_api.clCreateKernel(g_program, "argon2id_aqm64", &rc);
    if (!g_kernel || rc != CL_SUCCESS) {
        set_error(err, err_cap, "clCreateKernel failed (rc=" + std::to_string(rc) + ")");
        aqm64_opencl_shutdown();
        return 1;
    }

    int memory_lanes = static_cast<int>((global_mem / 5) / kBytesPerCandidate);
    memory_lanes = std::max(1, memory_lanes);
    int rec = std::min({8, memory_lanes, std::max(1, static_cast<int>(compute_units))});
    if (recommended_batch) *recommended_batch = rec;

    const std::string vendor = device_string(g_device, CL_DEVICE_VENDOR);
    const std::string name = device_string(g_device, CL_DEVICE_NAME);
    if (device_name && device_name_cap > 0) {
        std::snprintf(device_name, static_cast<size_t>(device_name_cap),
            "%s %s (OpenCL, %.1f GiB)",
            vendor.c_str(), name.c_str(),
            static_cast<double>(global_mem) / (1024.0 * 1024.0 * 1024.0));
        device_name[device_name_cap - 1] = '\0';
    }
    if (err && err_cap > 0) err[0] = '\0';
    return 0;
}

AQ_EXPORT int aqm64_opencl_run(
    const uint64_t* initial_blocks,
    int count,
    uint64_t* final_blocks,
    char* err,
    int err_cap)
{
    if (!g_context || !g_queue || !g_kernel || !g_device) {
        set_error(err, err_cap, "OpenCL backend is not initialized");
        return 1;
    }
    if (!initial_blocks || !final_blocks || count < 1 || count > 64) {
        set_error(err, err_cap, "invalid AQM64 OpenCL batch");
        return 1;
    }
    if (ensure_capacity(count, err, err_cap) != 0) return 1;

    const size_t initial_bytes = static_cast<size_t>(count) * kInitialWords * sizeof(uint64_t);
    const size_t final_bytes = static_cast<size_t>(count) * kBlockWords * sizeof(uint64_t);

    cl_int rc = g_api.clEnqueueWriteBuffer(g_queue, g_initial, CL_TRUE, 0, initial_bytes,
                                           initial_blocks, 0, nullptr, nullptr);
    if (rc != CL_SUCCESS) {
        set_error(err, err_cap, "clEnqueueWriteBuffer failed (rc=" + std::to_string(rc) + ")");
        return 1;
    }

    rc = g_api.clSetKernelArg(g_kernel, 0, sizeof(cl_mem), &g_workspace);
    if (rc == CL_SUCCESS) rc = g_api.clSetKernelArg(g_kernel, 1, sizeof(cl_mem), &g_initial);
    if (rc == CL_SUCCESS) rc = g_api.clSetKernelArg(g_kernel, 2, sizeof(int), &count);
    if (rc == CL_SUCCESS) rc = g_api.clSetKernelArg(g_kernel, 3, sizeof(cl_mem), &g_final);
    if (rc != CL_SUCCESS) {
        set_error(err, err_cap, "clSetKernelArg failed (rc=" + std::to_string(rc) + ")");
        return 1;
    }

    const size_t local = 128;
    const size_t global = static_cast<size_t>(count) * local;
    rc = g_api.clEnqueueNDRangeKernel(g_queue, g_kernel, 1, nullptr, &global, &local, 0, nullptr, nullptr);
    if (rc != CL_SUCCESS) {
        set_error(err, err_cap, "OpenCL AQM64 kernel launch failed (rc=" + std::to_string(rc) + ")");
        return 1;
    }
    rc = g_api.clFinish(g_queue);
    if (rc != CL_SUCCESS) {
        set_error(err, err_cap, "OpenCL AQM64 kernel execution failed (rc=" + std::to_string(rc) + ")");
        return 1;
    }
    rc = g_api.clEnqueueReadBuffer(g_queue, g_final, CL_TRUE, 0, final_bytes,
                                   final_blocks, 0, nullptr, nullptr);
    if (rc != CL_SUCCESS) {
        set_error(err, err_cap, "clEnqueueReadBuffer failed (rc=" + std::to_string(rc) + ")");
        return 1;
    }
    if (err && err_cap > 0) err[0] = '\0';
    return 0;
}
