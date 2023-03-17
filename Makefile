.PHONY: all
all: chat-client.exe chat-server.exe

.PHONY: test
test:
	go test -run=^Test ./protocol

.PHONY: fuzz
fuzz:
	go test -run=^Fuzz -fuzz=Fuzz -fuzztime=20s ./protocol

.PHONY: all
clean:
	rm -rf chat-client.exe chat-server.exe protocol/testdata

.PHONY: run-client
run-client: chat-client.exe
	./chat-client.exe

.PHONY: run-server
run-server: chat-server.exe
	./chat-server.exe

chat-client.exe: $(wildcard client/*) $(wildcard bin/client/*)
	go build -o chat-client.exe ./bin/client

chat-server.exe: $(wildcard server/*) $(wildcard bin/server/*)
	go build -o chat-server.exe ./bin/server
