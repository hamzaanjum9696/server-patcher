all: server

BINARY_NAME=server-patcher

local:
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_NAME)_local cmd/server-patcher/main.go
	@echo "Done. Binary is located at $(BINARY_NAME)_local"

server:
	@echo "Building $(BINARY_NAME)..."
	@GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME) cmd/server-patcher/main.go
	@echo "Done. Binary is located at $(BINARY_NAME)"
	@scp $(BINARY_NAME) root@139.59.44.246:/root/server-patcher
	@echo "Shipped binary to server at /root/server-patcher"