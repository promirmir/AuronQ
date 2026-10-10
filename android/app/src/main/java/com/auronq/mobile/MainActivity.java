package com.auronq.mobile;

import android.app.Activity;
import android.app.AlertDialog;
import android.content.ClipData;
import android.content.ClipboardManager;
import android.content.Context;
import android.content.Intent;
import android.content.SharedPreferences;
import android.graphics.Color;
import android.graphics.Typeface;
import android.graphics.drawable.GradientDrawable;
import android.net.Uri;
import android.os.Bundle;
import android.os.Handler;
import android.os.Looper;
import android.text.InputType;
import android.view.Gravity;
import android.view.View;
import android.widget.Button;
import android.widget.EditText;
import android.widget.FrameLayout;
import android.widget.LinearLayout;
import android.widget.ScrollView;
import android.widget.TextView;
import android.widget.Toast;

import com.auronq.mobilebind.bridge.Bridge;

import org.json.JSONArray;
import org.json.JSONObject;

import java.io.File;
import java.io.FileOutputStream;
import java.io.InputStream;
import java.io.OutputStream;
import java.text.SimpleDateFormat;
import java.util.Date;
import java.util.LinkedHashSet;
import java.util.Locale;
import java.util.Set;
import java.util.TimeZone;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

public class MainActivity extends Activity {
    private static final int IMPORT_WALLET = 1001;
    private static final int EXPORT_WALLET = 1002;

    private static final int BG = Color.rgb(7, 11, 18);
    private static final int PANEL = Color.rgb(14, 21, 33);
    private static final int PANEL2 = Color.rgb(18, 28, 43);
    private static final int LINE = Color.rgb(33, 48, 71);
    private static final int TEXT = Color.rgb(237, 244, 255);
    private static final int MUTED = Color.rgb(143, 162, 186);
    private static final int ACCENT = Color.rgb(104, 226, 183);
    private static final int BLUE = Color.rgb(102, 166, 255);
    private static final int DANGER = Color.rgb(255, 111, 130);

    private final ExecutorService executor = Executors.newSingleThreadExecutor();
    // AQM64 replay must NEVER occupy the fast account/network executor.
    private final ExecutorService deepVerifier = Executors.newSingleThreadExecutor();
    private final ExecutorService historyExecutor = Executors.newSingleThreadExecutor();
    private boolean historyBusy = false;
    private boolean deepBusy = false;
    private long deepVerifiedHeight = -1;
    private long lastDeepAttemptMillis = 0;
    private String deepVerifiedTip = "";
    private String deepVerifiedWork = "";
    private TextView verificationStageText;
    private final Handler handler = new Handler(Looper.getMainLooper());

    private SharedPreferences prefs;
    private boolean english;

    private File walletDir;
    private File walletFile;
    private File headerCacheFile;
    private String nodeUrl = "";
    private String walletAddress = "";
    private String verifiedTip = "";
    private String verifiedWork = "";
    private long verifiedHeight = -1;
    private String currentScreen = "home";
    private String lastHistoryKey = "";
    private boolean liveBusy = false;
    // These describe a THREE-PEER observation, not independently verified AQM64.
    private boolean quickAccountReady = false;
    private long quickHeight = -1;
    private String quickTip = "";
    private long walletGeneration = 0;
    private boolean networkReachable = false;

    private FrameLayout content;
    private LinearLayout dashboardScreen;
    private LinearLayout walletScreen;
    private LinearLayout sendScreen;
    private LinearLayout networkScreen;

    private Button navHome;
    private Button navWallet;
    private Button navSend;
    private Button navNetwork;

    private TextView topTitle;
    private TextView topNetworkDot;

    private TextView dashBalance;
    private TextView dashHeight;
    private TextView dashPeers;
    private TextView dashMempool;
    private TextView dashNode;
    private TextView dashTip;

    private TextView walletAddressText;
    private TextView walletBalance;
    private EditText passwordInput;
    private Button backupButton;
    private Button copyButton;
    private Button sendShortcutButton;
    private Button sendButton;
    private TextView walletTrustStatus;
    private Button deleteButton;
    private TextView walletHistoryStatus;
    private LinearLayout walletHistory;

    private EditText sendPassword;
    private EditText recipientInput;
    private EditText amountInput;
    private TextView sendResult;

    private TextView netStatus;
    private TextView netHeight;
    private TextView netPeers;
    private TextView netMempool;
    private TextView netIssued;
    private TextView netNode;
    private TextView netNetworkId;
    private TextView netTip;
    private TextView netWork;
    private TextView netObserved;
    private LinearLayout recentBlocks;

    private final Runnable liveLoop = new Runnable() {
        @Override
        public void run() {
            refreshAll();
            handler.postDelayed(this, 5000);
        }
    };

