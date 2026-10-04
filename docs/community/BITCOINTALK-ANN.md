# Bitcointalk ANN — AuronQ (AURQ)

Suggested title:

`[ANN][AURQ][PoW] AuronQ — ML-DSA-87 signatures, AQM64 PoW, Windows/Linux/Android`

---

[center][size=24pt][b]AuronQ (AURQ)[/b][/size]

[size=14pt][b]A new open-source Proof-of-Work network built around independent validation and post-quantum transaction signatures[/b][/size]

[b]Mainnet is live • Windows / Linux full node • Android wallet • Source code public[/b][/center]

Hi everyone.

I have been working on a new cryptocurrency project called [b]AuronQ[/b], ticker [b]AURQ[/b].

I did not want to make another token that exists only as a contract on someone else's chain. The idea was to build an actual network from the ground up: its own blockchain, its own full node, wallet, mining, peer-to-peer networking, Explorer and mobile wallet.

Bitcoin was a major inspiration for me, especially because of how resilient and difficult to break its basic architecture has proven to be over many years of real-world operation. I did not want to copy Bitcoin, but its decentralization, independent validation and overall robustness strongly influenced the direction I wanted AuronQ to take.

AuronQ is still young and experimental, but it is already running as a real Mainnet and the source code is public.

The things that matter most to me are decentralization, security and preparing the project for future threats created by increasingly powerful computers and new computing technologies.

The project uses a UTXO model, Proof-of-Work, ML-DSA-87 transaction signatures and an AuronQ-specific mining construction called AQM64.

I am not presenting this as a finished replacement for Bitcoin or as something that is already battle-tested. It is not. The code and cryptographic composition still need independent review. What I want now is to get more technically interested people to actually run it, mine it, break it, review it and help expose weak points while the project is still young.

[hr]

[size=16pt][b]Why I started AuronQ[/b][/size]

What interested me most was the idea of building a cryptocurrency where users can actually verify things themselves instead of depending on one central website, API or Explorer.

A full AuronQ node validates its own blockchain locally. It checks transactions, UTXOs, signatures, Proof-of-Work, timestamps, difficulty and chain work.

The bootstrap nodes only help a fresh installation find the network. They do not decide what is valid and they do not have special consensus authority.

The built-in Explorer also reads from the full node's own validated chain. There is no idea that one public Explorer is the "truth".

I also wanted to experiment with post-quantum signatures at the transaction level. AuronQ currently uses ML-DSA-87, standardized in FIPS 204.

That does [b]not[/b] mean I am claiming the entire cryptocurrency is magically "quantum-proof". It means specifically that transaction authorization uses a post-quantum signature scheme.

[hr]

[size=16pt][b]Main links[/b][/size]

[b]Website:[/b]
https://promirmir.github.io/AuronQ/

[b]GitHub:[/b]
https://github.com/promirmir/AuronQ

[b]Latest Desktop / Full Node release:[/b]
https://github.com/promirmir/AuronQ/releases/tag/v1.7.13

[b]Public Explorer:[/b]
https://mir.taild63f46.ts.net/explorer

[b]Protocol specification:[/b]
https://github.com/promirmir/AuronQ/blob/main/PROTOCOL.md

[b]AQM64:[/b]
https://github.com/promirmir/AuronQ/blob/main/AQM64.md

[b]Decentralization model:[/b]
https://github.com/promirmir/AuronQ/blob/main/DECENTRALIZATION.md

[b]Roadmap:[/b]
https://github.com/promirmir/AuronQ/blob/main/ROADMAP.md

[hr]

[size=16pt][b]Current Mainnet parameters[/b][/size]

