package p2p

import (
	"fmt"
	torrentfile "go-bittorrent/torrent-file"
	"net"
)

type PeerState struct {
	Conn     net.Conn
	Chocked  bool
	Bitfield []byte
	Torrent  *torrentfile.TorrentFile
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
	p.Chocked = true
}

func (p *PeerState) handleUnchoke() {
	p.Chocked = false
}

func (p *PeerState) handleBitfield(msg *Message) {
	p.Bitfield = make([]byte, len(msg.Payload))
	copy(p.Bitfield, msg.Payload)

	interestedMsg := &Message{ID: 2}
	p.SendMessage(interestedMsg)

}

func (p *PeerState) handlePiece(msg *Message) {

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
