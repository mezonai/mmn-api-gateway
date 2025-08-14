# MMN API Gateway

A high-performance gRPC gateway with load balancing capabilities for blockchain nodes. This gateway acts as a single entry point for blockchain operations, distributing requests across multiple blockchain nodes using various load balancing strategies.

## Features

- **Load Balancing**: Multiple strategies including Round Robin, Least Connections, and Health Check
- **Health Monitoring**: Automatic health checks for blockchain nodes
- **Retry Logic**: Configurable retry mechanisms with exponential backoff
- **gRPC Protocol**: Full gRPC support with reflection enabled
- **HTTP REST API**: RESTful endpoints for easy integration
- **Logging**: Structured JSON logging with configurable levels
- **Docker Support**: Containerized deployment with Docker
- **Clean Architecture**: Well-organized code structure with separation of concerns

## Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Client       │    │   Gateway        │    │   Blockchain    │
│   (gRPC/HTTP)  │───▶│   (Load Balancer)│───▶│   Node 1        │
└─────────────────┘    └──────────────────┘    ├─────────────────┤
                                                │   Blockchain    │
                                                │   Node 2        │
                                                ├─────────────────┤
                                                │   Blockchain    │
                                                │   Node 3        │
                                                └─────────────────┘
```

## Quick Start

### Prerequisites

- Go 1.23 or later
- Docker (optional)

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/mezonai/mmn-api-gateway.git
   cd mmn-api-gateway
   ```

2. **Install Go dependencies**
   ```bash
   go mod download
   go mod tidy
   ```

3. **Build the application**
   ```bash
   make build-docker
   ```

### Running the Gateway

#### Development Mode
```bash
make run-server
```

#### Docker Mode
```bash
# Build and run with Docker
make build-docker
make run-docker
```

## Configuration

The gateway is configured via `config.yaml`. Key configuration options:

```yaml
gateway:
  grpc_port: 9090               # gRPC server port
  http_port: 8081               # HTTP server port
  enable_http: true             # Enable HTTP server
  enable_grpc: true             # Enable gRPC server
  timeout: 10s                  # Request timeout
  max_retries: 3                # Maximum retry attempts
  log_level: "info"             # Logging level

loadbalancer:
  strategy: "round_robin"       # Load balancing strategy
  health_check_interval: 30s    # Health check frequency
  max_failures: 3               # Max failures before marking node unhealthy

nodes:
  - address: "localhost:9001"   # Blockchain node address
    weight: 1                   # Node weight for load balancing
  - address: "localhost:9002"   # Blockchain node address
    weight: 1                   # Node weight for load balancing
  - address: "localhost:9003"   # Blockchain node address
    weight: 1                   # Node weight for load balancing
```

## Load Balancing Strategies

### Round Robin
Distributes requests sequentially across all healthy nodes.

### Least Connections
Routes requests to the node with the fewest active connections.

### Health Check
Routes requests to the healthiest node based on health check results.

## API Endpoints

### gRPC Endpoints
- `AddTx` - Submit a new transaction
- `Check` - Health status endpoint
- `GetStats` - Get comprehensive health statistics for all nodes
- `GetAccount` - Get account information
- `GetTxHistory` - Get transaction history for an account

### HTTP REST API Endpoints
- `GET /` - Service information and available endpoints
- `GET /health` - Health check endpoint
- `GET /stats` - Get comprehensive health statistics for all nodes

### GetStats Endpoint

#### gRPC
```bash
# Using grpcurl to get stats
grpcurl -plaintext localhost:9090 grpc.health.v1.Health/Check -d '{}'
```

#### HTTP REST API
```bash
# Get stats via HTTP
curl http://localhost:8081/stats
```

**Response includes:**
- Total number of nodes
- Number of healthy nodes
- Load balancing strategy
- Detailed information for each node:
  - Address
  - Weight
  - Health status (🟢 healthy, 🔴 unhealthy)
  - Failure count
  - Last health check timestamp

