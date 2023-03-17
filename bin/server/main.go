package main

import (
	"fmt"

	"github.com/mnxn/chat/server"
)

func main() {
	s := server.NewServer("localhost", 5555)
	err := s.Run()
	if err != nil {
		fmt.Printf("server error: %s", err.Error())
	}
}
