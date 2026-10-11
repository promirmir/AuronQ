package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "strings"
    "testing"

    aq "auronq/internal/auronq"
)

// Disposable wallets only. No mainnet connection or real user funds.
func TestWalletDecryptionRegression1713(t *testing.T) {
    tests := []struct { name, password string }{
        {"ascii", "AuronQ-Wallet-Test-123!"},
        {"unicode-spaces", "Zażółć gęślą jaźń 2026!  "},
        {"symbols", "AuronQ_+@#$%^&*()-=|2026"},
    }
    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            dir := t.TempDir()
            source := filepath.Join(dir, "original.wallet")
            restored := filepath.Join(dir, "restored.wallet")

            created, err := aq.NewWallet(source, tc.password, aq.MainnetNetworkByte)
            if err != nil { t.Fatalf("create disposable wallet: %v", err) }
            wantAddress, wantPublicKey := created.Address(), created.PublicKey()
            created.Close()
            before, err := os.ReadFile(source)
            if err != nil { t.Fatal(err) }

            opened, err := aq.LoadWallet(source, tc.password)
            if err != nil { t.Fatalf("correct password failed after reopening: %v", err) }
            if opened.Address() != wantAddress || !bytes.Equal(opened.PublicKey(), wantPublicKey) {
                t.Fatal("wallet identity changed after reopening")
            }
            opened.Close()

            if wrong, err := aq.LoadWallet(source, tc.password+"!"); err == nil {
                wrong.Close()
                t.Fatal("wrong password unexpectedly unlocked wallet")
            }
            after, err := os.ReadFile(source)
            if err != nil { t.Fatal(err) }
            if !bytes.Equal(before, after) {
                t.Fatal("failed unlock modified encrypted original")
            }

            if err := os.WriteFile(restored, before, 0600); err != nil { t.Fatal(err) }
            imported, err := aq.LoadWallet(restored, tc.password)
            if err != nil { t.Fatalf("correct password failed on restored backup: %v", err) }
            if imported.Address() != wantAddress || !bytes.Equal(imported.PublicKey(), wantPublicKey) {
                t.Fatal("restored wallet identity changed")
            }
            imported.Close()
            if wrong, err := aq.LoadWallet(restored, tc.password+"!"); err == nil {
                wrong.Close()
                t.Fatal("wrong password unlocked restored backup")
            }
        })
    }
}

func TestWalletDecryptionRejectsCorruptedBackup(t *testing.T) {
    dir := t.TempDir()
    source, damaged := filepath.Join(dir, "original.wallet"), filepath.Join(dir, "damaged.wallet")
    const password = "Not-A-Real-Funded-Wallet!"
    created, err := aq.NewWallet(source, password, aq.MainnetNetworkByte)
    if err != nil { t.Fatal(err) }
    created.Close()
    original, err := os.ReadFile(source)
    if err != nil { t.Fatal(err) }

    var data map[string]any
    if err := json.Unmarshal(original, &data); err != nil { t.Fatal(err) }
    ciphertext, ok := data["ciphertext"].(string)
    if !ok || len(ciphertext) < 2 { t.Fatal("unexpected wallet format") }
    replacement := "0"
    if strings.HasPrefix(ciphertext, "0") { replacement = "1" }
    data["ciphertext"] = replacement + ciphertext[1:]
    raw, err := json.Marshal(data)
    if err != nil { t.Fatal(err) }
    if err := os.WriteFile(damaged, raw, 0600); err != nil { t.Fatal(err) }

    if opened, err := aq.LoadWallet(damaged, password); err == nil {
        opened.Close()
        t.Fatal("modified ciphertext unexpectedly accepted")
    }
    intact, err := aq.LoadWallet(source, password)
    if err != nil { t.Fatalf("original wallet no longer decryptable: %v", err) }
    intact.Close()
}
