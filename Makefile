all: upload

# set environment variables from .env file
include .env

BINARY_NAME=server-patcher

local:
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_NAME)_local cmd/server-patcher/main.go
	@echo "Done. Binary is located at $(BINARY_NAME)_local"

SCP_COMMAND := sshpass -p $(SERVER_CREDS_PASSWORD) scp -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null $(BINARY_NAME) $(SERVER_CREDS_USER)@$(SERVER_IP):/$(SERVER_CREDS_USER)/$(BINARY_NAME)""
ifeq ($(OS),Windows_NT)
	SCP_COMMAND := pscp -pw $(SERVER_CREDS_PASSWORD) $(BINARY_NAME) $(SERVER_CREDS_USER)@$(SERVER_IP):/$(SERVER_CREDS_USER)/$(BINARY_NAME)
endif

build:
	@echo "Building $(BINARY_NAME)..."
	@GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME) cmd/server-patcher/main.go
	@echo "Done building."

upload: build
	@echo "Uploading $(BINARY_NAME)..."
	@$(shell $(SCP_COMMAND))
	@echo "Done uploading."

.PHONY: build upload
