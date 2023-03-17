package server

import (
	"fmt"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/mnxn/chat/protocol"
)

type user struct {
	name     string
	incoming chan protocol.ClientRequest
	outgoing chan protocol.ServerResponse
	done     chan struct{}
}

type room struct {
	users      map[string]*user
	usersMutex sync.RWMutex
}

type Server struct {
	host string
	port int

	rooms      map[string]*room
	roomsMutex sync.RWMutex

	room
}

func NewServer(host string, port int) *Server {
	return &Server{
		host: host,
		port: port,

		rooms:      make(map[string]*room),
		roomsMutex: sync.RWMutex{},

		room: room{
			users:      make(map[string]*user),
			usersMutex: sync.RWMutex{},
		},
	}
}

func (s *Server) Run() error {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.host, s.port))
	if err != nil {
		return fmt.Errorf("error listening: %w", err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			return fmt.Errorf("error accepting connection: %w", err)
		}

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	request, err := protocol.DecodeClientRequest(conn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "decode error: %s\n", err)
		return
	}
	connect, ok := request.(*protocol.ConnectRequest)
	if !ok {
		err := protocol.EncodeServerResponse(conn, &protocol.FatalErrorResponse{
			Error: protocol.NotConnected,
			Info:  "client not connected",
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "error encoding response: %v", err)
		}
		return
	}
	if connect.Version != 1 {
		err := protocol.EncodeServerResponse(conn, &protocol.FatalErrorResponse{
			Error: protocol.UnsupportedVersion,
			Info:  "expected version 1",
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "error encoding response: %v", err)
		}
		return
	}
	if strings.ContainsRune(connect.Name, ' ') {
		err := protocol.EncodeServerResponse(conn, &protocol.FatalErrorResponse{
			Error: protocol.InvalidUser,
			Info:  "username cannot contain spaces",
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "error encoding response: %v", err)
		}
		return
	}

	user := &user{
		name:     connect.Name,
		incoming: make(chan protocol.ClientRequest),
		outgoing: make(chan protocol.ServerResponse),
		done:     make(chan struct{}),
	}
	s.usersMutex.Lock()
	s.users[user.name] = user
	s.usersMutex.Unlock()

	go func() {
		for {
			request, err := protocol.DecodeClientRequest(conn)
			if err != nil {
				fmt.Fprintf(os.Stderr, "decode error: %s\n", err)
				user.done <- struct{}{}
				return
			}
			user.incoming <- request
		}
	}()

	for {
		select {
		case request := <-user.incoming:
			go s.dispatch(user, request)
		case response := <-user.outgoing:
			err = protocol.EncodeServerResponse(conn, response)
			if err != nil {
				fmt.Fprintf(os.Stderr, "encode response error: %s\n", err)
				return
			}
		case <-user.done:
			s.usersMutex.Lock()
			delete(s.users, user.name)
			s.usersMutex.Unlock()
			fmt.Fprintf(os.Stderr, "user removed: %s\n", connect.Name)
			return
		}
	}
}

func (s *Server) dispatch(connectedUser *user, request protocol.ClientRequest) {
	fmt.Printf("received request: %#v\n", request)
	switch request := request.(type) {
	case *protocol.ConnectRequest:
		connectedUser.outgoing <- &protocol.FatalErrorResponse{
			Error: protocol.AlreadyConnected,
			Info:  "",
		}

	case *protocol.DisconnectRequest:
		connectedUser.done <- struct{}{}

	case *protocol.ListRoomsRequest:
		s.roomsMutex.RLock()
		rooms := make([]string, 0, len(s.rooms))
		for room := range s.rooms {
			rooms = append(rooms, room)
		}
		s.roomsMutex.RUnlock()

		connectedUser.outgoing <- &protocol.RoomListResponse{
			Count: uint32(len(rooms)),
			Rooms: rooms,
		}

	case *protocol.ListUsersRequest:
		var users []string
		if request.Room != "" {
			s.roomsMutex.RLock()
			room, ok := s.rooms[request.Room]
			s.roomsMutex.RUnlock()
			if !ok {
				connectedUser.outgoing <- &protocol.ErrorResponse{
					Error: protocol.MissingRoom,
					Info:  "",
				}
				break
			}

			room.usersMutex.RLock()
			users = make([]string, 0, len(room.users))
			for user := range room.users {
				users = append(users, user)
			}
			room.usersMutex.RUnlock()
		} else {
			s.usersMutex.RLock()
			users = make([]string, 0, len(s.users))
			for user := range s.users {
				users = append(users, user)
			}
			s.usersMutex.RUnlock()
		}

		connectedUser.outgoing <- &protocol.UserListResponse{
			Count: uint32(len(users)),
			Room:  request.Room,
			Users: users,
		}

	case *protocol.MessageRoomRequest:
		s.roomsMutex.RLock()
		room, ok := s.rooms[request.Room]
		s.roomsMutex.RUnlock()
		if !ok {
			connectedUser.outgoing <- &protocol.ErrorResponse{
				Error: protocol.MissingRoom,
				Info:  "",
			}
			break
		}

		room.usersMutex.RLock()
		for _, user := range room.users {
			user.outgoing <- &protocol.RoomMessageResponse{
				Room:   request.Room,
				Sender: connectedUser.name,
				Text:   request.Text,
			}
		}
		room.usersMutex.RUnlock()

	case *protocol.MessageUserRequest:
		s.usersMutex.RLock()
		user, ok := s.users[request.User]
		s.usersMutex.RUnlock()
		if !ok {
			connectedUser.outgoing <- &protocol.ErrorResponse{
				Error: protocol.MissingUser,
				Info:  "",
			}
			break
		}

		user.outgoing <- &protocol.UserMessageResponse{
			Sender: connectedUser.name,
			Text:   request.Text,
		}

	case *protocol.CreateRoomRequest:
		if strings.ContainsRune(request.Room, ' ') {
			connectedUser.outgoing <- &protocol.ErrorResponse{
				Error: protocol.InvalidRoom,
				Info:  "",
			}
			break
		}

		s.roomsMutex.Lock()
		if _, ok := s.rooms[request.Room]; ok {
			connectedUser.outgoing <- &protocol.ErrorResponse{
				Error: protocol.ExistingRoom,
				Info:  "",
			}
			s.roomsMutex.Unlock()
			break
		}
		s.rooms[request.Room] = &room{
			users:      make(map[string]*user),
			usersMutex: sync.RWMutex{},
		}
		s.roomsMutex.Unlock()

	case *protocol.JoinRoomRequest:
		s.roomsMutex.RLock()
		room, ok := s.rooms[request.Room]
		s.roomsMutex.RUnlock()
		if !ok {
			connectedUser.outgoing <- &protocol.ErrorResponse{
				Error: protocol.MissingRoom,
				Info:  "",
			}
			break
		}

		room.usersMutex.Lock()
		room.users[connectedUser.name] = connectedUser
		room.usersMutex.Unlock()

	case *protocol.LeaveRoomRequest:
		s.roomsMutex.Lock()
		room, ok := s.rooms[request.Room]

		if !ok {
			connectedUser.outgoing <- &protocol.ErrorResponse{
				Error: protocol.MissingRoom,
				Info:  "",
			}
			s.roomsMutex.Unlock()
			break
		}

		room.usersMutex.Lock()
		delete(room.users, connectedUser.name)
		if len(room.users) == 0 {
			delete(s.rooms, request.Room)
		}
		room.usersMutex.Unlock()

		s.roomsMutex.Unlock()
	}
}
