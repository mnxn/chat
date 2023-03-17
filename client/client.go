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

	input    chan string
	output   chan string
	incoming chan protocol.ServerResponse
	outgoing chan protocol.ClientRequest
	done     chan struct{}
}

func NewClient(host string, port int) *Client {
	return &Client{
		host: host,
		port: port,

		input:    make(chan string),
		output:   make(chan string),
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
		fmt.Print("you> ")
		for scanner.Scan() {
			c.input <- scanner.Text()
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
		case request := <-c.outgoing:
			err = protocol.EncodeClientRequest(conn, request)
			if err != nil {
				return fmt.Errorf("error sending request: %w", err)
			}

		case response := <-c.incoming:
			go c.dispatch(response)

		case input := <-c.input:
			go func() {
				c.outgoing <- &protocol.MessageRoomRequest{
					Room: "",
					Text: input,
				}
			}()
			fmt.Print("you> ")

		case output := <-c.output:
			fmt.Println()
			fmt.Println(output)
			fmt.Print("you> ")

		case <-c.done:
			return nil
		}
	}
}

func (c *Client) dispatch(response protocol.ServerResponse) {
	switch response := response.(type) {
	case *protocol.ErrorResponse:
		if len(response.Info) > 0 {
			c.output <- fmt.Sprintf("[error] %s: %s", response.Error, response.Info)
		} else {
			c.output <- fmt.Sprintf("[error] %s", response.Error)
		}

	case *protocol.FatalErrorResponse:
		if len(response.Info) > 0 {
			c.output <- fmt.Sprintf("[fatal error] %s: %s", response.Error, response.Info)
		} else {
			c.output <- fmt.Sprintf("[fatal error] %s", response.Error)
		}

	case *protocol.RoomListResponse:
		var sb strings.Builder
		sb.WriteString("Room Listing:\n")
		for _, room := range response.Rooms {
			sb.WriteRune('\t')
			sb.WriteString(room)
			sb.WriteRune('\n')
		}
		c.output <- sb.String()

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
		c.output <- sb.String()

	case *protocol.RoomMessageResponse:
		c.output <- fmt.Sprintf("%s in %s> %s", response.Sender, response.Room, response.Text)

	case *protocol.UserMessageResponse:
		c.output <- fmt.Sprintf("%s to you> %s", response.Sender, response.Text)
	}
}
