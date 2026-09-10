package tracker

import (
	"encoding/binary"
	"fmt"
	"go-bittorrent/bencode"
	"go-bittorrent/p2p"
	torrentfile "go-bittorrent/torrent-file"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
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

func buildTrackerRequest(torrentFile *torrentfile.TorrentFile, announce string, peerId [20]byte, port uint16) (string, error) {
	base, err := url.Parse(announce)
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

func GetPeers(torrentFile *torrentfile.TorrentFile, peerId [20]byte, port uint16) ([]p2p.Peer, error) {
	for _, announce := range torrentFile.AnnounceList {
		peers, err := sendTrackerRequest(torrentFile, announce, peerId, port)
		if err == nil {
			return peers, nil
		}
	}
	return nil, fmt.Errorf("no trackers could be reached")
}

func sendTrackerRequest(torrentFile *torrentfile.TorrentFile, announce string, peerId [20]byte, port uint16) ([]p2p.Peer, error) {
	if strings.HasPrefix(announce, "http") {
		return sendHTTPTrackerRequest(torrentFile, announce, peerId, port)
	} else if strings.HasPrefix(announce, "udp") {
		return sendUDPTrackerRequest(torrentFile, announce, peerId, port)
	}

	return nil, fmt.Errorf("unknown protocol: %s", announce)
}

func sendHTTPTrackerRequest(torrentFile *torrentfile.TorrentFile, announce string, peerId [20]byte, port uint16) ([]p2p.Peer, error) {
	trackerUrl, err := buildTrackerRequest(torrentFile, announce, peerId, port)
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(trackerUrl)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tracker returned status %d", resp.StatusCode)
	}

	decoded, err := bencode.Decode(resp.Body)
	if err != nil {
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

func sendUDPTrackerRequest(torrentFile *torrentfile.TorrentFile, announce string, peerId [20]byte, port uint16) ([]p2p.Peer, error) {
	base, err := url.Parse(announce)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTimeout("udp", base.Host, 5*time.Second)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(15 * time.Second))

	connectionRequest, transactionID := generateUDPConnectRequest()
	_, err = conn.Write(connectionRequest)
	if err != nil {
		return nil, fmt.Errorf("error sending connect: %v", err)
	}

	connectionResponse := make([]byte, 16)
	length, err := conn.Read(connectionResponse)
	if err != nil {
		return nil, fmt.Errorf("error reading connect: %v", err)
	}
	if length < 16 {
		return nil, fmt.Errorf("connect response too short")
	}

	action := binary.BigEndian.Uint32(connectionResponse[0:4])
	if action != 0 {
		return nil, fmt.Errorf("tracker returned error on connect")
	}

	transactionIDReceived := binary.BigEndian.Uint32(connectionResponse[4:8])
	if transactionIDReceived != transactionID {
		return nil, fmt.Errorf("invalid transaction id on connect")
	}

	connectionID := binary.BigEndian.Uint64(connectionResponse[8:16])

	announceRequest, transactionID := generateUDPAnnounceRequest(connectionID, torrentFile, peerId, uint32(port))
	_, err = conn.Write(announceRequest)
	if err != nil {
		return nil, fmt.Errorf("error sending announce: %v", err)
	}

	announceResponse := make([]byte, 2048)
	length, err = conn.Read(announceResponse)
	if err != nil {
		return nil, fmt.Errorf("error reading announce: %v", err)
	}

	announceResponse = announceResponse[:length]

	if len(announceResponse) < 20 {
		return nil, fmt.Errorf("announce response too short")
	}

	action = binary.BigEndian.Uint32(announceResponse[0:4])
	if action != 1 {
		if action == 3 {
			return nil, fmt.Errorf("tracker error: %s", string(announceResponse[8:]))
		}
		return nil, fmt.Errorf("unknown tracker action")
	}

	transactionIDReceived = binary.BigEndian.Uint32(announceResponse[4:8])
	if transactionIDReceived != transactionID {
		return nil, fmt.Errorf("invalid transaction id on announce")
	}

	peersData := announceResponse[20:]
	return parsePeers(string(peersData))
}

func generateUDPConnectRequest() ([]byte, uint32) {
	const PROTOCOL_ID uint64 = 0x41727101980

	message := make([]byte, 16)

	transactionID := rand.Uint32()

	binary.BigEndian.PutUint64(message[0:8], PROTOCOL_ID)
	binary.BigEndian.PutUint32(message[8:12], 0)
	binary.BigEndian.PutUint32(message[12:16], transactionID)

	return message, transactionID
}

func generateUDPAnnounceRequest(connectionID uint64, torrentFile *torrentfile.TorrentFile, peerId [20]byte, port uint32) ([]byte, uint32) {
	/*
		Message format:
		OFFSET --- SIZE --- NAME
		0 --- 8 --- Connection ID
		8 --- 4 --- Action (1 for Announce)
		12 --- 4 --- Transaction ID (random)
		16 --- 20 --- InfoHash
		36 --- 20 --- PeerID
		56 --- 8 --- Downloaded
		64 --- 8 --- Left
		72 --- 8 --- Uploaded
		80 --- 4 --- Event: 0(None), 1(Completed), 2(Started), 3(Stopped)
		84 --- 4 --- IP
		88 --- 4 --- Key (random)
		92 --- 4 --- Num_Want (number of peers wanted)
		96 --- 2 --- Port
	*/

	message := make([]byte, 98)

	transactionID := rand.Uint32()
	key := rand.Uint32()
	peersWanted := 50

	binary.BigEndian.PutUint64(message[0:8], connectionID)
	binary.BigEndian.PutUint32(message[8:12], 1) // Announce
	binary.BigEndian.PutUint32(message[12:16], transactionID)
	copy(message[16:36], torrentFile.InfoHash[:])
	copy(message[36:56], peerId[:])
	binary.BigEndian.PutUint64(message[56:64], 0)
	binary.BigEndian.PutUint64(message[64:72], uint64(torrentFile.TotalFileSize()))
	binary.BigEndian.PutUint64(message[72:80], 0)
	binary.BigEndian.PutUint32(message[80:84], 2)
	binary.BigEndian.PutUint32(message[84:88], 0)
	binary.BigEndian.PutUint32(message[88:92], key)
	binary.BigEndian.PutUint32(message[92:96], uint32(peersWanted))
	binary.BigEndian.PutUint16(message[96:98], uint16(port))

	return message, transactionID
}