[table]
[tr][td][b]Ticker[/b][/td][td]AURQ[/td][/tr]
[tr][td][b]Model[/b][/td][td]UTXO[/td][/tr]
[tr][td][b]Consensus[/b][/td][td]Proof-of-Work[/td][/tr]
[tr][td][b]PoW[/b][/td][td]AQM64 v1[/td][/tr]
[tr][td][b]Transaction signatures[/b][/td][td]ML-DSA-87[/td][/tr]
[tr][td][b]Block target[/b][/td][td]600 seconds[/td][/tr]
[tr][td][b]Initial block reward[/b][/td][td]49.5 AURQ[/td][/tr]
[tr][td][b]Halving interval[/b][/td][td]210,000 mined blocks[/td][/tr]
[tr][td][b]Nominal maximum supply[/b][/td][td]21,000,000 AURQ[/td][/tr]
[tr][td][b]Genesis founder allocation[/b][/td][td]210,000 AURQ (1%, counted inside the supply cap)[/td][/tr]
[tr][td][b]Coinbase maturity[/b][/td][td]100 blocks[/td][/tr]
[tr][td][b]Default P2P port[/b][/td][td]18444[/td][/tr]
[tr][td][b]License[/b][/td][td]MIT[/td][/tr]
[/table]

I want to be completely open about the genesis allocation: [b]210,000 AURQ, or 1% of the nominal supply cap, was created in genesis for the founder address[/b]. It is part of the 21 million cap, not additional supply on top of it.

The normal block reward starts at 49.5 AURQ instead of 50 AURQ because the 1% founder allocation was carved out from the issuance model rather than added above the cap.

[b]Network ID:[/b]
[code]44e62c2ace002a6660c14e252173c1aa303529c68e40c998e92da2b453f44f30b1e58c94d533587e2186004593fb856c433fcdb5418ed430ec8617e29529365c[/code]

[b]Genesis hash:[/b]
[code]5750a455c04bfe93c9edfef1a12744b05e29ac6da1a9dd5b790566629dea2080c581beb2f0324efba2067c9efb113ed7a29598fd3ffbed965ced31f265d0cec4[/code]

[hr]

[size=16pt][b]What is AQM64?[/b][/size]

AQM64 is the Proof-of-Work construction used by AuronQ.

It is built from standardized primitives: SHAKE256 and Argon2id. The goal was to make mining memory-hard rather than simply create another SHA-style hash race.

At a high level:

[code]
pre  = SHAKE256-512(header)
salt = SHAKE256-256(previous_hash + height + algorithm)
mid  = Argon2id(pre, salt, 64 MiB memory, 2 iterations)
pow  = SHAKE256-512(pre + mid)
[/code]

The exact consensus serialization and domain separation are documented in the protocol specification.

I want to be clear about one thing: [b]AQM64 itself has not yet been independently cryptographically reviewed[/b], and I am not claiming it is ASIC-proof.

[hr]

[size=16pt][b]What already works[/b][/size]

At this point AuronQ is more than a protocol document.

There is already:

[list]
[li]a working Mainnet;[/li]
[li]Windows and Linux full-node releases;[/li]
[li]Desktop wallet;[/li]
[li]CPU mining;[/li]
[li]peer discovery and peer gossip;[/li]
[li]wallet transaction history;[/li]
[li]a built-in blockchain Explorer;[/li]
[li]Android light wallet;[/li]
[li]local ML-DSA-87 transaction signing;[/li]
[li]chain reorganization based on validated cumulative work;[/li]
[li]reproducible-build checks;[/li]
[li]fuzzing and race-detector tests;[/li]
[li]automated network health checks.[/li]
[/list]

I have also tested scenarios with multiple nodes being split into partitions, reconnecting and converging again.

One important test permanently removes the original bootstrap node and then verifies that the remaining network continues and that a fresh node can join using another surviving peer.

Recently a real mining bug was reported publicly: a miner could keep hashing an old block template after another miner had already advanced the chain.

That bug was fixed in v1.7.12. The miner now watches the current tip and abandons stale work when the chain advances or changes through a reorg.

I mention this because I would rather show real bugs and real fixes than pretend the software is perfect.

[hr]

[size=16pt][b]Downloads[/b][/size]

[b]Windows x64 — AuronQ 1.7.13[/b]
https://github.com/promirmir/AuronQ/releases/download/v1.7.13/AuronQ-1.7.13-Windows-x64.zip

SHA-256:
[code]7a7ce3d0b31b290bb6a3e43ab357ec6977ee4e7098a7d950594eda6cb64b2ba1[/code]

[b]Linux amd64 — AuronQ 1.7.13[/b]
https://github.com/promirmir/AuronQ/releases/download/v1.7.13/AuronQ-1.7.13-Linux-amd64.tar.gz

