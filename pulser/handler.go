package pulser

import (
	"bufio"
	"io"
	"log"
	"net"
	"sync"
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

	wg := new(sync.WaitGroup)

	_ = conn.SetDeadline(time.Now().Add(ReadWriteTimeout))

	lim := io.LimitReader(conn, MaxReadBytes)
	r := bufio.NewReader(lim)

	w := io.Writer(conn)

	log.Printf("Connection established with %v", conn.RemoteAddr())

	wg.Add(1)
	go pulse(w, 3, wg)

	wg.Add(1)
	go readUpdates(r, wg)

	wg.Wait()
}
