package announcer

import (
	"log"
	"net"
)

func HandleConnection(conn net.Conn) {
	defer func() {
		log.Printf("Connection %v closed", conn.RemoteAddr())
		conn.Close()
	}()

	log.Printf("Connection %v established", conn.RemoteAddr())
}
