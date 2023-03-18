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
	sb.WriteString("   Room Listing:\n")
	for _, room := range response.Rooms {
		sb.WriteString("      ")
		sb.WriteString(room)
		sb.WriteRune('\n')
	}
	c.output <- sb.String()
}

func (c *Client) UserList(response *protocol.UserListResponse) {
	var sb strings.Builder
	if len(response.Room) > 0 {
		sb.WriteString(fmt.Sprintf("   User Listing in Room %s:\n", response.Room))
	} else {
		sb.WriteString("   User Listing in Server:\n")
	}
	for _, room := range response.Users {
		sb.WriteString("      ")
		sb.WriteString(room)
		sb.WriteRune('\n')
	}
	c.output <- sb.String()
}

func (c *Client) RoomMessage(response *protocol.RoomMessageResponse) {
	c.output <- fmt.Sprintf("<%s@%s> %s\n", response.Sender, response.Room, response.Text)
}

func (c *Client) UserMessage(response *protocol.UserMessageResponse) {
	c.output <- fmt.Sprintf("(%s) %s\n", response.Sender, response.Text)
}
