package client

import (
	"fmt"
	"strings"

	"github.com/mnxn/chat/protocol"
)

func (c *Client) VisitError(response *protocol.ErrorResponse) {
	if len(response.Info) > 0 {
		c.output <- fmt.Sprintf("[server error] %s: %s", response.Error, response.Info)
	} else {
		c.output <- fmt.Sprintf("[server error] %s", response.Error)
	}
}

func (c *Client) VisitFatalError(response *protocol.FatalErrorResponse) {
	if len(response.Info) > 0 {
		c.output <- fmt.Sprintf("[fatal error] %s: %s", response.Error, response.Info)
	} else {
		c.output <- fmt.Sprintf("[fatal error] %s", response.Error)
	}
}

func (c *Client) VisitRoomList(response *protocol.RoomListResponse) {
	var sb strings.Builder
	sb.WriteString("Room Listing:\n")
	for _, room := range response.Rooms {
		sb.WriteRune('\t')
		sb.WriteString(room)
		sb.WriteRune('\n')
	}
	c.output <- sb.String()
}

func (c *Client) VisitUserList(response *protocol.UserListResponse) {
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
}

func (c *Client) VisitRoomMessage(response *protocol.RoomMessageResponse) {
	c.output <- fmt.Sprintf("%s in %s> %s", response.Sender, response.Room, response.Text)
}

func (c *Client) VisitUserMessage(response *protocol.UserMessageResponse) {
	c.output <- fmt.Sprintf("%s to you> %s", response.Sender, response.Text)
}
