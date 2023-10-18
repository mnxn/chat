package protocol

import (
	"errors"
	"fmt"
	"io"
)

var ErrInvalidRequestType = errors.New("invalid RequestType value")

// ClientRequest messages originate in the clients before being received by the server and responded to.
type ClientRequest interface {
	RequestType() RequestType
	Accept(RequestVisitor)
	encodeRequest(io.Writer) error
	decodeRequest(io.Reader) error
}

func EncodeClientRequest(w io.Writer, request ClientRequest) error {
	err := encodeRequestType(w, request.RequestType())
	if err != nil {
		return fmt.Errorf("encode ClientRequest.Type: %w", err)
	}

	err = request.encodeRequest(w)
	if err != nil {
		return fmt.Errorf("encode ClientRequest: %w", err)
	}

	return nil
}

func DecodeClientRequest(r io.Reader) (ClientRequest, error) {
	var requestType RequestType
	err := decodeRequestType(r, &requestType)
	if err != nil {
		return nil, fmt.Errorf("decode ClientRequest.Type: %w", err)
	}

	var request ClientRequest
	switch requestType {
	case Keepalive:
		request = new(KeepaliveRequest)
	case Connect:
		request = new(ConnectRequest)
	case Disconnect:
		request = new(DisconnectRequest)
	case ListRooms:
		request = new(ListRoomsRequest)
	case ListUsers:
		request = new(ListUsersRequest)
	case MessageRoom:
		request = new(MessageRoomRequest)
	case MessageUser:
		request = new(MessageUserRequest)
	case CreateRoom:
		request = new(CreateRoomRequest)
	case JoinRoom:
		request = new(JoinRoomRequest)
	case LeaveRoom:
		request = new(LeaveRoomRequest)
	}

	err = request.decodeRequest(r)
	if err != nil {
		return nil, fmt.Errorf("decode ClientRequest: %w", err)
	}

	return request, nil
}

type RequestType uint32

const (
	Keepalive RequestType = iota
	Connect
	Disconnect
	ListRooms
	ListUsers
	MessageRoom
	MessageUser
	CreateRoom
	JoinRoom
	LeaveRoom
)

func (r RequestType) GoString() string {
	switch r {
	case Keepalive:
		return "Keepalive"
	case Connect:
		return "Connect"
	case Disconnect:
		return "Disconnect"
	case ListRooms:
		return "ListRooms"
	case ListUsers:
		return "ListUsers"
	case MessageRoom:
		return "MessageRoom"
	case MessageUser:
		return "MessageUser"
	case CreateRoom:
		return "CreateRoom"
	case JoinRoom:
		return "JoinRoom"
	case LeaveRoom:
		return "LeaveRoom"
	default:
		return fmt.Sprintf("RequestType(%d)", r)
	}
}

func (r RequestType) String() string { return r.GoString() }

func encodeRequestType(w io.Writer, typ RequestType) error {
	switch typ {
	case Keepalive,
		Connect, Disconnect,
		ListRooms, ListUsers,
		MessageRoom, MessageUser,
		CreateRoom, JoinRoom, LeaveRoom:
		break
	default:
		return fmt.Errorf("encode RequestType(%d): %w", typ, ErrInvalidRequestType)
	}

	err := encodeInt(w, typ)
	if err != nil {
		return fmt.Errorf("encode RequestType(%d): %w", typ, err)
	}

	return nil
}

func decodeRequestType(r io.Reader, typ *RequestType) error {
	err := decodeInt(r, typ)
	if err != nil {
		return fmt.Errorf("decode RequestType: %w", err)
	}

	switch *typ {
	case
		Keepalive,
		Connect, Disconnect,
		ListRooms, ListUsers,
		MessageRoom, MessageUser,
		CreateRoom, JoinRoom, LeaveRoom:
		break
	default:
		return fmt.Errorf("decode RequestType(0x%08X): %w", uint32(*typ), ErrInvalidRequestType)
	}

	return nil
}

// KeepaliveRequest messages MUST be sent to the server at least every 30 seconds to prevent the TCP connection from closing.
type KeepaliveRequest struct{}

func (*KeepaliveRequest) RequestType() RequestType { return Keepalive }

func (*KeepaliveRequest) encodeRequest(w io.Writer) error { return nil }

