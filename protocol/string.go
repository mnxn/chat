package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"unicode/utf8"
)

var ErrInvalidUtf8String = errors.New("invalid UTF-8 string")

var byteOrder = binary.BigEndian

func encodeString(w io.Writer, s string) error {
	bytes := []byte(s)

	err := binary.Write(w, byteOrder, uint32(len(bytes)))
	if err != nil {
		return fmt.Errorf("encode StringData.Length: %w", err)
	}

	if !utf8.ValidString(s) {
		return fmt.Errorf("encode StringData.Data: %w", ErrInvalidUtf8String)
	}

	err = binary.Write(w, byteOrder, bytes)
	if err != nil {
		return fmt.Errorf("encode StringData.Data: %w", err)
	}

	return nil
}

func decodeString(r io.Reader, s *string) error {
	var length uint32
	err := binary.Read(r, byteOrder, &length)
	if err != nil {
		return fmt.Errorf("decode StringData.Length: %w", err)
	}

	bytes := make([]byte, length)
	err = binary.Read(r, byteOrder, &bytes)
	if err != nil {
		return fmt.Errorf("decode StringData.Data: %w", err)
	}

	if !utf8.Valid(bytes) {
		return fmt.Errorf("decode StringData.Data: %w", ErrInvalidUtf8String)
	}

	*s = string(bytes)
	return nil
}
