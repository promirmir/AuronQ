# AuronQ 1.7.0 Mainnet — Windows

## Normalny użytkownik

1. Rozpakuj cały ZIP.
2. Uruchom `START-AURONQ.cmd`.
3. AuronQ Desktop uruchamia pełny węzeł, wczytuje niezmienny mainnet `network.json`, pobiera oficjalny manifest bootstrap z GitHuba, weryfikuje Network ID i automatycznie poznaje/zapisuje kolejnych publicznych peerów.
4. Użytkownik za NAT/CGNAT nadal uczestniczy normalnie przez połączenia wychodzące. Przekierowanie portu nie jest wymagane do zwykłego używania, synchronizacji, transakcji ani kopania.

## Oficjalne bootstrap nody operatora

Na dwóch komputerach operatora uruchom `START-SEED-NODE.cmd` zamiast zwykłego launchera. Skrypt uruchamia Desktop oraz, jeżeli Tailscale jest zainstalowany i Funnel został wcześniej zatwierdzony dla tailnetu, włącza publiczny HTTPS Funnel w tle. Nie trzeba zostawiać PowerShella otwartego.

Aktualne oficjalne bootstrap hosty są publikowane w `bootstrap.json` w repozytorium `promirmir/AuronQ`. To są tylko punkty pierwszego kontaktu. Nie przechowują centralnego stanu sieci i nie decydują o ważności bloków — każdy pełny node waliduje łańcuch samodzielnie.

## Tożsamość sieci

Network ID:
`44e62c2ace002a6660c14e252173c1aa303529c68e40c998e92da2b453f44f30b1e58c94d533587e2186004593fb856c433fcdb5418ed430ec8617e29529365c`

Genesis:
`5750a455c04bfe93c9edfef1a12744b05e29ac6da1a9dd5b790566629dea2080c581beb2f0324efba2067c9efb113ed7a29598fd3ffbed965ced31f265d0cec4`

## Bezpieczeństwo

AuronQ 1.7.0 jest publicznym mainnetem technicznie uruchamianym przez pełne węzły P2P, ale AQM64 i cały konsensus nie przeszły niezależnego profesjonalnego audytu kryptograficznego/bezpieczeństwa. Nie należy przedstawiać projektu jako zewnętrznie audytowanej infrastruktury finansowej.