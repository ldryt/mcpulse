package gateway

import (
	"io"
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
	session      *Session
	bi           BackendInfo
	clientReader io.Reader
	clientConn   net.Conn
	hd           slp.HandshakeData
}

type BackendInfo struct {
	destAddr string
	ready    bool
}

func New(cfg *config.Config) *Gateway {
	return &Gateway{Config: cfg}
}

func (gw *Gateway) proxyConnection(c ConnectionInfo) {
	s := c.session

	backendConn, err := net.DialTimeout("tcp", c.bi.destAddr, BackendDialTimeout)
	if err != nil {
		s.Error("while reaching backend [%s]: %v", c.bi.destAddr, err)
		return
	}
	defer backendConn.Close()

	if err := slp.SendHandshake(backendConn, c.hd); err != nil {
		s.Error("while sending Handshake to backend [%s]: %v", c.bi.destAddr, err)
		return
	}
	if err := slp.SendLoginStart(backendConn, s.User.Name, s.User.UUID.MSB, s.User.UUID.LSB); err != nil {
		s.Error("while sending LoginStart to backend [%s]: %v", c.bi.destAddr, err)
		return
	}
	s.Log("Successfully authenticated to backend [%s].", c.bi.destAddr)
	c.clientConn.SetDeadline(time.Time{})

	s.Log("Tunnel started.")
	serverClosed := make(chan error, 1)
	clientClosed := make(chan error, 1)
	go func() {
		_, err := io.Copy(c.clientConn, backendConn)
		serverClosed <- err
	}()
	go func() {
		_, err := io.Copy(backendConn, io.MultiReader(c.clientReader, c.clientConn))
		clientClosed <- err
	}()

	select {
	case err := <-clientClosed:
		if err != nil {
			s.Warn("Client connection error: %v", err)
		} else {
			s.Log("Client quitted.")
		}
		backendConn.Close()

	case err := <-serverClosed:
		if err != nil {
			s.Warn("Backend connection error: %v", err)
		} else {
			s.Log("Backend server closed connection.")
		}
		c.clientConn.Close()
	}

	s.Log("Stopped tunnelling.")
}
