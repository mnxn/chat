package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

var (
	ErrInvalidServerResponse = errors.New("invalid ServerResponse")
	ErrInvalidResponseType   = errors.New("invalid ResponseType")
	ErrInvalidErrorType      = errors.New("invalid ErrorType")
)

type ServerResponse interface {
	encodeResponse(io.Writer) error
	decodeResponse(io.Reader) error
}

func EncodeServerResponse(w io.Writer, response ServerResponse) error {
	err := encodeResponseType(w, response)
	if err != nil {
		return fmt.Errorf("EncodeServerResponse: %w", err)
	}

	response.encodeResponse(w)
	if err != nil {
		return fmt.Errorf("EncodeServerResponse: %w", err)
	}

	return nil
}

func DecodeServerResponse(r io.Reader) (ServerResponse, error) {
	response, err := decodeResponseType(r)
	if err != nil {
		return nil, fmt.Errorf("DecodeServerResponse: %w", err)
	}

	err = response.decodeResponse(r)
	if err != nil {
		return nil, fmt.Errorf("DecodeServerResponse: %w", err)
	}

	return response, nil
}

type ResponseType uint32

const (
	Error ResponseType = 1 + iota
	FatalError
	RoomList
	UserList
	RoomMessage
	UserMessage
)

func encodeResponseType(w io.Writer, response ServerResponse) error {
	var responseType ResponseType
	switch response.(type) {
	case *ErrorResponse:
		responseType = Error
	case *FatalErrorResponse:
		responseType = FatalError
	case *RoomListResponse:
		responseType = RoomList
	case *UserListResponse:
		responseType = UserList
	case *RoomMessageResponse:
		responseType = RoomMessage
	case *UserMessageResponse:
		responseType = UserMessage
	default:
		return ErrInvalidServerResponse
	}

	err := binary.Write(w, byteOrder, responseType)
	if err != nil {
		return fmt.Errorf("encode ResponseType: %w", err)
	}

	return nil
}

func decodeResponseType(r io.Reader) (ServerResponse, error) {
	var responseType ResponseType
	err := binary.Read(r, byteOrder, &responseType)
	if err != nil {
		return nil, fmt.Errorf("decode ResponseType: %w", err)
	}

	switch responseType {
	case Error:
		return new(ErrorResponse), nil
	case FatalError:
		return new(FatalErrorResponse), nil
	case RoomList:
		return new(RoomListResponse), nil
	case UserList:
		return new(UserListResponse), nil
	case RoomMessage:
		return new(RoomMessageResponse), nil
	case UserMessage:
		return new(UserMessageResponse), nil
	default:
		return nil, ErrInvalidResponseType
	}
}

type ErrorType uint32

const (
	Disconnection ErrorType = 1 + iota
	InternalError
	MalformedRequest
	UnsupportedVersion
	MissingRoom
	MissingUser
	ExistingRoom
	ExistingUser
	InvalidRoom
	InvalidUser
	InvalidText
)

func encodeErrorType(w io.Writer, e ErrorType) error {
	switch e {
	case Disconnection,
		InternalError,
		MalformedRequest,
		UnsupportedVersion,
		MissingRoom, MissingUser,
		ExistingRoom, ExistingUser,
		InvalidRoom, InvalidUser, InvalidText:
		break
	default:
		return ErrInvalidErrorType
	}

	err := binary.Write(w, byteOrder, e)
	if err != nil {
		return fmt.Errorf("encode ErrorType: %w", err)
	}

	return nil
}

func decodeErrorType(r io.Reader, e *ErrorType) error {
	err := binary.Read(r, byteOrder, e)
	if err != nil {
		return fmt.Errorf("decode ErrorType: %w", err)
	}

	switch *e {
	case Disconnection,
		InternalError,
		MalformedRequest,
		UnsupportedVersion,
		MissingRoom, MissingUser,
		ExistingRoom, ExistingUser,
		InvalidRoom, InvalidUser, InvalidText:
		break
	default:
		return ErrInvalidErrorType
	}

	return nil
}

type ErrorResponse struct {
	Error ErrorType
	Info  string
}

func (e *ErrorResponse) encodeResponse(w io.Writer) error {
	err := encodeErrorType(w, e.Error)
	if err != nil {
		return fmt.Errorf("encode ErrorResponse.Error: %w", err)
	}

	err = encodeString(w, e.Info)
	if err != nil {
		return fmt.Errorf("encode ErrorResponse.Info: %w", err)
	}

	return nil
}

func (e *ErrorResponse) decodeResponse(r io.Reader) error {
	err := decodeErrorType(r, &e.Error)
	if err != nil {
		return fmt.Errorf("decode ErrorResponse.Error: %w", err)
	}

	err = decodeString(r, &e.Info)
	if err != nil {
		return fmt.Errorf("decode ErrorResponse.Info: %w", err)
	}

	return nil
}

type FatalErrorResponse struct {
	Error ErrorType
	Info  string
}

func (fe *FatalErrorResponse) encodeResponse(w io.Writer) error {
	err := encodeErrorType(w, fe.Error)
	if err != nil {
		return fmt.Errorf("encode FatalErrorResponse.Error: %w", err)
	}

	err = encodeString(w, fe.Info)
	if err != nil {
		return fmt.Errorf("encode FatalErrorResponse.Info: %w", err)
	}

	return nil
}

