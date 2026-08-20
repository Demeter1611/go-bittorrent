package main

import (
	"fmt"
	"go-bittorrent/p2p"
	torrentfile "go-bittorrent/torrent-file"
	"go-bittorrent/tracker"
)

func main() {
	tf, err := torrentfile.Open(`C:\vscode\go-bittorrent\test-torrents\enwiki-20260601-pages-articles-multistream.xml.bz2-bac5df1f39fd83fc87826a8dc546e56db34f2322.torrent`)
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

	for _, peer := range peers {
		conn, err := p2p.Connection(peer, tf.InfoHash, peerId)
		if err != nil {
			fmt.Printf("failed with %s: %v\n", peer.IP, err)
			continue
		}
		fmt.Printf("handshake OK with %s\n", peer.IP)
		conn.Close()
	}
}
