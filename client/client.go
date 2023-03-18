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
	name string

	host string
	port int

	input    chan string
	output   chan string
	incoming chan protocol.ServerResponse
	outgoing chan protocol.ClientRequest
	done     chan struct{}
}

func NewClient(name, host string, port int) *Client {
	return &Client{
		name: name,

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
		Name:    c.name,
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
			go response.Accept(c)

		case input := <-c.input:
			go c.parse(input)
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

func (c *Client) parse(input string) {
	if !strings.HasPrefix(input, "/") {
		c.outgoing <- &protocol.MessageRoomRequest{
			Room: "",
			Text: input,
		}
		return
	}

	split := strings.SplitN(input[1:], " ", 3)
	if len(split) < 1 {
		c.output <- "[command error] invalid command: use /help to see all commands"
		return
	}

	switch split[0] {
	default:
		c.output <- "[command error] invalid command: use /help to see all commands"

	case "help":
		c.output <- `command help:
    /help             show this message
    /rooms            list rooms in the server
    /users            list users in the server
    /users  [room]    list users in a room
    /msg    [room]    send a message to a specific room
    /dm     [user]    send a direct message to a user
    /create [room]    create a room
    /join   [room]    join a room
    /leave  [room]    leave a room
    /quit             quit the chat program`

	case "rooms":
		c.outgoing <- &protocol.ListRoomsRequest{}

	case "users":
		var room string
		if len(split) >= 2 {
			room = split[1]
		}
		c.outgoing <- &protocol.ListUsersRequest{
			Room: room,
		}

	case "msg":
		if len(split) < 2 {
			c.output <- "[command error] missing command arguments: use /help to see usage"
			return
		}
		c.outgoing <- &protocol.MessageRoomRequest{
			Room: split[1],
			Text: split[2],
		}

	case "dm":
		if len(split) < 2 {
			c.output <- "[command error] missing command arguments: use /help to see usage"
			return
		}
		c.outgoing <- &protocol.MessageUserRequest{
			User: split[1],
			Text: split[2],
		}

	case "create":
		if len(split) < 2 {
			c.output <- "[command error] missing command argument: use /help to see usage"
			return
		}
		c.outgoing <- &protocol.CreateRoomRequest{
			Room: split[1],
		}

	case "join":
		if len(split) < 2 {
			c.output <- "[command error] missing command argument: use /help to see usage"
			return
		}
		c.outgoing <- &protocol.JoinRoomRequest{
			Room: split[1],
		}

	case "leave":
		if len(split) < 2 {
			c.output <- "[command error] missing command argument: use /help to see usage"
			return
		}
		c.outgoing <- &protocol.LeaveRoomRequest{
			Room: split[1],
		}

	case "quit":
		c.outgoing <- &protocol.DisconnectRequest{}
		c.done <- struct{}{}
	}
}
