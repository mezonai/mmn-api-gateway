package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	mmnpb "github.com/mezonai/mmn/proto"

	"github.com/mezonai/mmn-api-gateway/internal/config"
	"github.com/mezonai/mmn-api-gateway/internal/gateway"
	"github.com/mezonai/mmn-api-gateway/internal/handlers"
	"github.com/mezonai/mmn-api-gateway/internal/loadbalancer"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	defaultConfigFile = "config.yaml"
	shutdownTimeout   = 30 * time.Second
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.LoadConfig(defaultConfigFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	logger := setupLogger(cfg.Gateway.LogLevel)
	logger.Info("Starting MMN API Gateway")

	lb, err := setupLoadBalancer(cfg, logger)
	if err != nil {
		return fmt.Errorf("failed to setup load balancer: %w", err)
	}
	defer lb.Close()

	gateway := gateway.NewBlockchainGateway(
		lb,
		cfg.Gateway.Timeout,
		cfg.Gateway.MaxRetries,
		logger,
	)

	var grpcServer *grpc.Server
	if cfg.Gateway.EnableGRPC {
		grpcServer, err = startGRPCServer(cfg, gateway, logger)
		if err != nil {
			return fmt.Errorf("failed to start gRPC server: %w", err)
		}
		defer grpcServer.GracefulStop()
	}

	var httpServer *handlers.HTTPServer
	if cfg.Gateway.EnableHTTP {
		httpServer, err = handlers.NewHTTPServer(cfg, gateway, logger)
		if err != nil {
			return fmt.Errorf("failed to start HTTP server: %w", err)
		}
		defer httpServer.Shutdown(context.Background())
	}

	_, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down servers...")

	if grpcServer != nil {
		grpcServer.GracefulStop()
	}
	if httpServer != nil {
		httpServer.Shutdown(context.Background())
	}

	logger.Info("Servers stopped successfully")
	return nil
}

func setupLogger(level string) *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	switch level {
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "info":
		logger.SetLevel(logrus.InfoLevel)
	case "warn":
		logger.SetLevel(logrus.WarnLevel)
	case "error":
		logger.SetLevel(logrus.ErrorLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}

	return logger
}

func setupLoadBalancer(cfg *config.Config, logger *logrus.Logger) (*loadbalancer.LoadBalancer, error) {
	lb := loadbalancer.NewLoadBalancer(
		cfg.LoadBalancer.Strategy,
		cfg.LoadBalancer.HealthCheck,
		cfg.LoadBalancer.MaxFailures,
		logger,
	)

	// Load nodes from config
	for _, nodeConfig := range cfg.Nodes {
		if err := lb.AddNode(nodeConfig.Address, nodeConfig.Weight); err != nil {
			logger.Warnf("Failed to add node %s: %v", nodeConfig.Address, err)
		} else {
			logger.Infof("Successfully added node %s with weight %d", nodeConfig.Address, nodeConfig.Weight)
		}
	}

	// Start health checking
	lb.StartHealthCheck()
	return lb, nil
}

func startGRPCServer(cfg *config.Config, gateway *gateway.BlockchainGateway, logger *logrus.Logger) (*grpc.Server, error) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Gateway.GRPCPort))
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %d: %w", cfg.Gateway.GRPCPort, err)
	}

	server := grpc.NewServer()
	mmnpb.RegisterTxServiceServer(server, gateway)
	mmnpb.RegisterHealthServiceServer(server, gateway)
	mmnpb.RegisterAccountServiceServer(server, gateway)
	reflection.Register(server)

	logger.Infof("Starting gRPC server on port %d", cfg.Gateway.GRPCPort)

	go func() {
		if err := server.Serve(lis); err != nil {
			logger.Errorf("Failed to serve gRPC: %v", err)
		}
	}()

	return server, nil
}
