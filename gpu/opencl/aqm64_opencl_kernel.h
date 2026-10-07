#pragma once

static const char* kAQM64OpenCLSource = R"CLC(
typedef ulong u64;

#define KBLOCK_WORDS 128
#define KINITIAL_WORDS 256
#define KMEMORY_BLOCKS (64 * 1024)
#define KSYNC_POINTS 4
#define KSEGMENTS (KMEMORY_BLOCKS / KSYNC_POINTS)
#define KTIME_COST 2

inline u64 aq_rotr64(u64 x, uint n) {
    return (x >> n) | (x << (64 - n));
}

inline u64 aq_blamka(u64 x, u64 y) {
    return x + y + (u64)2 * (u64)((uint)x) * (u64)((uint)y);
}

inline void aq_gmix(__private u64* a, __private u64* b, __private u64* c, __private u64* d) {
    *a = aq_blamka(*a, *b);
    *d = aq_rotr64(*d ^ *a, 32);
    *c = aq_blamka(*c, *d);
    *b = aq_rotr64(*b ^ *c, 24);
    *a = aq_blamka(*a, *b);
    *d = aq_rotr64(*d ^ *a, 16);
    *c = aq_blamka(*c, *d);
    *b = aq_rotr64(*b ^ *c, 63);
}

inline void aq_round16(__private u64 v[16]) {
    aq_gmix(&v[0], &v[4], &v[8],  &v[12]);
    aq_gmix(&v[1], &v[5], &v[9],  &v[13]);
    aq_gmix(&v[2], &v[6], &v[10], &v[14]);
    aq_gmix(&v[3], &v[7], &v[11], &v[15]);
    aq_gmix(&v[0], &v[5], &v[10], &v[15]);
    aq_gmix(&v[1], &v[6], &v[11], &v[12]);
    aq_gmix(&v[2], &v[7], &v[8],  &v[13]);
    aq_gmix(&v[3], &v[4], &v[9],  &v[14]);
}

inline void aq_compress_local(
    __local u64* out,
    __local const u64* in1,
    __local const u64* in2,
    int xor_old,
    __local u64* r,
    __local u64* z)
{
    uint t = get_local_id(0);
    r[t] = in1[t] ^ in2[t];
    z[t] = r[t];
    barrier(CLK_LOCAL_MEM_FENCE);

    if (t < 8) {
        u64 v[16];
        for (uint j = 0; j < 16; ++j) v[j] = z[t * 16 + j];
        aq_round16(v);
        for (uint j = 0; j < 16; ++j) z[t * 16 + j] = v[j];
    }
    barrier(CLK_LOCAL_MEM_FENCE);

    if (t < 8) {
        u64 v[16];
        for (uint j = 0; j < 8; ++j) {
            v[2 * j]     = z[16 * j + 2 * t];
            v[2 * j + 1] = z[16 * j + 2 * t + 1];
        }
        aq_round16(v);
        for (uint j = 0; j < 8; ++j) {
            z[16 * j + 2 * t]     = v[2 * j];
            z[16 * j + 2 * t + 1] = v[2 * j + 1];
        }
    }
    barrier(CLK_LOCAL_MEM_FENCE);

    if (xor_old) out[t] ^= r[t] ^ z[t];
    else out[t] = r[t] ^ z[t];
    barrier(CLK_LOCAL_MEM_FENCE);
}

inline void aq_compress_global(
    __global u64* out,
    __global const u64* in1,
    __global const u64* in2,
    int xor_old,
    __local u64* r,
    __local u64* z)
{
    uint t = get_local_id(0);
    r[t] = in1[t] ^ in2[t];
    z[t] = r[t];
    barrier(CLK_LOCAL_MEM_FENCE);

    if (t < 8) {
        u64 v[16];
        for (uint j = 0; j < 16; ++j) v[j] = z[t * 16 + j];
        aq_round16(v);
        for (uint j = 0; j < 16; ++j) z[t * 16 + j] = v[j];
    }
    barrier(CLK_LOCAL_MEM_FENCE);

    if (t < 8) {
        u64 v[16];
        for (uint j = 0; j < 8; ++j) {
            v[2 * j]     = z[16 * j + 2 * t];
            v[2 * j + 1] = z[16 * j + 2 * t + 1];
        }
        aq_round16(v);
        for (uint j = 0; j < 8; ++j) {
            z[16 * j + 2 * t]     = v[2 * j];
            z[16 * j + 2 * t + 1] = v[2 * j + 1];
        }
    }
    barrier(CLK_LOCAL_MEM_FENCE);

    if (xor_old) out[t] ^= r[t] ^ z[t];
    else out[t] = r[t] ^ z[t];
    barrier(CLK_LOCAL_MEM_FENCE | CLK_GLOBAL_MEM_FENCE);
}

