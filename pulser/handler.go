package pulser

import (
	"bufio"
	"io"
	"log"
	"net"
	"time"
)

const (
	MaxReadBytes     int64         = 4096
	ReadWriteTimeout time.Duration = 20 * time.Second
)

func HandleConnection(conn net.Conn) {
	defer func() {
		log.Printf("Connection closed with %v", conn.RemoteAddr())
		conn.Close()
	}()

	_ = conn.SetDeadline(time.Now().Add(ReadWriteTimeout))

	lim := io.LimitReader(conn, MaxReadBytes)
	r := bufio.NewReader(lim)

	w := bufio.NewWriter(conn)

	log.Printf("Connection established with %v", conn.RemoteAddr())

	go pulse(w, 3)
	go readUpdates(r)

	select {}
}
