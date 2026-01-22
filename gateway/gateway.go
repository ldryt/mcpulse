package gateway

import (
	"io"
	"log"
	"net"
	"time"

	"github.com/ldryt/mcpulse/config"
	"github.com/ldryt/mcpulse/slp"
)

const (
	BackendDialTimeout time.Duration = 3 * time.Second
)

type Gateway struct {
	Config *config.Config
}

type ConnectionInfo struct {
	bi               BackendInfo
	clientConn       net.Conn
	hd               slp.HandshakeData
	username         string
	clientReader     io.Reader
	uuidMSB, uuidLSB uint64
}

type BackendInfo struct {
	destAddr string
	ready    bool
}

func New(cfg *config.Config) *Gateway {
	return &Gateway{Config: cfg}
}

func (gw *Gateway) proxyConnection(c ConnectionInfo) {
	log.Printf("[%s] Dialing backend %s...", c.username, c.bi.destAddr)
	backendConn, err := net.DialTimeout("tcp", c.bi.destAddr, BackendDialTimeout)
	if err != nil {
		log.Printf("Backend unreachable: %v", err)
		return
	}
	defer backendConn.Close()

	if err := slp.SendHandshake(backendConn, c.hd); err != nil {
		log.Printf("Error sending handshake: %v", err)
		return
	}

	if err := slp.SendLoginStart(backendConn, c.username, c.uuidMSB, c.uuidLSB); err != nil {
		log.Printf("Error sending login start: %v", err)
		return
	}

	log.Printf("[%s] Proxying...", c.username)

	c.clientConn.SetDeadline(time.Time{})

	done := make(chan struct{})
	go func() {
		defer close(done)
		io.Copy(c.clientConn, backendConn)
	}()

	io.Copy(backendConn, io.MultiReader(c.clientReader, c.clientConn))

	<-done
}
