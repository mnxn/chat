package server

import (
	"fmt"
	"net"
	"os"

	"github.com/mnxn/chat/protocol"
)

type Server struct {
	host string
	port int
}

func NewServer(host string, port int) *Server {
	return &Server{
		host: host,
		port: port,
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

	for {
		request, err := protocol.DecodeClientRequest(conn)
		if err != nil {
			fmt.Fprintf(os.Stderr, "decode error: %s\n", err)
			return
		}

		fmt.Printf("%#v\n", request)
	}
}
