package slp

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/ldryt/mcpulse/pulser"
)

const (
	MaxReadBytes     int64         = 4096
	ReadWriteTimeout time.Duration = 20 * time.Second
)

func HandleConnection(conn net.Conn) {
	defer func() {
		log.Printf("Connection closed")
		conn.Close()
	}()

	_ = conn.SetDeadline(time.Now().Add(ReadWriteTimeout))

	lim := io.LimitReader(conn, MaxReadBytes)
	r := bufio.NewReader(lim)

	w := io.Writer(conn)

	log.Printf("Connection established")

	hd, err := handleHandshake(r)
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
		handleStatus(r, w)
	case 2:
		handleLogin(r, w)
	}
}

func handleStatus(r io.Reader, w io.Writer) {
	err := handleStatusRequest(r)
	if err != nil {
		log.Printf("Error handling status request: %v", err)
		return
	}
	log.Printf("Received status request")

	err = sendStatusResponse(w)
	if err != nil {
		log.Printf("Error sending status response: %v", err)
		return
	}
	log.Printf("Sent status response")

	payload, err := handlePingRequest(r)
	if err != nil {
		log.Printf("Error handling ping request: %v", err)
		return
	}
	log.Printf("Received ping request")

	err = sendPongResponse(w, payload)
	if err != nil {
		log.Printf("Error sending ping response: %v", err)
		return
	}
	log.Printf("Sent ping response")
}

func handleLogin(r io.Reader, w io.Writer) {
	player, err := handleLoginStart(r)
	if err != nil {
		log.Printf("Error handling login start request: %v", err)
		return
	}
	log.Printf(
		"Received login request: [Username: %v] [UUID: %v]",
		player.Name,
		toUUID(player.UUID.MSB, player.UUID.LSB),
	)

	pulser.AddStartRequest(toUUID(player.UUID.MSB, player.UUID.LSB))

	err = sendDisconnect(w)
	if err != nil {
		log.Printf("Error sending disconnect: %v", err)
		return
	}
	log.Printf("Sent login disconnect")
}

func toUUID(a, b uint64) string {
	var bytes [16]byte
	binary.BigEndian.PutUint64(bytes[:8], uint64(a))
	binary.BigEndian.PutUint64(bytes[8:], uint64(b))
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%12x",
		bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:16])
}
