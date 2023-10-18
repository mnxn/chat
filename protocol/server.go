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

// ServerResponse messages originate in the server before being received by relevant clients.
type ServerResponse interface {
	ResponseType() ResponseType
	Accept(ResponseVisitor)
	encodeResponse(io.Writer) error
	decodeResponse(io.Reader) error
}

func EncodeServerResponse(w io.Writer, response ServerResponse) error {
	err := encodeResponseType(w, response.ResponseType())
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
	err := decodeResponseType(r, &responseType)
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

func (r ResponseType) GoString() string {
	switch r {
	case Error:
		return "Error"
	case FatalError:
		return "FatalError"
	case RoomList:
		return "RoomList"
	case UserList:
		return "UserList"
	case RoomMessage:
		return "RoomMessage"
	case UserMessage:
		return "UserMessage"
	default:
		return fmt.Sprintf("ResponseType(%d)", r)
	}
}

func (r ResponseType) String() string { return r.GoString() }

func encodeResponseType(w io.Writer, typ ResponseType) error {
	switch typ {
	case Error, FatalError,
		RoomList, UserList,
		RoomMessage, UserMessage:
		break
	default:
		return fmt.Errorf("encode ResponseType(%d): %w", typ, ErrInvalidResponseType)
	}

	err := encodeInt(w, typ)
	if err != nil {
		return fmt.Errorf("encode ResponseType(%d): %w", typ, err)
	}

	return nil
}

func decodeResponseType(r io.Reader, typ *ResponseType) error {
	err := decodeInt(r, typ)
	if err != nil {
		return fmt.Errorf("decode ResponseType: %w", err)
	}

	switch *typ {
	case Error, FatalError,
		RoomList, UserList,
		RoomMessage, UserMessage:
		break
	default:
		return fmt.Errorf("decode ResponseType(0x%08X): %w", uint32(*typ), ErrInvalidResponseType)
	}

	return nil
}

type ErrorType uint32

const (
	// The client is attempting to send a request without first sending a ConnectRequest.
	NotConnected ErrorType = 1 + iota

	// The client is already connected and is trying to connect again.
	AlreadyConnected

	// The server had an internal error that prevented it from completing the request.
	InternalError

	// The client sent a request that was not able to be decoded.
	MalformedRequest

	// The client is attempting to connect with a version of the protocol that does not match the server's supported versions.
	//   - This error MUST be sent in a FatalError server message.
	UnsupportedVersion

	// The requested room has not been created on the server or has been removed from the room list.
	MissingRoom

	// The request user has not connected to the server or has already disconnected.
	MissingUser

	// The requested room has already been created. It cannot be created again.
	ExistingRoom

	// The client is attempting to connect with a name that is already in-use on the server.
	//   - This error MUST be sent in a FatalError server message.
	ExistingUser

	// The requested room name does not satisfy the server's room naming requirements.
	//  - The server SHOULD include additional information that explains the room naming requirements.
	InvalidRoom

	// The client is attempting to connect with a name that does not satisfy the server's user naming requirements.
	//   - The server SHOULD include additional information that explains the user naming requirements.
	//   - This error MUST be sent in a FatalError server message.
	InvalidUser

	// The client is attempting to send a chat message with text that does not satisfy the server's text content requirements.
	//   - The server SHOULD include additional information that explains the text content requirements.
	InvalidText
)

func (e ErrorType) GoString() string {
	switch e {
	case NotConnected:
		return "NotConnected"
	case AlreadyConnected:
		return "AlreadyConnected"
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
		return "ExistingUser"
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
	case NotConnected, AlreadyConnected,
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
	case NotConnected, AlreadyConnected,
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

// An ErrorResponse is sent to clients when there is an error performing an operation.
type ErrorResponse struct {
	Error ErrorType // The error code corresponding to the error. See ErrorType.
	Info  string    // Additional information about the cause of the error
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

// A FatalErrorResponse is sent to clients when there is an error performing an operation.
//   - Clients MUST disconnect from the server following this message.
type FatalErrorResponse struct {
	Error ErrorType // The error code corresponding to the error. See ErrorType.
	Info  string    // Additional information about the cause of the error.
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

// A RoomListResponse is sent as a response to clients that ask for the list of rooms.
type RoomListResponse struct {
	User  string   // The user who has joined the rooms. Empty if the response is for the list of rooms in the entire server.
	Count uint32   // The number of rooms in the response.
	Rooms []string // The array of room names.
}

func (*RoomListResponse) ResponseType() ResponseType { return RoomList }

func (rl *RoomListResponse) encodeResponse(w io.Writer) error {
	err := encodeString(w, rl.User)
	if err != nil {
		return fmt.Errorf("encode RoomListResponse.User: %w", err)
	}

	count := uint32(len(rl.Rooms))
	err = encodeInt(w, count)
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
	err := decodeString(r, &rl.User)
	if err != nil {
		return fmt.Errorf("decode RoomListResponse.User: %w", err)
	}

	err = decodeInt(r, &rl.Count)
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

// A UserListResponse is sent as a response to clients that ask for a list of users.
type UserListResponse struct {
	Room  string   // The room the users are located in. Empty if the response is for the list of users in the entire server.
	Count uint32   // The number of users in the room/server.
	Users []string // The array of user names.
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

// A RoomMessageResponse is sent when another user has sent a message in a room that the client user has joined.
type RoomMessageResponse struct {
	Room   string // The name of the room the chat message was sent from.
	Sender string // The name of the user that sent the direct message.
	Text   string // The text content of the chat message.
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

// A UserMessageResponse is sent when another user has sent a direct message to the client user.
type UserMessageResponse struct {
	Sender string // The name of the user that sent the direct message.
	Text   string // The text content of the chat message.
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
