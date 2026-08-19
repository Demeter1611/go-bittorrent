package tracker

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"go-bittorrent/bencode"
	"go-bittorrent/p2p"
	torrentfile "go-bittorrent/torrent-file"
	"net"
	"net/http"
	"net/url"
	"strconv"
)

func GeneratePeerId() ([20]byte, error) {
	peerId := [20]byte{}
	copy(peerId[:], "-GO0001-")
	_, err := rand.Read(peerId[8:])
	if err != nil {
		return [20]byte{}, err
	}
	return peerId, nil
}

func buildTrackerRequest(torrentFile *torrentfile.TorrentFile, peerId [20]byte, port uint16) (string, error) {
	base, err := url.Parse(torrentFile.Announce)
	if err != nil {
		return "", err
	}

	newValues := url.Values{
		"info_hash":  []string{string(torrentFile.InfoHash[:])},
		"peer_id":    []string{string(peerId[:])},
		"port":       []string{strconv.Itoa(int(port))},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"left":       []string{strconv.FormatInt(torrentFile.TotalFileSize(), 10)},
		"compact":    []string{"1"},
	}

	base.RawQuery = newValues.Encode()
	return base.String(), nil
}

func SendTrackerRequest(torrentFile *torrentfile.TorrentFile, peerId [20]byte, port uint16) ([]p2p.Peer, error) {
	trackerUrl, err := buildTrackerRequest(torrentFile, peerId, port)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	resp, err := http.Get(trackerUrl)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println(resp.StatusCode)
		return nil, fmt.Errorf("tracker returned status %d", resp.StatusCode)
	}

	decoded, err := bencode.Decode(resp.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	trackerDict, ok := decoded.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid tracker response")
	}

	if reason, ok := trackerDict["failure reason"].(string); ok {
		return nil, fmt.Errorf("tracker error: %s", reason)
	}

	peersValue, ok := trackerDict["peers"]
	if !ok {
		return nil, fmt.Errorf("tracker response missing 'peers' key")
	}

	peersList, err := parsePeers(peersValue)
	if err != nil {
		return nil, err
	}

	return peersList, nil
}

func parsePeers(peersValue any) ([]p2p.Peer, error) {
	var peers []p2p.Peer

	switch v := peersValue.(type) {
	case string:
		peerSize := 6
		if len(v)%6 != 0 {
			return nil, fmt.Errorf("invalid binary peer length")
		}

		numPeers := len(v) / peerSize
		for i := 0; i < numPeers; i++ {
			offset := i * peerSize
			ip := net.IP([]byte(v[offset : offset+4]))
			port := binary.BigEndian.Uint16([]byte(v[offset+4 : offset+6]))
			peers = append(peers, p2p.Peer{IP: ip, Port: port})
		}

	case []any:
		for _, peerItem := range v {
			peerDict, ok := peerItem.(map[string]any)
			if !ok {
				continue
			}

			ipString, ipOk := peerDict["ip"].(string)
			portInt, portOk := peerDict["port"].(int64)

			if ipOk && portOk {
				peers = append(peers, p2p.Peer{IP: net.ParseIP(ipString), Port: uint16(portInt)})
			}
		}

	default:
		return nil, fmt.Errorf("unknown peer format")
	}
	return peers, nil

}
