package client

import (
	"bufio"
	"fmt"
	"net"
	"os"

	"github.com/mnxn/chat/protocol"
)

type Client struct {
	host string
	port int
}

func NewClient(host string, port int) *Client {
	return &Client{
		host: host,
		port: port,
	}
}

func (s *Client) Run() error {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", s.host, s.port))
	if err != nil {
		return fmt.Errorf("error dialing: %w", err)
	}
	defer conn.Close()

	err = protocol.EncodeClientRequest(conn, &protocol.ConnectRequest{
		Version: 1,
		Name:    "me",
	})
	if err != nil {
		return fmt.Errorf("error initiating connection: %w", err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		err := protocol.EncodeClientRequest(conn, &protocol.MessageRoomRequest{
			Room: "",
			Text: scanner.Text(),
		})
		if err != nil {
			return fmt.Errorf("error sending message: %w", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading input: %w", err)
	}

	return nil
}
