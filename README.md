# go-bittorrent

A BitTorrent client implemented from scratch in Go — no third-party BitTorrent or bencode libraries. Built as a learning project to explore binary protocol design, concurrent networking, and low-level file I/O.

The client connects to a real BitTorrent swarm, performs peer handshakes, downloads pieces from multiple peers concurrently, verifies each piece's integrity via SHA-1, and reassembles the final file(s) on disk.

**Verified working end-to-end**: a full Ubuntu Desktop ISO was downloaded via this client and its SHA-256 checksum matched Ubuntu's official `SHA256SUMS` exactly, confirming byte-for-byte correctness against a real, multi-GB torrent.

## What it does

- Parses `.torrent` files using a custom-built bencode decoder/encoder
- Computes the torrent's `info_hash` (SHA-1 over the raw bencoded `info` dictionary)
- Announces to HTTP trackers and parses both compact and dictionary-style peer lists
- Performs the BitTorrent peer wire handshake (BEP 3) with real peers over TCP
- Implements the peer message protocol (`choke`, `unchoke`, `interested`, `bitfield`, `request`, `piece`)
- Downloads pieces concurrently from multiple peers via a shared work queue
- Verifies every downloaded piece against its expected SHA-1 hash before writing to disk
- Supports both single-file and multi-file torrents, correctly mapping pieces across file boundaries
- Detects when all pieces are downloaded and shuts down cleanly
