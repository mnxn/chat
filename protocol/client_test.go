package protocol

import (
	"bytes"
	"testing"

	"github.com/mnxn/chat/generic"
)

var clientRequestTests = []struct {
	ClientRequest
	bytes []byte
}{
	{
		&ConnectRequest{
			Version: 1,
			Name:    "me",
		},
		[]byte{
			0, 0, 0, 1, // Connect
			0, 0, 0, 1, // uint32(1)

			0, 0, 0, 2, // uint32(2)
			109, 101, // "me"
		},
	},

	{
		&DisconnectRequest{},
		[]byte{
			0, 0, 0, 2, // Disconnect
		},
	},

	{
		&ListRoomsRequest{},
		[]byte{
			0, 0, 0, 3, // ListRooms
		},
	},

	{
		&ListUsersRequest{
			Room: "",
		},
		[]byte{
			0, 0, 0, 4, // ListUsers
			0, 0, 0, 0, // uint32(0)
		},
	},
	{
		&ListUsersRequest{
			Room: "general",
		},
		[]byte{
			0, 0, 0, 4, // ListUsers
			0, 0, 0, 7, // uint32(7)
			103, 101, 110, 101, 114, 97, 108, // "general"
		},
	},

	{
		&MessageRoomRequest{
			Room: "room",
			Text: "hello",
		},
		[]byte{
			0, 0, 0, 5, // MessageRoom

			0, 0, 0, 4, // uint32(4)
			114, 111, 111, 109, // "room"

			0, 0, 0, 5, // uint32(5)
			104, 101, 108, 108, 111, // "hello"
		},
	},

	{
		&MessageUserRequest{
			User: "other",
			Text: "hi",
		},
		[]byte{
			0, 0, 0, 6, // MessageUser

			0, 0, 0, 5, // uint32(5)
			111, 116, 104, 101, 114, // "other"

			0, 0, 0, 2, // uint32(2)
			104, 105, // "hi"
		},
	},

	{
		&CreateRoomRequest{
			Room: "create",
		},
		[]byte{
			0, 0, 0, 7, // CreateRoom

			0, 0, 0, 6, // uint32(6)
			99, 114, 101, 97, 116, 101, // "create"
		},
	},

	{
		&JoinRoomRequest{
			Room: "join",
		},
		[]byte{
			0, 0, 0, 8, // JoinRoom

			0, 0, 0, 4, // uint32(4)
			106, 111, 105, 110, // "join"
		},
	},

	{
		&LeaveRoomRequest{
			Room: "leave",
		},
		[]byte{
			0, 0, 0, 9, // LeaveRoom

			0, 0, 0, 5, // uint32(5)
			108, 101, 97, 118, 101, // "leave"
		},
	},
}

func TestEncodeClientRequest(t *testing.T) {
	for _, test := range clientRequestTests {
		var buf bytes.Buffer
		err := EncodeClientRequest(&buf, test.ClientRequest)
		if !generic.TestError(t, "EncodeClientRequest", test.ClientRequest, nil, err) {
			continue
		}
		actual := buf.Bytes()

		if !generic.TestEqual(t, "EncodeClientRequest", test.ClientRequest, test.bytes, actual) {
			continue
		}
	}
}

func TestDecodeClientRequest(t *testing.T) {
	for _, test := range clientRequestTests {
		actual, err := DecodeClientRequest(bytes.NewReader(test.bytes))
		if !generic.TestError(t, "DecodeClientRequest", test.bytes, nil, err) {
			continue
		}

		if !generic.TestEqual(t, "DecodeClientRequest", test.bytes, test.ClientRequest, actual) {
			continue
		}
	}
}
