package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

var (
	ErrInvalidClientRequest = errors.New("invalid ClientRequest")
	ErrInvalidRequestType   = errors.New("invalid RequestType")
)

type ClientRequest interface {
	encodeRequest(io.Writer) error
	decodeRequest(io.Reader) error
}

func EncodeClientRequest(w io.Writer, request ClientRequest) error {
	err := encodeRequestType(w, request)
	if err != nil {
		return fmt.Errorf("EncodeClientRequest: %w", err)
	}

	request.encodeRequest(w)
	if err != nil {
		return fmt.Errorf("EncodeClientRequest: %w", err)
	}

	return nil
}

func DecodeClientRequest(r io.Reader) (ClientRequest, error) {
	request, err := decodeRequestType(r)
	if err != nil {
		return nil, fmt.Errorf("DecodeClientRequest: %w", err)
	}

	err = request.decodeRequest(r)
	if err != nil {
		return nil, fmt.Errorf("DecodeClientRequest: %w", err)
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

func encodeRequestType(w io.Writer, request ClientRequest) error {
	var requestType RequestType
	switch request.(type) {
	case *ConnectRequest:
		requestType = Connect
	case *DisconnectRequest:
		requestType = Disconnect
	case *ListRoomsRequest:
		requestType = ListRooms
	case *ListUsersRequest:
		requestType = ListUsers
	case *MessageRoomRequest:
		requestType = MessageRoom
	case *MessageUserRequest:
		requestType = MessageUser
	case *CreateRoomRequest:
		requestType = CreateRoom
	case *JoinRoomRequest:
		requestType = JoinRoom
	case *LeaveRoomRequest:
		requestType = LeaveRoom
	default:
		return ErrInvalidClientRequest
	}

	err := binary.Write(w, byteOrder, requestType)
	if err != nil {
		return fmt.Errorf("encode RequestType: %w", err)
	}

	return nil
}

func decodeRequestType(r io.Reader) (ClientRequest, error) {
	var requestType RequestType
	err := binary.Read(r, byteOrder, &requestType)
	if err != nil {
		return nil, fmt.Errorf("decode RequestType: %w", err)
	}

	switch requestType {
	case Connect:
		return new(ConnectRequest), nil
	case Disconnect:
		return new(DisconnectRequest), nil
	case ListRooms:
		return new(ListRoomsRequest), nil
	case ListUsers:
		return new(ListUsersRequest), nil
	case MessageRoom:
		return new(MessageRoomRequest), nil
	case MessageUser:
		return new(MessageUserRequest), nil
	case CreateRoom:
		return new(CreateRoomRequest), nil
	case JoinRoom:
		return new(JoinRoomRequest), nil
	case LeaveRoom:
		return new(LeaveRoomRequest), nil
	default:
		return nil, ErrInvalidRequestType
	}
}

type ConnectRequest struct {
	Version uint32
	Name    string
}

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

func (*DisconnectRequest) encodeRequest(w io.Writer) error { return nil }

func (*DisconnectRequest) decodeRequest(r io.Reader) error { return nil }

type CreateRoomRequest struct {
	Room string
}

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

func (*ListRoomsRequest) encodeRequest(w io.Writer) error { return nil }

func (*ListRoomsRequest) decodeRequest(r io.Reader) error { return nil }

type ListUsersRequest struct {
	Room string
}

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
