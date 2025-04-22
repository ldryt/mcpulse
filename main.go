package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/ldryt/mcpulse/config"
	"github.com/ldryt/mcpulse/slp"
)

const (
	MaxReadBytes     int64         = 4096
	ReadWriteTimeout time.Duration = 20 * time.Second
)

func main() {
	log.Println("Loading configuration...")
	err := config.Load()
	if err != nil {
		log.Fatalln("Error loading configuration:", err)
	}

	address := config.Get().ListenAddress
	listener, err := net.Listen("tcp4", address)
	if err != nil {
		log.Fatalf("Error creating listener on %v: %v", address, err)
	}
	defer listener.Close()

	log.Printf("Listening on %v", listener.Addr())

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection on %v: %v", listener.Addr(), err)
			continue
		}

		go HandleConnection(conn)
	}
}

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
		handleStatus(r, w)
	case 2:
		handleLogin(r, w)
	}
}

func handleStatus(r io.Reader, w io.Writer) {
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

func handleLogin(r io.Reader, w io.Writer) {
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

	err = slp.SendDisconnect(w, "Login successful.")
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
