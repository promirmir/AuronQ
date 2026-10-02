# AuronQ 1.7.3 Mainnet — Windows

## Normalny użytkownik

1. Pobierz `AuronQ-1.7.3-Windows-x64.zip` wyłącznie z GitHub Releases.
2. Sprawdź SHA-256 pliku.
3. Rozpakuj cały ZIP do jednego folderu.
4. Uruchom `START-AURONQ.cmd`.
5. AuronQ Desktop uruchomi pełny node, zweryfikuje `network.json`, pobierze oficjalny manifest bootstrap i sam rozpocznie synchronizację.

Normalny użytkownik **nie potrzebuje** Tailscale, Cloudflare, Go, Dockera, ręcznego wpisywania peerów ani przekierowania portu.

Użytkownik za NAT/CGNAT nadal może synchronizować, wysyłać transakcje, kopać i relayować dane przez połączenia wychodzące.

## Suma kontrolna oficjalnego v1.7.3

Oficjalny SHA-256 jest publikowany w pliku `SHA256SUMS.txt` dołączonym do wydania GitHub.

Bieżące binaria Windows nie mają podpisu Authenticode. Windows SmartScreen może więc wyświetlić ostrzeżenie przy pierwszym uruchomieniu. Weryfikuj sumę SHA-256 i źródło pobrania.

## Niezależny publiczny node

Jeżeli użytkownik chce, aby jego komputer pomagał utrzymywać AuronQ również jako publicznie osiągalny pełny node, może użyć `START-PUBLIC-NODE.cmd`. Skrypt otwiera TCP/18444 w Zaporze Windows i uruchamia AuronQ Desktop.

To jest tryb opcjonalny. Przy routerze nadal może być potrzebne przekierowanie TCP/18444 na ten komputer. Przy CGNAT zwykle potrzebny jest VPS lub publiczny relay. AuronQ nie uznaje samej deklaracji adresu — inny node musi potwierdzić, że peer jest rzeczywiście osiągalny.

Publiczne nody są wymieniane przez peer gossip i zapisywane lokalnie. Repozytorium ma też automatyczny crawler, który okresowo sprawdza osiągalne publiczne peery AuronQ i może dopisywać zweryfikowane adresy do `bootstrap.json`. Dzięki temu wraz ze wzrostem sieci nowe instalacje mogą przestać zależeć od pierwszego komputera bootstrap.

## Operator publicznego bootstrap noda

`START-SEED-NODE.cmd` jest przeznaczony wyłącznie dla operatora noda, który ma być publicznym punktem pierwszego kontaktu. Skrypt uruchamia AuronQ oraz, jeśli Tailscale jest zainstalowany i Funnel wcześniej autoryzowany, wystawia lokalny port AuronQ przez Funnel w tle.

Zwykły użytkownik uruchamia `START-AURONQ.cmd`, nie `START-SEED-NODE.cmd`.

Aktualne bootstrapy są publikowane w:

`https://raw.githubusercontent.com/promirmir/AuronQ/main/bootstrap.json`

Na moment publikacji dokumentu potwierdzony jest jeden publiczny bootstrap: `https://mir.taild63f46.ts.net`. Bootstrap nie ma uprawnień konsensusowych.

## Tożsamość sieci

Network ID:

`44e62c2ace002a6660c14e252173c1aa303529c68e40c998e92da2b453f44f30b1e58c94d533587e2186004593fb856c433fcdb5418ed430ec8617e29529365c`

Genesis:

`5750a455c04bfe93c9edfef1a12744b05e29ac6da1a9dd5b790566629dea2080c581beb2f0324efba2067c9efb113ed7a29598fd3ffbed965ced31f265d0cec4`

## Bezpieczeństwo

AuronQ 1.7.3 jest publicznym mainnetem, ale AQM64 i cały konsensus/network stack nie przeszły niezależnego profesjonalnego audytu kryptograficznego i bezpieczeństwa.

## Usuwanie portfela

W wersji 1.7.3 zakładka **Portfele** ma przycisk **Usuń**. Usuwanie dotyczy wyłącznie lokalnego zaszyfrowanego pliku `.wallet`; nie przenosi ani nie niszczy monet zapisanych w blockchainie.

Aplikacja wyświetla ostrzeżenie i wymaga ręcznego wpisania dokładnej nazwy portfela. Portfela używanego aktualnie przez koparkę nie można usunąć, dopóki mining nie zostanie zatrzymany. Przed usunięciem ważnego portfela wykonaj **Backup** — usunięcie ostatniej kopii pliku może oznaczać trwałą utratę dostępu do środków.
