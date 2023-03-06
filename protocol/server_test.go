package protocol

import (
	"bytes"
	"io"
	"testing"

	"github.com/mnxn/chat/generic"
)

var errorTypeTests = []struct {
	ErrorType
	bytes []byte
}{
	{Disconnection, []byte{
		0, 0, 0, 1, // uint32(1)
	}},
	{InternalError, []byte{
		0, 0, 0, 2, // uint32(2)
	}},
	{MalformedRequest, []byte{
		0, 0, 0, 3, // uint32(3)
	}},
	{UnsupportedVersion, []byte{
		0, 0, 0, 4, // uint32(4)
	}},
	{MissingRoom, []byte{
		0, 0, 0, 5, // uint32(5)
	}},
	{MissingUser, []byte{
		0, 0, 0, 6, // uint32(6)
	}},
	{ExistingRoom, []byte{
		0, 0, 0, 7, // uint32(7)
	}},
	{ExistingUser, []byte{
		0, 0, 0, 8, // uint32(8)
	}},
	{InvalidRoom, []byte{
		0, 0, 0, 9, // uint32(9)
	}},
	{InvalidUser, []byte{
		0, 0, 0, 10, // uint32(10)
	}},
	{InvalidText, []byte{
		0, 0, 0, 11, // uint32(11)
	}},
}

func TestEncodeErrorType(t *testing.T) {
	for _, test := range errorTypeTests {
		var buf bytes.Buffer
		err := encodeErrorType(&buf, test.ErrorType)
		if !generic.TestError(t, "encodeErrorType", test.ErrorType, nil, err) {
			continue
		}
		actual := buf.Bytes()

		if !generic.TestEqual(t, "encodeErrorType", test.ErrorType, test.bytes, actual) {
			continue
		}
	}

	invalidValue := ErrorType(1000)
	err := encodeErrorType(io.Discard, invalidValue)
	generic.TestError(t, "encodeErrorType", invalidValue, ErrInvalidErrorType, err)
}

func TestDecodeErrorType(t *testing.T) {
	for _, test := range errorTypeTests {
		var actual ErrorType
		err := decodeErrorType(bytes.NewReader(test.bytes), &actual)
		if !generic.TestError(t, "decodeErrorType", test.bytes, nil, err) {
			continue
		}

		if !generic.TestEqual(t, "decodeErrorType", test.bytes, test.ErrorType, actual) {
			continue
		}
	}

	invalidBytes := []byte{0xFF, 0xFF, 0xFF, 0xFF}
	err := decodeErrorType(bytes.NewBuffer(invalidBytes), new(ErrorType))
	generic.TestError(t, "decodeErrorType", invalidBytes, ErrInvalidErrorType, err)
}
