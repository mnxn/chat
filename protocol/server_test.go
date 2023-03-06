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

var serverResponseTests = []struct {
	ServerResponse
	bytes []byte
}{
	{
		&ErrorResponse{
			Error: UnsupportedVersion,
			Info:  "info",
		},
		[]byte{
			0, 0, 0, 1, // Error
			0, 0, 0, 4, // UnsupportedVersion
			0, 0, 0, 4, // uint32(4)
			105, 110, 102, 111, // "info"
		},
	},

	{
		&FatalErrorResponse{
			Error: InternalError,
			Info:  "fatal",
		},
		[]byte{
			0, 0, 0, 2, // FatalError
			0, 0, 0, 2, // InternalError
			0, 0, 0, 5, // uint32(5)
			102, 97, 116, 97, 108, // "fatal"
		},
	},

	{
		&RoomListResponse{
			Count: 0,
			Rooms: []string{},
		},
		[]byte{
			0, 0, 0, 3, // RoomList
			0, 0, 0, 0, // uint32(0)
		},
	},
	{
		&RoomListResponse{
			Count: 3,
			Rooms: []string{
				"A",
				"BB",
				"CCC",
			},
		},
		[]byte{
			0, 0, 0, 3, // RoomList

			0, 0, 0, 3, // uint32(3)

			0, 0, 0, 1, // uint32(1)
			65, // "A"

			0, 0, 0, 2, // uint32(2)
			66, 66, // "BB"

			0, 0, 0, 3, // uint32(3)
			67, 67, 67, // "CCC"
		},
	},

	{
		&UserListResponse{
			Room:  "",
			Count: 0,
			Users: []string{},
		},
		[]byte{
			0, 0, 0, 4, // UserList
			0, 0, 0, 0, // uint32(0)
			0, 0, 0, 0, // uint32(0)
		},
	},
	{
		&UserListResponse{
			Room:  "main",
			Count: 2,
			Users: []string{
				"1",
				"2",
			},
		},
		[]byte{
			0, 0, 0, 4, // UserList

			0, 0, 0, 4, // uint32(4)
			109, 97, 105, 110, // "main"

			0, 0, 0, 2, // uint32(2)

			0, 0, 0, 1, // uint32(1)
			49, // "1"

			0, 0, 0, 1, // uint32(1)
			50, // "2"
		},
	},

	{
		&RoomMessageResponse{
			Room:   "room",
			Sender: "sender",
			Text:   "text",
		},
		[]byte{
			0, 0, 0, 5, // RoomMessage

			0, 0, 0, 4, // uint32(4)
			114, 111, 111, 109, // "room"

			0, 0, 0, 6, // uint32(6)
			115, 101, 110, 100, 101, 114, // "sender"

			0, 0, 0, 4, // uint32(4)
			116, 101, 120, 116, // "text"
		},
	},

	{
		&UserMessageResponse{
			Sender: "SENDER",
			Text:   "TEXT",
		},
		[]byte{
			0, 0, 0, 6, // UserMessage

			0, 0, 0, 6, // uint32(6)
			83, 69, 78, 68, 69, 82, // "SENDER"

			0, 0, 0, 4, // uint32(4)
			84, 69, 88, 84, // "TEXT"
		},
	},
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

func TestEncodeServerResponse(t *testing.T) {
	for _, test := range serverResponseTests {
		var buf bytes.Buffer
		err := EncodeServerResponse(&buf, test.ServerResponse)
		if !generic.TestError(t, "EncodeServerResponse", test.ServerResponse, nil, err) {
			continue
		}
		actual := buf.Bytes()

		if !generic.TestEqual(t, "EncodeServerResponse", test.ServerResponse, test.bytes, actual) {
			continue
		}
	}
}

func TestDecodeServerResponse(t *testing.T) {
	for _, test := range serverResponseTests {
		actual, err := DecodeServerResponse(bytes.NewReader(test.bytes))
		if !generic.TestError(t, "DecodeServerResponse", test.bytes, nil, err) {
			continue
		}

		if !generic.TestEqual(t, "DecodeServerResponse", test.bytes, test.ServerResponse, actual) {
			continue
		}
	}
}
