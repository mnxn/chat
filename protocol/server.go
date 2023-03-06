package protocol

import (
	"errors"
	"fmt"
	"io"
)

var (
	ErrInvalidResponseType = errors.New("invalid ResponseType value")
	ErrInvalidErrorType    = errors.New("invalid ErrorType value")
)

type ServerResponse interface {
	ResponseType() ResponseType
	encodeResponse(io.Writer) error
	decodeResponse(io.Reader) error
}

func EncodeServerResponse(w io.Writer, response ServerResponse) error {
	err := encodeInt(w, response.ResponseType())
	if err != nil {
		return fmt.Errorf("encode ServerResponse.Type: %w", err)
	}

	err = response.encodeResponse(w)
	if err != nil {
		return fmt.Errorf("encode ServerResponse: %w", err)
	}

	return nil
}

func DecodeServerResponse(r io.Reader) (ServerResponse, error) {
	var responseType ResponseType
	err := decodeInt(r, &responseType)
	if err != nil {
		return nil, fmt.Errorf("decode ServerResponse.Type: %w", err)
	}

	var response ServerResponse
	switch responseType {
	case Error:
		response = new(ErrorResponse)
	case FatalError:
		response = new(FatalErrorResponse)
	case RoomList:
		response = new(RoomListResponse)
	case UserList:
		response = new(UserListResponse)
	case RoomMessage:
		response = new(RoomMessageResponse)
	case UserMessage:
		response = new(UserMessageResponse)
	default:
		return nil, fmt.Errorf("decode ServerResponse.Type(%d): %w", responseType, ErrInvalidResponseType)
	}

	err = response.decodeResponse(r)
	if err != nil {
		return nil, fmt.Errorf("decode ServerResponse: %w", err)
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

func (e ErrorType) GoString() string {
	switch e {
	case Disconnection:
		return "Disconnection"
	case InternalError:
		return "InternalError"
	case MalformedRequest:
		return "MalformedRequest"
	case UnsupportedVersion:
		return "UnsupportedVersion"
	case MissingRoom:
		return "MissingRoom"
	case MissingUser:
		return "MissingUser"
	case ExistingRoom:
		return "ExistingRoom"
	case ExistingUser:
		return "ExistingRoom"
	case InvalidRoom:
		return "InvalidRoom"
	case InvalidUser:
		return "InvalidUser"
	case InvalidText:
		return "InvalidText"
	default:
		return fmt.Sprintf("ErrorType(%d)", e)
	}
}

func (e ErrorType) String() string { return e.GoString() }

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
		return fmt.Errorf("encode ErrorType(%d): %w", e, ErrInvalidErrorType)
	}

	err := encodeInt(w, e)
	if err != nil {
		return fmt.Errorf("encode ErrorType(%d): %w", e, err)
	}

	return nil
}

func decodeErrorType(r io.Reader, e *ErrorType) error {
	err := decodeInt(r, e)
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
		return fmt.Errorf("decode ErrorType(0x%08X): %w", uint32(*e), ErrInvalidErrorType)
	}

	return nil
}

type ErrorResponse struct {
	Error ErrorType
	Info  string
}

func (*ErrorResponse) ResponseType() ResponseType { return Error }

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

func (*FatalErrorResponse) ResponseType() ResponseType { return FatalError }

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

func (*RoomListResponse) ResponseType() ResponseType { return RoomList }

func (rl *RoomListResponse) encodeResponse(w io.Writer) error {
	count := uint32(len(rl.Rooms))
	err := encodeInt(w, count)
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
	err := decodeInt(r, &rl.Count)
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

func (*UserListResponse) ResponseType() ResponseType { return UserList }

func (ul *UserListResponse) encodeResponse(w io.Writer) error {
	err := encodeString(w, ul.Room)
	if err != nil {
		return fmt.Errorf("encode UserListResponse.Room: %w", err)
	}

	count := uint32(len(ul.Users))
	err = encodeInt(w, count)
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

	err = decodeInt(r, &ul.Count)
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

func (*RoomMessageResponse) ResponseType() ResponseType { return RoomMessage }

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

func (*UserMessageResponse) ResponseType() ResponseType { return UserMessage }

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
