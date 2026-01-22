package gateway

import (
	"bufio"
	"io"
	"net"
	"time"

	"github.com/ldryt/mcpulse/slp"
)

const (
	MaxHandshakeSize int64         = 4096
	HandshakeTimeout time.Duration = 20 * time.Second
)

func (gw *Gateway) HandleConnection(conn net.Conn) {
	s := NewSession(conn)
	s.Log("Connection established with %s.", s.UUID.String())

	defer func() {
		conn.Close()
		s.Log("Connection closed cleanly.")
	}()

	_ = conn.SetDeadline(time.Now().Add(HandshakeTimeout))
	r := bufio.NewReader(io.LimitReader(conn, MaxHandshakeSize))
	w := io.Writer(conn)

	hd, err := slp.HandleHandshake(r)
	if err != nil {
		s.Warn("while handling Handshake: %v", err)
		return
	}

	s.Log(
		"[Protocol: %d] [Next State: %d] [Target: %s:%d]",
		hd.ProtocolVersion,
		hd.NextState,
		hd.ServerAddress,
		hd.ServerPort,
	)

	switch hd.NextState {
	case 1:
		gw.handleStatus(s, r, w)
	case 2:
		gw.handleLogin(s, conn, r, w, hd)
	default:
		s.Error("while parsing NextState %d", hd.NextState)
	}
}

func (gw *Gateway) handleStatus(s *Session, r io.Reader, w io.Writer) {
	if err := slp.HandleStatusRequest(r); err != nil {
		s.Error("while handling Status: %v", err)
		return
	}

	if err := slp.SendStatusResponse(w); err != nil {
		s.Error("while sending Status: %v", err)
		return
	}

	payload, err := slp.HandlePingRequest(r)
	if err != nil {
		s.Error("while handling Ping: %v", err)
		return
	}

	err = slp.SendPongResponse(w, payload)
	if err != nil {
		s.Error("while sending Ping: %v", err)
		return
	}
}

func (gw *Gateway) handleLogin(s *Session, conn net.Conn, r io.Reader, w io.Writer, hd slp.HandshakeData) {
	if u, err := slp.HandleLoginStart(r); err != nil {
		s.Error("while reading LoginStart: %v", err)
		return
	} else {
		s.User = u
		s.Log("[Username: %s] [UUID: %s]", s.User.Name, s.User.UUID.Repr.String())
	}

	backend := BackendInfo{destAddr: "127.0.0.1:25565", ready: true}
	if !backend.ready {
		slp.SendDisconnect(w, "§eTon serveur démarre...\n§7Reviens dans 20 secondes !")
		return
	}

	gw.proxyConnection(
		ConnectionInfo{
			session:      s,
			bi:           backend,
			clientConn:   conn,
			clientReader: r,
			hd:           hd,
		},
	)
}
