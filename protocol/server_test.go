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
	{NotConnected, []byte{
		0, 0, 0, 1, // uint32(1)
	}},
	{AlreadyConnected, []byte{
		0, 0, 0, 2, // uint32(2)
	}},
	{InternalError, []byte{
		0, 0, 0, 3, // uint32(3)
	}},
	{MalformedRequest, []byte{
		0, 0, 0, 4, // uint32(4)
	}},
	{UnsupportedVersion, []byte{
		0, 0, 0, 5, // uint32(5)
	}},
	{MissingRoom, []byte{
		0, 0, 0, 6, // uint32(6)
	}},
	{MissingUser, []byte{
		0, 0, 0, 7, // uint32(7)
	}},
	{ExistingRoom, []byte{
		0, 0, 0, 8, // uint32(8)
	}},
	{ExistingUser, []byte{
		0, 0, 0, 9, // uint32(9)
	}},
	{InvalidRoom, []byte{
		0, 0, 0, 10, // uint32(10)
	}},
	{InvalidUser, []byte{
		0, 0, 0, 11, // uint32(11)
	}},
	{InvalidText, []byte{
		0, 0, 0, 12, // uint32(12)
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
			0, 0, 0, 5, // UnsupportedVersion
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
			0, 0, 0, 3, // InternalError
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
	t.Parallel()

	for i := range errorTypeTests {
		test := errorTypeTests[i]
		t.Run("encodeErrorType", func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			err := encodeErrorType(&buf, test.ErrorType)
			if !generic.TestError(t, "encode", test.ErrorType, nil, err) {
				return
			}
			actual := buf.Bytes()

			generic.TestEqual(t, "encode", test.ErrorType, test.bytes, actual)
		})
	}

	t.Run("encodeErrorType", func(t *testing.T) {
		t.Parallel()

		invalidValue := ErrorType(1000)
		err := encodeErrorType(io.Discard, invalidValue)
		generic.TestError(t, "encode", invalidValue, ErrInvalidErrorType, err)
	})
}

func TestDecodeErrorType(t *testing.T) {
	t.Parallel()

	for i := range errorTypeTests {
		test := errorTypeTests[i]
		t.Run("decodeErrorType", func(t *testing.T) {
			t.Parallel()

			var actual ErrorType
			err := decodeErrorType(bytes.NewReader(test.bytes), &actual)
			if !generic.TestError(t, "decode", test.bytes, nil, err) {
				return
			}

			generic.TestEqual(t, "decode", test.bytes, test.ErrorType, actual)
		})
	}

	t.Run("decodeErrorType", func(t *testing.T) {
		t.Parallel()

		invalidBytes := []byte{0xFF, 0xFF, 0xFF, 0xFF}
		err := decodeErrorType(bytes.NewBuffer(invalidBytes), new(ErrorType))
		generic.TestError(t, "decode", invalidBytes, ErrInvalidErrorType, err)
	})
}

func TestEncodeServerResponse(t *testing.T) {
	t.Parallel()

	for i := range serverResponseTests {
		test := serverResponseTests[i]
		t.Run("EncodeServerResponse", func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			err := EncodeServerResponse(&buf, test.ServerResponse)
			if !generic.TestError(t, "encode", test.ServerResponse, nil, err) {
				return
			}
			actual := buf.Bytes()

			generic.TestEqual(t, "encode", test.ServerResponse, test.bytes, actual)
		})
	}
}

func TestDecodeServerResponse(t *testing.T) {
	t.Parallel()

	for i := range serverResponseTests {
		test := serverResponseTests[i]
		t.Run("DecodeServerResponse", func(t *testing.T) {
			t.Parallel()

			actual, err := DecodeServerResponse(bytes.NewReader(test.bytes))
			if !generic.TestError(t, "decode", test.bytes, nil, err) {
				return
			}

			generic.TestEqual(t, "decode", test.bytes, test.ServerResponse, actual)
		})
	}
}

func TestDecodeServerResponseSequential(t *testing.T) {
	t.Parallel()

	var source bytes.Buffer
	expected := make([]ServerResponse, len(serverResponseTests))
	for i, test := range serverResponseTests {
		source.Write(test.bytes)
		expected[i] = test.ServerResponse
	}

	actual := make([]ServerResponse, 0, len(serverResponseTests))
	response, err := DecodeServerResponse(&source)
	for err == nil {
		actual = append(actual, response)
		response, err = DecodeServerResponse(&source)
	}
	if !generic.TestError(t, "sequential", len(actual), io.EOF, err) {
		return
	}

	generic.TestEqual(t, "sequential", len(actual), expected, actual)
}
