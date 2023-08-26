all: server

# set environment variables from .env file
include .env

BINARY_NAME=server-patcher


local:
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_NAME)_local cmd/server-patcher/main.go
	@echo "Done. Binary is located at $(BINARY_NAME)_local"

server:
	@echo "Building $(BINARY_NAME)..."
	@GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME) cmd/server-patcher/main.go
	@echo "Done building. Binary is located at $(BINARY_NAME)"
	@SCP_COMMAND="sshpass -p $(SERVER_CREDS_PASSWORD) scp -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null $(BINARY_NAME) $(SERVER_CREDS_USER)@$(SERVER_IP):/$(SERVER_CREDS_USER)/$(BINARY_NAME)"
	@if [ "$(OS)" = "Windows_NT" ]; then \
		SCP_COMMAND="pscp -pw $(SERVER_CREDS_PASSWORD) $(BINARY_NAME) $(SERVER_CREDS_USER)@$(SERVER_IP):/$(SERVER_CREDS_USER)/$(BINARY_NAME)"; \
	fi
	@$(SCP_COMMAND)
	@echo "Done uploading. Binary is located at $(SERVER_IP):/$(SERVER_CREDS_USER)/$(BINARY_NAME)"