package slp

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
)

func HandleConnection(conn net.Conn) {
	defer func() {
		log.Printf("Connection %v closed", conn.RemoteAddr())
		conn.Close()
	}()

	log.Printf("Connection %v established", conn.RemoteAddr())

	hd, err := handleHandshake(conn)
	if err != nil {
		log.Printf("Error handshaking with %v: %v", conn.RemoteAddr(), err)
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
		handleStatus(conn)
	case 2:
		handleLogin(conn)
	}
}

func handleStatus(conn net.Conn) {
	err := handleStatusRequest(conn)
	if err != nil {
		log.Printf("Error handling status request from %v: %v", conn.RemoteAddr(), err)
		return
	}
	log.Printf("Received status request from %v", conn.RemoteAddr())

	err = sendStatusResponse(conn)
	if err != nil {
		log.Printf("Error sending status response to %v: %v", conn.RemoteAddr(), err)
		return
	}
	log.Printf("Sent status response to %v", conn.RemoteAddr())

	payload, err := handlePingRequest(conn)
	if err != nil {
		log.Printf("Error handling ping request from %v: %v", conn.RemoteAddr(), err)
		return
	}
	log.Printf("Received ping request from %v", conn.RemoteAddr())

	err = sendPongResponse(conn, payload)
	if err != nil {
		log.Printf("Error sending ping response to %v: %v", conn.RemoteAddr(), err)
		return
	}
	log.Printf("Sent ping response to %v", conn.RemoteAddr())
}

func handleLogin(conn net.Conn) {
	player, err := handleLoginStart(conn)
	if err != nil {
		log.Printf("Error handling login start request from %v: %v", conn.RemoteAddr(), err)
		return
	}
	log.Printf(
		"Received login request from %v: [Username: %v] [UUID: %v]",
		conn.RemoteAddr(),
		player.Name,
		toUUID(player.UUID.MSB, player.UUID.LSB),
	)

	err = sendDisconnect(conn)
	if err != nil {
		log.Printf("Error sending disconnect to %v: %v", conn.RemoteAddr(), err)
		return
	}
	log.Printf("Sent login disconnect to %v", conn.RemoteAddr())
}

func toUUID(a, b uint64) string {
	var bytes [16]byte
	binary.BigEndian.PutUint64(bytes[:8], uint64(a))
	binary.BigEndian.PutUint64(bytes[8:], uint64(b))
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%12x",
		bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:16])
}
