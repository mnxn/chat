package client

import (
	"fmt"
	"strings"
	"time"

	"github.com/mnxn/chat/protocol"
)

func (c *Client) Error(response *protocol.ErrorResponse) {
	if len(response.Info) > 0 {
		c.output <- fmt.Sprintf("[server error] %s: %s\n", response.Error, response.Info)
	} else {
		c.output <- fmt.Sprintf("[server error] %s\n", response.Error)
	}
}

func (c *Client) FatalError(response *protocol.FatalErrorResponse) {
	if len(response.Info) > 0 {
		c.output <- fmt.Sprintf("[fatal error] %s: %s\n", response.Error, response.Info)
	} else {
		c.output <- fmt.Sprintf("[fatal error] %s\n", response.Error)
	}
	_ = c.conn.SetReadDeadline(time.Now())
}

func (c *Client) RoomList(response *protocol.RoomListResponse) {
	var sb strings.Builder
	if response.User == "" {
		fmt.Fprintln(&sb, "   Room Listing in Server:")
	} else {
		fmt.Fprintf(&sb, "   Room Listing for User %s:\n", response.User)
	}
	for _, room := range response.Rooms {
		fmt.Fprintf(&sb, "      %s\n", room)
	}
	c.output <- sb.String()
}

func (c *Client) UserList(response *protocol.UserListResponse) {
	var sb strings.Builder
	if response.Room == "" {
		fmt.Fprintln(&sb, "   User Listing in Server:")
	} else {
		fmt.Fprintf(&sb, "   User Listing in Room %s:\n", response.Room)
	}
	for _, user := range response.Users {
		fmt.Fprintf(&sb, "      %s\n", user)
	}
	c.output <- sb.String()
}

func (c *Client) RoomMessage(response *protocol.RoomMessageResponse) {
	c.output <- fmt.Sprintf("<%s@%s> %s\n", response.Sender, response.Room, response.Text)
}

func (c *Client) UserMessage(response *protocol.UserMessageResponse) {
	c.output <- fmt.Sprintf("(%s) %s\n", response.Sender, response.Text)
}
