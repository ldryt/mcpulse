package slp

import (
	"errors"
	"io"
)

// https://wiki.vg/Protocol#Type:VarInt

const (
	segmentBits int32 = 0x7F
	continueBit int32 = 0x80
)

func readVarInt(r PacketReader) (value int32, err error) {
	var position int

	for {
		currentByte, err := r.ReadByte()
		if err != nil {
			return 0, err
		}

		value |= int32(currentByte) & segmentBits << position

		if (int32(currentByte) & continueBit) == 0 {
			break
		}

		position += 7

		if position >= 32 {
			return 0, errors.New("varint too big")
		}
	}

	return value, nil
}

func writeVarInt(w io.Writer, value int32) (err error) {
	for {
		if (value & ^segmentBits) == 0 {
			_, err = w.Write([]byte{byte(value)})
			return err
		}

		_, err = w.Write([]byte{byte((value & segmentBits) | continueBit)})
		if err != nil {
			return err
		}

		value = int32(uint32(value) >> 7)
	}
}

func sizeVarInt(value int32) (size int) {
	for {
		size++
		value >>= 7
		if value == 0 {
			break
		}
	}

	return size
}
