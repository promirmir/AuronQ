# AuronQ 1.7.8 Mainnet — Windows

## Normalny użytkownik

1. Pobierz `AuronQ-1.7.8-Windows-x64.zip` wyłącznie z GitHub Releases.
2. Sprawdź SHA-256 pliku.
3. Rozpakuj cały ZIP do jednego folderu.
4. Uruchom `START-AURONQ.cmd`.
5. AuronQ Desktop uruchomi pełny node, zweryfikuje `network.json`, pobierze oficjalny manifest bootstrap i sam rozpocznie synchronizację.

Normalny użytkownik **nie potrzebuje** Tailscale, Cloudflare, Go, Dockera, ręcznego wpisywania peerów ani przekierowania portu.

Użytkownik za NAT/CGNAT nadal może synchronizować, wysyłać transakcje, kopać i relayować dane przez połączenia wychodzące.

## Pierwsze uruchomienie — skrót

1. Zamknij starszą wersję AuronQ przyciskiem **Sieć → Zamknij AuronQ**.
2. Rozpakuj cały oficjalny ZIP 1.7.8 do nowego folderu.
3. Uruchom `START-AURONQ.cmd` tylko raz.
4. **Poczekaj cierpliwie.** Pierwsze uruchomienie AuronQ Desktop może trwać dłużej, ponieważ aplikacja uruchamia pełny node, wczytuje lokalny blockchain, sprawdza stan sieci i rozpoczyna synchronizację. Nie uruchamiaj programu wielokrotnie tylko dlatego, że okno nie pojawiło się od razu.
5. Poczekaj, aż node rozpocznie synchronizację i pojawi się wysokość blockchaina.
6. Utwórz lub zaimportuj portfel i wykonaj jego backup.
7. Dane użytkownika są przechowywane w `%AppData%\AuronQ`; nie usuwaj tego katalogu podczas zwykłej aktualizacji.

## Suma kontrolna oficjalnego v1.7.8

Oficjalny SHA-256 pliku Windows:

```text
cb74453965d641cc699e34037f1cb66fcaf4cc40757ade84cab78b350d5ae093  AuronQ-1.7.5-Windows-x64.zip
```

Pełny `SHA256SUMS.txt` jest również dołączony do wydania GitHub.

Bieżące binaria Windows nie mają podpisu Authenticode. Windows SmartScreen może więc wyświetlić ostrzeżenie przy pierwszym uruchomieniu. Weryfikuj sumę SHA-256 i źródło pobrania.

## Explorer blockchaina

Od wersji **1.7.5** pełny node udostępnia wbudowany, tylko-do-odczytu AuronQ Explorer pod ścieżką `/explorer`. Od **1.7.6** AuronQ Desktop ma osobną zakładkę **Explorer**, która osadza lokalny explorer z `http://127.0.0.1:18444/explorer`. W **1.7.8** wzmocniono bootstrap i synchronizację sieci; Explorer pozostaje tylko-do-odczytu. Jeżeli node jest publicznie osiągalny przez HTTPS, możesz też otworzyć jego adres i dopisać `/explorer`.

## Niezależny publiczny node

Zwykły użytkownik nie musi wystawiać publicznego portu i powinien uruchamiać `START-AURONQ.cmd`.

Jeżeli operator świadomie chce udostępnić publicznie osiągalny pełny node/bootstrap, może użyć `START-SEED-NODE.cmd`. W aktualnym wydaniu skrypt korzysta z Tailscale Funnel, odczytuje publiczną nazwę `*.ts.net`, ustawia `AURONQ_ADVERTISE` i uruchamia AuronQ Desktop.

Publiczne nody są wymieniane przez peer gossip i zapisywane lokalnie. Repozytorium ma też automatyczny crawler, który okresowo sprawdza osiągalne publiczne peery AuronQ i dopisuje zweryfikowane adresy do `bootstrap.json`.

## Operator publicznego bootstrap noda

`START-SEED-NODE.cmd` jest przeznaczony wyłącznie dla operatora noda, który ma być publicznym punktem pierwszego kontaktu. Skrypt uruchamia AuronQ oraz, jeśli Tailscale jest zainstalowany i Funnel wcześniej autoryzowany, wystawia lokalny port AuronQ przez Funnel w tle.

Zwykły użytkownik uruchamia `START-AURONQ.cmd`, nie `START-SEED-NODE.cmd`.

Aktualne bootstrapy są publikowane w:

`https://raw.githubusercontent.com/promirmir/AuronQ/main/bootstrap.json`

Na moment aktualizacji dokumentu crawler potwierdza dwa publiczne bootstrapy:

- `https://mir.taild63f46.ts.net`
- `https://desktop-4nifg1j.taild63f46.ts.net`

Bootstrapy nie mają żadnych dodatkowych uprawnień konsensusowych.

## Tożsamość sieci

Network ID:

`44e62c2ace002a6660c14e252173c1aa303529c68e40c998e92da2b453f44f30b1e58c94d533587e2186004593fb856c433fcdb5418ed430ec8617e29529365c`

Genesis:

`5750a455c04bfe93c9edfef1a12744b05e29ac6da1a9dd5b790566629dea2080c581beb2f0324efba2067c9efb113ed7a29598fd3ffbed965ced31f265d0cec4`

## Bezpieczeństwo

AuronQ 1.7.8 jest publicznym mainnetem, ale AQM64 i cały konsensus/network stack nie przeszły niezależnego profesjonalnego audytu kryptograficznego i bezpieczeństwa.

## Usuwanie portfela

W wersji 1.7.8 zakładka **Portfele** ma przycisk **Usuń**. Usuwanie dotyczy wyłącznie lokalnego zaszyfrowanego pliku `.wallet`; nie przenosi ani nie niszczy monet zapisanych w blockchainie.

Aplikacja wyświetla ostrzeżenie i wymaga ręcznego wpisania dokładnej nazwy portfela. Portfela używanego aktualnie przez koparkę nie można usunąć, dopóki mining nie zostanie zatrzymany. Przed usunięciem ważnego portfela wykonaj **Backup** — usunięcie ostatniej kopii pliku może oznaczać trwałą utratę dostępu do środków.
