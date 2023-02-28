package protocol

import (
	"io"
)

type ServerResponse interface {
	EncodeResponse(io.Writer) error
	DecodeResponse(io.Reader) error
}

type ResponseType uint32

const (
	Error ResponseType = iota
	FatalError
	RoomList
	UserList
	RoomMessage
	UserMessage
)

type ErrorType uint32

const (
	Disconnection ErrorType = iota
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

type ErrorResponse struct {
	Error ErrorType
	Info  string
}

type FatalErrorResponse struct {
	Error ErrorType
	Info  string
}

type RoomListResponse struct {
	Count uint32
	Rooms []string
}

type UserListResponse struct {
	Room  string
	Count uint32
	Users []string
}

type RoomMessageResponse struct {
	Room   string
	Sender string
	Text   string
}

type UserMessageResponse struct {
	Sender string
	Text   string
}
