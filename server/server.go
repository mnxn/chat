package server

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"sync/atomic"

	"github.com/mnxn/chat/protocol"
)

type Server struct {
	port int

	general *room

	rooms      map[string]*room
	roomsMutex sync.RWMutex

	room

	logger *log.Logger
}

type room struct {
	users      map[string]*user
	usersMutex sync.RWMutex
}

func (r *room) contains(userName string) bool {
	r.usersMutex.RLock()
	_, ok := r.users[userName]
	r.usersMutex.RUnlock()
	return ok
}

type user struct {
	atomicName atomic.Pointer[string]
	incoming   chan protocol.ClientRequest
	outgoing   chan protocol.ServerResponse
}

func (u *user) name() string {
	return *u.atomicName.Load()
}

func (u *user) connected() bool {
	return u.name() != ""
}

type connectedUser struct {
	*user
	server *Server
	conn   net.Conn
}

func NewServer(port int, logger *log.Logger) *Server {
	general := &room{
		users:      make(map[string]*user),
		usersMutex: sync.RWMutex{},
	}

	return &Server{
		port: port,

		general: general,

		rooms:      map[string]*room{"general": general},
		roomsMutex: sync.RWMutex{},

		room: room{
			users:      make(map[string]*user),
			usersMutex: sync.RWMutex{},
		},

		logger: logger,
	}
}

func (s *Server) Run() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("error starting tcp server: %w", err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			s.logger.Printf("error accepting connection: %s", err.Error())
			return nil
		}

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	cu := &connectedUser{
		user: &user{
			atomicName: atomic.Pointer[string]{},
			incoming:   make(chan protocol.ClientRequest),
			outgoing:   make(chan protocol.ServerResponse),
		},
		server: s,
		conn:   conn,
	}
	cu.atomicName.Store(new(string))

	defer func() {
		if !cu.connected() {
			return
		}

		s.usersMutex.Lock()
		delete(s.users, cu.name())
		s.usersMutex.Unlock()

		s.roomsMutex.RLock()
		for roomName, room := range s.rooms {
			s.removeRoomUser(roomName, room, cu.user)
		}
		s.roomsMutex.RUnlock()

		s.logger.Printf("user removed: %s\n", cu.name())
	}()

	decodeErr := make(chan error)
	go func() {
		response, err := protocol.DecodeClientRequest(conn)
		for err == nil {
			cu.incoming <- response
			response, err = protocol.DecodeClientRequest(conn)
		}
		decodeErr <- err
	}()

	for {
		select {
		case response := <-cu.outgoing:
			s.logger.Printf("sent response to %s: %#v\n", cu.name(), response)
			err := protocol.EncodeServerResponse(conn, response)
			if err != nil {
				s.logger.Printf("encode response error: %s\n", err)
				return
			}

		case request := <-cu.incoming:
			s.logger.Printf("received request from %s: %#v\n", cu.name(), request)
			go request.Accept(cu)

		case err := <-decodeErr:
			if !errors.Is(err, os.ErrDeadlineExceeded) {
				s.logger.Printf("error receiving request: %s\n", err)
			}
			return
		}
	}
}

func (s *Server) removeRoomUser(roomName string, room *room, user *user) {
	room.usersMutex.Lock()
	_, inRoom := room.users[user.name()]
	if inRoom && len(room.users) == 1 && room != s.general {
		delete(s.rooms, roomName)
		s.logger.Printf("removed room: %s\n", roomName)
	} else {
		delete(room.users, user.name())
	}
	room.usersMutex.Unlock()
}
