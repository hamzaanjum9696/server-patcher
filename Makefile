all: server

BINARY_NAME=server-patcher
REMOTE_USER=root
REMOTE_HOST=139.59.44.246
DESTINATION_PATH=/root/server-patcher
PRIVATE_KEY="C:\Users\sam9696\.ssh\do.ppk"
PASSWORD="Energize-Hypnotize8-Legislate"

local:
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_NAME)_local cmd/server-patcher/main.go
	@echo "Done. Binary is located at $(BINARY_NAME)_local"

server:
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_NAME) cmd/server-patcher/main.go
	@echo "Done. Binary is located at $(BINARY_NAME)"
	@pscp -pw $(PASSWORD) $(BINARY_NAME) $(REMOTE_USER)@$(REMOTE_HOST):$(DESTINATION_PATH)
	@echo "Shipped binary to server at /root/server-patcher"