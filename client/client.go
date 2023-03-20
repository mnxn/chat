package client

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"sync/atomic"
	"time"

	"github.com/mnxn/chat/protocol"
)

type Client struct {
	name string

	host string
	port int

	atomicCurrent atomic.Pointer[string]

	input    chan string
	output   chan string
	incoming chan protocol.ServerResponse
	outgoing chan protocol.ClientRequest

	ticker *time.Ticker
	conn   net.Conn
}

func NewClient(name, host string, port int, keepalive int) *Client {
	client := &Client{
		name: name,
		host: host,
		port: port,

		atomicCurrent: atomic.Pointer[string]{},

		input:    make(chan string),
		output:   make(chan string),
		incoming: make(chan protocol.ServerResponse),
		outgoing: make(chan protocol.ClientRequest),

		ticker: time.NewTicker(time.Duration(keepalive) * time.Second),
		conn:   nil,
	}
	current := "general"
	client.atomicCurrent.Store(&current)
	return client
}

func (c *Client) Run() error {
	var err error
	c.conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", c.host, c.port))
	if err != nil {
		return fmt.Errorf("error dialing: %w", err)
	}
	defer c.conn.Close()

	err = protocol.EncodeClientRequest(c.conn, &protocol.ConnectRequest{
		Version: 1,
		Name:    c.name,
	})
	if err != nil {
		return fmt.Errorf("error initiating connection: %w", err)
	}

	fmt.Println("connected.")
	fmt.Println()

	scannerErr := make(chan error)
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			c.input <- scanner.Text()
		}
		scannerErr <- scanner.Err()
	}()

	decodeErr := make(chan error)
	go func() {
		response, err := protocol.DecodeServerResponse(c.conn)
		for err == nil {
			c.incoming <- response
			response, err = protocol.DecodeServerResponse(c.conn)
		}
		decodeErr <- err
	}()

	for {
		select {
		case input := <-c.input:
			go c.parse(input)

		case output := <-c.output:
			fmt.Print(output)

		case <-c.ticker.C:
			err = protocol.EncodeClientRequest(c.conn, &protocol.KeepaliveRequest{})
			if err != nil {
				return fmt.Errorf("error sending request: %w", err)
			}

		case request := <-c.outgoing:
			err = protocol.EncodeClientRequest(c.conn, request)
			if err != nil {
				return fmt.Errorf("error sending request: %w", err)
			}

		case response := <-c.incoming:
			go response.Accept(c)

		case err := <-decodeErr:
			if errors.Is(err, os.ErrDeadlineExceeded) {
				return nil
			}
			return fmt.Errorf("error receiving response: %w", err)

		case err := <-scannerErr:
			if err != nil {
				return fmt.Errorf("error reading input: %w", err)
			}
			return nil
		}
	}
}