inline uint aq_index_alpha_single_lane(u64 random, uint pass, uint slice, uint index) {
    u64 m;
    u64 s;
    if (pass == 0) {
        m = (u64)slice * (u64)KSEGMENTS + (u64)index;
        s = 0;
    } else {
        m = (u64)(3 * KSEGMENTS) + (u64)index;
        s = (u64)((slice + 1) % KSYNC_POINTS) * (u64)KSEGMENTS;
    }
    --m;
    u64 p = random & (u64)0xffffffffUL;
    p = (p * p) >> 32;
    p = (p * m) >> 32;
    return (uint)((s + m - (p + 1)) % (u64)KMEMORY_BLOCKS);
}

__kernel void argon2id_aqm64(
    __global u64* workspace,
    __global const u64* initial,
    int count,
    __global u64* final_out)
{
    uint candidate = get_group_id(0);
    uint t = get_local_id(0);
    if ((int)candidate >= count || get_local_size(0) != KBLOCK_WORDS) return;

    __global u64* B = workspace +
        ((size_t)candidate * (size_t)KMEMORY_BLOCKS * (size_t)KBLOCK_WORDS);
    __global const u64* init = initial + ((size_t)candidate * (size_t)KINITIAL_WORDS);

    __local u64 r[KBLOCK_WORDS];
    __local u64 z[KBLOCK_WORDS];
    __local u64 addresses[KBLOCK_WORDS];
    __local u64 input[KBLOCK_WORDS];
    __local u64 zero[KBLOCK_WORDS];
    __local u64 random_word;
    __local uint ref_offset;

    B[t] = init[t];
    B[KBLOCK_WORDS + t] = init[KBLOCK_WORDS + t];
    barrier(CLK_GLOBAL_MEM_FENCE);

    for (uint pass = 0; pass < KTIME_COST; ++pass) {
        for (uint slice = 0; slice < KSYNC_POINTS; ++slice) {
            int independent = (pass == 0 && slice < (KSYNC_POINTS / 2));

            addresses[t] = 0;
            input[t] = 0;
            zero[t] = 0;
            barrier(CLK_LOCAL_MEM_FENCE);

            if (independent && t == 0) {
                input[0] = pass;
                input[1] = 0;
                input[2] = slice;
                input[3] = KMEMORY_BLOCKS;
                input[4] = KTIME_COST;
                input[5] = 2;
            }
            barrier(CLK_LOCAL_MEM_FENCE);

            uint index = 0;
            if (pass == 0 && slice == 0) {
                index = 2;
                if (t == 0) ++input[6];
                barrier(CLK_LOCAL_MEM_FENCE);
                aq_compress_local(addresses, input, zero, 0, r, z);
                aq_compress_local(addresses, addresses, zero, 0, r, z);
            }

            uint offset = slice * KSEGMENTS + index;
            while (index < KSEGMENTS) {
                uint prev = offset - 1;
                if (index == 0 && slice == 0) prev += KMEMORY_BLOCKS;

                if (independent && (index % KBLOCK_WORDS) == 0) {
                    if (t == 0) ++input[6];
                    barrier(CLK_LOCAL_MEM_FENCE);
                    aq_compress_local(addresses, input, zero, 0, r, z);
                    aq_compress_local(addresses, addresses, zero, 0, r, z);
                }

                if (t == 0) {
                    random_word = independent
                        ? addresses[index % KBLOCK_WORDS]
                        : B[(size_t)prev * KBLOCK_WORDS];
                    ref_offset = aq_index_alpha_single_lane(random_word, pass, slice, index);
                }
                barrier(CLK_LOCAL_MEM_FENCE);

                aq_compress_global(
                    B + (size_t)offset * KBLOCK_WORDS,
                    B + (size_t)prev * KBLOCK_WORDS,
                    B + (size_t)ref_offset * KBLOCK_WORDS,
                    pass != 0,
                    r,
                    z);

                ++index;
                ++offset;
            }
        }
    }

    final_out[(size_t)candidate * KBLOCK_WORDS + t] =
        B[(size_t)(KMEMORY_BLOCKS - 1) * KBLOCK_WORDS + t];
}
)CLC";
