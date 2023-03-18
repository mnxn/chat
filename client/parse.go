package client

import (
	"strings"

	"github.com/mnxn/chat/protocol"
)

const helpMessage = `    command help:
        /help             show this message
        /switch [room]    switch current room
        /rooms            list rooms in the server
        /users            list users in the server
        /users  [room]    list users in a room
        /msg    [room]    send a message to a specific room
        /dm     [user]    send a direct message to a user
        /create [room]    create a room
        /join   [room]    join a room
        /leave  [room]    leave a room
        /quit             quit the chat program`

func (c *Client) parse(input string) {
	defer c.prompt()

	if !strings.HasPrefix(input, "/") {
		c.currentMutex.RLock()
		current := c.current
		c.currentMutex.RUnlock()

		c.outgoing <- &protocol.MessageRoomRequest{
			Room: current,
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
		c.output <- helpMessage

	case "switch":
		if len(split) < 2 {
			c.output <- "[command error] missing command argument: use /help to see usage"
			return
		}
		c.currentMutex.Lock()
		c.current = split[1]
		c.currentMutex.Unlock()

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
		c.currentMutex.Lock()
		c.current = split[1]
		c.currentMutex.Unlock()

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
