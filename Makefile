# Makefile for blockchain-node-gateway

# Environment variables
GO=go
IMAGE_NAME=mmn-api-gateway
CONTAINER_NAME=mmn-api-gateway-container

# Default commands
.PHONY: help
help: ## Show help
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: run
run: ## Run gRPC server application
	@echo "Starting mmn api gateway..."
	$(GO) run cmd/main.go

.PHONY: run-client
run-client: ## Run client application
	@echo "Running client application..."
	$(GO) run client/main.go

.PHONY: run-client-test
run-client-test: ## Run client tests
	@echo "Running client tests..."
	$(GO) test ./client/...

.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(IMAGE_NAME) .

.PHONY: docker-run
docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run -d --name $(CONTAINER_NAME) -p 8080:8080 -p 9090:9090 -p 9100:9100 -p 9101:9101 $(IMAGE_NAME)

# Default goal when running make without parameters
.DEFAULT_GOAL := help
