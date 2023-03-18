package server

import (
	"strings"
	"sync"
	"time"

	"github.com/mnxn/chat/protocol"
)

func (cu connectedUser) Connect(request *protocol.ConnectRequest) {
	cu.outgoing <- &protocol.FatalErrorResponse{
		Error: protocol.AlreadyConnected,
		Info:  "",
	}
}

func (cu connectedUser) Disconnect(request *protocol.DisconnectRequest) {
	cu.conn.SetReadDeadline(time.Now())
}

func (cu connectedUser) ListRooms(request *protocol.ListRoomsRequest) {
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

func (cu connectedUser) ListUsers(request *protocol.ListUsersRequest) {
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

func (cu connectedUser) MessageRoom(request *protocol.MessageRoomRequest) {
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
		if user.name != cu.name {
			user.outgoing <- &protocol.RoomMessageResponse{
				Room:   request.Room,
				Sender: cu.name,
				Text:   request.Text,
			}
		}
	}
	room.usersMutex.RUnlock()
}

func (cu connectedUser) MessageUser(request *protocol.MessageUserRequest) {
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
		Sender: cu.name,
		Text:   request.Text,
	}
}

func (cu connectedUser) CreateRoom(request *protocol.CreateRoomRequest) {
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

func (cu connectedUser) JoinRoom(request *protocol.JoinRoomRequest) {
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
	room.users[cu.name] = cu.user
	room.usersMutex.Unlock()
}

func (cu connectedUser) LeaveRoom(request *protocol.LeaveRoomRequest) {
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
	delete(room.users, cu.name)
	if len(room.users) == 0 {
		delete(cu.server.rooms, request.Room)
	}
	room.usersMutex.Unlock()

	cu.server.roomsMutex.Unlock()
}
