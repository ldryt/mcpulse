package gateway

import (
	"encoding/base64"
	"fmt"
	"log"
	"net"

	"github.com/google/uuid"
	"github.com/ldryt/mcpulse/slp"
)

type Session struct {
	UUID       uuid.UUID
	RemoteAddr string
	User       slp.PlayerData
}

func NewSession(conn net.Conn) *Session {
	return &Session{
		UUID:       uuid.New(),
		RemoteAddr: conn.RemoteAddr().String(),
	}
}

func (s *Session) Log(format string, v ...any) {
	prefix := fmt.Sprintf("[%s] [%s] ", s.UUID.String()[:8], s.RemoteAddr)
	if s.User.Name != "" {
		prefix += fmt.Sprintf("[%s/%s] ", base64.RawURLEncoding.EncodeToString([]byte(s.User.Name)), s.User.UUID.Repr.String()[:4])
	}
	log.Printf(prefix+format, v...)
}

func (s *Session) Error(format string, v ...any) {
	s.Log("ERROR: "+format, v...)
}

func (s *Session) Warn(format string, v ...any) {
	s.Log("WARN: "+format, v...)
}
