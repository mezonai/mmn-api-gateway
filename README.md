# Blockchain Node Gateway

A gRPC API Gateway for the MMN blockchain system, providing load balancing and routing capabilities to different blockchain nodes.

## 🚀 Key Features

- **gRPC Gateway**: Provides gRPC API to interact with blockchain nodes
- **Load Balancing**: Supports round-robin load balancing for multiple nodes
- **Retry Mechanism**: Automatic retry on connection failures
- **Health Check**: Monitors node status
- **Transaction Management**: Handles blockchain transactions
- **Account Management**: Manages accounts and balances
- **Prometheus Metrics**: Monitoring and observability

## 🏗️ Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   gRPC Client  │───▶│  Gateway Server  │───▶│  MMN Node 1     │
│                 │    │                  │    │  (localhost:9001)│
└─────────────────┘    └──────────────────┘    ├─────────────────┘
                                               │  MMN Node 2     │
                                               │  (localhost:9002)│
                                               ├─────────────────┘
                                               │  MMN Node 3     │
                                               │  (localhost:9003)│
                                               └─────────────────┘
```

## 📁 Project Structure

```
blockchain-node-gateway/
├── cmd/                    # Main application entry point
│   └── main.go           # Initialize and run gRPC server
├── client/                # Client test and demo
│   ├── main.go           # Client demo with health check
│   └── client_test.go    # Unit tests for client
├── internal/              # Internal code
│   ├── config/           # Application configuration
│   │   └── config.go     # Configuration structure definition
│   └── rpc/              # gRPC server implementation
│       ├── gateway.go     # Gateway service logic
│       └── server.go      # Server setup and management
├── etc/                   # Configuration files
│   └── gateway.yaml      # Server and endpoints configuration
├── go.mod                 # Go module dependencies
├── go.sum                 # Go module checksums
├── Makefile               # Build and run commands
└── README.md              # This documentation
```

## 🛠️ System Requirements

- Go 1.24 or higher
- Git

## 📦 Installation

1. Clone repository:
```bash
git clone <repository-url>
cd blockchain-node-gateway
```

2. Install dependencies:
```bash
go mod download
```

## 🚀 Running the Application

### Run Gateway Server

```bash
make run
```

Or:
```bash
go run cmd/main.go
```

### Run Client Demo

```bash
make run-client
```

Or:
```bash
go run client/main.go
```

### Run Tests

```bash
make run-client-test
```

Or:
```bash
go test ./client/...
```

## ⚙️ Configuration

Main configuration file: `etc/gateway.yaml`

```yaml
Name: api.gateway
Host: 0.0.0.0
Port: 8080

# Prometheus metrics
Prometheus:
  Host: 0.0.0.0
  Port: 9100
  Path: /metrics

# gRPC server config
MMNRpc:
  Name: mmn.grpc
  ListenOn: 0.0.0.0:9090
  MaxRetries: 3
  
  # Target nodes for load balancing
  Target: "localhost:9001,localhost:9002,localhost:9003"
  ServiceConfig: '{"loadBalancingPolicy":"round_robin"}'
  
  # Prometheus metrics for gRPC
  Prometheus:
    Host: 0.0.0.0
    Port: 9101
    Path: /metrics
```

## 🔧 API Endpoints

### Health Check
- **Endpoint**: `Check`
- **Purpose**: Kiểm tra trạng thái của node
- **Response**: Thông tin về node ID, status, slot, height, mempool size

### Transaction Management
- **Endpoint**: `AddTx`
- **Purpose**: Thêm giao dịch mới vào blockchain
- **Request**: `SignedTxMsg` với transaction đã được ký

### Account Management
- **Endpoint**: `GetAccount`
- **Purpose**: Lấy thông tin tài khoản và số dư
- **Request**: `GetAccountRequest` với địa chỉ

- **Endpoint**: `GetTxHistory`
- **Purpose**: Lấy lịch sử giao dịch của tài khoản
- **Request**: `GetTxHistoryRequest` với địa chỉ, limit, offset

## 🔄 Load Balancing

The Gateway uses round-robin load balancing to distribute requests to nodes:

1. **Manual Resolver**: Creates manual resolver for multiple endpoints
2. **Round-Robin**: Uses gRPC built-in round-robin policy
3. **Retry Logic**: Automatic retry with exponential backoff
4. **Connection Management**: Manages connections to nodes

## 📊 Monitoring

- **Prometheus Metrics**: Both REST and gRPC endpoints
- **Health Checks**: Monitor node status
- **Logging**: Structured logging with configurable levels

## 🧪 Testing

### Unit Tests
```bash
go test ./client/...
```

### Integration Tests
Client tests include:
- Health check testing
- Faucet token transfer
- Transaction history retrieval

## 🚨 Troubleshooting

### Common Issues

1. **Connection Failed**: Check node endpoints in config
2. **Load Balancing Not Working**: Ensure multiple endpoints are configured
3. **Retry Failures**: Check `MaxRetries` and network connectivity

### Logs
- Server logs: Console output with configurable levels
- Client logs: Detailed operation logging
- Error logs: Retry attempts and failure reasons

## 🤝 Contributing

1. Fork repository
2. Tạo feature branch
3. Commit changes
4. Push to branch
5. Tạo Pull Request

## 📄 License

[License information]

## 📞 Support

[Contact information hoặc support channels]
