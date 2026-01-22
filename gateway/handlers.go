package gateway

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/ldryt/mcpulse/slp"
)

const (
	MaxHandshakeSize int64         = 4096
	HandshakeTimeout time.Duration = 20 * time.Second
)

func (gw *Gateway) HandleConnection(conn net.Conn) {
	defer func() {
		log.Printf("Connection closed")
		conn.Close()
	}()

	_ = conn.SetDeadline(time.Now().Add(HandshakeTimeout))

	lim := io.LimitReader(conn, MaxHandshakeSize)
	r := bufio.NewReader(lim)

	w := io.Writer(conn)

	log.Printf("Connection established")

	hd, err := slp.HandleHandshake(r)
	if err != nil {
		log.Printf("Error handshaking: %v", err)
		return
	}

	log.Printf(
		"Handshaked with %v: [Protocol: %v] [Address: %v] [Port: %v] [Next State: %v]",
		conn.RemoteAddr(),
		hd.ProtocolVersion,
		hd.ServerAddress,
		hd.ServerPort,
		hd.NextState,
	)

	switch hd.NextState {
	case 1:
		gw.handleStatus(r, w)
	case 2:
		gw.handleLogin(conn, r, w, hd)
	}
}

func (gw *Gateway) handleStatus(r io.Reader, w io.Writer) {
	err := slp.HandleStatusRequest(r)
	if err != nil {
		log.Printf("Error handling status request: %v", err)
		return
	}
	log.Printf("Received status request")

	err = slp.SendStatusResponse(w)
	if err != nil {
		log.Printf("Error sending status response: %v", err)
		return
	}
	log.Printf("Sent status response")

	payload, err := slp.HandlePingRequest(r)
	if err != nil {
		log.Printf("Error handling ping request: %v", err)
		return
	}
	log.Printf("Received ping request")

	err = slp.SendPongResponse(w, payload)
	if err != nil {
		log.Printf("Error sending ping response: %v", err)
		return
	}
	log.Printf("Sent ping response")
}

func (gw *Gateway) handleLogin(conn net.Conn, r io.Reader, w io.Writer, hd slp.HandshakeData) {
	player, err := slp.HandleLoginStart(r)
	if err != nil {
		log.Printf("Error handling login start request: %v", err)
		return
	}
	log.Printf(
		"Received login request: [Username: %v] [UUID: %v]",
		player.Name,
		toUUID(player.UUID.MSB, player.UUID.LSB),
	)

	bi := BackendInfo{destAddr: "127.0.0.1:25565", ready: true}

	if !bi.ready {
		slp.SendDisconnect(w, "§eTon serveur démarre...\n§7Reviens dans 20 secondes !")
		return
	}

	gw.proxyConnection(
		ConnectionInfo{
			bi:           bi,
			clientConn:   conn,
			clientReader: r,
			hd:           hd,
			username:     player.Name,
			uuidMSB:      player.UUID.MSB,
			uuidLSB:      player.UUID.LSB,
		},
	)
}
func toUUID(a, b uint64) string {
	var bytes [16]byte
	binary.BigEndian.PutUint64(bytes[:8], uint64(a))
	binary.BigEndian.PutUint64(bytes[8:], uint64(b))
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%12x",
		bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:16])
}
