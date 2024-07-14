package slp

import (
	"bytes"
	"io"
)

// https://wiki.vg/Protocol#Packet_format

type Packet struct {
	ID   int32
	Data bytes.Buffer
}

func readPacket(r io.Reader) (p Packet, err error) {
	var PacketLength int32

	PacketLength, err = readVarInt(r)
	if err != nil {
		return Packet{}, err
	}

	p.ID, err = readVarInt(r)
	if err != nil {
		return Packet{}, err
	}

	_, err = io.CopyN(&p.Data, r, int64(PacketLength-int32(sizeVarInt(p.ID))))
	if err != nil {
		return Packet{}, err
	}

	return p, nil
}

func sendPacket(w io.Writer, p Packet) (err error) {
	var PacketLength int = sizeVarInt(p.ID) + p.Data.Len()

	err = writeVarInt(w, int32(PacketLength))
	if err != nil {
		return err
	}

	err = writeVarInt(w, p.ID)
	if err != nil {
		return err
	}

	_, err = p.Data.WriteTo(w)
	if err != nil {
		return err
	}

	return nil
}
