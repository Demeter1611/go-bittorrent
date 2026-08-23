package client

import (
	"fmt"
	"go-bittorrent/p2p"
	"go-bittorrent/storage"
	torrentfile "go-bittorrent/torrent-file"
	"sync"
)

type TorrentClient struct {
	Torrent   *torrentfile.TorrentFile
	Storage   *storage.TorrentStorage
	PeerId    [20]byte
	Peers     []p2p.Peer
	WorkQueue chan *p2p.PieceWork
	wg        sync.WaitGroup
}

func (c *TorrentClient) Download() {
	c.initWorkQueue()

	for _, peer := range c.Peers {
		c.wg.Add(1)
		go func(p p2p.Peer) {
			defer c.wg.Done()
			if err := c.startWorker(p); err != nil {
				fmt.Println(err)
			}
		}(peer)
	}

	c.wg.Wait()
}

func (c *TorrentClient) initWorkQueue() {
	numOfPieces := uint32(len(c.Torrent.PieceHashes))
	c.WorkQueue = make(chan *p2p.PieceWork, numOfPieces)

	for index := uint32(0); index < numOfPieces; index++ {
		c.WorkQueue <- &p2p.PieceWork{
			Index:  index,
			Length: uint32(c.Torrent.PieceSize(index)),
			Hash:   c.Torrent.PieceHashes[index],
		}
	}
}

func (c *TorrentClient) startWorker(peer p2p.Peer) error {
	peerState, err := p2p.Connection(peer, c.PeerId, c.Torrent, c.Storage, c.WorkQueue)
	if err != nil {
		return err
	}
	defer peerState.Conn.Close()

	err = peerState.RunEventLoop()
	if peerState.CurrentBuffer != nil {
		c.WorkQueue <- &p2p.PieceWork{
			Index:  peerState.CurrentBuffer.Index,
			Length: uint32(len(peerState.CurrentBuffer.Buffer)),
			Hash:   peerState.Torrent.PieceHashes[peerState.CurrentBuffer.Index],
		}
	}
	return err
}
