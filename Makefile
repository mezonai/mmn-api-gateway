.PHONY: run-server run-client run-client-test build-docker run-docker

# Minimal Makefile with only four commands

BINARY_NAME=mmn-api-gateway
MAIN_SERVER=cmd/server/main.go
MAIN_CLIENT=cmd/client/main.go

# Run server
run-server:
	@echo "Running server..."
	go run $(MAIN_SERVER)

# Run client
run-client:
	@echo "Running client..."
	go run $(MAIN_CLIENT)

# Run client tests
run-client-test:
	@echo "Running client tests..."
	go test ./cmd/client/...

# Build Docker image
build-docker:
	@echo "Building Docker image $(BINARY_NAME):latest..."
	docker build -t $(BINARY_NAME):latest .

# Run Docker container
run-docker:
	@echo "Running Docker container $(BINARY_NAME):latest..."
	docker run -d --rm -p 9090:9090 -p 8081:8081 -p 9091:9091 $(BINARY_NAME):latest
