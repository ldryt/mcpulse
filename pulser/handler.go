package pulser

import (
	"bufio"
	"io"
	"log"
	"net"
)

const MaxReadBytes int64 = 4096

func HandleConnection(conn net.Conn) {
	defer func() {
		log.Printf("Connection closed with %v", conn.RemoteAddr())
		conn.Close()
	}()

	lim := io.LimitReader(conn, MaxReadBytes)
	r := bufio.NewReader(lim)

	w := bufio.NewWriter(conn)

	log.Printf("Connection established with %v", conn.RemoteAddr())

	go pulse(w, 3)
	go readUpdates(r)

	select {}
}
