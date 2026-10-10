package com.auronq.mobile;

/** Run with plain Java; uses no Android device or external dependencies. */
public final class AccountSyncGuardSelfTest {
    private static void check(boolean ok, String why) {
        if (!ok) throw new AssertionError(why);
    }
    public static void main(String[] args) {
        String address = "aurq1mywallet";
        String tip = "0123456789abcdef";
        String work = "1f42";
        check(AccountSyncGuard.sameSession(7, 7, address, address, 1331, 1331,
                tip, tip, work, work), "valid account response rejected");
        check(!AccountSyncGuard.sameSession(6, 7, address, address, 1331, 1331,
                tip, tip, work, work), "delayed response survived invalidation");
        check(!AccountSyncGuard.sameSession(7, 7, "aurq1old", address, 1331, 1331,
                tip, tip, work, work), "previous wallet response accepted");
        check(!AccountSyncGuard.sameSession(7, 7, address, address, 1330, 1331,
                tip, tip, work, work), "old block response accepted");
        check(!AccountSyncGuard.sameSession(7, 7, address, address, 1331, 1331,
                "wrong", tip, work, work), "different chain tip accepted");
        check(!AccountSyncGuard.sameSession(7, 7, address, address, 1331, 1331,
                tip, tip, "oldwork", work), "different cumulative work accepted");
        check(AccountSyncGuard.acceptedPeerReport(address, address, 4, 3, true),
                "two-plus peer agreement rejected");
        check(!AccountSyncGuard.acceptedPeerReport(address, address, 1, 1, false),
                "single node balance treated as verified");
        check(!AccountSyncGuard.acceptedPeerReport(address, address, 3, 1, false),
                "conflicting node balance accepted");
        check(!AccountSyncGuard.acceptedPeerReport("aurq1other", address, 4, 3, true),
                "other account balance accepted");
        check(!AccountSyncGuard.acceptedPeerReport(address, address, 4, 5, true),
                "peer vote count greater than observations accepted");
        check(AccountSyncGuard.isFormattedAmount("0.00000000"), "real verified zero disallowed");
        check(AccountSyncGuard.isFormattedAmount("1234.00000001"), "real positive disallowed");
        check(!AccountSyncGuard.isFormattedAmount(null), "missing balance read as zero");
        check(!AccountSyncGuard.isFormattedAmount(""), "empty balance read as zero");
        check(!AccountSyncGuard.isFormattedAmount("-1.00000000"), "negative balance accepted");
        check(!AccountSyncGuard.isFormattedAmount("12.3"), "invalid precision accepted");
        check(!AccountSyncGuard.isFormattedAmount("1e12"), "noncanonical formatted amount accepted");
        check(!AccountSyncGuard.previewSupersedes(1331, tip, work, 1330, tip, work),
                "old preview invalidated validated account");
        check(!AccountSyncGuard.previewSupersedes(1331, tip, work, 1331, tip, work),
                "same preview invalidated validated account");
        check(AccountSyncGuard.previewSupersedes(1331, tip, work, 1332, "new", work),
                "new chain height failed to invalidate previous balance");
        check(AccountSyncGuard.previewSupersedes(1331, tip, work, 1331, "fork", work),
                "same-height fork failed to invalidate previous balance");
        check(AccountSyncGuard.previewSupersedes(1331, tip, work, 1331, tip, "forkwork"),
                "chainwork change failed to invalidate previous balance");
        System.out.println("AccountSyncGuard: 22 standalone fail-closed checks passed");
    }
}
