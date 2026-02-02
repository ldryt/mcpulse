package slp

import "io"

// https://wiki.vg/Protocol#Type:String

func ReadString(r PacketReader) (value string, err error) {
	var len int32
	var buf []byte

	len, err = readVarInt(r)
	if err != nil {
		return "", err
	}

	buf = make([]byte, len)

	_, err = r.Read(buf)
	if err != nil {
		return "", err
	}

	return string(buf), nil
}

func WriteString(w io.Writer, value string) (err error) {
	var len int = len(value)

	err = writeVarInt(w, int32(len))
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(value))
	if err != nil {
		return err
	}

	return nil
}
