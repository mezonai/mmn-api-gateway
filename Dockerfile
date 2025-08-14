# Build stage
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" \
    -o mmn-api-gateway \
    cmd/server/main.go

# Final stage
FROM alpine:latest

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/mmn-api-gateway .

# Copy configuration file
COPY --from=builder /app/config.yaml .

# Expose gRPC port
EXPOSE 9090

# Set default command
CMD ["./mmn-api-gateway"]
