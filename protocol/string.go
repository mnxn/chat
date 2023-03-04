package protocol

import (
	"encoding/binary"
	"fmt"
	"io"
)

var byteOrder = binary.BigEndian

func encodeString(w io.Writer, s string) error {
	bytes := []byte(s)

	err := binary.Write(w, byteOrder, int32(len(bytes)))
	if err != nil {
		return fmt.Errorf("encode StringData: %w", err)
	}

	err = binary.Write(w, byteOrder, bytes)
	if err != nil {
		return fmt.Errorf("encode StringData: %w", err)
	}

	return nil
}

func decodeString(r io.Reader, s *string) error {
	var length uint32
	err := binary.Read(r, byteOrder, &length)
	if err != nil {
		return fmt.Errorf("decode StringData.length: %w", err)
	}

	bytes := make([]byte, length)
	err = binary.Read(r, byteOrder, &bytes)
	if err != nil {
		return fmt.Errorf("decode StringData.data: %w", err)
	}

	*s = string(bytes)
	return nil
}
