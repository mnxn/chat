package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

var ErrInvalidRequestType = errors.New("invalid RequestType")

type ClientRequest interface {
	RequestType() RequestType
	encodeRequest(io.Writer) error
	decodeRequest(io.Reader) error
}

func EncodeClientRequest(w io.Writer, request ClientRequest) error {
	err := binary.Write(w, byteOrder, request.RequestType())
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
	err := binary.Read(r, byteOrder, &requestType)
	if err != nil {
		return nil, fmt.Errorf("decode ClientRequest.Type: %w", err)
	}

	var request ClientRequest
	switch requestType {
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
	default:
		return nil, fmt.Errorf("decode ClientRequest.Type: %w", ErrInvalidRequestType)
	}

	err = request.decodeRequest(r)
	if err != nil {
		return nil, fmt.Errorf("decode ClientRequest: %w", err)
	}

	return request, nil
}

type RequestType uint32

const (
	Connect RequestType = 1 + iota
	Disconnect
	ListRooms
	ListUsers
	MessageRoom
	MessageUser
	CreateRoom
	JoinRoom
	LeaveRoom
)

type ConnectRequest struct {
	Version uint32
	Name    string
}

func (*ConnectRequest) RequestType() RequestType { return Connect }

func (c *ConnectRequest) encodeRequest(w io.Writer) error {
	err := binary.Write(w, byteOrder, c.Version)
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
	err := binary.Read(r, byteOrder, &c.Version)
	if err != nil {
		return fmt.Errorf("decode ConnectRequest.Version: %w", err)
	}

	err = decodeString(r, &c.Name)
	if err != nil {
		return fmt.Errorf("decode ConnectRequest.Name: %w", err)
	}

	return nil
}

type DisconnectRequest struct{}

func (*DisconnectRequest) RequestType() RequestType { return Disconnect }

func (*DisconnectRequest) encodeRequest(w io.Writer) error { return nil }

func (*DisconnectRequest) decodeRequest(r io.Reader) error { return nil }

type CreateRoomRequest struct {
	Room string
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

type JoinRoomRequest struct {
	Room string
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

type LeaveRoomRequest struct {
	Room string
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

type ListRoomsRequest struct{}

func (*ListRoomsRequest) RequestType() RequestType { return ListRooms }

func (*ListRoomsRequest) encodeRequest(w io.Writer) error { return nil }

func (*ListRoomsRequest) decodeRequest(r io.Reader) error { return nil }

type ListUsersRequest struct {
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

type MessageRoomRequest struct {
	Room string
	Text string
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
