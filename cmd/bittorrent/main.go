package main

import (
	"fmt"
	"go-bittorrent/client"
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
		return
	}

	ts, err := storage.NewTorrentStorage(tf)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer ts.Destroy()

	client := client.TorrentClient{
		Torrent: tf,
		Storage: ts,
		PeerId:  peerId,
		Peers:   peers,
	}

	client.Download()
}
