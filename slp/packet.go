package slp

import (
	"bytes"
	"errors"
	"io"
	"math"
)

// https://wiki.vg/Protocol#Packet_format

type Packet struct {
	ID   int32
	Data bytes.Buffer
}

type PacketReader interface {
	io.Reader
	io.ByteReader
}

func ReadPacket(r PacketReader) (p Packet, err error) {
	var PacketLength int32

	PacketLength, err = readVarInt(r)
	if err != nil {
		return Packet{}, err
	}
	if sizeVarInt(PacketLength) > 3 {
		return Packet{}, errors.New("packet length field must not be longer than 3 bytes")
	}

	p.ID, err = readVarInt(r)
	if err != nil {
		return Packet{}, err
	}

	n, err := io.CopyN(&p.Data, r, int64(PacketLength-int32(sizeVarInt(p.ID))))
	if err != nil {
		return Packet{}, err
	}
	if n != int64(PacketLength-int32(sizeVarInt(p.ID))) {
		return Packet{}, errors.New("packet length differs from read length")
	}
	if n+int64(sizeVarInt(PacketLength)+sizeVarInt(p.ID)) >= int64(math.Pow(2, 21)) {
		return Packet{}, errors.New("packets cannot be larger than 2^21 - 1 bytes")
	}

	return p, nil
}

func SendPacket(w io.Writer, p Packet) (err error) {
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