    interface Task { String run() throws Exception; }
    interface Success { void accept(String value); }

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);

        prefs = getSharedPreferences("auronq_mobile", MODE_PRIVATE);
        english = prefs.getBoolean("english", false);
        nodeUrl = prefs.getString("last_node", "");

        walletDir = new File(getFilesDir(), "wallets");
        if (!walletDir.exists()) walletDir.mkdirs();
        walletFile = new File(walletDir, "main.wallet");
        File headerDir = new File(getFilesDir(), "headers");
        if (!headerDir.exists()) headerDir.mkdirs();
        headerCacheFile = new File(headerDir, "mainnet-v1.json");

        setContentView(buildUi());
        loadWalletState();
        showScreen("home");
        refreshAll();
    }

    @Override
    protected void onResume() {
        super.onResume();
        invalidateQuickAccount(tr("Ponownie sprawdzam bieżący stan konta…",
                "Rechecking current account state…"));
        handler.removeCallbacks(liveLoop);
        handler.post(liveLoop);
    }

    @Override
    protected void onPause() {
        handler.removeCallbacks(liveLoop);
        super.onPause();
    }

    private String tr(String pl, String en) {
        return english ? en : pl;
    }

    private void toggleLanguage() {
        prefs.edit().putBoolean("english", !english).apply();
        recreate();
    }

    private View buildUi() {
        LinearLayout root = new LinearLayout(this);
        root.setOrientation(LinearLayout.VERTICAL);
        root.setBackgroundColor(BG);

        LinearLayout top = new LinearLayout(this);
        top.setOrientation(LinearLayout.HORIZONTAL);
        top.setGravity(Gravity.CENTER_VERTICAL);
        top.setPadding(dp(18), dp(12), dp(18), dp(12));
        top.setBackgroundColor(Color.rgb(9, 14, 22));

        TextView mark = text("A", 18, true);
        mark.setTextColor(Color.rgb(7, 16, 23));
        mark.setGravity(Gravity.CENTER);
        mark.setBackground(roundGradient(ACCENT, BLUE, 13));
        top.addView(mark, new LinearLayout.LayoutParams(dp(42), dp(42)));

        LinearLayout brand = new LinearLayout(this);
        brand.setOrientation(LinearLayout.VERTICAL);
        brand.setPadding(dp(12), 0, 0, 0);
        TextView name = text("AuronQ", 19, true);
        topTitle = text(tr("Pulpit", "Dashboard"), 11, false);
        topTitle.setTextColor(MUTED);
        brand.addView(name);
        brand.addView(topTitle);
        top.addView(brand, new LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f));

        Button lang = secondaryButton(english ? "PL" : "EN");
        lang.setTextSize(11);
        lang.setMinWidth(0);
        lang.setMinimumWidth(0);
        lang.setPadding(dp(10), dp(6), dp(10), dp(6));
        lang.setOnClickListener(v -> toggleLanguage());
        top.addView(lang);

        topNetworkDot = text("●", 16, true);
        topNetworkDot.setPadding(dp(10), 0, 0, 0);
        topNetworkDot.setTextColor(Color.rgb(86, 101, 121));
        top.addView(topNetworkDot);
        root.addView(top);

        content = new FrameLayout(this);
        dashboardScreen = buildDashboard();
        walletScreen = buildWallet();
        sendScreen = buildSend();
        networkScreen = buildNetwork();
        content.addView(wrapScroll(dashboardScreen));
        content.addView(wrapScroll(walletScreen));
        content.addView(wrapScroll(sendScreen));
        content.addView(wrapScroll(networkScreen));
        root.addView(content, new LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT, 0, 1f));

        LinearLayout nav = new LinearLayout(this);
        nav.setOrientation(LinearLayout.HORIZONTAL);
        nav.setPadding(dp(6), dp(6), dp(6), dp(8));
        nav.setBackgroundColor(Color.rgb(9, 14, 22));

        navHome = navButton("⌂\n" + tr("Pulpit", "Home"), () -> showScreen("home"));
        navWallet = navButton("◫\n" + tr("Portfel", "Wallet"), () -> showScreen("wallet"));
        navSend = navButton("↗\n" + tr("Wyślij", "Send"), () -> showScreen("send"));
        navNetwork = navButton("◎\n" + tr("Sieć", "Network"), () -> showScreen("network"));
        nav.addView(navHome, weight());
        nav.addView(navWallet, weight());
        nav.addView(navSend, weight());
        nav.addView(navNetwork, weight());
        root.addView(nav);

        return root;
    }

    private LinearLayout buildDashboard() {
        LinearLayout root = screenRoot();

        root.addView(text(tr("Witaj w AuronQ", "Welcome to AuronQ"), 26, true));
        TextView sub = text(tr(
                "Mobilny portfel AuronQ Mainnet. Klucze pozostają lokalnie na telefonie.",
                "AuronQ Mainnet mobile wallet. Your keys stay local on the phone."), 13, false);
        sub.setTextColor(MUTED);
        root.addView(sub, mt(5));

        LinearLayout stats1 = row();
        dashBalance = statCard(stats1, tr("SALDO", "BALANCE"), "—", "AURQ");
        dashHeight = statCard(stats1, tr("WYSOKOŚĆ", "HEIGHT"), "—", tr("łańcuch", "chain"));
        root.addView(stats1, mt(20));

        LinearLayout stats2 = row();
        dashPeers = statCard(stats2, "PEERS", "0", "node");
        dashMempool = statCard(stats2, "MEMPOOL", "0", tr("transakcje", "transactions"));
        root.addView(stats2, mt(10));

        LinearLayout networkCard = card();
        networkCard.addView(section(tr("STAN SIECI", "NETWORK STATUS")));
        dashNode = text(tr("Szukanie AuronQ Mainnet…", "Finding AuronQ Mainnet…"), 14, true);
        networkCard.addView(dashNode, mt(10));
        dashTip = text("Tip: —", 11, false);
        dashTip.setTextColor(MUTED);
        dashTip.setTextIsSelectable(true);
        networkCard.addView(dashTip, mt(6));
        verificationStageText = text(tr("Etap 1/2 — oczekiwanie na 3 węzły",
                "Stage 1/2 — waiting for 3 peers"), 11, false);
        verificationStageText.setTextColor(BLUE);
        networkCard.addView(verificationStageText, mt(8));

        Button openNetwork = primaryButton(tr("Szczegóły sieci na żywo", "Live network details"));
        openNetwork.setOnClickListener(v -> showScreen("network"));
        networkCard.addView(openNetwork, mt(14));
        root.addView(networkCard, mt(16));

        LinearLayout walletCard = card();
        walletCard.addView(section(tr("TWÓJ PORTFEL", "YOUR WALLET")));
        TextView note = text(tr(
                "Saldo jest pokazywane dopiero po potwierdzeniu zgodności przez trzy grupy sieci. Nie jest to pełna lokalna weryfikacja blockchaina.",
                "Your balance is shown only after three public network groups agree; this is not full local blockchain verification."), 13, false);
        note.setTextColor(MUTED);
        walletCard.addView(note, mt(8));
        Button openWallet = secondaryButton(tr("Otwórz portfel", "Open wallet"));
        openWallet.setOnClickListener(v -> showScreen("wallet"));
        walletCard.addView(openWallet, mt(14));
        root.addView(walletCard, mt(16));

        return root;
    }

    private LinearLayout buildWallet() {
        LinearLayout root = screenRoot();
        root.addView(title(tr("Portfel", "Wallet")));
        TextView sub = text(tr(
                "Ten sam format zaszyfrowanego pliku .wallet co w AuronQ Desktop.",
                "Uses the same encrypted .wallet file format as AuronQ Desktop."), 13, false);
        sub.setTextColor(MUTED);
        root.addView(sub, mt(5));

        LinearLayout balanceCard = card();
        balanceCard.addView(section(tr("SALDO DOSTĘPNE", "SPENDABLE BALANCE")));
        walletBalance = text("— AURQ", 29, true);
        balanceCard.addView(walletBalance, mt(8));
        walletTrustStatus = text(tr("Sprawdzam zgodność 3 grup węzłów…",
                "Checking agreement across 3 network groups…"), 11, false);
        walletTrustStatus.setTextColor(BLUE);
        balanceCard.addView(walletTrustStatus, mt(8));
        root.addView(balanceCard, mt(18));

        LinearLayout addressCard = card();
        addressCard.addView(section(tr("ADRES", "ADDRESS")));
        walletAddressText = text(tr("Brak lokalnego portfela", "No local wallet"), 12, false);
        walletAddressText.setTextColor(MUTED);
        walletAddressText.setTextIsSelectable(true);
        addressCard.addView(walletAddressText, mt(9));
        root.addView(addressCard, mt(12));

        passwordInput = input(tr("Hasło portfela (min. 12 znaków)", "Wallet password (min. 12 characters)"));
        passwordInput.setInputType(InputType.TYPE_CLASS_TEXT | InputType.TYPE_TEXT_VARIATION_PASSWORD);
        root.addView(passwordInput, mt(14));

        LinearLayout r1 = row();
        Button createButton = primaryButton(tr("Utwórz", "Create"));
        createButton.setOnClickListener(v -> createWallet());
        Button importButton = secondaryButton(tr("Importuj", "Import"));
        importButton.setOnClickListener(v -> importWallet());
        r1.addView(createButton, weight());
        r1.addView(importButton, weight());
        root.addView(r1, mt(12));

        LinearLayout r2 = row();
        backupButton = secondaryButton("Backup");
        backupButton.setOnClickListener(v -> backupWallet());
        copyButton = secondaryButton(tr("Kopiuj adres", "Copy address"));
        copyButton.setOnClickListener(v -> copyAddress());
        r2.addView(backupButton, weight());
        r2.addView(copyButton, weight());
        root.addView(r2, mt(8));

        sendShortcutButton = primaryButton(tr("Wyślij AURQ", "Send AURQ"));
        sendShortcutButton.setOnClickListener(v -> showScreen("send"));
        sendShortcutButton.setEnabled(false);
        root.addView(sendShortcutButton, mt(12));

        LinearLayout historyCard = card();
        historyCard.addView(section(tr("HISTORIA TRANSAKCJI", "TRANSACTION HISTORY")));
        walletHistoryStatus = text(tr("Otwórz portfel, aby pobrać historię.", "Open the wallet to load history."), 11, false);
        walletHistoryStatus.setTextColor(MUTED);
        historyCard.addView(walletHistoryStatus, mt(6));
        walletHistory = new LinearLayout(this);
        walletHistory.setOrientation(LinearLayout.VERTICAL);
        historyCard.addView(walletHistory, mt(10));
        Button historyRefresh = secondaryButton(tr("Odśwież historię", "Refresh history"));
        historyRefresh.setOnClickListener(v -> {
            lastHistoryKey = "";
            refreshAll();
        });
        historyCard.addView(historyRefresh, mt(10));
        root.addView(historyCard, mt(18));

        deleteButton = dangerButton(tr("Usuń lokalny portfel", "Delete local wallet"));
        deleteButton.setOnClickListener(v -> deleteWallet());
        root.addView(deleteButton, mt(28));

        TextView warning = text(tr(
                "Zrób backup przed usunięciem. Usunięcie jedynej kopii portfela może oznaczać trwałą utratę dostępu do środków.",
                "Create a backup before deleting. Deleting the only wallet copy can permanently remove access to funds."), 12, false);
        warning.setTextColor(MUTED);
        root.addView(warning, mt(10));

        return root;
    }

    private LinearLayout buildSend() {
        LinearLayout root = screenRoot();
        root.addView(title(tr("Wyślij AURQ", "Send AURQ")));
        TextView sub = text(tr(
                "Transakcja jest podpisywana ML-DSA-87 lokalnie na telefonie i dopiero potem wysyłana do AuronQ Mainnet.",
                "The transaction is signed locally on the phone with ML-DSA-87 and only then submitted to AuronQ Mainnet."), 13, false);
        sub.setTextColor(MUTED);
        root.addView(sub, mt(5));

        LinearLayout card = card();
        recipientInput = input(tr("Adres odbiorcy aurq1…", "Recipient address aurq1…"));
        card.addView(label(tr("ADRES ODBIORCY", "RECIPIENT ADDRESS")));
        card.addView(recipientInput, mt(6));

        amountInput = input(tr("Kwota, np. 1.25000000", "Amount, e.g. 1.25000000"));
        amountInput.setInputType(InputType.TYPE_CLASS_NUMBER | InputType.TYPE_NUMBER_FLAG_DECIMAL);
        card.addView(label(tr("KWOTA AURQ", "AURQ AMOUNT")), mt(14));
        card.addView(amountInput, mt(6));

        sendPassword = input(tr("Hasło portfela", "Wallet password"));
        sendPassword.setInputType(InputType.TYPE_CLASS_TEXT | InputType.TYPE_TEXT_VARIATION_PASSWORD);
        card.addView(label(tr("HASŁO PORTFELA", "WALLET PASSWORD")), mt(14));
        card.addView(sendPassword, mt(6));

        sendButton = primaryButton(tr("Wyślij transakcję", "Send transaction"));
        sendButton.setEnabled(false);
        sendButton.setOnClickListener(v -> sendTransaction());
        card.addView(sendButton, mt(16));
        root.addView(card, mt(18));

        sendResult = text("", 12, false);
        sendResult.setTextColor(MUTED);
        sendResult.setTextIsSelectable(true);
        root.addView(sendResult, mt(14));

        TextView security = text(tr(
                "Przed wysłaniem sprawdź cały adres i kwotę. AuronQ Mobile nie wysyła hasła ani klucza prywatnego do publicznego noda.",
                "Before sending, verify the full address and amount. AuronQ Mobile never sends your password or private key to a public node."), 12, false);
        security.setTextColor(MUTED);
        root.addView(security, mt(16));

        return root;
    }

    private LinearLayout buildNetwork() {
        LinearLayout root = screenRoot();
        root.addView(title(tr("Sieć AuronQ na żywo", "AuronQ Network Live")));
        TextView sub = text(tr(
                "Podgląd tej samej sieci AuronQ Mainnet, z którą łączą się inne nody i portfele.",
                "Live view of the same AuronQ Mainnet used by other nodes and wallets."), 13, false);
        sub.setTextColor(MUTED);
        root.addView(sub, mt(5));

        netStatus = text(tr("● Łączenie…", "● Connecting…"), 14, true);
        netStatus.setTextColor(MUTED);
        root.addView(netStatus, mt(16));

        TextView firstSyncNote = text(tr(
                "Pierwsze uruchomienie może potrwać dłużej, ponieważ telefon lokalnie weryfikuje łańcuch nagłówków AQM64 od genesis. Kolejne uruchomienia korzystają z zapisanego, zweryfikowanego stanu i powinny być wyraźnie szybsze.",
                "The first launch can take longer because the phone locally verifies the AQM64 header chain from genesis. Later launches reuse the saved verified state and should be noticeably faster."), 11, false);
        firstSyncNote.setTextColor(MUTED);
        root.addView(firstSyncNote, mt(8));

        LinearLayout stats1 = row();
        netHeight = statCard(stats1, tr("WYSOKOŚĆ", "HEIGHT"), "—", tr("blok", "block"));
        netPeers = statCard(stats1, "PEERS", "0", tr("połączenia", "connections"));
        root.addView(stats1, mt(12));

        LinearLayout stats2 = row();
        netMempool = statCard(stats2, "MEMPOOL", "0", tr("transakcje", "transactions"));
        netIssued = statCard(stats2, tr("WYEMITOWANO", "ISSUED"), "—", "AURQ");
        root.addView(stats2, mt(10));

        LinearLayout details = card();
        details.addView(section(tr("SZCZEGÓŁY MAINNETU", "MAINNET DETAILS")));
        netNode = kv(details, "Node", "—");
        netNetworkId = kv(details, "Network ID", "—");
        netTip = kv(details, "Tip", "—");
        netWork = kv(details, "Chain work", "—");
        netObserved = kv(details, tr("Ostatni odczyt", "Last update"), "—");
        root.addView(details, mt(16));

        LinearLayout live = card();
        live.addView(section(tr("OSTATNIE BLOKI", "RECENT BLOCKS")));
        TextView info = text(tr("Aktualizacja co 5 sekund", "Updates every 5 seconds"), 11, false);
        info.setTextColor(MUTED);
        live.addView(info, mt(4));
        recentBlocks = new LinearLayout(this);
        recentBlocks.setOrientation(LinearLayout.VERTICAL);
        live.addView(recentBlocks, mt(10));
        root.addView(live, mt(16));

        Button refresh = secondaryButton(tr("Odśwież teraz", "Refresh now"));
        refresh.setOnClickListener(v -> refreshAll());
        root.addView(refresh, mt(14));

        TextView model = text(tr(
                "AuronQ Mobile w tym trybie nie weryfikuje całego AQM64 od genesis. Wymaga zgodnych danych z 3 publicznych grup sieci IP, co nie gwarantuje niezależności ich operatorów. Klucze pozostają lokalnie, podpisy są wykonywane na telefonie. Pełną walidację transakcji i UTXO wykonują pełne węzły. Nie traktuj tego modelu jak samodzielnej weryfikacji blockchaina.",
                "This AuronQ Mobile mode does not independently verify the entire AQM64 chain from genesis. It requires agreement from 3 public IP network groups, which does not prove different node ownership. Keys and signing stay local. Full nodes validate transactions and UTXOs. Peer agreement is not independent cryptographic blockchain verification."), 12, false);
        model.setTextColor(MUTED);
        root.addView(model, mt(18));

        return root;
    }

    private void showScreen(String which) {
        currentScreen = which;
        ((View) dashboardScreen.getParent()).setVisibility("home".equals(which) ? View.VISIBLE : View.GONE);
        ((View) walletScreen.getParent()).setVisibility("wallet".equals(which) ? View.VISIBLE : View.GONE);
        ((View) sendScreen.getParent()).setVisibility("send".equals(which) ? View.VISIBLE : View.GONE);
        ((View) networkScreen.getParent()).setVisibility("network".equals(which) ? View.VISIBLE : View.GONE);

        navStyle(navHome, "home".equals(which));
        navStyle(navWallet, "wallet".equals(which));
        navStyle(navSend, "send".equals(which));
        navStyle(navNetwork, "network".equals(which));

        if ("home".equals(which)) topTitle.setText(tr("Pulpit", "Dashboard"));
        if ("wallet".equals(which)) topTitle.setText(tr("Portfel", "Wallet"));
        if ("send".equals(which)) topTitle.setText(tr("Wyślij", "Send"));
        if ("network".equals(which)) topTitle.setText(tr("Sieć na żywo", "Network Live"));
        if ("wallet".equals(which)) refreshAll();
    }

    private String discoverResilientNode() throws Exception {
        Set<String> candidates = new LinkedHashSet<>();
        if (nodeUrl != null && !nodeUrl.isEmpty()) candidates.add(nodeUrl);

        String cached = prefs.getString("known_nodes", "[]");
        try {
            JSONArray a = new JSONArray(cached);
            for (int i = 0; i < a.length() && i < 32; i++) {
                String p = a.optString(i, "").trim();
                if (p.startsWith("https://") || p.startsWith("http://")) candidates.add(p);
            }
        } catch (Exception ignored) {
        }

        for (String candidate : candidates) {
            try {
                Bridge.status(candidate);
                return candidate;
            } catch (Exception ignored) {
            }
        }
        return Bridge.discoverNode();
    }

    private void rememberNetwork(String node) {
        try {
            Set<String> nodes = new LinkedHashSet<>();
            if (node != null && node.startsWith("https://")) nodes.add(node);

            try {
                String learned = Bridge.peerCandidates(node);
                JSONArray a = new JSONArray(learned);
                for (int i = 0; i < a.length() && nodes.size() < 32; i++) {
                    String p = a.optString(i, "").trim();
                    if (p.startsWith("https://") || p.startsWith("http://")) nodes.add(p);
                }
            } catch (Exception ignored) {
            }

            try {
                JSONArray old = new JSONArray(prefs.getString("known_nodes", "[]"));
                for (int i = 0; i < old.length() && nodes.size() < 32; i++) {
                    String p = old.optString(i, "").trim();
                    if (p.startsWith("https://") || p.startsWith("http://")) nodes.add(p);
                }
            } catch (Exception ignored) {
            }

            JSONArray out = new JSONArray();
            for (String p : nodes) out.put(p);
            prefs.edit()
                    .putString("last_node", node)
                    .putString("known_nodes", out.toString())
                    .apply();
        } catch (Exception ignored) {
        }
    }

    private void updateSendButtons() {
        boolean conflictingValidatedTip = deepVerifiedHeight == quickHeight
                && deepVerifiedHeight >= 0 && !deepVerifiedTip.isEmpty()
                && !deepVerifiedTip.equalsIgnoreCase(quickTip);
        boolean ready = quickAccountReady && !conflictingValidatedTip
                && walletFile.exists() && !walletAddress.isEmpty();
        if (sendButton != null) sendButton.setEnabled(ready);
        if (sendShortcutButton != null) sendShortcutButton.setEnabled(ready);
    }

    private void invalidateQuickAccount(String reason) {
        quickAccountReady = false;
        if (dashBalance != null) dashBalance.setText("—");
        if (walletBalance != null) walletBalance.setText("— AURQ");
        if (walletTrustStatus != null) {
            walletTrustStatus.setText(reason);
            walletTrustStatus.setTextColor(BLUE);
        }
        if (walletHistory != null) walletHistory.removeAllViews();
        if (walletHistoryStatus != null) {
            walletHistoryStatus.setText(tr("Historia niepotwierdzona — sprawdzanie węzłów",
                    "History unconfirmed — checking peers"));
            walletHistoryStatus.setTextColor(BLUE);
        }
        lastHistoryKey = "";
        updateSendButtons();
    }

    private void refreshAll() {
        if (liveBusy) return;
        liveBusy = true;
        final long requestGeneration = walletGeneration;
        executor.execute(() -> {
            try {
                String known = prefs.getString("known_nodes", "[]");
                String raw = Bridge.quickNetworkSnapshot(known);
                JSONObject state = new JSONObject(raw);
                String node = state.optString("node", "");
                long height = state.optLong("height", -1);
                String tip = state.optString("tip", "");
                runOnUiThread(() -> {
                    if (requestGeneration != walletGeneration) return;
                    applyQuickNetwork(raw);
                });
                String address = "";
                if (walletFile.exists()) address = Bridge.walletAddress(walletFile.getAbsolutePath());
                final String requestedAddress = address;
                if (!address.isEmpty()) {
                    String account = Bridge.quickAccountSnapshot(known, address, 50);
                    runOnUiThread(() -> {
                        if (requestGeneration != walletGeneration
                                || !walletAddress.equals(requestedAddress)) return;
                        applyQuickAccount(account);
                        if (quickAccountReady) {
                            scheduleVerifiedPeerHistory(known, requestedAddress, requestGeneration);
                        }
                        maybeStartDeepVerification(known);
                    });
                } else {
                    runOnUiThread(() -> {
                        if (requestGeneration == walletGeneration) maybeStartDeepVerification(known);
                    });
                }
                // Full recent-block downloads are read-only, secondary UI.
                // Never delay account reads with bulky transaction signatures.
                if ("network".equals(currentScreen) && !node.isEmpty()) {
                    try {
                        String blocks = Bridge.networkSnapshot(node, 6);
                        runOnUiThread(() -> {
                            if (requestGeneration != walletGeneration
                                    || quickHeight != height || !quickTip.equalsIgnoreCase(tip))
                                return;
                            applyQuickRecentBlocks(blocks);
                        });
                    } catch (Exception ignored) { }
                }
            } catch (Exception e) {
                String message = e.getMessage();
                runOnUiThread(() -> {
                    if (requestGeneration != walletGeneration) return;
                    invalidateQuickAccount(tr("Brak zgodności 3 grup węzłów — saldo nieustalone",
                            "Three network groups disagree — balance unavailable"));
                    dashNode.setText(tr("● Niepotwierdzona sieć", "● Network unconfirmed"));
                    dashNode.setTextColor(BLUE);
                    netStatus.setText(tr("● Oczekiwanie na 3 zgodne grupy węzłów",
                            "● Waiting for 3 agreeing network groups"));
                    netStatus.setTextColor(BLUE);
                    netObserved.setText(message == null ? "—" : message);
                });
            } finally {
                liveBusy = false;
            }
        });
    }

    // Account history can become expensive on a large full-node blockchain.
    // It must not block fast balances or the original transaction signer.
    private void scheduleVerifiedPeerHistory(String known, String address, long generation) {
        if (historyBusy || !"wallet".equals(currentScreen) || !quickAccountReady) return;
        final long height = quickHeight;
        final String tip = quickTip;
        final String key = address + ":" + height + ":" + tip;
        if (key.equals(lastHistoryKey)) return;
        historyBusy = true;
        walletHistoryStatus.setText(tr("Pobieram potwierdzoną historię w tle…",
                "Loading confirmed history in background…"));
        walletHistoryStatus.setTextColor(BLUE);
        historyExecutor.execute(() -> {
            String response = null, failure = null;
            try {
                response = Bridge.quickHistorySnapshot(known, address, 50);
            } catch (Exception e) {
                failure = e.getMessage();
            }
            final String raw = response;
            final String error = failure;
            runOnUiThread(() -> {
                historyBusy = false;
                if (generation != walletGeneration || !quickAccountReady
                        || !walletAddress.equals(address) || quickHeight != height
                        || !quickTip.equalsIgnoreCase(tip)) return;
                if (raw != null) {
                    try {
                        JSONObject j = new JSONObject(raw);
                        if (j.optLong("height", -1) == height
                                && tip.equalsIgnoreCase(j.optString("tip", ""))
                                && address.equals(j.optString("address", ""))
                                && j.optInt("peer_agreement", 0) == 3
                                && j.optJSONArray("items") != null) {
                            applyHistory(raw);
                            lastHistoryKey = key;
                            return;
                        }
                    } catch (Exception ignored) { }
                }
                walletHistoryStatus.setText(tr("Nie udało się potwierdzić historii: ",
                        "Unable to corroborate history: ") + (error == null ? "—" : error));
                walletHistoryStatus.setTextColor(BLUE);
            });
        });
    }

    // Stage 2 replays EVERY AQM64 header from the embedded Mainnet genesis.
    // It runs only after fast peer observations become usable. It NEVER gives
    // spending authority by itself and never overwrites the quorum balance.
    private void maybeStartDeepVerification(String known) {
        if (deepBusy || !networkReachable || quickHeight < 1) return;
        long now = android.os.SystemClock.elapsedRealtime();
        if (now - lastDeepAttemptMillis < 120000) return;
        if (deepVerifiedHeight == quickHeight
                && deepVerifiedTip.equalsIgnoreCase(quickTip)) return;
        deepBusy = true;
        lastDeepAttemptMillis = now;
        if (verificationStageText != null) {
            verificationStageText.setText(tr("Etap 1/2 aktywny • niezależna weryfikacja AQM64 trwa w tle",
                    "Stage 1/2 active • independent AQM64 verification runs in background"));
        }
        deepVerifier.execute(() -> {
            String report = null, failure = null;
            try {
                report = Bridge.quorumSnapshotVerified(known, headerCacheFile.getAbsolutePath(), 0);
            } catch (Exception e) {
                failure = e.getMessage();
            }
            final String completedReport = report;
            final String completedError = failure;
            runOnUiThread(() -> {
                deepBusy = false;
                if (completedReport != null) {
                    try {
                        JSONObject j = new JSONObject(completedReport);
                        if (j.optBoolean("header_verified", false)) {
                            long h = j.optLong("verified_height", -1);
                            String tip = j.optString("verified_tip", "");
                            String work = j.optString("verified_chain_work", "");
                            if (h >= 0 && tip.length() == 128 && !work.isEmpty()) {
                                deepVerifiedHeight = h;
                                deepVerifiedTip = tip;
                                deepVerifiedWork = work;
                                // Validated AQM64 contradicts an equally high
                                // quorum-reported tip: fail closed for spending.
                                if (quickHeight == h && !quickTip.equalsIgnoreCase(tip)) {
                                    invalidateQuickAccount(tr("Sprzeczny łańcuch — wysyłanie zablokowane",
                                            "Conflicting chain — spending disabled"));
                                }
                                updateSendButtons();
                            }
                        }
                    } catch (Exception ignored) { }
                }
                if (verificationStageText == null) return;
                if (deepVerifiedHeight == quickHeight && quickHeight >= 1
                        && deepVerifiedTip.equalsIgnoreCase(quickTip)) {
                    verificationStageText.setText(tr(
                        "Etap 2/2 — cała historia nagłówków AQM64 zweryfikowana lokalnie • saldo nadal z 3 węzłów",
                        "Stage 2/2 — entire AQM64 header history verified locally • balance still from 3 peers"));
                    verificationStageText.setTextColor(ACCENT);
                } else {
                    verificationStageText.setText(tr(
                        "Etap 1/2 — tryb 3 węzłów; pełna walidacja AQM64 nieukończona",
                        "Stage 1/2 — three-peer mode; full AQM64 verification incomplete"));
                    verificationStageText.setTextColor(BLUE);
                    // Old, newer or unavailable peer status is NOT equivalent
                    // to an independently validated tip.
                    if (completedError != null) netObserved.setText(completedError);
                }
            });
        });
    }

    private void applyQuickNetwork(String raw) {
        try {
            JSONObject state = new JSONObject(raw);
            if (!state.has("height") || !state.has("tip")
                    || state.optInt("peer_agreement", 0) < 3
                    || state.optJSONArray("network_groups") == null
                    || state.optJSONArray("network_groups").length() < 3) {
                invalidateQuickAccount(tr("Nie można potwierdzić stanu sieci",
                        "Could not confirm network state"));
                return;
            }
            long height = state.optLong("height", -1);
            String tip = state.optString("tip", "");
            if (height < 1 || tip.length() != 128) {
                invalidateQuickAccount(tr("Nieprawidłowe dane węzłów", "Invalid peer data"));
                return;
            }
            if (height != quickHeight || !tip.equalsIgnoreCase(quickTip)) {
                invalidateQuickAccount(tr("Nowy stan sieci — sprawdzam konto",
                        "New network state — verifying account"));
            }
            networkReachable = true;
            quickHeight = height;
            quickTip = tip;
            nodeUrl = state.optString("node", "");
            verifiedHeight = -1;
            verifiedTip = "";
            verifiedWork = "";
            if (verificationStageText != null) {
                if (deepVerifiedHeight == height && deepVerifiedTip.equalsIgnoreCase(tip)) {
                    verificationStageText.setText(tr("Etap 2/2 — AQM64 lokalnie zweryfikowane (stan konta z 3 węzłów)",
                        "Stage 2/2 — AQM64 verified locally (account data from 3 peers)"));
                    verificationStageText.setTextColor(ACCENT);
                } else {
                    verificationStageText.setText(tr("Etap 1/2 — 3 węzły zgodne; AQM64 weryfikowane w tle",
                        "Stage 1/2 — 3 peers agree; AQM64 checking in background"));
                    verificationStageText.setTextColor(BLUE);
                }
            }
            dashHeight.setText(String.valueOf(height));
            dashPeers.setText(String.valueOf(state.optInt("peers")));
            dashMempool.setText(String.valueOf(state.optInt("mempool")));
            dashNode.setText(tr("● Zgodność 3 grup sieci • bez pełnej weryfikacji AQM64",
                    "● 3 network groups agree • full AQM64 not verified"));
            dashNode.setTextColor(BLUE);
            dashTip.setText("Tip: " + shortHash(tip));
            netStatus.setText(tr("● Potwierdzone przez 3 grupy IP, nie kryptograficznie od genesis",
                    "● Observed by 3 IP groups, not cryptographically verified from genesis"));
            netStatus.setTextColor(BLUE);
            netHeight.setText(String.valueOf(height));
            netPeers.setText(String.valueOf(state.optInt("peers")));
            netMempool.setText(String.valueOf(state.optInt("mempool")));
            netIssued.setText(String.format(Locale.US, "%.8f", state.optDouble("issued_coins")));
            netNode.setText(nodeUrl);
            netNetworkId.setText(state.optString("network_id"));
            netTip.setText(tip);
            netWork.setText(state.optString("chain_work"));
            netObserved.setText(state.optString("observed_at"));
            topNetworkDot.setTextColor(BLUE);
        } catch (Exception e) {
            invalidateQuickAccount(tr("Odpowiedź sieci niepoprawna", "Invalid network response"));
        }
    }

    private void applyQuickAccount(String raw) {
        try {
            JSONObject state = new JSONObject(raw);
            String address = state.optString("address", "");
            String tip = state.optString("tip", "");
            long height = state.optLong("height", -1);
            String value = state.optString("spendable", "");
            JSONArray items = state.optJSONArray("items");
            if (!walletAddress.equals(address) || !quickTip.equalsIgnoreCase(tip)
                    || quickHeight != height || (deepVerifiedHeight == height
                        && !deepVerifiedTip.isEmpty() && !deepVerifiedTip.equalsIgnoreCase(tip))
                    || state.optInt("peer_agreement", 0) != 3
                    || !value.matches("[0-9]+\\.[0-9]{8}") || items == null) {
                invalidateQuickAccount(tr("Brak spójnych danych konta z 3 węzłów",
                        "Missing consistent account data from 3 peers"));
                return;
            }
            // Rendering MUST be atomic for this current wallet and chain tip.
            dashBalance.setText(value);
            walletBalance.setText(value + " AURQ");
            walletTrustStatus.setText(tr("Zgodne 3 grupy sieci • nie jest to pełny dowód blockchaina",
                    "3 network groups agree • not a full blockchain proof"));
            walletTrustStatus.setTextColor(BLUE);
            quickAccountReady = true;
            updateSendButtons();
        } catch (Exception e) {
            invalidateQuickAccount(tr("Nie można zweryfikować danych konta",
                    "Could not verify account data"));
        }
    }

    private void applyQuickRecentBlocks(String raw) {
        try {
            JSONObject s = new JSONObject(raw);
            if (s.optLong("height", -1) != quickHeight
                    || !quickTip.equalsIgnoreCase(s.optString("tip", ""))) return;
            JSONArray blocks = s.optJSONArray("blocks");
            if (blocks == null) return;
            recentBlocks.removeAllViews();
            for (int i = 0; i < blocks.length(); i++) {
                JSONObject b = blocks.optJSONObject(i);
                if (b == null || !b.has("transactions")) continue;
                recentBlocks.addView(blockRow(b.optLong("height"), b.optString("hash"),
                        b.optLong("timestamp"), b.optInt("transactions")));
            }
        } catch (Exception ignored) { }
    }

    private void applySnapshot(String raw) {
        try {
            JSONObject j = new JSONObject(raw);
            networkReachable = true;
            verifiedTip = j.optString("verified_tip", "");
            verifiedWork = j.optString("verified_chain_work", "");
            verifiedHeight = j.has("verified_height") ? j.optLong("verified_height", -1) : -1;
            long height = j.optLong("height");
            int peers = j.optInt("peers");
            int mempool = j.optInt("mempool");
            double issued = j.optDouble("issued_coins");

            int observedPeers = Math.max(1, j.optInt("peer_observed", 1));
            int agreeingPeers = Math.max(1, j.optInt("peer_agreement", 1));
            boolean multiPeerConfirmed = j.optBoolean("multi_peer_confirmed", false);
            boolean headerVerified = j.optBoolean("header_verified", false);
            long headersCheckedNow = j.optLong("headers_checked_now", 0);
            String agreementText = agreeingPeers + "/" + observedPeers;

            dashHeight.setText(String.valueOf(height));
            dashPeers.setText(String.valueOf(peers));
            dashMempool.setText(String.valueOf(mempool));
            dashNode.setText(headerVerified
                    ? tr("● AQM64 zweryfikowane • peery " + agreementText, "● AQM64 verified • peers " + agreementText)
                    : tr("● Nagłówki niezweryfikowane", "● Headers unverified"));
            dashNode.setTextColor(headerVerified ? ACCENT : DANGER);
            dashTip.setText("Tip: " + shortHash(j.optString("tip")));

            String verifyDetail = headersCheckedNow > 0
                    ? tr(" • nowych nagłówków: ", " • new headers: ") + headersCheckedNow
                    : "";
            netStatus.setText(headerVerified
                    ? tr("● Lokalnie zweryfikowano PoW/difficulty • zgodność peerów ", "● PoW/difficulty verified locally • peer agreement ")
                        + agreementText + verifyDetail
                    : tr("● Brak niezależnej weryfikacji nagłówków", "● No independent header verification"));
            netStatus.setTextColor(headerVerified ? ACCENT : DANGER);
            netHeight.setText(String.valueOf(height));
            netPeers.setText(String.valueOf(peers));
            netMempool.setText(String.valueOf(mempool));
            netIssued.setText(String.format(Locale.US, "%.8f", issued));
            netNode.setText(j.optString("node", nodeUrl));
            netNetworkId.setText(j.optString("network_id"));
            netTip.setText(j.optString("tip"));
            netWork.setText(j.optString("chain_work"));
            netObserved.setText(j.optString("observed_at"));

            topNetworkDot.setTextColor(headerVerified ? ACCENT : DANGER);

            recentBlocks.removeAllViews();
            JSONArray blocks = j.optJSONArray("blocks");
            if (blocks == null || blocks.length() == 0) {
                TextView empty = text(tr("Brak danych o blokach", "No block data"), 12, false);
                empty.setTextColor(MUTED);
                recentBlocks.addView(empty);
            } else {
                for (int i = 0; i < blocks.length(); i++) {
                    JSONObject b = blocks.optJSONObject(i);
                    if (b == null) continue;
                    recentBlocks.addView(blockRow(
                            b.optLong("height"),
                            b.optString("hash"),
                            b.optLong("timestamp"),
                            b.optInt("transactions")));
                }
            }
        } catch (Exception e) {
            setNetworkOffline(tr("Nieprawidłowa odpowiedź sieci", "Invalid network response"));
        }
    }

    private void applyNetworkPreview(String raw) {
        try {
            JSONObject j = new JSONObject(raw);
            networkReachable = true;
            nodeUrl = j.optString("node", nodeUrl);
            long height = j.optLong("height");
            int peers = j.optInt("peers");
            int mempool = j.optInt("mempool");

            dashHeight.setText(String.valueOf(height));
            dashPeers.setText(String.valueOf(peers));
            dashMempool.setText(String.valueOf(mempool));
            dashNode.setText(headerCacheFile != null && headerCacheFile.exists()
                    ? tr("● Połączono • trwa weryfikacja AQM64…", "● Connected • verifying AQM64…")
                    : tr("● Połączono • pierwsza weryfikacja AQM64 może potrwać dłużej…", "● Connected • first AQM64 verification may take longer…"));
            dashNode.setTextColor(BLUE);
            dashTip.setText("Tip: " + shortHash(j.optString("tip")));

            netStatus.setText(tr("● Sieć osiągalna • lokalna weryfikacja nagłówków trwa", "● Network reachable • local header verification in progress"));
            netStatus.setTextColor(BLUE);
            netHeight.setText(String.valueOf(height));
            netPeers.setText(String.valueOf(peers));
            netMempool.setText(String.valueOf(mempool));
            netNode.setText(j.optString("node", nodeUrl));
            netTip.setText(j.optString("tip"));
            netWork.setText(j.optString("chain_work"));
            netObserved.setText(j.optString("observed_at"));
            topNetworkDot.setTextColor(BLUE);
        } catch (Exception ignored) {
        }
    }

    private void handleNetworkRefreshFailure(String error) {
        if (networkReachable) {
            topNetworkDot.setTextColor(BLUE);
            dashNode.setText(verifiedHeight >= 0
                    ? tr("● Ostatni stan zweryfikowany • ponawiam połączenie", "● Last state verified • reconnecting")
                    : tr("● Sieć osiągalna • weryfikacja przerwana, ponawiam", "● Network reachable • verification interrupted, retrying"));
            dashNode.setTextColor(BLUE);
            netStatus.setText(verifiedHeight >= 0
                    ? tr("● Chwilowy błąd odświeżania • zachowano zweryfikowany stan", "● Temporary refresh error • verified state preserved")
                    : tr("● Połączono z peerem • ponawiam lokalną weryfikację AQM64", "● Peer reachable • retrying local AQM64 verification"));
            netStatus.setTextColor(BLUE);
            netObserved.setText(error == null ? "—" : error);
            return;
        }
        setNetworkOffline(error);
    }

    private void setNetworkOffline(String error) {
        networkReachable = false;
        topNetworkDot.setTextColor(DANGER);
        dashNode.setText(tr("● Brak połączenia", "● Offline"));
        dashNode.setTextColor(DANGER);
        netStatus.setText(tr("● Brak połączenia z AuronQ Mainnet", "● No connection to AuronQ Mainnet"));
        netStatus.setTextColor(DANGER);
        netObserved.setText(error == null ? "—" : error);
        nodeUrl = "";
        verifiedTip = "";
        verifiedWork = "";
        verifiedHeight = -1;
    }

    private View blockRow(long height, String hash, long timestamp, int txs) {
        LinearLayout row = new LinearLayout(this);
        row.setOrientation(LinearLayout.VERTICAL);
        row.setPadding(dp(12), dp(10), dp(12), dp(10));
        row.setBackground(round(PANEL2, LINE, 12));

        LinearLayout top = new LinearLayout(this);
        top.setOrientation(LinearLayout.HORIZONTAL);
        TextView h = text(tr("Blok #", "Block #") + height, 14, true);
        top.addView(h, new LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f));
        TextView tx = text(txs + " tx", 11, true);
        tx.setTextColor(BLUE);
        top.addView(tx);
        row.addView(top);

        TextView hs = text(shortHash(hash), 11, false);
        hs.setTextColor(MUTED);
        hs.setTextIsSelectable(true);
        row.addView(hs, mt(4));

        TextView time = text(formatTime(timestamp), 10, false);
        time.setTextColor(MUTED);
        row.addView(time, mt(3));

        LinearLayout.LayoutParams p = new LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT,
                LinearLayout.LayoutParams.WRAP_CONTENT);
        p.bottomMargin = dp(8);
        row.setLayoutParams(p);
        return row;
    }

    private void applyBalance(String raw) {
        try {
            JSONObject j = new JSONObject(raw);
            String spendable = j.optString("spendable", "0.00000000");
            dashBalance.setText(spendable);
            walletBalance.setText(spendable + " AURQ");
        } catch (Exception ignored) {
        }
    }

    private void applyHistory(String raw) {
        try {
            JSONObject j = new JSONObject(raw);
            JSONArray items = j.optJSONArray("items");
            walletHistory.removeAllViews();
            int count = items == null ? 0 : items.length();
            int observed = j.optInt("peer_observed", 1);
            int agreeing = j.optInt("peer_agreement", 1);
            boolean multi = j.optBoolean("multi_peer_confirmed", false);
            walletHistoryStatus.setText(count + " " + tr("transakcji", "transactions")
                    + " • " + tr("zgodność ", "agreement ") + agreeing + "/" + observed);
            walletHistoryStatus.setTextColor(multi ? ACCENT : MUTED);
            if (count == 0) {
                TextView empty = text(tr("Brak transakcji dla tego portfela.", "No transactions for this wallet."), 12, false);
                empty.setTextColor(MUTED);
                walletHistory.addView(empty);
                return;
            }
            for (int i = 0; i < count; i++) {
                JSONObject item = items.optJSONObject(i);
                if (item != null) walletHistory.addView(historyRow(item));
            }
        } catch (Exception e) {
            walletHistoryStatus.setText(tr("Nie udało się odczytać historii.", "Could not read transaction history."));
            walletHistoryStatus.setTextColor(DANGER);
        }
    }

    private View historyRow(JSONObject item) {
        LinearLayout row = new LinearLayout(this);
        row.setOrientation(LinearLayout.VERTICAL);
        row.setPadding(dp(12), dp(10), dp(12), dp(10));
        row.setBackground(round(PANEL2, LINE, 12));

        String type = item.optString("type", "received");
        String status = item.optString("status", "confirmed");
        String typeText;
        if ("genesis".equals(type)) typeText = tr("Genesis", "Genesis");
        else if ("mining".equals(type)) typeText = tr("Nagroda z kopania", "Mining reward");
        else if ("sent".equals(type)) typeText = tr("Wysłano", "Sent");
        else typeText = tr("Odebrano", "Received");

        String statusText;
        if ("pending".equals(status)) statusText = tr("Oczekuje", "Pending");
        else if ("immature".equals(status)) statusText = tr("Dojrzewa", "Immature");
        else statusText = tr("Potwierdzona", "Confirmed");

        LinearLayout top = new LinearLayout(this);
        top.setOrientation(LinearLayout.HORIZONTAL);
        TextView kind = text(typeText, 14, true);
        top.addView(kind, new LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f));
        String amount = item.optString("amount", "0.00000000");
        TextView value = text(("sent".equals(type) ? "-" : "+") + amount + " AURQ", 13, true);
        value.setTextColor("sent".equals(type) ? DANGER : ACCENT);
        top.addView(value);
        row.addView(top);

        TextView stat = text(statusText + " • " + item.optLong("confirmations", 0) + " " + tr("potw.", "conf."), 10, true);
        stat.setTextColor("pending".equals(status) ? BLUE : MUTED);
        row.addView(stat, mt(4));

        String block = item.isNull("height") ? "mempool" : "#" + item.optLong("height");
        TextView meta = text(formatTime(item.optLong("timestamp")) + " • " + block, 10, false);
        meta.setTextColor(MUTED);
        row.addView(meta, mt(3));

        TextView txid = text(shortHash(item.optString("txid")), 10, false);
        txid.setTextColor(MUTED);
        txid.setTextIsSelectable(true);
        row.addView(txid, mt(3));

        LinearLayout.LayoutParams p = new LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT,
                LinearLayout.LayoutParams.WRAP_CONTENT);
        p.bottomMargin = dp(8);
        row.setLayoutParams(p);
        return row;
    }

    private void createWallet() {
        if (walletFile.exists()) {
            toast(tr(
                    "Portfel już istnieje. Zrób backup albo usuń go przed utworzeniem nowego.",
                    "A wallet already exists. Back it up or delete it before creating a new one."));
            return;
        }
        String password = passwordInput.getText().toString();
        run(tr("Tworzenie portfela…", "Creating wallet…"),
                () -> Bridge.createWallet(walletFile.getAbsolutePath(), password),
                value -> {
                    walletAddress = value;
                    lastHistoryKey = "";
                    loadWalletState();
                    refreshAll();
                    toast(tr("Portfel utworzony", "Wallet created"));
                });
    }

    private void importWallet() {
        Intent i = new Intent(Intent.ACTION_OPEN_DOCUMENT);
        i.setType("*/*");
        i.addCategory(Intent.CATEGORY_OPENABLE);
        startActivityForResult(i, IMPORT_WALLET);
    }

    private void backupWallet() {
        if (!walletFile.exists()) {
            toast(tr("Brak portfela do backupu", "No wallet to back up"));
            return;
        }
        Intent i = new Intent(Intent.ACTION_CREATE_DOCUMENT);
        i.setType("application/json");
        i.putExtra(Intent.EXTRA_TITLE, "AuronQ-main.wallet");
        startActivityForResult(i, EXPORT_WALLET);
    }

    private void copyAddress() {
        if (walletAddress.isEmpty()) {
            toast(tr("Brak adresu", "No address"));
            return;
        }
        ClipboardManager cm = (ClipboardManager) getSystemService(Context.CLIPBOARD_SERVICE);
        cm.setPrimaryClip(ClipData.newPlainText("AuronQ address", walletAddress));
        toast(tr("Adres skopiowany", "Address copied"));
    }

    private void sendTransaction() {
        if (!walletFile.exists() || walletAddress.isEmpty()) {
            toast(tr("Najpierw utwórz albo zaimportuj portfel", "Create or import a wallet first"));
            return;
        }
        if (!quickAccountReady || nodeUrl.isEmpty() || quickHeight < 0 || quickTip.isEmpty()) {
            toast(tr("Zaczekaj na zgodność 3 grup węzłów i bieżące saldo",
                    "Wait for three network groups and current wallet balance"));
            return;
        }
        String password = sendPassword.getText().toString();
        String to = recipientInput.getText().toString().trim();
        String amount = amountInput.getText().toString().trim();

        new AlertDialog.Builder(this)
                .setTitle(tr("Potwierdź wysyłkę", "Confirm transaction"))
                .setMessage(tr(
                        "Wyślij " + amount + " AURQ na:\n\n" + to
                        + "\n\nUwaga: saldo potwierdzają 3 grupy węzłów, nie pełna weryfikacja łańcucha. Transakcja jest nieodwracalna.",
                        "Send " + amount + " AURQ to:\n\n" + to
                        + "\n\nWarning: 3 network groups confirm the balance; full chain verification is not performed. Sending is irreversible."))
                .setNegativeButton(tr("Anuluj", "Cancel"), null)
                .setPositiveButton(tr("Wyślij", "Send"), (d, w) ->
                        run(tr("Podpisywanie i wysyłanie…", "Signing and sending…"),
                                () -> Bridge.sendWithQuickQuorum(
                                        prefs.getString("known_nodes", "[]"),
                                        walletFile.getAbsolutePath(),
                                        password, to, amount),
                                value -> {
                                    try {
                                        JSONObject j = new JSONObject(value);
                                        int accepted = j.optInt("direct_accepted", 1);
                                        int attempted = j.optInt("broadcast_attempted", 1);
                                        int utxoAgree = j.optInt("utxo_peer_agreement", 1);
                                        int utxoObserved = j.optInt("utxo_peer_observed", 1);
                                        sendResult.setText("TXID:\n" + j.optString("txid")
                                                + "\nFee: " + j.optString("fee") + " AURQ"
                                                + "\n" + tr("UTXO zgodne: ", "UTXO agreement: ") + utxoAgree + "/" + utxoObserved
                                                + "\n" + tr("Rozgłoszenie bezpośrednie: ", "Direct broadcast: ") + accepted + "/" + attempted);
                                    } catch (Exception e) {
                                        sendResult.setText(value);
                                    }
                                    sendPassword.setText("");
                                    refreshAll();
                                }))
                .show();
    }

    private void deleteWallet() {
        if (!walletFile.exists()) {
            toast(tr("Brak lokalnego portfela", "No local wallet"));
            return;
        }
        new AlertDialog.Builder(this)
                .setTitle(tr("Usunąć portfel?", "Delete wallet?"))
                .setMessage(tr(
                        "Najpierw upewnij się, że masz backup. Usunięcie pliku bez backupu może oznaczać trwałą utratę dostępu do środków.",
                        "Make sure you have a backup first. Deleting the wallet file without a backup can permanently remove access to funds."))
                .setNegativeButton(tr("Anuluj", "Cancel"), null)
                .setPositiveButton(tr("Usuń", "Delete"), (d, w) -> {
                    if (walletFile.delete()) {
                        walletAddress = "";
                        lastHistoryKey = "";
                        passwordInput.setText("");
                        loadWalletState();
                        refreshAll();
                        toast(tr("Lokalny portfel usunięty", "Local wallet deleted"));
                    } else {
                        toast(tr("Nie udało się usunąć pliku", "Could not delete the file"));
                    }
                })
                .show();
    }

    private void loadWalletState() {
        walletGeneration++;
        invalidateQuickAccount(tr("Sprawdzam stan nowego portfela…",
                "Checking newly selected wallet…"));
        if (!walletFile.exists()) {
            walletAddress = "";
            walletAddressText.setText(tr("Brak lokalnego portfela", "No local wallet"));
            invalidateQuickAccount(tr("Brak lokalnego portfela", "No local wallet"));
            backupButton.setEnabled(false);
            copyButton.setEnabled(false);
            sendShortcutButton.setEnabled(false);
            deleteButton.setEnabled(false);
            if (walletHistory != null) walletHistory.removeAllViews();
            if (walletHistoryStatus != null) {
                walletHistoryStatus.setText(tr("Brak lokalnego portfela", "No local wallet"));
                walletHistoryStatus.setTextColor(MUTED);
            }
            return;
        }
        final long generation = walletGeneration;
        walletAddress = "";
        executor.execute(() -> {
            try {
                String addr = Bridge.walletAddress(walletFile.getAbsolutePath());
                runOnUiThread(() -> {
                    if (generation != walletGeneration) return;
                    walletAddress = addr;
                    walletAddressText.setText(addr);
                    backupButton.setEnabled(true);
                    copyButton.setEnabled(true);
                    updateSendButtons();
                    deleteButton.setEnabled(true);
                    refreshAll();
                });
            } catch (Exception e) {
                runOnUiThread(() -> toast(tr("Błąd portfela: ", "Wallet error: ") + e.getMessage()));
            }
        });
    }

    private void run(String busy, Task task, Success success) {
        if (sendResult != null) sendResult.setText(busy);
        executor.execute(() -> {
            try {
                String result = task.run();
                runOnUiThread(() -> success.accept(result));
            } catch (Exception e) {
                runOnUiThread(() -> {
                    if (sendResult != null) sendResult.setText(tr("Błąd: ", "Error: ") + e.getMessage());
                    toast(e.getMessage());
                });
            }
        });
    }

    @Override
    @SuppressWarnings("deprecation")
    protected void onActivityResult(int requestCode, int resultCode, Intent data) {
        super.onActivityResult(requestCode, resultCode, data);
        if (resultCode != RESULT_OK || data == null || data.getData() == null) return;
        Uri uri = data.getData();

        if (requestCode == IMPORT_WALLET) {
            File tmp = new File(walletDir, "import.tmp");
            try (InputStream in = getContentResolver().openInputStream(uri);
                 OutputStream out = new FileOutputStream(tmp)) {
                copy(in, out);
                String addr = Bridge.walletAddress(tmp.getAbsolutePath());
                Runnable install = () -> {
                    if (walletFile.exists() && !walletFile.delete()) {
                        toast(tr("Nie udało się zastąpić obecnego portfela", "Could not replace the current wallet"));
                        tmp.delete();
                        return;
                    }
                    if (!tmp.renameTo(walletFile)) {
                        toast(tr("Nie udało się zapisać importowanego portfela", "Could not save the imported wallet"));
                        tmp.delete();
                        return;
                    }
                    walletAddress = addr;
                    lastHistoryKey = "";
                    loadWalletState();
                    refreshAll();
                    toast(tr("Portfel zaimportowany", "Wallet imported"));
                };
                if (walletFile.exists()) {
                    new AlertDialog.Builder(this)
                            .setTitle(tr("Zastąpić portfel?", "Replace wallet?"))
                            .setMessage(tr(
                                    "Obecny lokalny portfel zostanie zastąpiony. Zrób backup przed kontynuacją.",
                                    "The current local wallet will be replaced. Create a backup before continuing."))
                            .setNegativeButton(tr("Anuluj", "Cancel"), (d, w) -> tmp.delete())
                            .setPositiveButton(tr("Zastąp", "Replace"), (d, w) -> install.run())
                            .show();
                } else {
                    install.run();
                }
            } catch (Exception e) {
                tmp.delete();
                toast(tr("Import nieudany: ", "Import failed: ") + e.getMessage());
            }
        } else if (requestCode == EXPORT_WALLET) {
            try (InputStream in = new java.io.FileInputStream(walletFile);
                 OutputStream out = getContentResolver().openOutputStream(uri, "w")) {
                copy(in, out);
                toast(tr("Backup zapisany", "Backup saved"));
            } catch (Exception e) {
                toast(tr("Backup nieudany: ", "Backup failed: ") + e.getMessage());
            }
        }
    }

    private static void copy(InputStream in, OutputStream out) throws Exception {
        if (in == null || out == null) throw new Exception("cannot open file");
        byte[] buf = new byte[8192];
        int n;
        while ((n = in.read(buf)) >= 0) out.write(buf, 0, n);
        out.flush();
    }

    private ScrollView wrapScroll(LinearLayout child) {
        ScrollView scroll = new ScrollView(this);
        scroll.setFillViewport(true);
        scroll.setBackgroundColor(BG);
        scroll.addView(child);
        return scroll;
    }

    private LinearLayout screenRoot() {
        LinearLayout l = new LinearLayout(this);
        l.setOrientation(LinearLayout.VERTICAL);
        l.setPadding(dp(16), dp(20), dp(16), dp(28));
        return l;
    }

    private LinearLayout card() {
        LinearLayout l = new LinearLayout(this);
        l.setOrientation(LinearLayout.VERTICAL);
        l.setPadding(dp(16), dp(16), dp(16), dp(16));
        l.setBackground(round(PANEL, LINE, 18));
        return l;
    }

    private LinearLayout row() {
        LinearLayout l = new LinearLayout(this);
        l.setOrientation(LinearLayout.HORIZONTAL);
        l.setGravity(Gravity.TOP);
        return l;
    }

    private TextView statCard(LinearLayout row, String label, String value, String foot) {
        LinearLayout c = card();
        TextView a = text(label, 10, true);
        a.setTextColor(MUTED);
        c.addView(a);
        TextView v = text(value, 21, true);
        c.addView(v, mt(7));
        TextView f = text(foot, 10, false);
        f.setTextColor(MUTED);
        c.addView(f, mt(4));

        LinearLayout.LayoutParams p = new LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f);
        p.setMarginEnd(dp(5));
        row.addView(c, p);
        return v;
    }

    private TextView kv(LinearLayout parent, String key, String initial) {
        TextView k = text(key, 10, true);
        k.setTextColor(MUTED);
        parent.addView(k, mt(12));
        TextView v = text(initial, 11, false);
        v.setTextIsSelectable(true);
        parent.addView(v, mt(4));
        return v;
    }

    private TextView title(String s) {
        return text(s, 25, true);
    }

    private TextView section(String s) {
        TextView t = text(s, 11, true);
        t.setTextColor(BLUE);
        return t;
    }

    private TextView label(String s) {
        TextView t = text(s, 10, true);
        t.setTextColor(MUTED);
        return t;
    }

    private TextView text(String s, int sp, boolean bold) {
        TextView t = new TextView(this);
        t.setText(s);
        t.setTextSize(sp);
        t.setTextColor(TEXT);
        if (bold) t.setTypeface(Typeface.DEFAULT_BOLD);
        return t;
    }

    private EditText input(String hint) {
        EditText e = new EditText(this);
        e.setHint(hint);
        e.setHintTextColor(Color.rgb(100, 117, 138));
        e.setTextColor(TEXT);
        e.setSingleLine(true);
        e.setPadding(dp(12), dp(12), dp(12), dp(12));
        e.setBackground(round(Color.rgb(8, 17, 27), LINE, 11));
        return e;
    }

    private Button navButton(String label, Runnable action) {
        Button b = new Button(this);
        b.setText(label);
        b.setTextSize(10);
        b.setAllCaps(false);
        b.setGravity(Gravity.CENTER);
        b.setTextColor(MUTED);
        b.setBackgroundColor(Color.TRANSPARENT);
        b.setOnClickListener(v -> action.run());
        return b;
    }

    private void navStyle(Button b, boolean active) {
        b.setTextColor(active ? TEXT : MUTED);
        b.setBackground(active ? round(Color.rgb(17, 27, 41), Color.TRANSPARENT, 12) : null);
    }

    private Button primaryButton(String label) {
        Button b = new Button(this);
        b.setText(label);
        b.setAllCaps(false);
        b.setTextColor(Color.rgb(6, 18, 15));
        b.setTypeface(Typeface.DEFAULT_BOLD);
        b.setBackground(roundGradient(ACCENT, Color.rgb(89, 205, 167), 11));
        return b;
    }

    private Button secondaryButton(String label) {
        Button b = new Button(this);
        b.setText(label);
        b.setAllCaps(false);
        b.setTextColor(TEXT);
        b.setBackground(round(Color.rgb(26, 38, 56), LINE, 11));
        return b;
    }

    private Button dangerButton(String label) {
        Button b = new Button(this);
        b.setText(label);
        b.setAllCaps(false);
        b.setTextColor(Color.rgb(255, 158, 170));
        b.setBackground(round(Color.rgb(50, 24, 31), Color.rgb(100, 44, 55), 11));
        return b;
    }

    private GradientDrawable round(int fill, int stroke, int radius) {
        GradientDrawable g = new GradientDrawable();
        g.setColor(fill);
        if (stroke != Color.TRANSPARENT) g.setStroke(dp(1), stroke);
        g.setCornerRadius(dp(radius));
        return g;
    }

    private GradientDrawable roundGradient(int a, int b, int radius) {
        GradientDrawable g = new GradientDrawable(GradientDrawable.Orientation.TL_BR, new int[]{a, b});
        g.setCornerRadius(dp(radius));
        return g;
    }

    private LinearLayout.LayoutParams mt(int top) {
        LinearLayout.LayoutParams p = new LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT,
                LinearLayout.LayoutParams.WRAP_CONTENT);
        p.topMargin = dp(top);
        return p;
    }

    private LinearLayout.LayoutParams weight() {
        LinearLayout.LayoutParams p = new LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f);
        p.setMarginEnd(dp(5));
        return p;
    }

    private String shortHash(String s) {
        if (s == null) return "—";
        if (s.length() <= 24) return s;
        return s.substring(0, 12) + "…" + s.substring(s.length() - 10);
    }

    private String formatTime(long unix) {
        if (unix <= 0) return "—";
        SimpleDateFormat f = new SimpleDateFormat("yyyy-MM-dd HH:mm:ss 'UTC'", Locale.US);
        f.setTimeZone(TimeZone.getTimeZone("UTC"));
        return f.format(new Date(unix * 1000L));
    }

    private int dp(int v) {
        return Math.round(v * getResources().getDisplayMetrics().density);
    }

    private void toast(String msg) {
        Toast.makeText(this, msg == null ? tr("Błąd", "Error") : msg, Toast.LENGTH_LONG).show();
    }

    @Override
    protected void onDestroy() {
        handler.removeCallbacks(liveLoop);
        executor.shutdownNow();
        deepVerifier.shutdownNow();
        historyExecutor.shutdownNow();
        super.onDestroy();
    }
}
