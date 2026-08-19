package torrentfile

import (
	"bytes"
	"crypto/sha1"
	"testing"
)

func TestSplitPieceHAshes(t *testing.T) {
	validPieces := "1234567890123456789012345678901234567890"

	tests := []struct {
		name        string
		pieces      string
		expectError bool
		expectedLen int
	}{
		{"Valid pieces (40 bytes)", validPieces, false, 2},
		{"Invalid pieces (25 bytes)", "1234567890123456789012345", true, 0},
		{"Empty pieces (0 bytes)", "", false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashes, err := splitPieceHashes(tt.pieces)

			if tt.expectError {
				if err == nil {
					t.Fatalf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if len(hashes) != tt.expectedLen {
					t.Errorf("Expected %d hashes, got %d", tt.expectedLen, len(hashes))
				}
			}
		})
	}
}

func TestComputeInfoHash(t *testing.T) {
	infoDict := map[string]any{
		"name": "test.txt",
	}

	expectedHash := sha1.Sum([]byte("d4:name8:test.txte"))

	hash, err := computeInfoHash(infoDict)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !bytes.Equal(hash[:], expectedHash[:]) {
		t.Errorf("Incorrect hash. Expected: %x, got %x", expectedHash, hash)
	}
}

func TestParseTorrentFile(t *testing.T) {
	dict := map[string]any{
		"announce": "http://tracker.example.com/announce",
		"info": map[string]any{
			"name":         "test.txt",
			"length":       100,
			"piece length": 32,
			"pieces":       "12345678901234567890",
		},
	}

	tf, err := parseTorrentFile(dict)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if tf.Announce != "http://tracker.example.com/announce" {
		t.Errorf("Incorrect announce: %v", tf.Announce)
	}

	if tf.Name != "test.txt" {
		t.Errorf("Incorrect name: %v", tf.Name)
	}

	if tf.Length != 100 {
		t.Errorf("Incorrect length: %v", tf.Length)
	}
	if len(tf.PieceHashes) != 1 {
		t.Errorf("Expected 1 piece hash, got: %d", len(tf.PieceHashes))
	}
}