SHA-256:
[code]2bdd1ffa72ca99c7395a146d21943e4aedcabe24cde123b5af8ff3c6c640ecb6[/code]

[b]Android — AuronQ Mobile 0.5.1 Alpha[/b]
https://github.com/promirmir/AuronQ/releases/download/android-v0.5.1-alpha/AuronQ-Mobile-0.5.1-alpha.apk

SHA-256:
[code]c224dd9f94d52132d71e7f781dfadcbbc3a7bc92c2c07bf32d38dc4929328c73[/code]

Windows binaries are not currently Authenticode-signed, so SmartScreen may show a warning on first launch.

[hr]

[size=16pt][b]Quick start on Windows[/b][/size]

The easiest way to try the network is:

[list=1]
[li]Download the Windows ZIP.[/li]
[li]Check the SHA-256 checksum.[/li]
[li]Extract the whole archive.[/li]
[li]Run [code]START-AURONQ.cmd[/code].[/li]
[li]Wait for the node to connect and synchronize.[/li]
[li]Create or import a wallet and back it up.[/li]
[li]If you want to mine, enable CPU mining from the Desktop application.[/li]
[/list]

Full Windows guide:
https://github.com/promirmir/AuronQ/blob/main/README-WINDOWS.md

[hr]

[size=16pt][b]Android wallet[/b][/size]

The Android application is a light wallet, not a full node.

It keeps private keys and signing on the phone and independently verifies the Mainnet header chain from the embedded genesis, including AQM64 Proof-of-Work, difficulty, timestamps and hash continuity.

For wallet balances, transaction history and UTXO data it compares full nodes that match the locally verified header state.

It does not rebuild the complete UTXO set from every full block locally, so I do not describe it as equivalent to a full node.

[hr]

[size=16pt][b]What I am looking for now[/b][/size]

The next stage is not about pretending AuronQ is already a huge network.

I have high hopes for AuronQ. I think it has the potential to become a genuinely interesting project, and perhaps one day something much bigger, but only time, independent testing and real network growth can prove that.

I also wanted to build something that could outlive the initial development phase and perhaps one day be genuinely useful to someone. A strong decentralized network cannot be created by one person alone. It becomes stronger when independent people run nodes, mine, test, report problems and help expose weak points.

What the project needs most right now is independent people.

I would like to see:

[list]
[li]more independent full nodes;[/li]
[li]more independent miners;[/li]
[li]people testing synchronization and reorg behavior;[/li]
[li]developers reading the code;[/li]
[li]people looking specifically for consensus and networking bugs;[/li]
[li]security and cryptography reviewers willing to criticize the design.[/li]
[/list]

If someone finds a serious flaw, I want it reported publicly and fixed publicly.

GitHub issues:
https://github.com/promirmir/AuronQ/issues

[hr]

[size=16pt][b]Current limitations[/b][/size]

I think it is important to say what AuronQ is [b]not[/b] yet.

[list]
[li]It has not been independently audited.[/li]
[li]AQM64 has not received independent cryptographic review.[/li]
[li]The network is still small and needs more independent operators.[/li]
[li]The Android application is still Alpha.[/li]
[li]Windows binaries are not yet code-signed.[/li]
[li]There is no claim that the entire system is fully "quantum-proof".[/li]
[/list]

So if you are looking only for a finished product with years of battle testing, this is not there yet.

If you like testing young protocols and finding problems before they become big problems, that is exactly the kind of feedback I am looking for.

[hr]

[size=16pt][b]Feedback is welcome[/b][/size]

If you run it, I would genuinely like to know:

[list]
[li]what OS you used;[/li]
[li]whether the node found peers on its own;[/li]
[li]whether synchronization completed correctly;[/li]
[li]how mining behaved;[/li]
[li]whether the Android wallet connected and verified the chain;[/li]
[li]anything that looks wrong, confusing or unsafe.[/li]
[/list]

Technical criticism is welcome. I would rather have someone point out a real problem now than hide weaknesses behind marketing.

Bitcoin showed me how important it is for a network to remain useful without depending on a single operator. That kind of robustness is one of the strongest ideas behind AuronQ as well. The more independent people who run, mine and test the network, the stronger it can become against failures and attacks.

Thanks for taking a look at the project.
