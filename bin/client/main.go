package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	"github.com/mnxn/chat/client"
)

var (
	name      = flag.String("name", "", "display name")
	host      = flag.String("host", "localhost", "chat server hostname")
	port      = flag.Int("port", 5555, "chat server port number")
	keepalive = flag.Int("keepalive", 15, "how often to send keepalive request to the server in seconds")
)

func main() {
	flag.Parse()

	fmt.Printf("connecting to %s:%d\n", *host, *port)

	if *name == "" {
		fmt.Print("   enter display name: ")
		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() {
			fmt.Fprintln(os.Stderr, "error reading name.")
			return
		}
		*name = scanner.Text()
	}

	c := client.NewClient(*name, *host, *port, *keepalive)
	err := c.Run()
	if err != nil {
		fmt.Printf("client error: %s", err.Error())
	}
}
