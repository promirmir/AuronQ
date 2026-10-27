# AuronQ Desktop na Linux — instalacja przez dwuklik

## Ubuntu, Linux Mint i Debian (64-bit x86 / amd64)

Jest to **dokładnie ten sam kod AuronQ Desktop v1.7.14 co na Windows**: portfel, pełny węzeł Mainnet, sieć, kopanie oraz lokalny Explorer. Nie jest to wyłącznie CLI. W Linuksie interfejs otwiera się w domyślnej przeglądarce, zamiast w oknie Microsoft Edge jak na niektórych komputerach z Windows.

1. Pobierz [AuronQ Desktop dla Linux (.deb)](https://github.com/promirmir/AuronQ/releases/download/linux-desktop-v1.7.14.1/AuronQ-Desktop-1.7.14-1-Linux-amd64.deb) i plik `SHA256SUMS.txt` z [wydania v1.7.14.1](https://github.com/promirmir/AuronQ/releases/tag/linux-desktop-v1.7.14.1). Suma SHA-256 pliku `.deb`: `79f2859dc1ced75beb321b98bc9ed1edd0e5ab0c18a07be2c0d8ddd7ea90d73a`.
2. **Kliknij dwukrotnie plik .deb** i naciśnij **Zainstaluj**. Jeśli system wymaga hasła administratora, potwierdź instalację.
3. Otwórz **menu aplikacji → AuronQ Desktop** i kliknij ikonę. Nie trzeba terminala, Go, Dockera, ręcznej konfiguracji peerów ani przekierowania portu.
4. Program otworzy panel w przeglądarce i uruchomi lokalny pełny węzeł. Pierwsze wczytanie/zatwierdzanie łańcucha może zająć czas, a przeglądarka otworzy się dopiero po wstępnej inicjalizacji noda.
5. W programie utwórz lub zaimportuj portfel i wykonaj jego bezpieczną kopię zapasową. Do zakończenia synchronizacji informacje o wysokości łańcucha, saldzie i transakcjach mogą być nieaktualne.

**Ważne:** wcześniej publikowane archiwum `AuronQ-1.7.14-Linux-amd64.tar.gz` zawiera tylko `auronq` (CLI), bez graficznego Desktopu. Do uruchamiania przez kliknięcie wybierz nowy pakiet **.deb**.

### Jeśli przeglądarka nie otworzyła się automatycznie

Otwórz `http://127.0.0.1:18445/` na **tym samym komputerze**, na którym uruchomiłeś AuronQ. Jeśli panel nie odpowiada, zamknij AuronQ i uruchom ponownie **tylko raz** z menu aplikacji. Nie uruchamiaj dwóch osobnych nodów na porcie 18444. Port 18445 to lokalny panel Desktop, a 18444 to lokalny węzeł Mainnet.

### Aktualizacja, zamykanie i usuwanie

- Zamknij aplikację przez jej opcję zamknięcia/zatrzymania, a nie samo zamknięcie karty przeglądarki: node i koparka mogą nadal działać.
- Przy aktualizacji zainstaluj nowy pakiet systemowo. Danych portfela i łańcucha **nie kasuj**.
- Dane użytkownika pozostają w katalogu konfiguracyjnym AuronQ, zazwyczaj `~/.config/AuronQ/` (jeśli `XDG_CONFIG_HOME` jest ustawione, w jego odpowiednim podkatalogu).
- Usunięcie pakietu systemowego nie powinno być używane jako metoda kasowania portfeli. Zadbaj o co najmniej dwie kopie zaszyfrowanego portfela i osobną kopię hasła.

### Pozostałe dystrybucje Linux

Pakiet **.deb** jest przeznaczony dla Ubuntu/Mint/Debian i systemów zgodnych z Debianem. Nie próbuj go bezpośrednio instalować w Fedora/Arch. Zgłoś potrzebę natywnego pakietu dla innego systemu w [Issues](https://github.com/promirmir/AuronQ/issues). Stare archiwum `.tar.gz` to wyłącznie program konsolowy.

### Audyt

AuronQ wymaga niezależnego audytu bezpieczeństwa. [Szczegóły techniczne](SECURITY.md).

## English — Ubuntu / Mint / Debian x86-64

Download the `.deb` and checksum from the [Linux Desktop release](https://github.com/promirmir/AuronQ/releases/tag/linux-desktop-v1.7.14.1). Verify its SHA-256, double-click the .deb, select **Install**, then launch **AuronQ Desktop** from your applications menu. This is the **same graphical Desktop code as Windows**, opening in your default browser. No terminal or Go required. On first launch blockchain verification can take time before the UI opens; do not start additional instances. Your wallet/chain data remain under the user's configuration directory (normally `~/.config/AuronQ/`). The older `AuronQ-1.7.14-Linux-amd64.tar.gz` includes only CLI, not Desktop. AuronQ requires an independent security audit. DEB SHA-256: `79f2859dc1ced75beb321b98bc9ed1edd0e5ab0c18a07be2c0d8ddd7ea90d73a`.
