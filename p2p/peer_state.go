package p2p

import (
	"fmt"
	"net"
)

type PeerState struct {
	Conn     net.Conn
	Chocked  bool
	Bitfield []byte
}

func (p *PeerState) handleMessage(msg *Message) {
	if msg == nil {
		return
	}

	switch int(msg.ID) {
	case 0:
		fmt.Println("choke")
	case 1:
		fmt.Println("unchoke")
	case 2:
		fmt.Println("interested")
	case 3:
		fmt.Println("not interested")
	case 4:
		fmt.Println("have")
	case 5:
		fmt.Println("bitfield")
	case 6:
		fmt.Println("request")
	case 7:
		fmt.Println("piece")
	case 8:
		fmt.Println("cancel")
	}
}
