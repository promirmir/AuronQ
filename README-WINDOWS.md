# AuronQ Mainnet — Windows

## Normalny użytkownik

1. Otwórz [GitHub Releases](https://github.com/promirmir/AuronQ/releases/latest) i pobierz bieżący `AuronQ-<version>-Windows-x64.zip`.
2. Sprawdź SHA-256 pliku.
3. Rozpakuj cały ZIP do jednego folderu.
4. Uruchom `START-AURONQ.cmd`.
5. AuronQ Desktop uruchomi własny pełny node, zweryfikuje `network.json`, odnajdzie peery i rozpocznie synchronizację.
6. Utwórz lub zaimportuj portfel i wykonaj backup.

Normalny użytkownik **nie potrzebuje** Tailscale, Cloudflare, Go, Dockera, ręcznego wpisywania peerów ani przekierowania portu. Użytkownik za NAT/CGNAT nadal jest pełnym nodem: sam waliduje blockchain, utrzymuje mempool i relayuje dane przez połączenia wychodzące. NAT może jedynie ograniczyć możliwość przyjmowania nowych połączeń przychodzących.

## Pierwsze uruchomienie / aktualizacja

1. Zamknij starszą wersję AuronQ przyciskiem **Sieć → Zamknij AuronQ**.
2. Rozpakuj cały ZIP bieżącego wydania do nowego folderu.
3. Uruchom `START-AURONQ.cmd` tylko raz.
4. Poczekaj na uruchomienie pełnego noda i synchronizację.
5. **Nie usuwaj `%AppData%\AuronQ`** podczas zwykłej aktualizacji — znajdują się tam portfele, blockchain i zapamiętane peery.

## Weryfikacja pobranego wydania

Każde oficjalne wydanie Desktop / Full Node publikuje plik `SHA256SUMS.txt`. Zweryfikuj pobrany ZIP względem sumy z **tego samego wydania GitHub**, zamiast kopiować sumę kontrolną ze starej instrukcji.

PowerShell:

```powershell
Get-FileHash .\AuronQ-<version>-Windows-x64.zip -Algorithm SHA256
```

Porównaj wynik z odpowiednim wpisem w `SHA256SUMS.txt`.

Bieżące binaria Windows nie mają podpisu Authenticode, więc SmartScreen może wyświetlić ostrzeżenie.

## Explorer blockchaina

Każdy pełny node ma **własny** Explorer pod `/explorer`. W AuronQ Desktop zakładka **Explorer** otwiera lokalny `http://127.0.0.1:18444/explorer`.

Explorer nie jest centralną usługą projektu i nie ma uprawnień konsensusowych. Wyświetla blockchain, który **ten konkretny pełny node sam zweryfikował**. Publiczny adres jednego noda może zniknąć bez wpływu na Explorera i blockchain innych użytkowników.

## Bootstrap i decentralizacja

Seed/bootstrap jest wyłącznie adresem pierwszego kontaktu. Od hardeningu wprowadzonego w v1.7.10 skonfigurowany seed nie jest trwałym uprzywilejowanym peerem: po wielokrotnych błędach może zostać usunięty z aktywnego zestawu. Peery poznane przez P2P są zapisywane lokalnie i mogą całkowicie zastąpić początkowe komputery.

CI zawiera test rozproszony, w którym pierwotny/bootstrapowy node zostaje wyłączony na stałe, późniejsze nody dalej tworzą i synchronizują blockchain, a świeży node dołącza przez ocalałego peera niebędącego nodem założycielskim.

To nie oznacza, że sieć może działać bez **jakichkolwiek** komputerów. Tak jak każda sieć P2P potrzebuje działających uczestników. Celem jest brak zależności od **konkretnego** komputera lub operatora.

## Publiczny node

Zwykły użytkownik powinien uruchamiać `START-AURONQ.cmd`. Jeżeli świadomie chce wystawić publicznie osiągalny full node, powinien użyć `START-PUBLIC-NODE.cmd`, przekierować TCP/18444 na routerze i upewnić się, że nie jest za CGNAT. `START-SEED-NODE.cmd` pozostaje opcjonalnym wariantem wykorzystującym Tailscale Funnel.

Publiczne peery są wymieniane przez gossip i zapisywane lokalnie. Node uruchomiony na serwerze/VPS z publicznym adresem przypisanym bezpośrednio do interfejsu próbuje automatycznie ogłosić swój publiczny endpoint; odbiorca nadal wykonuje callback-verification przed dodaniem go do gossip. Automatyczny crawler repozytorium co godzinę sprawdza osiągalne publiczne peery AuronQ i przygotowuje aktualizację rejestru przez chroniony proces PR/CI. Niezależny, publicznie osiągalny node jest szczególnie cenny, ponieważ może stać się alternatywną drogą wejścia do sieci, gdy nody projektu są wyłączone.

Aktualne bootstrapy są publikowane w `bootstrap.json`; bootstrap nie może zatwierdzić nieważnego bloku ani zmienić Network ID.

## Tożsamość sieci

Network ID:

`44e62c2ace002a6660c14e252173c1aa303529c68e40c998e92da2b453f44f30b1e58c94d533587e2186004593fb856c433fcdb5418ed430ec8617e29529365c`

Genesis:

`5750a455c04bfe93c9edfef1a12744b05e29ac6da1a9dd5b790566629dea2080c581beb2f0324efba2067c9efb113ed7a29598fd3ffbed965ced31f265d0cec4`

## Portfele

Usuwanie portfela dotyczy wyłącznie lokalnego zaszyfrowanego pliku `.wallet`; nie przenosi ani nie niszczy monet zapisanych w blockchainie. Przed usunięciem ostatniej kopii portfela wykonaj **Backup**.

## Bezpieczeństwo

AuronQ Mainnet jest publiczną siecią, ale AQM64 i cały konsensus/network stack **nie przeszły niezależnego profesjonalnego audytu kryptograficznego i bezpieczeństwa**. Testy CI, fuzzing, race detector i testy sieci rozproszonej istotnie podnoszą jakość, ale nie zastępują zewnętrznego audytu.
