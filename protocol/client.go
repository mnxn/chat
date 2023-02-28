package protocol

type RequestType uint32

const (
	Connect RequestType = iota
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

type DisconnectRequest struct{}

type CreateRoomRequest struct {
	Room string
}

type JoinRoomRequest struct {
	Room string
}

type LeaveRoomRequest struct {
	Room string
}

type ListRoomsRequest struct{}

type ListUsersRequest struct {
	Room string
}

type MessageRoomRequest struct {
	Room string
	Text string
}

type MessageUserRequest struct {
	User string
	Text string
}
