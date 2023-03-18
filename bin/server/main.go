package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/mnxn/chat/server"
)

var (
	host = flag.String("host", "localhost", "chat server hostname")
	port = flag.Int("port", 5555, "chat server port number")
)

func main() {
	flag.Usage = func() {
		exe := filepath.Base(os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "usage: %s [options] name\noptions:\n", exe)
		flag.PrintDefaults()
	}
	flag.Parse()

	logger := log.Default()
	logger.Printf("serving on %s:%d\n", *host, *port)

	s := server.NewServer(*host, *port, logger)
	err := s.Run()
	if err != nil {
		logger.Fatalf("server error: %s\n", err.Error())
	}
}
