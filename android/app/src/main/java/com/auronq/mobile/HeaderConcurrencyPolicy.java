package com.auronq.mobile;

/** Conservative native-memory policy: each AQM64 worker uses 64 MiB plus overhead. */
final class HeaderConcurrencyPolicy {
    private HeaderConcurrencyPolicy() {}

    static int workers(long totalMemoryBytes, long freeMemoryBytes, boolean lowMemory) {
        final long mib = 1024L * 1024L;
        return !lowMemory && totalMemoryBytes >= 3L * 1024L * mib
                && freeMemoryBytes >= 768L * mib ? 2 : 1;
    }
}
