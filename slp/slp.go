package slp

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"

	"github.com/ldryt/mcpulse/config"
)

// https://wiki.vg/Server_List_Ping

type HandshakeData struct {
	ProtocolVersion int32
	ServerAddress   string
	ServerPort      uint16
	NextState       int32
}

type PlayerData struct {
	Name string
	UUID struct {
		MSB uint64
		LSB uint64
	}
}

type StatusResponse struct {
	Version struct {
		Name     string `json:"name"`
		Protocol int    `json:"protocol"`
	} `json:"version"`
	Description struct {
		Text string `json:"text"`
	} `json:"description"`
	Favicon string `json:"favicon"`
}

func HandleHandshake(r io.Reader) (h HandshakeData, err error) {
	var p Packet

	p, err = readPacket(r)
	if err != nil {
		return HandshakeData{}, err
	}
	if p.ID != 0 {
		return HandshakeData{}, errors.New("not an handshake packet")
	}

	h.ProtocolVersion, err = readVarInt(&p.Data)
	if err != nil {
		return HandshakeData{}, err
	}

	h.ServerAddress, err = readString(&p.Data)
	if err != nil {
		return HandshakeData{}, err
	}

	err = binary.Read(&p.Data, binary.BigEndian, &h.ServerPort)
	if err != nil {
		return HandshakeData{}, err
	}

	h.NextState, err = readVarInt(&p.Data)
	if err != nil {
		return HandshakeData{}, err
	}
	if !(h.NextState == 1 || h.NextState == 2) {
		return HandshakeData{}, errors.New("state is neither status or login")
	}

	return h, nil
}

func HandleStatusRequest(r io.Reader) (err error) {
	var p Packet

	p, err = readPacket(r)
	if err != nil {
		return err
	}
	if p.ID != 0 {
		return errors.New("not a status request packet")
	}

	return nil
}

func SendStatusResponse(w io.Writer) (err error) {
	var p Packet
	var sr StatusResponse
	var srMarshalled []byte
	var cfg config.Config = *config.Get()

	p.ID = 0

	sr = StatusResponse{
		Version: struct {
			Name     string `json:"name"`
			Protocol int    `json:"protocol"`
		}{
			Name:     cfg.Version.Name,
			Protocol: cfg.Version.Protocol,
		},
		Description: struct {
			Text string `json:"text"`
		}{
			Text: cfg.Motd,
		},
		Favicon: cfg.FaviconB64,
	}

	srMarshalled, err = json.Marshal(sr)
	if err != nil {
		return err
	}

	err = writeString(&p.Data, string(srMarshalled))
	if err != nil {
		return err
	}

	err = sendPacket(w, p)
	if err != nil {
		return err
	}

	return nil
}

func HandlePingRequest(r io.Reader) (pl int64, err error) {
	var p Packet

	p, err = readPacket(r)
	if err != nil {
		return 0, err
	}
	if p.ID != 1 {
		return 0, errors.New("not a ping request packet")
	}

	err = binary.Read(&p.Data, binary.BigEndian, &pl)
	if err != nil {
		return 0, err
	}

	return pl, nil
}

func SendPongResponse(w io.Writer, pl int64) (err error) {
	var p Packet

	p.ID = 1

	err = binary.Write(&p.Data, binary.BigEndian, pl)
	if err != nil {
		return err
	}

	err = sendPacket(w, p)
	if err != nil {
		return err
	}

	return nil
}

func HandleLoginStart(r io.Reader) (pr PlayerData, err error) {
	var p Packet

	p, err = readPacket(r)
	if err != nil {
		return PlayerData{}, err
	}
	if p.ID != 0 {
		return PlayerData{}, errors.New("not a login start packet")
	}

	pr.Name, err = readString(&p.Data)
	if err != nil {
		return PlayerData{}, err
	}

	err = binary.Read(&p.Data, binary.BigEndian, &pr.UUID.MSB)
	if err != nil {
		return PlayerData{}, err
	}
	err = binary.Read(&p.Data, binary.BigEndian, &pr.UUID.LSB)
	if err != nil {
		return PlayerData{}, err
	}

	return pr, nil
}

func SendDisconnect(w io.Writer, reason string) (err error) {
	var p Packet

	p.ID = 0

	err = writeString(&p.Data, reason)
	if err != nil {
		return err
	}

	err = sendPacket(w, p)
	if err != nil {
		return err
	}

	return nil
}