func (*KeepaliveRequest) decodeRequest(r io.Reader) error { return nil }

// This ConnectRequest MUST be sent to a server at the beginning of a connection.
//   - The server MAY respond with an error message.
//   - The server MUST update the user list if the client connected successfully.
type ConnectRequest struct {
	Version uint32 // The version of the protocol that the client is using.
	Name    string // The display name the user wishes to connect with.
}

func (*ConnectRequest) RequestType() RequestType { return Connect }

func (c *ConnectRequest) encodeRequest(w io.Writer) error {
	err := encodeInt(w, c.Version)
	if err != nil {
		return fmt.Errorf("encode ConnectRequest.Version: %w", err)
	}

	err = encodeString(w, c.Name)
	if err != nil {
		return fmt.Errorf("encode ConnectRequest.Name: %w", err)
	}

	return nil
}

func (c *ConnectRequest) decodeRequest(r io.Reader) error {
	err := decodeInt(r, &c.Version)
	if err != nil {
		return fmt.Errorf("decode ConnectRequest.Version: %w", err)
	}

	err = decodeString(r, &c.Name)
	if err != nil {
		return fmt.Errorf("decode ConnectRequest.Name: %w", err)
	}

	return nil
}

// The DisconnectRequest MUST be sent before a client intentionally disconnects from the server.
//   - The server MUST remove the user from the active user list, notify other users,
//     and close the TCP connection immediately upon receiving this message.
//   - If the server notices that the client closed the TCP connection without sending this message,
//     it MUST also remove the user from the active user list and notify the other users.
type DisconnectRequest struct{}

func (*DisconnectRequest) RequestType() RequestType { return Disconnect }

func (*DisconnectRequest) encodeRequest(w io.Writer) error { return nil }

func (*DisconnectRequest) decodeRequest(r io.Reader) error { return nil }

// A CreateRoomRequest should be sent by the client to create a new room.
//   - The server MAY respond with an error message.
//   - The server MUST update the room list if the room was created successfully.
//   - The server MUST NOT add the user to the newly created room until the client joins with a JoinRoomRequest.
type CreateRoomRequest struct {
	Room string // Desired name of the new room.
}

func (*CreateRoomRequest) RequestType() RequestType { return CreateRoom }

func (cr *CreateRoomRequest) encodeRequest(w io.Writer) error {
	err := encodeString(w, cr.Room)
	if err != nil {
		return fmt.Errorf("encode CreateRoomRequest.Room: %w", err)
	}

	return nil
}

func (cr *CreateRoomRequest) decodeRequest(r io.Reader) error {
	err := decodeString(r, &cr.Room)
	if err != nil {
		return fmt.Errorf("decode CreateRoomRequest.Name: %w", err)
	}

	return nil
}

// A CreateRoomRequest should be sent by the client to join a room.
//   - The server MAY respond with an error message.
//   - The server MUST update the room's list of users if the room was joined successfully.
type JoinRoomRequest struct {
	Room string // Desired name of the room to join.
}

func (*JoinRoomRequest) RequestType() RequestType { return JoinRoom }

func (jr *JoinRoomRequest) encodeRequest(w io.Writer) error {
	err := encodeString(w, jr.Room)
	if err != nil {
		return fmt.Errorf("encode JoinRoomRequest.Room: %w", err)
	}

	return nil
}

func (jr *JoinRoomRequest) decodeRequest(r io.Reader) error {
	err := decodeString(r, &jr.Room)
	if err != nil {
		return fmt.Errorf("decode JoinRoomRequest.Name: %w", err)
	}

	return nil
}

// A LeaveRoomRequest should be sent by the client to leave a room.
//   - The server MAY respond with an error message.
//   - The server MUST update the room's list of users if the room was left successfully.
//   - The server MUST remove a room from the room list if there are no users remaining.
type LeaveRoomRequest struct {
	Room string // Desired name of the room to leave.
}

func (*LeaveRoomRequest) RequestType() RequestType { return LeaveRoom }

func (lr *LeaveRoomRequest) encodeRequest(w io.Writer) error {
	err := encodeString(w, lr.Room)
	if err != nil {
		return fmt.Errorf("encode LeaveRoomRequest.Room: %w", err)
	}

	return nil
}

