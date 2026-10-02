package com.auronq.mobile;

import android.app.Activity;
import android.app.AlertDialog;
import android.content.ClipData;
import android.content.ClipboardManager;
import android.content.Context;
import android.content.Intent;
import android.graphics.Color;
import android.net.Uri;
import android.os.Bundle;
import android.text.InputType;
import android.view.Gravity;
import android.view.View;
import android.widget.Button;
import android.widget.EditText;
import android.widget.LinearLayout;
import android.widget.ScrollView;
import android.widget.TextView;
import android.widget.Toast;

import com.auronq.mobilebind.bridge.Bridge;

import org.json.JSONObject;

import java.io.File;
import java.io.FileOutputStream;
import java.io.InputStream;
import java.io.OutputStream;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

public class MainActivity extends Activity {
    private static final int IMPORT_WALLET = 1001;
    private static final int EXPORT_WALLET = 1002;

    private final ExecutorService executor = Executors.newSingleThreadExecutor();

    private File walletDir;
    private File walletFile;
    private String nodeUrl = "";
    private String walletAddress = "";

    private TextView networkText;
    private TextView addressText;
    private TextView balanceText;
    private TextView resultText;
    private EditText passwordInput;
    private EditText recipientInput;
    private EditText amountInput;
    private Button createButton;
    private Button importButton;
    private Button backupButton;
    private Button copyButton;
    private Button sendButton;
    private Button deleteButton;

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
        refreshNetwork();
    }

    private View buildUi() {
        ScrollView scroll = new ScrollView(this);
        scroll.setBackgroundColor(Color.rgb(11, 15, 20));

        LinearLayout root = new LinearLayout(this);
        root.setOrientation(LinearLayout.VERTICAL);
        root.setPadding(dp(20), dp(24), dp(20), dp(40));
        scroll.addView(root);

        TextView title = text("AuronQ Mobile", 28, true);
        root.addView(title);

        TextView sub = text("Mainnet • 0.1.0 Alpha", 14, false);
        sub.setTextColor(Color.rgb(128, 194, 255));
        root.addView(sub, marginTop(4));

        networkText = text("Sieć: łączenie…", 15, false);
        root.addView(networkText, marginTop(24));

        root.addView(section("PORTFEL"), marginTop(28));
        addressText = text("Brak lokalnego portfela", 15, false);
        addressText.setTextIsSelectable(true);
        root.addView(addressText, marginTop(8));

        balanceText = text("Saldo: —", 22, true);
        root.addView(balanceText, marginTop(12));

        LinearLayout walletButtons = horizontal();
        createButton = button("Utwórz");
        importButton = button("Import");
        backupButton = button("Backup");
        copyButton = button("Kopiuj adres");
        walletButtons.addView(createButton, weight());
        walletButtons.addView(importButton, weight());
        root.addView(walletButtons, marginTop(12));

        LinearLayout walletButtons2 = horizontal();
        walletButtons2.addView(backupButton, weight());
        walletButtons2.addView(copyButton, weight());
        root.addView(walletButtons2, marginTop(8));

        passwordInput = input("Hasło portfela (min. 12 znaków)");
        passwordInput.setInputType(InputType.TYPE_CLASS_TEXT | InputType.TYPE_TEXT_VARIATION_PASSWORD);
        root.addView(passwordInput, marginTop(16));

        root.addView(section("WYŚLIJ AURQ"), marginTop(30));
        recipientInput = input("Adres odbiorcy aurq1…");
        root.addView(recipientInput, marginTop(8));
        amountInput = input("Kwota, np. 1.25000000");
        amountInput.setInputType(InputType.TYPE_CLASS_NUMBER | InputType.TYPE_NUMBER_FLAG_DECIMAL);
        root.addView(amountInput, marginTop(8));

        sendButton = button("Wyślij transakcję");
        root.addView(sendButton, marginTop(12));

        Button refresh = button("Odśwież sieć i saldo");
        root.addView(refresh, marginTop(16));

        resultText = text("", 13, false);
        resultText.setTextIsSelectable(true);
        root.addView(resultText, marginTop(12));

        deleteButton = button("Usuń lokalny portfel");
        root.addView(deleteButton, marginTop(28));

        TextView warning = text(
                "Alpha do testów. Klucz prywatny pozostaje lokalnie w zaszyfrowanym pliku wallet. " +
                "Aplikacja korzysta z publicznego noda AuronQ i nie jest pełnym nodem/SPV. " +
                "Nie przechowuj tu dużej wartości przed niezależnym audytem projektu.",
                12, false);
        warning.setTextColor(Color.rgb(180, 180, 180));
        root.addView(warning, marginTop(20));

        createButton.setOnClickListener(v -> createWallet());
        importButton.setOnClickListener(v -> importWallet());
        backupButton.setOnClickListener(v -> backupWallet());
        copyButton.setOnClickListener(v -> copyAddress());
        sendButton.setOnClickListener(v -> sendTransaction());
        refresh.setOnClickListener(v -> refreshNetwork());
        deleteButton.setOnClickListener(v -> deleteWallet());

        return scroll;
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
                    refreshBalance();
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
        String password = passwordInput.getText().toString();
        String to = recipientInput.getText().toString().trim();
        String amount = amountInput.getText().toString().trim();

        new AlertDialog.Builder(this)
                .setTitle("Potwierdź wysyłkę")
                .setMessage("Wyślij " + amount + " AURQ na:\n\n" + to + "\n\nTransakcja po wysłaniu może być nieodwracalna.")
                .setNegativeButton("Anuluj", null)
                .setPositiveButton("Wyślij", (d, w) ->
                        run("Podpisywanie i wysyłanie…",
                                () -> Bridge.send(nodeUrl, walletFile.getAbsolutePath(), password, to, amount),
                                value -> {
                                    try {
                                        JSONObject j = new JSONObject(value);
                                        resultText.setText("TXID:\n" + j.optString("txid") + "\nFee: " + j.optString("fee") + " AURQ");
                                    } catch (Exception e) {
                                        resultText.setText(value);
                                    }
                                    refreshBalance();
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
            addressText.setText("Brak lokalnego portfela");
            balanceText.setText("Saldo: —");
            backupButton.setEnabled(false);
            copyButton.setEnabled(false);
            sendButton.setEnabled(false);
            deleteButton.setEnabled(false);
            return;
        }
        run("Sprawdzanie portfela…",
                () -> Bridge.walletAddress(walletFile.getAbsolutePath()),
                value -> {
                    walletAddress = value;
                    addressText.setText(value);
                    backupButton.setEnabled(true);
                    copyButton.setEnabled(true);
                    sendButton.setEnabled(true);
                    deleteButton.setEnabled(true);
                });
    }

    private void refreshNetwork() {
        run("Szukanie AuronQ Mainnet…",
                Bridge::discoverNode,
                value -> {
                    nodeUrl = value;
                    networkText.setText("Sieć: połączono\n" + nodeUrl);
                    refreshBalance();
                });
    }

    private void refreshBalance() {
        if (walletAddress.isEmpty() || nodeUrl.isEmpty()) return;
        run("Pobieranie salda…",
                () -> Bridge.balance(nodeUrl, walletAddress),
                value -> {
                    try {
                        JSONObject j = new JSONObject(value);
                        balanceText.setText("Saldo: " + j.optString("spendable") + " AURQ");
                        resultText.setText("Łącznie: " + j.optString("total") + " AURQ");
                    } catch (Exception e) {
                        resultText.setText(value);
                    }
                });
    }

    private void run(String busy, Task task, Success success) {
        resultText.setText(busy);
        executor.execute(() -> {
            try {
                String result = task.run();
                runOnUiThread(() -> success.accept(result));
            } catch (Exception e) {
                runOnUiThread(() -> {
                    resultText.setText("Błąd: " + e.getMessage());
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
                    refreshBalance();
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

    private TextView section(String s) {
        TextView t = text(s, 12, true);
        t.setTextColor(Color.rgb(120, 180, 235));
        return t;
    }

    private TextView text(String s, int sp, boolean bold) {
        TextView t = new TextView(this);
        t.setText(s);
        t.setTextSize(sp);
        t.setTextColor(Color.WHITE);
        if (bold) t.setTypeface(android.graphics.Typeface.DEFAULT_BOLD);
        return t;
    }

    private EditText input(String hint) {
        EditText e = new EditText(this);
        e.setHint(hint);
        e.setHintTextColor(Color.rgb(130, 130, 130));
        e.setTextColor(Color.WHITE);
        e.setSingleLine(true);
        e.setPadding(dp(12), dp(12), dp(12), dp(12));
        return e;
    }

    private Button button(String label) {
        Button b = new Button(this);
        b.setText(label);
        b.setAllCaps(false);
        return b;
    }

    private LinearLayout horizontal() {
        LinearLayout l = new LinearLayout(this);
        l.setOrientation(LinearLayout.HORIZONTAL);
        l.setGravity(Gravity.CENTER_VERTICAL);
        return l;
    }

    private LinearLayout.LayoutParams weight() {
        LinearLayout.LayoutParams p = new LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f);
        p.setMarginEnd(dp(6));
        return p;
    }

    private LinearLayout.LayoutParams marginTop(int dp) {
        LinearLayout.LayoutParams p = new LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT,
                LinearLayout.LayoutParams.WRAP_CONTENT);
        p.topMargin = dp(dp);
        return p;
    }

    private int dp(int v) {
        return Math.round(v * getResources().getDisplayMetrics().density);
    }

    private void toast(String msg) {
        Toast.makeText(this, msg == null ? "Błąd" : msg, Toast.LENGTH_LONG).show();
    }

    @Override
    protected void onDestroy() {
        executor.shutdownNow();
        super.onDestroy();
    }
}