**Example HTTP Response:**
```json
{
  "status": "success",
  "stats": {
    "total_nodes": 3,
    "healthy_nodes": 2,
    "strategy": "round_robin",
    "nodes": [
      {
        "address": "localhost:9001",
        "weight": 1,
        "is_healthy": true,
        "failures": 0,
        "last_health": "2024-01-15T10:30:00Z"
      },
      {
        "address": "localhost:9002",
        "weight": 1,
        "is_healthy": true,
        "failures": 0,
        "last_health": "2024-01-15T10:30:00Z"
      },
      {
        "address": "localhost:9003",
        "weight": 1,
        "is_healthy": false,
        "failures": 2,
        "last_health": "2024-01-15T10:25:00Z"
      }
    ]
  }
}
```

## Monitoring

### Health Check
```bash
# gRPC health check
grpcurl -plaintext localhost:9090 grpc.health.v1.Health/Check

# HTTP health check
curl http://localhost:8081/health
```

### Node Statistics
```bash
# Get comprehensive node health statistics via gRPC
grpcurl -plaintext localhost:9090 grpc.health.v1.Health/Check -d '{}'

# Get comprehensive node health statistics via HTTP
curl http://localhost:8081/stats
```

### Logs
Logs are output in JSON format with configurable levels (debug, info, warn, error)

## Development

### Project Structure
```
.
├── cmd/                    # Application entry points
│   ├── server/            # Main server
│   └── client/            # Test client
├── internal/               # Internal packages
│   ├── config/            # Configuration management
│   ├── gateway/            # gRPC gateway implementation
│   │   ├── gateway.go     # Business logic
│   │   └── types.go       # Type definitions
│   ├── handlers/           # HTTP endpoint handlers
│   │   └── handlers.go    # HTTP request handlers
│   └── loadbalancer/      # Load balancing logic
├── config.yaml            # Configuration file
├── Dockerfile             # Docker image definition
├── Makefile               # Build and development commands
└── README.md              # This file
```

### Testing

```bash
# Run server
make run-server

# Run client
make run-client

# Build Docker image
make build-docker

# Run Docker container
make run-docker
```

#### Testing the Gateway

1. **Start the gateway:**
   ```bash
   # Development mode
   make run-server
   
   # Or with Docker
   make build-docker
   make run-docker
   ```

2. **Test gRPC endpoints:**
   ```bash
   # Run test client
   make run-client
   ```

3. **Test health check:**
   ```bash
   # Using grpcurl
   grpcurl -plaintext localhost:9090 grpc.health.v1.Health/Check
   ```

4. **Test HTTP endpoints:**
   ```bash
   # Health check
   curl http://localhost:8081/health
   
   # Get stats
   curl http://localhost:8081/stats
   ```

## Deployment

### Docker Deployment
```bash
# Build and deploy
make build-docker
make run-docker
```

### Production Considerations

- **Security**: Use TLS certificates for gRPC and HTTP
- **Authentication**: Implement API key or JWT authentication
- **Rate Limiting**: Add rate limiting middleware
- **Logging**: Configure centralized logging (ELK stack)
- **Monitoring**: Set up alerting for health checks
- **Backup**: Regular backup of configuration and state

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

For support and questions:
- Create an issue on GitHub
- Contact the development team
- Check the documentation

## Roadmap

### Completed ✅
- [x] gRPC server implementation
- [x] Load balancing with multiple strategies
- [x] Health checking for nodes
- [x] Basic configuration management
- [x] Docker containerization
- [x] Test client implementation
- [x] Retry logic with exponential backoff
- [x] Structured JSON logging
- [x] HTTP REST API server with /stats endpoint
- [x] Clean architecture with separated concerns
- [x] Custom types and structured responses
- [x] Organized HTTP handlers package

### Planned 📋
- [ ] Prometheus metrics integration
- [ ] Circuit breaker implementation
- [ ] WebSocket support for real-time updates
- [ ] Rate limiting and throttling
- [ ] Authentication and authorization
- [ ] TLS/SSL support
- [ ] Service discovery with Consul
- [ ] Comprehensive monitoring dashboard

## Known Issues

- Metrics server is not yet implemented
- Some advanced load balancing features are in development

## Troubleshooting

### Common Issues

1. **Port conflicts**: Ensure ports 9090 (gRPC) and 8081 (HTTP) are available
2. **Node connection failures**: Check that blockchain nodes are running and accessible
3. **Health check failures**: Verify node addresses in config.yaml are correct

### Debug Mode
```bash
# Set log level to debug in config.yaml
log_level: "debug"
make run-server
```