func (fe *FatalErrorResponse) decodeResponse(r io.Reader) error {
	err := decodeErrorType(r, &fe.Error)
	if err != nil {
		return fmt.Errorf("decode FatalErrorResponse.Error: %w", err)
	}

	err = decodeString(r, &fe.Info)
	if err != nil {
		return fmt.Errorf("decode FatalErrorResponse.Info: %w", err)
	}

	return nil
}

type RoomListResponse struct {
	Count uint32
	Rooms []string
}

func (rl *RoomListResponse) encodeResponse(w io.Writer) error {
	count := uint32(len(rl.Rooms))
	err := binary.Write(w, byteOrder, count)
	if err != nil {
		return fmt.Errorf("encode RoomListResponse.Count: %w", err)
	}

	for i, room := range rl.Rooms {
		err = encodeString(w, room)
		if err != nil {
			return fmt.Errorf("encode RoomListResponse.Rooms[%d]: %w", i, err)
		}
	}

	return nil
}

func (rl *RoomListResponse) decodeResponse(r io.Reader) error {
	err := binary.Read(r, byteOrder, rl.Count)
	if err != nil {
		return fmt.Errorf("decode RoomListResponse.Count: %w", err)
	}
	rl.Rooms = make([]string, rl.Count)

	for i := uint32(0); i < rl.Count; i++ {
		err = decodeString(r, &rl.Rooms[i])
		if err != nil {
			return fmt.Errorf("decode RoomListResponse.Rooms[%d]: %w", i, err)
		}
	}

	return nil
}

type UserListResponse struct {
	Room  string
	Count uint32
	Users []string
}

func (ul *UserListResponse) encodeResponse(w io.Writer) error {
	err := encodeString(w, ul.Room)
	if err != nil {
		return fmt.Errorf("encode UserListResponse.Room: %w", err)
	}

	count := uint32(len(ul.Users))
	err = binary.Write(w, byteOrder, count)
	if err != nil {
		return fmt.Errorf("encode UserListResponse.Count: %w", err)
	}

	for i, user := range ul.Users {
		err = encodeString(w, user)
		if err != nil {
			return fmt.Errorf("encode UserListResponse.Users[%d]: %w", i, err)
		}
	}

	return nil
}

func (ul *UserListResponse) decodeResponse(r io.Reader) error {
	err := decodeString(r, &ul.Room)
	if err != nil {
		return fmt.Errorf("decode UserListResponse.Room: %w", err)
	}

	err = binary.Read(r, byteOrder, ul.Count)
	if err != nil {
		return fmt.Errorf("decode UserListResponse.Count: %w", err)
	}
	ul.Users = make([]string, ul.Count)

	for i := uint32(0); i < ul.Count; i++ {
		err = decodeString(r, &ul.Users[i])
		if err != nil {
			return fmt.Errorf("decode UserListResponse.Users[%d]: %w", i, err)
		}
	}

	return nil
}

type RoomMessageResponse struct {
	Room   string
	Sender string
	Text   string
}

func (rm *RoomMessageResponse) encodeResponse(w io.Writer) error {
	err := encodeString(w, rm.Room)
	if err != nil {
		return fmt.Errorf("encode RoomMessageResponse.Room: %w", err)
	}

	err = encodeString(w, rm.Sender)
	if err != nil {
		return fmt.Errorf("encode RoomMessageResponse.Sender: %w", err)
	}

	err = encodeString(w, rm.Text)
	if err != nil {
		return fmt.Errorf("encode RoomMessageResponse.Text: %w", err)
	}

	return nil
}

func (rm *RoomMessageResponse) decodeResponse(r io.Reader) error {
	err := decodeString(r, &rm.Room)
	if err != nil {
		return fmt.Errorf("decode RoomMessageResponse.Room: %w", err)
	}

	err = decodeString(r, &rm.Sender)
	if err != nil {
		return fmt.Errorf("decode RoomMessageResponse.Sender: %w", err)
	}

	err = decodeString(r, &rm.Text)
	if err != nil {
		return fmt.Errorf("decode RoomMessageResponse.Text: %w", err)
	}

	return nil
}

type UserMessageResponse struct {
	Sender string
	Text   string
}

func (um *UserMessageResponse) encodeResponse(w io.Writer) error {
	err := encodeString(w, um.Sender)
	if err != nil {
		return fmt.Errorf("encode UserMessageResponse.Sender: %w", err)
	}

	err = encodeString(w, um.Text)
	if err != nil {
		return fmt.Errorf("encode UserMessageResponse.Text: %w", err)
	}

	return nil
}

func (um *UserMessageResponse) decodeResponse(r io.Reader) error {
	err := decodeString(r, &um.Sender)
	if err != nil {
		return fmt.Errorf("decode UserMessageResponse.Sender: %w", err)
	}

	err = decodeString(r, &um.Text)
	if err != nil {
		return fmt.Errorf("decode UserMessageResponse.Text: %w", err)
	}

	return nil
}
