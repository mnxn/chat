package server

import (
	"strings"
	"sync"
	"time"

	"github.com/mnxn/chat/protocol"
)

func (cu *connectedUser) requireConnected() bool {
	if !cu.connected() {
		cu.outgoing <- &protocol.FatalErrorResponse{
			Error: protocol.NotConnected,
			Info:  "",
		}

		return false
	}

	return true
}

func (cu *connectedUser) sendFatalError(errorType protocol.ErrorType, info string) {
	cu.outgoing <- &protocol.FatalErrorResponse{
		Error: errorType,
		Info:  info,
	}
}

func (cu *connectedUser) Connect(request *protocol.ConnectRequest) {
	if cu.connected() {
		cu.outgoing <- &protocol.FatalErrorResponse{
			Error: protocol.AlreadyConnected,
			Info:  "",
		}
		return
	}

	if request.Version != 1 {
		cu.outgoing <- &protocol.FatalErrorResponse{
			Error: protocol.UnsupportedVersion,
			Info:  "expected version 1",
		}
		return
	}
	if strings.ContainsRune(request.Name, ' ') {
		cu.outgoing <- &protocol.FatalErrorResponse{
			Error: protocol.InvalidUser,
			Info:  "username cannot contain spaces",
		}
		return
	}

	cu.atomicName.Store(&request.Name)

	cu.server.usersMutex.Lock()
	if _, ok := cu.server.users[request.Name]; ok {
		cu.outgoing <- &protocol.FatalErrorResponse{
			Error: protocol.ExistingUser,
			Info:  "username already exists",
		}
		cu.server.usersMutex.Unlock()
		return
	}
	cu.server.users[request.Name] = cu.user
	cu.server.usersMutex.Unlock()

	cu.server.general.usersMutex.Lock()
	cu.server.general.users[request.Name] = cu.user
	cu.server.general.usersMutex.Unlock()
}

func (cu *connectedUser) Disconnect(request *protocol.DisconnectRequest) {
	if !cu.requireConnected() {
		return
	}

	cu.conn.SetReadDeadline(time.Now())
}

func (cu *connectedUser) ListRooms(request *protocol.ListRoomsRequest) {
	if !cu.requireConnected() {
		return
	}

	cu.server.roomsMutex.RLock()
	rooms := make([]string, 0, len(cu.server.rooms))
	for room := range cu.server.rooms {
		rooms = append(rooms, room)
	}
	cu.server.roomsMutex.RUnlock()

	cu.outgoing <- &protocol.RoomListResponse{
		Count: uint32(len(rooms)),
		Rooms: rooms,
	}
}

func (cu *connectedUser) ListUsers(request *protocol.ListUsersRequest) {
	if !cu.requireConnected() {
		return
	}

	var users []string
	if request.Room != "" {
		cu.server.roomsMutex.RLock()
		room, ok := cu.server.rooms[request.Room]
		cu.server.roomsMutex.RUnlock()
		if !ok {
			cu.outgoing <- &protocol.ErrorResponse{
				Error: protocol.MissingRoom,
				Info:  "",
			}
			return
		}

		room.usersMutex.RLock()
		users = make([]string, 0, len(room.users))
		for user := range room.users {
			users = append(users, user)
		}
		room.usersMutex.RUnlock()
	} else {
		cu.server.usersMutex.RLock()
		users = make([]string, 0, len(cu.server.users))
		for user := range cu.server.users {
			users = append(users, user)
		}
		cu.server.usersMutex.RUnlock()
	}

	cu.outgoing <- &protocol.UserListResponse{
		Count: uint32(len(users)),
		Room:  request.Room,
		Users: users,
	}
}

func (cu *connectedUser) MessageRoom(request *protocol.MessageRoomRequest) {
	if !cu.requireConnected() {
		return
	}

	cu.server.roomsMutex.RLock()
	room, ok := cu.server.rooms[request.Room]
	cu.server.roomsMutex.RUnlock()
	if !ok {
		cu.outgoing <- &protocol.ErrorResponse{
			Error: protocol.MissingRoom,
			Info:  "",
		}
		return
	}

	room.usersMutex.RLock()
	for _, user := range room.users {
		if user.name() != cu.name() {
			user.outgoing <- &protocol.RoomMessageResponse{
				Room:   request.Room,
				Sender: cu.name(),
				Text:   request.Text,
			}
		}
	}
	room.usersMutex.RUnlock()
}

func (cu *connectedUser) MessageUser(request *protocol.MessageUserRequest) {
	if !cu.requireConnected() {
		return
	}

	cu.server.usersMutex.RLock()
	user, ok := cu.server.users[request.User]
	cu.server.usersMutex.RUnlock()
	if !ok {
		cu.outgoing <- &protocol.ErrorResponse{
			Error: protocol.MissingUser,
			Info:  "",
		}
		return
	}

	user.outgoing <- &protocol.UserMessageResponse{
		Sender: cu.name(),
		Text:   request.Text,
	}
}

func (cu *connectedUser) CreateRoom(request *protocol.CreateRoomRequest) {
	if !cu.requireConnected() {
		return
	}

	if strings.ContainsRune(request.Room, ' ') {
		cu.outgoing <- &protocol.ErrorResponse{
			Error: protocol.InvalidRoom,
			Info:  "",
		}
		return
	}

	cu.server.roomsMutex.Lock()
	if _, ok := cu.server.rooms[request.Room]; ok {
		cu.outgoing <- &protocol.ErrorResponse{
			Error: protocol.ExistingRoom,
			Info:  "",
		}
		cu.server.roomsMutex.Unlock()
		return
	}
	cu.server.rooms[request.Room] = &room{
		users:      make(map[string]*user),
		usersMutex: sync.RWMutex{},
	}
	cu.server.roomsMutex.Unlock()
}

func (cu *connectedUser) JoinRoom(request *protocol.JoinRoomRequest) {
	if !cu.requireConnected() {
		return
	}

	cu.server.roomsMutex.RLock()
	room, ok := cu.server.rooms[request.Room]
	cu.server.roomsMutex.RUnlock()
	if !ok {
		cu.outgoing <- &protocol.ErrorResponse{
			Error: protocol.MissingRoom,
			Info:  "",
		}
		return
	}

	room.usersMutex.Lock()
	room.users[cu.name()] = cu.user
	room.usersMutex.Unlock()
}

func (cu *connectedUser) LeaveRoom(request *protocol.LeaveRoomRequest) {
	if !cu.requireConnected() {
		return
	}

	cu.server.roomsMutex.Lock()
	room, ok := cu.server.rooms[request.Room]

	if !ok {
		cu.outgoing <- &protocol.ErrorResponse{
			Error: protocol.MissingRoom,
			Info:  "",
		}
		cu.server.roomsMutex.Unlock()
		return
	}

	room.usersMutex.Lock()
	delete(room.users, cu.name())
	if len(room.users) == 0 {
		delete(cu.server.rooms, request.Room)
	}
	room.usersMutex.Unlock()

	cu.server.roomsMutex.Unlock()
}
