package p2p

import (
	"encoding/binary"
	"io"
	"net"
)

type Message struct {
	ID      byte
	Payload []byte
}

func readMessage(conn net.Conn) (*Message, error) {
	lengthBuffer := make([]byte, 4)

	_, err := io.ReadFull(conn, lengthBuffer)
	if err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint32(lengthBuffer)
	if length == 0 {
		return nil, nil
	}

	buffer := make([]byte, length)
	_, err = io.ReadFull(conn, buffer)
	if err != nil {
		return nil, err
	}

	message := Message{
		ID:      buffer[0],
		Payload: buffer[1:],
	}

	return &message, nil
}
