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

func Serialize(msg *Message) []byte {
	if msg == nil {
		return make([]byte, 4)
	}
	msgLength := uint32(1 + len(msg.Payload))

	buffer := make([]byte, msgLength+4)

	binary.BigEndian.PutUint32(buffer[0:4], msgLength)

	buffer[4] = msg.ID

	if len(msg.Payload) > 0 {
		copy(buffer[5:], msg.Payload)
	}

	return buffer
}

func buildRequestPayload(index uint32, begin uint32, length uint32) []byte {
	buffer := make([]byte, 12)

	binary.BigEndian.PutUint32(buffer[0:4], index)

	binary.BigEndian.PutUint32(buffer[4:8], begin)

	binary.BigEndian.PutUint32(buffer[8:12], length)

	return buffer
}
