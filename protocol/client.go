package protocol

import (
	"encoding/binary"
	"fmt"
	"io"
)

type ClientRequest interface {
	EncodeRequest(io.Writer) error
	DecodeRequest(io.Reader) error
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

func (c *ConnectRequest) EncodeRequest(w io.Writer) error {
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

func (c *ConnectRequest) DecodeRequest(r io.Reader) error {
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

func (*DisconnectRequest) EncodeRequest(w io.Writer) error { return nil }

func (*DisconnectRequest) DecodeRequest(r io.Reader) error { return nil }

type CreateRoomRequest struct {
	Room string
}

func (cr *CreateRoomRequest) EncodeRequest(w io.Writer) error {
	err := encodeString(w, cr.Room)
	if err != nil {
		return fmt.Errorf("encode CreateRoomRequest.Room: %w", err)
	}

	return nil
}

func (cr *CreateRoomRequest) DecodeRequest(r io.Reader) error {
	err := decodeString(r, &cr.Room)
	if err != nil {
		return fmt.Errorf("decode CreateRoomRequest.Name: %w", err)
	}

	return nil
}

type JoinRoomRequest struct {
	Room string
}

func (jr *JoinRoomRequest) EncodeRequest(w io.Writer) error {
	err := encodeString(w, jr.Room)
	if err != nil {
		return fmt.Errorf("encode JoinRoomRequest.Room: %w", err)
	}

	return nil
}

func (jr *JoinRoomRequest) DecodeRequest(r io.Reader) error {
	err := decodeString(r, &jr.Room)
	if err != nil {
		return fmt.Errorf("decode JoinRoomRequest.Name: %w", err)
	}

	return nil
}

type LeaveRoomRequest struct {
	Room string
}

func (lr *LeaveRoomRequest) EncodeRequest(w io.Writer) error {
	err := encodeString(w, lr.Room)
	if err != nil {
		return fmt.Errorf("encode LeaveRoomRequest.Room: %w", err)
	}

	return nil
}

func (lr *LeaveRoomRequest) DecodeRequest(r io.Reader) error {
	err := decodeString(r, &lr.Room)
	if err != nil {
		return fmt.Errorf("decode LeaveRoomRequest.Name: %w", err)
	}

	return nil
}

type ListRoomsRequest struct{}

func (*ListRoomsRequest) EncodeRequest(w io.Writer) error { return nil }

func (*ListRoomsRequest) DecodeRequest(r io.Reader) error { return nil }

type ListUsersRequest struct {
	Room string
}

func (lu *ListUsersRequest) EncodeRequest(w io.Writer) error {
	err := encodeString(w, lu.Room)
	if err != nil {
		return fmt.Errorf("encode ListUsersRequest.Room: %w", err)
	}

	return nil
}

func (lu *ListUsersRequest) DecodeRequest(r io.Reader) error {
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

func (mr *MessageRoomRequest) EncodeRequest(w io.Writer) error {
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

func (mr *MessageRoomRequest) DecodeRequest(r io.Reader) error {
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

func (mu *MessageUserRequest) EncodeRequest(w io.Writer) error {
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

func (mu *MessageUserRequest) DecodeRequest(r io.Reader) error {
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
