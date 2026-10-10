package com.auronq.mobile;

public final class HeaderConcurrencyPolicySelfTest {
    static void check(boolean yes) { if (!yes) throw new AssertionError("unsafe AQM64 worker setting"); }

    public static void main(String[] args) {
        long mib = 1024L*1024L;
        check(HeaderConcurrencyPolicy.workers(4L*1024*mib, 1400*mib, false)==2);
        check(HeaderConcurrencyPolicy.workers(2L*1024*mib, 1700*mib, false)==1);
        check(HeaderConcurrencyPolicy.workers(4L*1024*mib, 450*mib, false)==1);
        check(HeaderConcurrencyPolicy.workers(8L*1024*mib, 5000*mib, true)==1);
        check(HeaderConcurrencyPolicy.workers(0L,0L,false)==1);
        System.out.println("HeaderConcurrencyPolicy: conservative two-worker memory limit PASSED");
    }
}
