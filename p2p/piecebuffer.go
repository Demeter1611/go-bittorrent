package p2p

import (
	"crypto/sha1"
	"fmt"
)

type PieceBuffer struct {
	Index          uint32
	Buffer         []byte
	receivedBytes  uint32
	blocksReceived map[uint32]bool
}

func NewPieceBuffer(index uint32, pieceSize uint32) *PieceBuffer {
	return &PieceBuffer{
		Index:          index,
		Buffer:         make([]byte, pieceSize),
		receivedBytes:  0,
		blocksReceived: make(map[uint32]bool),
	}
}

func (pb *PieceBuffer) AddBlock(offset uint32, data []byte) error {
	if offset+uint32(len(data)) > uint32(len(pb.Buffer)) {
		return fmt.Errorf("data out of bounds: offset %d, len %d, buffer length %d", offset, len(data), len(pb.Buffer))
	}

	if pb.blocksReceived[offset] {
		return fmt.Errorf("block with offset %d already received", offset)
	}

	pb.receivedBytes += uint32(len(data))
	pb.blocksReceived[offset] = true

	copy(pb.Buffer[offset:], data)

	return nil
}

func (pb *PieceBuffer) IsDone() bool {
	return pb.receivedBytes == uint32(len(pb.Buffer))
}

func (pb *PieceBuffer) Verify(expectedHash [20]byte) bool {
	calculatedHash := sha1.Sum(pb.Buffer)

	return calculatedHash == expectedHash
}
