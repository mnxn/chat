package main

import (
	"flag"
	"log"

	"github.com/mnxn/chat/server"
)

var port = flag.Int("port", 5555, "chat server port number")

func main() {
	flag.Parse()

	logger := log.Default()
	logger.Printf("serving on port %d\n", *port)

	s := server.NewServer(*port, logger)
	err := s.Run()
	if err != nil {
		logger.Fatalf("server error: %s\n", err.Error())
	}
}
