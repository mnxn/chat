package server

import (
	"fmt"
	"log"
	"net"
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

	general *room

	rooms      map[string]*room
	roomsMutex sync.RWMutex

	room

	logger *log.Logger
}

type connectedUser struct {
	*user
	server *Server
}

func NewServer(host string, port int, logger *log.Logger) *Server {
	general := &room{
		users:      make(map[string]*user),
		usersMutex: sync.RWMutex{},
	}

	return &Server{
		host: host,
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
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.host, s.port))
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

	request, err := protocol.DecodeClientRequest(conn)
	if err != nil {
		s.logger.Printf("decode error: %s\n", err)
		return
	}
	s.logger.Printf("initial request from %v: %#v\n", conn.RemoteAddr(), request)

	connect, ok := request.(*protocol.ConnectRequest)
	if !ok {
		err := protocol.EncodeServerResponse(conn, &protocol.FatalErrorResponse{
			Error: protocol.NotConnected,
			Info:  "client not connected",
		})
		if err != nil {
			s.logger.Printf("error encoding response: %v", err)
		}
		return
	}
	if connect.Version != 1 {
		err := protocol.EncodeServerResponse(conn, &protocol.FatalErrorResponse{
			Error: protocol.UnsupportedVersion,
			Info:  "expected version 1",
		})
		if err != nil {
			s.logger.Printf("error encoding response: %v", err)
		}
		return
	}
	if strings.ContainsRune(connect.Name, ' ') {
		err := protocol.EncodeServerResponse(conn, &protocol.FatalErrorResponse{
			Error: protocol.InvalidUser,
			Info:  "username cannot contain spaces",
		})
		if err != nil {
			s.logger.Printf("error encoding response: %v", err)
		}
		return
	}

	cu := connectedUser{
		user: &user{
			name:     connect.Name,
			incoming: make(chan protocol.ClientRequest),
			outgoing: make(chan protocol.ServerResponse),
			done:     make(chan struct{}),
		},
		server: s,
	}

	s.usersMutex.Lock()
	if _, ok := s.users[cu.name]; ok {
		err := protocol.EncodeServerResponse(conn, &protocol.FatalErrorResponse{
			Error: protocol.ExistingUser,
			Info:  "username already exists",
		})
		if err != nil {
			s.logger.Printf("error encoding response: %v", err)
		}
		s.usersMutex.Unlock()
		return
	}
	s.users[cu.name] = cu.user
	s.usersMutex.Unlock()

	cu.server.general.usersMutex.Lock()
	cu.server.general.users[cu.name] = cu.user
	cu.server.general.usersMutex.Unlock()

	defer func() {
		s.usersMutex.Lock()
		delete(s.users, cu.name)
		s.usersMutex.Unlock()
		for _, room := range s.rooms {
			room.usersMutex.Lock()
			delete(s.users, cu.name)
			room.usersMutex.Unlock()
		}
		s.logger.Printf("user removed: %s\n", connect.Name)
	}()

	go func() {
		for {
			request, err := protocol.DecodeClientRequest(conn)
			if err != nil {
				s.logger.Printf("decode error: %s\n", err)
				cu.done <- struct{}{}
				return
			}
			cu.incoming <- request
		}
	}()

	for {
		select {
		case request := <-cu.incoming:
			s.logger.Printf("received request from %s: %#v\n", cu.name, request)
			go request.Accept(cu)
		case response := <-cu.outgoing:
			s.logger.Printf("sent response to %s: %#v\n", cu.name, response)
			err = protocol.EncodeServerResponse(conn, response)
			if err != nil {
				s.logger.Printf("encode response error: %s\n", err)
				return
			}
		case <-cu.done:

			return
		}
	}
}
