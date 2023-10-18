## Introduction

This project contains both the server and client implementations of a custom TCP
chat protocol.

Users can either send chat messages to other users or "rooms." Sending a chat
message to a room notifies every other user that is connected to the room.

Servers only act as relays and forward chat messages between clients. The server
maintains lists of rooms and users while clients can display message history.

Clients only interact with the server and are not made aware of other clients'
IP addresses. Clients can create, join, and leave rooms. The server shall
fulfill queries for lists of rooms and connected users.

## Usage

```
Usage of chat-client.exe:
  -host string
        chat server hostname (default "localhost")
  -keepalive int
        how often to send keepalive request to the server in seconds (default 15)
  -name string
        display name
  -port int
        chat server port number (default 5555)
```

```
Usage of chat-server.exe:
  -port int
        chat server port number (default 5555)
```

## Instructions

Build all or specific executables:

- `make`
- `make all`
- `make chat-client.exe`
- `make chat-server.exe`

Run protocol unit tests:

- `make test`

Run protocol fuzzing tests

- `make fuzz`

Clean files:

- `make clean`

Run with race detector:

- `make RACE=y run-client`
- `make RACE=y run-server`

# Protocol

Documentation for the protocol can be found here:
https://pkg.go.dev/github.com/mnxn/chat/protocol
