package pulser

import (
	"log"
	"net"
)

func HandleConnection(conn net.Conn) {
	defer func() {
		log.Printf("Connection closed with %v", conn.RemoteAddr())
		conn.Close()
	}()

	log.Printf("Connection established with %v", conn.RemoteAddr())

	go pulse(conn, 3)
	go readUpdates(conn)

	select {}
}
