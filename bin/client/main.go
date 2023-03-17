package main

import (
	"fmt"

	"github.com/mnxn/chat/client"
)

func main() {
	c := client.NewClient("localhost", 5555)
	err := c.Run()
	if err != nil {
		fmt.Printf("client error: %s", err.Error())
	}
}
