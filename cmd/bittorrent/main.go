package main

import (
	"fmt"
	"go-bittorrent/cmd/bencode"
	"os"
)

func main() {
	file, err := os.Open(`C:\vscode\go-bittorrent\cmd\bittorrent\hellooo.txt.torrent`)
	if err != nil {
		fmt.Print(err)
	} else {
		bencode.Encode(bencode.Decode(file))
	}
}
