package main

import (
	"fmt"
	torrentfile "go-bittorrent/cmd/torrent-file"
)

func main() {
	tf, err := torrentfile.Open(`C:\vscode\go-bittorrent\ubuntu-26.04-desktop-amd64.iso.torrent`)
	if err != nil {
		fmt.Print(err)
	}
	fmt.Print(tf)
}
