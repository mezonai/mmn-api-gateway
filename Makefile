# Makefile for mmn-api-gateway

# Default target
.DEFAULT_GOAL := help

# Variables
BINARY_NAME = mmn-api-gateway
MAIN_SERVER = cmd/server/main.go
MAIN_CLIENT = cmd/client/main.go

# Help target
help:
	@echo "Available commands:"
	@echo ""
	@echo "  build            - Build the application"
	@echo "  run-server       - Run the server"
	@echo "  run-client       - Run the client"
	@echo "  test             - Run all tests"
	@echo "  test-no-cache    - Run all tests without cache"
	@echo "  run-client-test  - Run client tests without cache"
	@echo "  build-docker     - Build Docker image"
	@echo "  run-docker       - Run Docker container"
	@echo "  help             - Show this help message"
	@echo ""

.PHONY: build run-server run-client run-client-test test test-no-cache build-docker run-docker help

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) $(MAIN_SERVER)

# Run server
run-server:
	@echo "Running server..."
	go run $(MAIN_SERVER)

# Run client
run-client:
	@echo "Running client..."
	go run $(MAIN_CLIENT)

# Run all tests
test:
	@echo "Running all tests..."
	go test ./...

# Run all tests without cache
test-no-cache:
	@echo "Running all tests without cache..."
	go test -count=1 ./...

# Run client tests
run-client-test:
	@echo "Running client tests without cache..."
	go test -count=1 ./cmd/client/...

# Build Docker image
build-docker:
	@echo "Building Docker image $(BINARY_NAME):latest..."
	docker build -t $(BINARY_NAME):latest .

# Run Docker container
run-docker:
	@echo "Running Docker container $(BINARY_NAME):latest..."
	docker run -d --rm -p 9090:9090 -p 8081:8081 -p 9091:9091 $(BINARY_NAME):latest
