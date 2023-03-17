package client

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/mnxn/chat/protocol"
)

type Client struct {
	host string
	port int

	incoming chan protocol.ServerResponse
	outgoing chan protocol.ClientRequest
	done     chan struct{}
}

func NewClient(host string, port int) *Client {
	return &Client{
		host:     host,
		port:     port,
		incoming: make(chan protocol.ServerResponse),
		outgoing: make(chan protocol.ClientRequest),
		done:     make(chan struct{}),
	}
}

func (c *Client) Run() error {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", c.host, c.port))
	if err != nil {
		return fmt.Errorf("error dialing: %w", err)
	}
	defer conn.Close()

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			c.outgoing <- &protocol.MessageRoomRequest{
				Room: "",
				Text: scanner.Text(),
			}
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "error reading input: %s", err)
		}
	}()

	go func() {
		for {
			response, err := protocol.DecodeServerResponse(conn)
			if err != nil {
				fmt.Fprintln(os.Stderr, "error receiving message: %w", err)
				log.Fatalln("exiting")
			}
			c.incoming <- response
		}
	}()

	err = protocol.EncodeClientRequest(conn, &protocol.ConnectRequest{
		Version: 1,
		Name:    "me",
	})
	if err != nil {
		return fmt.Errorf("error initiating connection: %w", err)
	}

	for {
		select {
		case response := <-c.incoming:
			go c.dispatch(response)
		case request := <-c.outgoing:
			err = protocol.EncodeClientRequest(conn, request)
			if err != nil {
				return fmt.Errorf("error sending request: %w", err)
			}
		case <-c.done:
			return nil
		}
	}
}

func (c *Client) dispatch(response protocol.ServerResponse) {
	switch response := response.(type) {
	case *protocol.ErrorResponse:
		if len(response.Info) > 0 {
			fmt.Printf("[error] %s: %s", response.Error, response.Info)
		} else {
			fmt.Printf("[error] %s", response.Error)
		}

	case *protocol.FatalErrorResponse:
		if len(response.Info) > 0 {
			fmt.Printf("[fatal error] %s: %s", response.Error, response.Info)
		} else {
			fmt.Printf("[fatal error] %s", response.Error)
		}

	case *protocol.RoomListResponse:
		var sb strings.Builder
		sb.WriteString("Room Listing:\n")
		for _, room := range response.Rooms {
			sb.WriteRune('\t')
			sb.WriteString(room)
			sb.WriteRune('\n')
		}
		fmt.Println(sb.String())

	case *protocol.UserListResponse:
		var sb strings.Builder
		if len(response.Room) > 0 {
			sb.WriteString(fmt.Sprintf("User Listing in Room %s:\n", response.Room))
		} else {
			sb.WriteString("User Listing in Server:\n")
		}
		for _, room := range response.Users {
			sb.WriteRune('\t')
			sb.WriteString(room)
			sb.WriteRune('\n')
		}
		fmt.Println(sb.String())

	case *protocol.RoomMessageResponse:
		fmt.Printf("%s in %s: %s\n", response.Sender, response.Room, response.Text)

	case *protocol.UserMessageResponse:
		fmt.Printf("%s to you: %s\n", response.Sender, response.Text)
	}
}
