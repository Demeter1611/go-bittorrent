package storage

import (
	"fmt"
	torrentfile "go-bittorrent/torrent-file"
	"os"
	"path/filepath"
)

type FileEntry struct {
	File        *os.File
	Length      int64
	GlobalStart int64
}

type TorrentStorage struct {
	Files []FileEntry
}

func (ts *TorrentStorage) WriteGlobal(data []byte, globalOffset int64) error {
	pieceEnd := globalOffset + int64(len(data))
	dataStart := 0

	for _, file := range ts.Files {
		fileGlobalEnd := file.GlobalStart + file.Length

		if globalOffset >= file.GlobalStart && globalOffset < fileGlobalEnd {

			var writeUntilGlobal int64
			if fileGlobalEnd <= pieceEnd {
				writeUntilGlobal = fileGlobalEnd
			} else {
				writeUntilGlobal = pieceEnd
			}

			bytesToWrite := writeUntilGlobal - globalOffset

			chunk := data[dataStart : dataStart+int(bytesToWrite)]

			localOffset := globalOffset - file.GlobalStart

			_, err := file.File.WriteAt(chunk, localOffset)
			if err != nil {
				return err
			}

			globalOffset = writeUntilGlobal
			dataStart += int(bytesToWrite)
		}

		if globalOffset == pieceEnd {
			return nil
		}
	}

	if globalOffset != pieceEnd {
		return fmt.Errorf("incomplete write: written %d out of %d", globalOffset, pieceEnd)
	}

	return nil
}

func NewTorrentStorage(tf *torrentfile.TorrentFile) (*TorrentStorage, error) {
	var globalOffset int64 = 0
	ts := &TorrentStorage{Files: []FileEntry{}}

	if len(tf.Files) == 0 {
		openedFile, err := os.OpenFile(tf.Name, os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			return nil, err
		}

		ts.Files = append(ts.Files, FileEntry{
			File:        openedFile,
			Length:      tf.Length,
			GlobalStart: 0,
		})

		return ts, nil
	}

	for _, f := range tf.Files {
		fullPath := filepath.Join(append([]string{tf.Name}, f.Path...)...)

		dir := filepath.Dir(fullPath)
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return nil, err
		}

		openedFile, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			ts.Destroy()
			return nil, err
		}

		err = openedFile.Truncate(f.Length)
		if err != nil {
			openedFile.Close()
			ts.Destroy()
			return nil, err
		}

		fileEntry := FileEntry{
			File:        openedFile,
			Length:      f.Length,
			GlobalStart: int64(globalOffset),
		}

		ts.Files = append(ts.Files, fileEntry)

		globalOffset += f.Length
	}

	return ts, nil
}

func (ts *TorrentStorage) Destroy() {
	for _, fileEntry := range ts.Files {
		fileEntry.File.Close()
	}
}
