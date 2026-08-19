package torrentfile

import (
	"crypto/sha1"
	"fmt"
	"go-bittorrent/cmd/bencode"
	"os"
)

type TorrentFile struct {
	Announce    string
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int64
	Length      int64
	Name        string
	Files       []File
}

type File struct {
	Length int64
	Path   []string
}

func Open(path string) (*TorrentFile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	raw, err := bencode.Decode(f)
	if err != nil {
		return nil, err
	}
	dict, ok := raw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("bencode value is not a dictionary")
	}

	torrentFile, err := parseTorrentFile(dict)
	if err != nil {
		return nil, err
	}
	return torrentFile, err
}

func parseTorrentFile(dict map[string]any) (*TorrentFile, error) {
	announce := ""

	if val, ok := dict["announce"].(string); ok {
		announce = val
	}

	infoDict, ok := dict["info"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid 'info' dictionary")
	}

	name, ok := infoDict["name"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid 'name'")
	}

	pieceLength, ok := infoDict["piece length"].(int64)
	if !ok {
		return nil, fmt.Errorf("invalid 'piece length'")
	}

	piecesStr, ok := infoDict["pieces"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid 'pieces'")
	}

	pieceHashes, err := splitPieceHashes(piecesStr)
	if err != nil {
		return nil, err
	}

	infoHash, err := computeInfoHash(infoDict)
	if err != nil {
		return nil, err
	}

	torrentFile := &TorrentFile{
		Announce:    announce,
		InfoHash:    infoHash,
		PieceHashes: pieceHashes,
		PieceLength: pieceLength,
		Name:        name,
	}

	if length, ok := infoDict["length"].(int64); ok {
		torrentFile.Length = length
	} else if filesRaw, ok := infoDict["files"].([]any); ok {
		files, err := parseFiles(filesRaw)
		if err != nil {
			return nil, err
		}
		torrentFile.Files = files
	} else {
		return nil, fmt.Errorf("torrent does not have 'length' or 'files'")
	}

	return torrentFile, nil
}

func splitPieceHashes(pieces string) ([][20]byte, error) {
	buf := []byte(pieces)
	if len(buf)%20 != 0 {
		return nil, fmt.Errorf("invalid piece length: %d", len(buf))
	}

	numHashes := len(buf) / 20
	hashes := make([][20]byte, numHashes)

	for i := 0; i < numHashes; i++ {
		copy(hashes[i][:], buf[i*20:(i+1)*20])
	}

	return hashes, nil
}

func computeInfoHash(infoDict map[string]any) ([20]byte, error) {
	encoded, err := bencode.Encode(infoDict)
	if err != nil {
		return [20]byte{}, err
	}
	return sha1.Sum(encoded), nil
}

func parseFiles(filesRaw []any) ([]File, error) {
	files := make([]File, 0, len(filesRaw))

	for _, fileRaw := range filesRaw {
		fDict, ok := fileRaw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid file")
		}

		length, ok := fDict["length"].(int64)
		if !ok {
			return nil, fmt.Errorf("invalid file length")
		}

		pathRaw, ok := fDict["path"].([]any)
		if !ok {
			return nil, fmt.Errorf("invalid file path")
		}

		path := make([]string, 0, len(pathRaw))
		for _, p := range pathRaw {
			s, ok := p.(string)
			if !ok {
				return nil, fmt.Errorf("invalid path component")
			}
			path = append(path, s)
		}

		files = append(files, File{Length: length, Path: path})
	}

	return files, nil
}

func (t *TorrentFile) TotalFileSize() int64 {
	if len(t.Files) == 0 {
		return t.Length
	}

	var totalLength int64 = 0
	for _, file := range t.Files {
		totalLength += file.Length
	}
	return totalLength
}
