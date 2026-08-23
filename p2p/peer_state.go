package p2p

import (
	"encoding/binary"
	"fmt"
	"go-bittorrent/storage"
	torrentfile "go-bittorrent/torrent-file"
	"net"
)

type PieceWork struct {
	Index  uint32
	Length uint32
	Hash   [20]byte
}

type PeerState struct {
	Conn          net.Conn
	Choked        bool
	Bitfield      []byte
	Torrent       *torrentfile.TorrentFile
	Storage       *storage.TorrentStorage
	CurrentBuffer *PieceBuffer
	WorkQueue     chan *PieceWork
}

func (p *PeerState) RunEventLoop() error {
	for {
		msg, err := readMessage(p.Conn)
		if err != nil {
			return err
		}

		p.HandleMessage(msg)
	}
}

func (p *PeerState) HandleMessage(msg *Message) {
	if msg == nil {
		return
	}

	switch int(msg.ID) {
	case 0:
		fmt.Println("choke")
		p.handleChoke()
	case 1:
		fmt.Println("unchoke")
		p.handleUnchoke()
	case 2:
		fmt.Println("interested")
	case 3:
		fmt.Println("not interested")
	case 4:
		fmt.Println("have")
	case 5:
		fmt.Println("bitfield")
		p.handleBitfield(msg)
	case 6:
		fmt.Println("request")
	case 7:
		fmt.Println("piece")
		p.handlePiece(msg)
	case 8:
		fmt.Println("cancel")
	}
}

func (p *PeerState) SendMessage(msg *Message) error {
	serialized := Serialize(msg)

	_, err := p.Conn.Write(serialized)
	if err != nil {
		return err
	}

	return nil
}

func (p *PeerState) handleChoke() {
	p.Choked = true
}

func (p *PeerState) handleUnchoke() {
	p.Choked = false

	p.startNextPiece()
}

func (p *PeerState) handleBitfield(msg *Message) {
	p.Bitfield = make([]byte, len(msg.Payload))
	copy(p.Bitfield, msg.Payload)

	interestedMsg := &Message{ID: 2}
	p.SendMessage(interestedMsg)

}

func (p *PeerState) handlePiece(msg *Message) {
	if p.CurrentBuffer == nil {
		return
	}
	if len(msg.Payload) < 8 {
		return
	}

	index := binary.BigEndian.Uint32(msg.Payload[0:4])
	begin := binary.BigEndian.Uint32(msg.Payload[4:8])
	block := msg.Payload[8:]

	if index != p.CurrentBuffer.Index {
		return
	}

	err := p.CurrentBuffer.AddBlock(begin, block)
	if err != nil {
		fmt.Println(err)
		return
	}

	if p.CurrentBuffer.IsDone() {
		expectedHash := p.Torrent.PieceHashes[index]

		if !p.CurrentBuffer.Verify(expectedHash) {
			p.WorkQueue <- &PieceWork{Index: index, Length: uint32(len(p.CurrentBuffer.Buffer)), Hash: expectedHash}
			p.CurrentBuffer = nil
			fmt.Println("corrupted data")
			return
		} else {
			globalOffset := int64(index) * p.Torrent.PieceLength

			err := p.Storage.WriteGlobal(p.CurrentBuffer.Buffer, globalOffset)

			if err != nil {
				p.WorkQueue <- &PieceWork{Index: index, Length: uint32(len(p.CurrentBuffer.Buffer)), Hash: expectedHash}
				fmt.Println(err)
			} else {
				fmt.Println("Piece succesfully written")
			}
		}

		p.CurrentBuffer = nil
		p.startNextPiece()
	}
}

func (p *PeerState) startNextPiece() {
	attempts := 0
	maxAttempts := len(p.Torrent.PieceHashes)

	for attempts < maxAttempts {
		select {
		case piece, ok := <-p.WorkQueue:
			if !ok {
				return
			}

			if !p.HasPiece(piece.Index) {
				p.WorkQueue <- piece
				attempts++
				continue
			}

			p.CurrentBuffer = NewPieceBuffer(piece.Index, piece.Length)
			p.requestPiece(piece.Index, piece.Length)
			return
		default:
			return
		}
	}
}

func (p *PeerState) HasPiece(index uint32) bool {
	byteIndex := index / 8

	bitOffset := index % 8

	if byteIndex >= uint32(len(p.Bitfield)) {
		return false
	}

	mask := byte(1 << (7 - bitOffset))

	return p.Bitfield[byteIndex]&mask != 0
}

func (p *PeerState) requestPiece(index uint32, pieceSize uint32) {
	const MAX_BLOCK_SIZE uint32 = 16384
	var begin uint32 = 0

	for begin < pieceSize {
		blockSize := MAX_BLOCK_SIZE

		if pieceSize-begin < MAX_BLOCK_SIZE {
			blockSize = pieceSize - begin
		}

		p.SendMessage(&Message{ID: 6, Payload: buildRequestPayload(index, begin, blockSize)})

		begin += blockSize
	}
}
