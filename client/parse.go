package client

import (
	"fmt"
	"strings"
	"time"

	"github.com/mnxn/chat/protocol"
)

const helpMessage = `   command help:
      /help              show this message
      /current           show current room
      /switch [room]     switch current room
      /rooms             list rooms in the server
      /users             list users in the server
      /users  [room]     list users in a room
      /msg    [rooms]    send a message to specific rooms
      /dm     [users]    send a direct message to specific users
      /create [rooms]    create rooms
      /join   [rooms]    join rooms
      /leave  [rooms]    leave rooms
      /quit              quit the chat program
`

func (c *Client) parse(input string) {
	if !strings.HasPrefix(input, "/") {
		current := *c.atomicCurrent.Load()

		c.outgoing <- &protocol.MessageRoomRequest{
			Room: current,
			Text: input,
		}
		return
	}

	split := strings.SplitN(input[1:], " ", 3)
	if len(split) < 1 {
		c.output <- "[command error] invalid command: use /help to see all commands\n"
		return
	}

	switch split[0] {
	default:
		c.output <- "[command error] invalid command: use /help to see all commands\n"

	case "help":
		c.output <- helpMessage

	case "current":
		c.output <- fmt.Sprintf("   Current room: %s\n", *c.atomicCurrent.Load())

	case "switch":
		if len(split) < 2 {
			c.output <- "[command error] missing command argument: use /help to see usage\n"
			return
		}
		c.atomicCurrent.Store(&split[1])

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
		if len(split) <= 2 {
			c.output <- "[command error] missing command arguments: use /help to see usage\n"
			return
		}
		for _, room := range strings.Split(split[1], ",") {
			c.outgoing <- &protocol.MessageRoomRequest{
				Room: room,
				Text: split[2],
			}
		}

	case "dm":
		if len(split) <= 2 {
			c.output <- "[command error] missing command arguments: use /help to see usage\n"
			return
		}
		for _, user := range strings.Split(split[1], ",") {
			c.outgoing <- &protocol.MessageUserRequest{
				User: user,
				Text: split[2],
			}
		}

	case "create":
		if len(split) < 2 {
			c.output <- "[command error] missing command argument: use /help to see usage\n"
			return
		}
		for _, room := range strings.Split(split[1], ",") {
			c.outgoing <- &protocol.CreateRoomRequest{
				Room: room,
			}
		}

	case "join":
		if len(split) < 2 {
			c.output <- "[command error] missing command argument: use /help to see usage\n"
			return
		}
		var room string
		for _, room = range strings.Split(split[1], ",") {
			c.outgoing <- &protocol.JoinRoomRequest{
				Room: room,
			}
		}
		c.atomicCurrent.Store(&room)

	case "leave":
		if len(split) < 2 {
			c.output <- "[command error] missing command argument: use /help to see usage\n"
			return
		}
		for _, room := range strings.Split(split[1], ",") {
			c.outgoing <- &protocol.LeaveRoomRequest{
				Room: room,
			}
		}

	case "quit":
		c.outgoing <- &protocol.DisconnectRequest{}
		_ = c.conn.SetReadDeadline(time.Now())
	}
}
