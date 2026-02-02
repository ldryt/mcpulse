package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/ldryt/mcpulse/slp"
)

type StatusResponse struct {
	Players struct {
		Online int `json:"online"`
	} `json:"players"`
}

func main() {
	host := flag.String("host", "localhost", "Target host")
	port := flag.Int("port", 25565, "Target port")
	timeout := flag.Duration("timeout", 2*time.Second, "Maximum connection timeout duration")
	flag.Parse()

	count, err := probe(*host, *port, *timeout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error probing %s:%d: %v\n", *host, *port, err)
		fmt.Println("-1")
		os.Exit(1)
	}

	fmt.Println(count)
}

func probe(host string, port int, timeout time.Duration) (int, error) {
	address := net.JoinHostPort(host, strconv.Itoa(port))

	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return -1, err
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(timeout))

	reader := bufio.NewReader(conn)

	err = slp.SendHandshake(conn, slp.HandshakeData{
		ProtocolVersion: -1,
		ServerAddress:   host,
		ServerPort:      uint16(port),
		NextState:       1,
	})
	if err != nil {
		return -1, fmt.Errorf("handshake failed: %w", err)
	}

	// Status Request
	err = slp.SendPacket(conn, slp.Packet{ID: 0x00})
	if err != nil {
		return -1, fmt.Errorf("status request failed: %w", err)
	}

	// Handle Status Response
	p, err := slp.ReadPacket(reader)
	if err != nil {
		return -1, fmt.Errorf("read response failed: %w", err)
	}

	if p.ID != 0x00 {
		return -1, fmt.Errorf("unexpected packet ID: %x", p.ID)
	}

	jsonStr, err := slp.ReadString(&p.Data)
	if err != nil {
		return -1, fmt.Errorf("failed to read json string: %w", err)
	}

	var status StatusResponse
	if err := json.Unmarshal([]byte(jsonStr), &status); err != nil {
		return -1, fmt.Errorf("invalid json: %w", err)
	}

	return status.Players.Online, nil
}
