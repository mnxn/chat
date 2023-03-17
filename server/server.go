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

type Server struct {
	host string
	port int

	users      map[string]*user
	usersMutex sync.RWMutex
}

func NewServer(host string, port int) *Server {
	return &Server{
		host:       host,
		port:       port,
		users:      make(map[string]*user),
		usersMutex: sync.RWMutex{},
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
		}
	}
}

func (s *Server) dispatch(user *user, request protocol.ClientRequest) {
	fmt.Printf("received request: %#v\n", request)
	user.outgoing <- &protocol.RoomMessageResponse{
		Room:   "general",
		Sender: user.name,
		Text:   "echo",
	}
}
