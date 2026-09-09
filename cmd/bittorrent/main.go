package main

import (
	"flag"
	"fmt"
	"go-bittorrent/client"
	"go-bittorrent/storage"
	torrentfile "go-bittorrent/torrent-file"
	"go-bittorrent/tracker"
	"go-bittorrent/tui"
	"os"

	tea "charm.land/bubbletea/v2"
)

func main() {
	torrentPath := flag.String("file", "", "Path to .torrent file")

	flag.Parse()

	if *torrentPath == "" {
		fmt.Println("Error: .torrent file not specified")
		fmt.Println(`Correct usage: go-bittorrent -file="C:\\path\\to\\file.torrent`)
		return
	}

	p := tea.NewProgram(tui.InitialModel())

	go func() {
		tf, err := torrentfile.Open(*torrentPath)
		p.Send(tui.FileSelectedMsg{FileName: *torrentPath, TotalFileSize: uint64(tf.TotalFileSize())})
		if err != nil {
			fmt.Println(err)
			p.Quit()
			return
		}

		peerId, err := tracker.GeneratePeerId()
		if err != nil {
			fmt.Println(err)
			p.Quit()
			return
		}

		peers, err := tracker.SendTrackerRequest(tf, peerId, 6881)
		if err != nil {
			fmt.Println(err)
			p.Quit()
			return
		}

		ts, err := storage.NewTorrentStorage(tf)
		if err != nil {
			fmt.Println(err)
			p.Quit()
			return
		}
		defer ts.Destroy()

		client := client.TorrentClient{
			Torrent: tf,
			Storage: ts,
			PeerId:  peerId,
			Peers:   peers,
			UI:      p,
		}

		client.Download()
	}()

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
