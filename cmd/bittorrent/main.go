package main

import (
	"fmt"
	torrentfile "go-bittorrent/cmd/torrent-file"
)

func main() {
	tf, err := torrentfile.Open(`C:\vscode\go-bittorrent\hellooo.txt.torrent`)
	if err != nil {
		fmt.Print(err)
	}
	fmt.Print(tf)
}
