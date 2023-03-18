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

	general *room

	rooms      map[string]*room
	roomsMutex sync.RWMutex

	room
}

type connectedUser struct {
	*user
	server *Server
}

func NewServer(host string, port int) *Server {
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
	s.users[cu.name] = cu.user
	s.usersMutex.Unlock()

	cu.server.general.usersMutex.Lock()
	cu.server.general.users[cu.name] = cu.user
	cu.server.general.usersMutex.Unlock()

	go func() {
		for {
			request, err := protocol.DecodeClientRequest(conn)
			if err != nil {
				fmt.Fprintf(os.Stderr, "decode error: %s\n", err)
				cu.done <- struct{}{}
				return
			}
			cu.incoming <- request
		}
	}()

	for {
		select {
		case request := <-cu.incoming:
			go request.Accept(cu)
		case response := <-cu.outgoing:
			err = protocol.EncodeServerResponse(conn, response)
			if err != nil {
				fmt.Fprintf(os.Stderr, "encode response error: %s\n", err)
				return
			}
		case <-cu.done:
			s.usersMutex.Lock()
			delete(s.users, cu.name)
			s.usersMutex.Unlock()
			fmt.Fprintf(os.Stderr, "user removed: %s\n", connect.Name)
			return
		}
	}
}
