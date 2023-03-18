package client

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	"github.com/mnxn/chat/protocol"
)

type Client struct {
	name string

	host string
	port int

	current      string
	currentMutex sync.RWMutex

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

		current:      "general",
		currentMutex: sync.RWMutex{},

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
		c.prompt()
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

		case output := <-c.output:
			fmt.Println()
			fmt.Println(output)
			c.prompt()

		case <-c.done:
			return nil
		}
	}
}

func (c *Client) prompt() {
	c.currentMutex.RLock()
	fmt.Printf("<%s@%s> ", c.name, c.current)
	c.currentMutex.RUnlock()
}
