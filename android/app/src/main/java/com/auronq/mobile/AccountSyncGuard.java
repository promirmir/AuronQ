package com.auronq.mobile;

import java.util.Locale;

/**
 * Pure-Java, side-effect-free safety checks for wallet UI presentation.
 * This class does NOT trust peer balances, update chain state, or sign funds.
 */
final class AccountSyncGuard {
    private AccountSyncGuard() {}

    static boolean sameSession(long requestedEpoch, long currentEpoch,
            String requestedAddress, String activeAddress,
            long requestedHeight, long activeHeight,
            String requestedTip, String activeTip,
            String requestedWork, String activeWork) {
        return requestedEpoch == currentEpoch
                && activeAddress != null && !activeAddress.isEmpty()
                && activeAddress.equals(requestedAddress)
                && requestedHeight >= 0 && requestedHeight == activeHeight
                && sameHex(requestedTip, activeTip)
                && sameHex(requestedWork, activeWork);
    }

    static boolean acceptedPeerReport(String returnedAddress, String walletAddress,
            int observed, int agreement, boolean multiPeerConfirmed) {
        return walletAddress != null && !walletAddress.isEmpty()
                && walletAddress.equals(returnedAddress)
                && observed >= 2 && agreement >= 2 && agreement <= observed
                && multiPeerConfirmed;
    }

    static boolean isFormattedAmount(String amount) {
        return amount != null && amount.length() <= 32
                && amount.matches("[0-9]+\\.[0-9]{8}");
    }

    static boolean previewSupersedes(long validatedHeight, String validatedTip,
            String validatedWork, long previewHeight, String previewTip,
            String previewWork) {
        if (validatedHeight < 0 || previewHeight < validatedHeight) return false;
        return previewHeight > validatedHeight || !sameHex(validatedTip, previewTip)
                || !sameHex(validatedWork, previewWork);
    }

    private static boolean sameHex(String left, String right) {
        return left != null && right != null && !left.isEmpty()
                && left.equalsIgnoreCase(right);
    }
}