func (lr *LeaveRoomRequest) decodeRequest(r io.Reader) error {
	err := decodeString(r, &lr.Room)
	if err != nil {
		return fmt.Errorf("decode LeaveRoomRequest.Name: %w", err)
	}

	return nil
}

// A ListRoomsRequest should be sent by the client to obtain a list of rooms.
//   - The server MUST respond with an error message or a RoomListResponse.
type ListRoomsRequest struct {
	//  The name of the user to get a list of joined rooms for.
	//  - If the user name is empty, the server MUST respond with a list of rooms for the entire server.
	User string
}

func (*ListRoomsRequest) RequestType() RequestType { return ListRooms }

func (lr *ListRoomsRequest) encodeRequest(w io.Writer) error {
	err := encodeString(w, lr.User)
	if err != nil {
		return fmt.Errorf("encode ListRoomsRequest.User: %w", err)
	}

	return nil
}

func (lr *ListRoomsRequest) decodeRequest(r io.Reader) error {
	err := decodeString(r, &lr.User)
	if err != nil {
		return fmt.Errorf("decode ListRoomsRequest.User: %w", err)
	}

	return nil
}

// A ListUsersRequest should be sent by the client to obtain a list of users in  room or the entire server.
//   - The server must respond with an error message or a RoomListResponse.
type ListUsersRequest struct {
	// The name of the room to get a list of users for.
	//   - If the room name is empty, the server MUST respond with a list of users for the entire server.
	Room string
}

func (*ListUsersRequest) RequestType() RequestType { return ListUsers }

func (lu *ListUsersRequest) encodeRequest(w io.Writer) error {
	err := encodeString(w, lu.Room)
	if err != nil {
		return fmt.Errorf("encode ListUsersRequest.Room: %w", err)
	}

	return nil
}

func (lu *ListUsersRequest) decodeRequest(r io.Reader) error {
	err := decodeString(r, &lu.Room)
	if err != nil {
		return fmt.Errorf("decode ListUsersRequest.Name: %w", err)
	}

	return nil
}

// A MessageRoomRequest should be sent by the client to send a chat message to a room.
//   - The server MAY respond with an error message.
//   - The server MUST forward the message to other users in the room if sending the chat message was successful.
type MessageRoomRequest struct {
	Room string // The name of the room to send the chat message to.
	Text string // The text content of the chat message.
}

func (*MessageRoomRequest) RequestType() RequestType { return MessageRoom }

func (mr *MessageRoomRequest) encodeRequest(w io.Writer) error {
	err := encodeString(w, mr.Room)
	if err != nil {
		return fmt.Errorf("encode MessageRoomRequest.Room: %w", err)
	}

	err = encodeString(w, mr.Text)
	if err != nil {
		return fmt.Errorf("encode MessageRoomRequest.Text: %w", err)
	}

	return nil
}

func (mr *MessageRoomRequest) decodeRequest(r io.Reader) error {
	err := decodeString(r, &mr.Room)
	if err != nil {
		return fmt.Errorf("decode MessageRoomRequest.Name: %w", err)
	}

	err = decodeString(r, &mr.Text)
	if err != nil {
		return fmt.Errorf("decode MessageRoomRequest.Text: %w", err)
	}

	return nil
}

// A MessageUserRequest should be sent by the client to send a chat message to another user.
//   - The server MAY respond with an error message.
//   - The server MUST forward the message to the other user if sending the chat message was successful.
type MessageUserRequest struct {
	User string
	Text string
}

func (*MessageUserRequest) RequestType() RequestType { return MessageUser }

func (mu *MessageUserRequest) encodeRequest(w io.Writer) error {
	err := encodeString(w, mu.User)
	if err != nil {
		return fmt.Errorf("encode MessageUserRequest.User: %w", err)
	}

	err = encodeString(w, mu.Text)
	if err != nil {
		return fmt.Errorf("encode MessageUserRequest.Text: %w", err)
	}

	return nil
}

func (mu *MessageUserRequest) decodeRequest(r io.Reader) error {
	err := decodeString(r, &mu.User)
	if err != nil {
		return fmt.Errorf("decode MessageUserRequest.User: %w", err)
	}

	err = decodeString(r, &mu.Text)
	if err != nil {
		return fmt.Errorf("decode MessageUserRequest.Text: %w", err)
	}

	return nil
}
