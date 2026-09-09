package client

import (
	"go-bittorrent/p2p"
	"go-bittorrent/storage"
	torrentfile "go-bittorrent/torrent-file"
	"go-bittorrent/tui"
	"net"
	"sync"
	"time"

	tea "charm.land/bubbletea/v2"
)

type TorrentClient struct {
	Torrent           *torrentfile.TorrentFile
	Storage           *storage.TorrentStorage
	PeerId            [20]byte
	Peers             []p2p.Peer
	WorkQueue         chan *p2p.PieceWork
	activeConnections []net.Conn
	bytesDownloaded   uint64
	completedPieces   uint32
	totalPieces       uint32
	mu                sync.Mutex
	wg                sync.WaitGroup

	UI *tea.Program
}

func (c *TorrentClient) Download() {
	c.initWorkQueue()

	go c.startSpeedMonitor()

	for _, peer := range c.Peers {
		c.wg.Add(1)
		go func(p p2p.Peer) {
			defer c.wg.Done()
			if err := c.startWorker(p); err != nil {
			}
		}(peer)
	}

	c.wg.Wait()
}

func (c *TorrentClient) initWorkQueue() {
	numOfPieces := uint32(len(c.Torrent.PieceHashes))
	c.totalPieces = numOfPieces
	c.completedPieces = 0

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
	c.mu.Lock()
	c.activeConnections = append(c.activeConnections, peerState.Conn)
	c.mu.Unlock()

	defer peerState.Conn.Close()

	peerState.OnPieceComplete = c.markPieceComplete

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

func (c *TorrentClient) markPieceComplete(pieceLength uint32) {
	c.mu.Lock()

	c.completedPieces++
	c.bytesDownloaded += uint64(pieceLength)

	isDone := c.completedPieces == c.totalPieces
	if isDone {
		c.UI.Send(tui.DownloadCompleteMsg{})
		close(c.WorkQueue)
		for _, conn := range c.activeConnections {
			conn.Close()
		}
	}
	c.mu.Unlock()
}

func (c *TorrentClient) startSpeedMonitor() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var lastBytes uint64 = 0

	for {
		<-ticker.C

		c.mu.Lock()
		currentBytes := c.bytesDownloaded
		completed := c.completedPieces
		total := c.totalPieces
		c.mu.Unlock()

		bytesInLastSecond := currentBytes - lastBytes
		lastBytes = currentBytes
		speedMBps := float64(bytesInLastSecond) / 1048576.0

		if total > 0 {
			progress := (float32(completed) / float32(total)) * 100.0

			c.UI.Send(tui.StatsUpdateMsg{
				Progress:        progress,
				Speed:           speedMBps,
				BytesDownloaded: c.bytesDownloaded,
				TotalPeers:      len(c.Peers),
				ActivePeers:     len(c.activeConnections),
			})
		}
	}
}
