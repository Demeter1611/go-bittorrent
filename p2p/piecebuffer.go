package p2p

import (
	"crypto/sha1"
	"fmt"
)

type PieceBuffer struct {
	Index      uint32
	Buffer     []byte
	Downloaded uint32
}

func NewPieceBuffer(index uint32, pieceSize uint32) *PieceBuffer {
	return &PieceBuffer{
		Index:      index,
		Buffer:     make([]byte, pieceSize),
		Downloaded: 0,
	}
}

func (pb *PieceBuffer) AddBlock(offset uint32, data []byte) error {
	if offset+uint32(len(data)) > uint32(len(pb.Buffer)) {
		return fmt.Errorf("data out of bounds: offset %d, len %d, buffer length %d", offset, len(data), len(pb.Buffer))
	}

	copy(pb.Buffer[offset:], data)
	pb.Downloaded += uint32(len(data))

	return nil
}

func (pb *PieceBuffer) IsDone() bool {
	return pb.Downloaded == uint32(len(pb.Buffer))
}

func (pb *PieceBuffer) Verify(expectedHash [20]byte) bool {
	calculatedHash := sha1.Sum(pb.Buffer)

	return calculatedHash == expectedHash
}
