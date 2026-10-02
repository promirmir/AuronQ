package com.auronq.mobile;

import android.app.Activity;
import android.app.AlertDialog;
import android.content.ClipData;
import android.content.ClipboardManager;
import android.content.Context;
import android.content.Intent;
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
import java.util.Locale;
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
    private final Handler handler = new Handler(Looper.getMainLooper());

    private File walletDir;
    private File walletFile;
    private String nodeUrl = "";
    private String walletAddress = "";
    private boolean liveBusy = false;

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
    private Button deleteButton;

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

        walletDir = new File(getFilesDir(), "wallets");
        if (!walletDir.exists()) walletDir.mkdirs();
        walletFile = new File(walletDir, "main.wallet");

        setContentView(buildUi());
        loadWalletState();
        showScreen("home");
        refreshAll();
    }

    @Override
    protected void onResume() {
        super.onResume();
        handler.removeCallbacks(liveLoop);
        handler.post(liveLoop);
    }

    @Override
    protected void onPause() {
        handler.removeCallbacks(liveLoop);
        super.onPause();
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
        topTitle = text("Pulpit", 11, false);
        topTitle.setTextColor(MUTED);
        brand.addView(name);
        brand.addView(topTitle);
        top.addView(brand, new LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f));

        topNetworkDot = text("●", 16, true);
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

        navHome = navButton("⌂\nPulpit", () -> showScreen("home"));
        navWallet = navButton("◫\nPortfel", () -> showScreen("wallet"));
        navSend = navButton("↗\nWyślij", () -> showScreen("send"));
        navNetwork = navButton("◎\nSieć", () -> showScreen("network"));
        nav.addView(navHome, weight());
        nav.addView(navWallet, weight());
        nav.addView(navSend, weight());
        nav.addView(navNetwork, weight());
        root.addView(nav);

        return root;
    }

    private LinearLayout buildDashboard() {
        LinearLayout root = screenRoot();

        TextView hello = text("Witaj w AuronQ", 26, true);
        root.addView(hello);
        TextView sub = text("Mobilny portfel AuronQ Mainnet. Klucze pozostają lokalnie na telefonie.", 13, false);
        sub.setTextColor(MUTED);
        root.addView(sub, mt(5));

        LinearLayout stats1 = row();
        dashBalance = statCard(stats1, "SALDO", "0.00000000", "AURQ");
        dashHeight = statCard(stats1, "WYSOKOŚĆ", "—", "łańcuch");
        root.addView(stats1, mt(20));

        LinearLayout stats2 = row();
        dashPeers = statCard(stats2, "PEERS", "0", "node");
        dashMempool = statCard(stats2, "MEMPOOL", "0", "transakcje");
        root.addView(stats2, mt(10));

        LinearLayout networkCard = card();
        networkCard.addView(section("STAN SIECI"));
        dashNode = text("Szukanie AuronQ Mainnet…", 14, true);
        networkCard.addView(dashNode, mt(10));
        dashTip = text("Tip: —", 11, false);
        dashTip.setTextColor(MUTED);
        dashTip.setTextIsSelectable(true);
        networkCard.addView(dashTip, mt(6));

        Button openNetwork = primaryButton("Szczegóły sieci na żywo");
        openNetwork.setOnClickListener(v -> showScreen("network"));
        networkCard.addView(openNetwork, mt(14));
        root.addView(networkCard, mt(16));

        LinearLayout walletCard = card();
        walletCard.addView(section("TWÓJ PORTFEL"));
        TextView note = text("Adres i saldo są odczytywane z tej samej sieci AuronQ Mainnet, z której korzystają inni użytkownicy.", 13, false);
        note.setTextColor(MUTED);
        walletCard.addView(note, mt(8));
        Button openWallet = secondaryButton("Otwórz portfel");
        openWallet.setOnClickListener(v -> showScreen("wallet"));
        walletCard.addView(openWallet, mt(14));
        root.addView(walletCard, mt(16));

        return root;
    }

    private LinearLayout buildWallet() {
        LinearLayout root = screenRoot();
        root.addView(title("Portfel"));
        TextView sub = text("Ten sam format zaszyfrowanego pliku .wallet co w AuronQ Desktop.", 13, false);
        sub.setTextColor(MUTED);
        root.addView(sub, mt(5));

        LinearLayout balanceCard = card();
        balanceCard.addView(section("SALDO DOSTĘPNE"));
        walletBalance = text("0.00000000 AURQ", 29, true);
        balanceCard.addView(walletBalance, mt(8));
        root.addView(balanceCard, mt(18));

        LinearLayout addressCard = card();
        addressCard.addView(section("ADRES"));
        walletAddressText = text("Brak lokalnego portfela", 12, false);
        walletAddressText.setTextColor(MUTED);
        walletAddressText.setTextIsSelectable(true);
        addressCard.addView(walletAddressText, mt(9));
        root.addView(addressCard, mt(12));

        passwordInput = input("Hasło portfela (min. 12 znaków)");
        passwordInput.setInputType(InputType.TYPE_CLASS_TEXT | InputType.TYPE_TEXT_VARIATION_PASSWORD);
        root.addView(passwordInput, mt(14));

        LinearLayout r1 = row();
        Button createButton = primaryButton("Utwórz");
        createButton.setOnClickListener(v -> createWallet());
        Button importButton = secondaryButton("Import");
        importButton.setOnClickListener(v -> importWallet());
        r1.addView(createButton, weight());
        r1.addView(importButton, weight());
        root.addView(r1, mt(12));

        LinearLayout r2 = row();
        backupButton = secondaryButton("Backup");
        backupButton.setOnClickListener(v -> backupWallet());
        copyButton = secondaryButton("Kopiuj adres");
        copyButton.setOnClickListener(v -> copyAddress());
        r2.addView(backupButton, weight());
        r2.addView(copyButton, weight());
        root.addView(r2, mt(8));

        sendShortcutButton = primaryButton("Wyślij AURQ");
        sendShortcutButton.setOnClickListener(v -> showScreen("send"));
        root.addView(sendShortcutButton, mt(12));

        deleteButton = dangerButton("Usuń lokalny portfel");
        deleteButton.setOnClickListener(v -> deleteWallet());
        root.addView(deleteButton, mt(28));

        TextView warning = text("Zrób backup przed usunięciem. Usunięcie jedynej kopii portfela może oznaczać trwałą utratę dostępu do środków.", 12, false);
        warning.setTextColor(MUTED);
        root.addView(warning, mt(10));

        return root;
    }

    private LinearLayout buildSend() {
        LinearLayout root = screenRoot();
        root.addView(title("Wyślij AURQ"));
        TextView sub = text("Transakcja jest podpisywana ML-DSA-87 lokalnie na telefonie i dopiero potem wysyłana do AuronQ Mainnet.", 13, false);
        sub.setTextColor(MUTED);
        root.addView(sub, mt(5));

        LinearLayout card = card();
        recipientInput = input("Adres odbiorcy aurq1…");
        card.addView(label("ADRES ODBIORCY"));
        card.addView(recipientInput, mt(6));

        amountInput = input("Kwota, np. 1.25000000");
        amountInput.setInputType(InputType.TYPE_CLASS_NUMBER | InputType.TYPE_NUMBER_FLAG_DECIMAL);
        card.addView(label("KWOTA AURQ"), mt(14));
        card.addView(amountInput, mt(6));

        sendPassword = input("Hasło portfela");
        sendPassword.setInputType(InputType.TYPE_CLASS_TEXT | InputType.TYPE_TEXT_VARIATION_PASSWORD);
        card.addView(label("HASŁO PORTFELA"), mt(14));
        card.addView(sendPassword, mt(6));

        Button sendButton = primaryButton("Wyślij transakcję");
        sendButton.setOnClickListener(v -> sendTransaction());
        card.addView(sendButton, mt(16));
        root.addView(card, mt(18));

        sendResult = text("", 12, false);
        sendResult.setTextColor(MUTED);
        sendResult.setTextIsSelectable(true);
        root.addView(sendResult, mt(14));

        TextView security = text("Przed wysłaniem sprawdź cały adres i kwotę. AuronQ Mobile nie wysyła hasła ani klucza prywatnego do publicznego noda.", 12, false);
        security.setTextColor(MUTED);
        root.addView(security, mt(16));

        return root;
    }

    private LinearLayout buildNetwork() {
        LinearLayout root = screenRoot();
        root.addView(title("Sieć AuronQ na żywo"));
        TextView sub = text("Podgląd tej samej sieci AuronQ Mainnet, z którą łączą się inne nody i portfele.", 13, false);
        sub.setTextColor(MUTED);
        root.addView(sub, mt(5));

        netStatus = text("● Łączenie…", 14, true);
        netStatus.setTextColor(MUTED);
        root.addView(netStatus, mt(16));

        LinearLayout stats1 = row();
        netHeight = statCard(stats1, "WYSOKOŚĆ", "—", "blok");
        netPeers = statCard(stats1, "PEERS", "0", "połączenia");
        root.addView(stats1, mt(12));

        LinearLayout stats2 = row();
        netMempool = statCard(stats2, "MEMPOOL", "0", "transakcje");
        netIssued = statCard(stats2, "WYEMITOWANO", "—", "AURQ");
        root.addView(stats2, mt(10));

        LinearLayout details = card();
        details.addView(section("SZCZEGÓŁY MAINNETU"));
        netNode = kv(details, "Node", "—");
        netNetworkId = kv(details, "Network ID", "—");
        netTip = kv(details, "Tip", "—");
        netWork = kv(details, "Chain work", "—");
        netObserved = kv(details, "Ostatni odczyt", "—");
        root.addView(details, mt(16));

        LinearLayout live = card();
        live.addView(section("OSTATNIE BLOKI"));
        TextView info = text("Aktualizacja co 5 sekund", 11, false);
        info.setTextColor(MUTED);
        live.addView(info, mt(4));
        recentBlocks = new LinearLayout(this);
        recentBlocks.setOrientation(LinearLayout.VERTICAL);
        live.addView(recentBlocks, mt(10));
        root.addView(live, mt(16));

        Button refresh = secondaryButton("Odśwież teraz");
        refresh.setOnClickListener(v -> refreshAll());
        root.addView(refresh, mt(14));

        TextView model = text("AuronQ Mobile jest klientem portfela, nie pełnym nodem. Sprawdza Network ID i odczytuje stan z publicznego noda AuronQ Mainnet. Pełne nody nadal niezależnie walidują blockchain.", 12, false);
        model.setTextColor(MUTED);
        root.addView(model, mt(18));

        return root;
    }

    private void showScreen(String which) {
        dashboardScreen.getParent().setVisibility("home".equals(which) ? View.VISIBLE : View.GONE);
        walletScreen.getParent().setVisibility("wallet".equals(which) ? View.VISIBLE : View.GONE);
        sendScreen.getParent().setVisibility("send".equals(which) ? View.VISIBLE : View.GONE);
        networkScreen.getParent().setVisibility("network".equals(which) ? View.VISIBLE : View.GONE);

        navStyle(navHome, "home".equals(which));
        navStyle(navWallet, "wallet".equals(which));
        navStyle(navSend, "send".equals(which));
        navStyle(navNetwork, "network".equals(which));

        if ("home".equals(which)) topTitle.setText("Pulpit");
        if ("wallet".equals(which)) topTitle.setText("Portfel");
        if ("send".equals(which)) topTitle.setText("Wyślij");
        if ("network".equals(which)) topTitle.setText("Sieć na żywo");
    }

    private void refreshAll() {
        if (liveBusy) return;
        liveBusy = true;
        executor.execute(() -> {
            try {
                String node = nodeUrl;
                if (node == null || node.isEmpty()) node = Bridge.discoverNode();

                String snapshot;
                try {
                    snapshot = Bridge.networkSnapshot(node, 6);
                } catch (Exception first) {
                    node = Bridge.discoverNode();
                    snapshot = Bridge.networkSnapshot(node, 6);
                }

                String balance = null;
                if (walletFile.exists()) {
                    try {
                        String addr = Bridge.walletAddress(walletFile.getAbsolutePath());
                        balance = Bridge.balance(node, addr);
                    } catch (Exception ignored) {
                    }
                }

                final String finalNode = node;
                final String finalSnapshot = snapshot;
                final String finalBalance = balance;
                runOnUiThread(() -> {
                    nodeUrl = finalNode;
                    applySnapshot(finalSnapshot);
                    if (finalBalance != null) applyBalance(finalBalance);
                });
            } catch (Exception e) {
                runOnUiThread(() -> setNetworkOffline(e.getMessage()));
            } finally {
                liveBusy = false;
            }
        });
    }

    private void applySnapshot(String raw) {
        try {
            JSONObject j = new JSONObject(raw);
            long height = j.optLong("height");
            int peers = j.optInt("peers");
            int mempool = j.optInt("mempool");
            double issued = j.optDouble("issued_coins");

            dashHeight.setText(String.valueOf(height));
            dashPeers.setText(String.valueOf(peers));
            dashMempool.setText(String.valueOf(mempool));
            dashNode.setText("● AuronQ Mainnet połączony");
            dashNode.setTextColor(ACCENT);
            dashTip.setText("Tip: " + shortHash(j.optString("tip")));

            netStatus.setText("● Połączono z AuronQ Mainnet");
            netStatus.setTextColor(ACCENT);
            netHeight.setText(String.valueOf(height));
            netPeers.setText(String.valueOf(peers));
            netMempool.setText(String.valueOf(mempool));
            netIssued.setText(String.format(Locale.US, "%.8f", issued));
            netNode.setText(j.optString("node", nodeUrl));
            netNetworkId.setText(j.optString("network_id"));
            netTip.setText(j.optString("tip"));
            netWork.setText(j.optString("chain_work"));
            netObserved.setText(j.optString("observed_at"));

            topNetworkDot.setTextColor(ACCENT);

            recentBlocks.removeAllViews();
            JSONArray blocks = j.optJSONArray("blocks");
            if (blocks == null || blocks.length() == 0) {
                TextView empty = text("Brak danych o blokach", 12, false);
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
            setNetworkOffline("Nieprawidłowa odpowiedź sieci");
        }
    }

    private void setNetworkOffline(String error) {
        topNetworkDot.setTextColor(DANGER);
        dashNode.setText("● Brak połączenia");
        dashNode.setTextColor(DANGER);
        netStatus.setText("● Brak połączenia z AuronQ Mainnet");
        netStatus.setTextColor(DANGER);
        netObserved.setText(error == null ? "—" : error);
        nodeUrl = "";
    }

    private View blockRow(long height, String hash, long timestamp, int txs) {
        LinearLayout row = new LinearLayout(this);
        row.setOrientation(LinearLayout.VERTICAL);
        row.setPadding(dp(12), dp(10), dp(12), dp(10));
        row.setBackground(round(PANEL2, LINE, 12));

        LinearLayout top = new LinearLayout(this);
        top.setOrientation(LinearLayout.HORIZONTAL);
        TextView h = text("Blok #" + height, 14, true);
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

    private void createWallet() {
        if (walletFile.exists()) {
            toast("Portfel już istnieje. Zrób backup albo usuń go przed utworzeniem nowego.");
            return;
        }
        String password = passwordInput.getText().toString();
        run("Tworzenie portfela…",
                () -> Bridge.createWallet(walletFile.getAbsolutePath(), password),
                value -> {
                    walletAddress = value;
                    loadWalletState();
                    refreshAll();
                    toast("Portfel utworzony");
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
            toast("Brak portfela do backupu");
            return;
        }
        Intent i = new Intent(Intent.ACTION_CREATE_DOCUMENT);
        i.setType("application/json");
        i.putExtra(Intent.EXTRA_TITLE, "AuronQ-main.wallet");
        startActivityForResult(i, EXPORT_WALLET);
    }

    private void copyAddress() {
        if (walletAddress.isEmpty()) {
            toast("Brak adresu");
            return;
        }
        ClipboardManager cm = (ClipboardManager) getSystemService(Context.CLIPBOARD_SERVICE);
        cm.setPrimaryClip(ClipData.newPlainText("AuronQ address", walletAddress));
        toast("Adres skopiowany");
    }

    private void sendTransaction() {
        if (!walletFile.exists() || walletAddress.isEmpty()) {
            toast("Najpierw utwórz albo zaimportuj portfel");
            return;
        }
        if (nodeUrl.isEmpty()) {
            toast("Brak połączenia z AuronQ Mainnet");
            return;
        }
        String password = sendPassword.getText().toString();
        String to = recipientInput.getText().toString().trim();
        String amount = amountInput.getText().toString().trim();

        new AlertDialog.Builder(this)
                .setTitle("Potwierdź wysyłkę")
                .setMessage("Wyślij " + amount + " AURQ na:\n\n" + to + "\n\nTransakcja po zatwierdzeniu w blockchainie jest nieodwracalna.")
                .setNegativeButton("Anuluj", null)
                .setPositiveButton("Wyślij", (d, w) ->
                        run("Podpisywanie i wysyłanie…",
                                () -> Bridge.send(nodeUrl, walletFile.getAbsolutePath(), password, to, amount),
                                value -> {
                                    try {
                                        JSONObject j = new JSONObject(value);
                                        sendResult.setText("TXID:\n" + j.optString("txid") + "\nFee: " + j.optString("fee") + " AURQ");
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
            toast("Brak lokalnego portfela");
            return;
        }
        new AlertDialog.Builder(this)
                .setTitle("Usunąć portfel?")
                .setMessage("Najpierw upewnij się, że masz backup. Usunięcie pliku bez backupu może oznaczać trwałą utratę dostępu do środków.")
                .setNegativeButton("Anuluj", null)
                .setPositiveButton("Usuń", (d, w) -> {
                    if (walletFile.delete()) {
                        walletAddress = "";
                        passwordInput.setText("");
                        loadWalletState();
                        refreshAll();
                        toast("Lokalny portfel usunięty");
                    } else {
                        toast("Nie udało się usunąć pliku");
                    }
                })
                .show();
    }

    private void loadWalletState() {
        if (!walletFile.exists()) {
            walletAddress = "";
            walletAddressText.setText("Brak lokalnego portfela");
            walletBalance.setText("0.00000000 AURQ");
            dashBalance.setText("0.00000000");
            backupButton.setEnabled(false);
            copyButton.setEnabled(false);
            sendShortcutButton.setEnabled(false);
            deleteButton.setEnabled(false);
            return;
        }
        executor.execute(() -> {
            try {
                String addr = Bridge.walletAddress(walletFile.getAbsolutePath());
                runOnUiThread(() -> {
                    walletAddress = addr;
                    walletAddressText.setText(addr);
                    backupButton.setEnabled(true);
                    copyButton.setEnabled(true);
                    sendShortcutButton.setEnabled(true);
                    deleteButton.setEnabled(true);
                });
            } catch (Exception e) {
                runOnUiThread(() -> toast("Błąd portfela: " + e.getMessage()));
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
                    if (sendResult != null) sendResult.setText("Błąd: " + e.getMessage());
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
                        toast("Nie udało się zastąpić obecnego portfela");
                        tmp.delete();
                        return;
                    }
                    if (!tmp.renameTo(walletFile)) {
                        toast("Nie udało się zapisać importowanego portfela");
                        tmp.delete();
                        return;
                    }
                    walletAddress = addr;
                    loadWalletState();
                    refreshAll();
                    toast("Portfel zaimportowany");
                };
                if (walletFile.exists()) {
                    new AlertDialog.Builder(this)
                            .setTitle("Zastąpić portfel?")
                            .setMessage("Obecny lokalny portfel zostanie zastąpiony. Zrób backup przed kontynuacją.")
                            .setNegativeButton("Anuluj", (d, w) -> tmp.delete())
                            .setPositiveButton("Zastąp", (d, w) -> install.run())
                            .show();
                } else {
                    install.run();
                }
            } catch (Exception e) {
                tmp.delete();
                toast("Import nieudany: " + e.getMessage());
            }
        } else if (requestCode == EXPORT_WALLET) {
            try (InputStream in = new java.io.FileInputStream(walletFile);
                 OutputStream out = getContentResolver().openOutputStream(uri, "w")) {
                copy(in, out);
                toast("Backup zapisany");
            } catch (Exception e) {
                toast("Backup nieudany: " + e.getMessage());
            }
        }
    }

    private static void copy(InputStream in, OutputStream out) throws Exception {
        if (in == null || out == null) throw new Exception("nie można otworzyć pliku");
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
        Toast.makeText(this, msg == null ? "Błąd" : msg, Toast.LENGTH_LONG).show();
    }

    @Override
    protected void onDestroy() {
        handler.removeCallbacks(liveLoop);
        executor.shutdownNow();
        super.onDestroy();
    }
}
