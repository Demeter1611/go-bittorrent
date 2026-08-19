package main

import (
	"fmt"
	torrentfile "go-bittorrent/cmd/torrent-file"
	"go-bittorrent/cmd/tracker"
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
	tracker.SendTrackerRequest(tf, peerId, 6881)
}
