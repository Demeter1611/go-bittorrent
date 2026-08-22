package p2p

import (
	"fmt"
	"go-bittorrent/storage"
	torrentfile "go-bittorrent/torrent-file"
	"io"
	"net"
	"strconv"
	"time"
)

type Peer struct {
	IP   net.IP
	Port uint16
}

func Connection(peer Peer, infoHash [20]byte, peerId [20]byte, torrentFile *torrentfile.TorrentFile, torrentStorage *storage.TorrentStorage) (net.Conn, error) {
	portStr := strconv.Itoa(int(peer.Port))

	address := net.JoinHostPort(peer.IP.String(), portStr)

	conn, err := net.DialTimeout("tcp", address, 3*time.Second)
	if err != nil {
		return nil, fmt.Errorf("connection failed: %v", err)
	}

	_, err = Handshake(conn, infoHash, peerId)
	if err != nil {
		conn.Close()
		return nil, err
	}

	state := &PeerState{
		Conn:    conn,
		Chocked: true,
		Torrent: torrentFile,
		Storage: torrentStorage,
	}
	//separa loop-ul asta de run intr-o functie diferita
	for {
		msg, err := readMessage(conn)
		if err != nil {
			return conn, err
		}

		state.HandleMessage(msg)
	}

}

func Handshake(conn net.Conn, infoHash [20]byte, peerId [20]byte) ([20]byte, error) {
	handshakeData := buildHandshake(infoHash, peerId)
	_, err := conn.Write(handshakeData)
	if err != nil {
		return [20]byte{}, err
	}

	resBuffer := make([]byte, 68)
	_, err = io.ReadFull(conn, resBuffer)
	if err != nil {
		return [20]byte{}, err
	}

	pstrLen := int(resBuffer[0])
	if pstrLen != 19 {
		return [20]byte{}, fmt.Errorf("invalid protocol string length: %d", pstrLen)
	}

	var resInfoHash [20]byte
	copy(resInfoHash[:], resBuffer[28:48])
	if resInfoHash != infoHash {
		return [20]byte{}, fmt.Errorf("info hash mismatch")
	}

	var receivedPeerId [20]byte
	copy(receivedPeerId[:], resBuffer[48:68])

	return receivedPeerId, nil
}

func buildHandshake(infoHash [20]byte, peerId [20]byte) []byte {
	buffer := make([]byte, 68)

	pstr := "BitTorrent protocol"

	buffer[0] = byte(len(pstr))
	offset := 1

	offset += copy(buffer[offset:], pstr)

	//Reserved (currently 0)
	offset += 8

	offset += copy(buffer[offset:], infoHash[:])

	offset += copy(buffer[offset:], peerId[:])

	return buffer
}
