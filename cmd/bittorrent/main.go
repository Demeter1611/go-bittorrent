package main

import (
	"fmt"
	"go-bittorrent/p2p"
	"go-bittorrent/storage"
	torrentfile "go-bittorrent/torrent-file"
	"go-bittorrent/tracker"
)

func main() {
	tf, err := torrentfile.Open(`C:\vscode\go-bittorrent\test-torrents\ubuntu-26.04-desktop-amd64.iso.torrent`)
	if err != nil {
		fmt.Print(err)
		return
	}

	peerId, err := tracker.GeneratePeerId()
	if err != nil {
		fmt.Print(err)
		return
	}

	peers, err := tracker.SendTrackerRequest(tf, peerId, 6881)
	if err != nil {
		fmt.Println(err)
	}

	ts, err := storage.NewTorrentStorage(tf)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer ts.Destroy()

	for _, peer := range peers {
		conn, err := p2p.Connection(peer, tf.InfoHash, peerId, tf, ts)
		if err != nil {
			fmt.Printf("failed with %s: %v\n", peer.IP, err)
			continue
		}
		fmt.Printf("handshake OK with %s\n", peer.IP)
		conn.Close()
	}
}
