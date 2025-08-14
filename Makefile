# Makefile for blockchain-node-gateway

# Environment variables
GO=go

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

# Default goal when running make without parameters
.DEFAULT_GOAL := help
