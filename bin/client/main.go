package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mnxn/chat/client"
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

	fmt.Printf("connecting to %s:%d\n", *host, *port)

	var name string
	if len(os.Args) > 1 {
		name = os.Args[1]
	} else {
		fmt.Print("enter display name: ")
		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() {
			fmt.Fprintln(os.Stderr, "error reading name.")
			return
		}
		name = scanner.Text()
	}

	c := client.NewClient(name, *host, *port)
	err := c.Run()
	if err != nil {
		fmt.Printf("client error: %s", err.Error())
	}
}
