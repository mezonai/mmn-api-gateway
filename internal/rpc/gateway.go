package rpc

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/mezonai/mmn-api-gateway/internal/config"
	mmnpb "github.com/mezonai/mmn/proto"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/resolver/manual"
)

// GatewayService handles requests and forwards them to appropriate gRPC nodes
type GatewayService struct {
	lbClient   *grpc.ClientConn
	maxRetries int
	mmnpb.UnimplementedTxServiceServer
	mmnpb.UnimplementedHealthServiceServer
	mmnpb.UnimplementedAccountServiceServer
}

func NewGatewayService(cfg *config.RpcConfig) *GatewayService {
	// Create load balancer client to connect to gRPC nodes
	lbClient, err := createLoadBalancerClient(cfg.Target, cfg.ServiceConfig)
	if err != nil {
		log.Fatalf("Failed to create load balancer client: %v", err)
	}

	return &GatewayService{
		lbClient:   lbClient,
		maxRetries: cfg.MaxRetries,
	}
}

func Register(grpcServer *grpc.Server, gateway *GatewayService) {
	mmnpb.RegisterTxServiceServer(grpcServer, gateway)
	mmnpb.RegisterHealthServiceServer(grpcServer, gateway)
	mmnpb.RegisterAccountServiceServer(grpcServer, gateway)
}

func createLoadBalancerClient(targets string, serviceConfig string) (*grpc.ClientConn, error) {
	// Create gRPC client with round-robin load balancing
	endpoints := strings.Split(targets, ",")
	addresses := make([]resolver.Address, 0, len(endpoints))

	for _, endpoint := range endpoints {
		endpoint = strings.TrimSpace(endpoint)
		if endpoint != "" {
			addresses = append(addresses, resolver.Address{Addr: endpoint})
		}
	}

	if len(addresses) == 0 {
		return nil, fmt.Errorf("no valid endpoints found in: %s", targets)
	}

	// Create manual resolver for multiple endpoints
	r := manual.NewBuilderWithScheme("mmn")
	r.InitialState(resolver.State{
		Addresses: addresses,
	})

	conn, err := grpc.NewClient(
		"mmn:///",
		grpc.WithResolvers(r),
		grpc.WithDefaultServiceConfig(serviceConfig),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to dial: %v", err)
	}

	return conn, nil
}

func (g *GatewayService) waitForRetry(ctx context.Context, attempt int, operation string) error {
	if attempt > 0 {
		log.Printf("Retrying %s, attempt %d/%d", operation, attempt, g.maxRetries)
		// Add exponential backoff delay
		backoffDelay := time.Duration(attempt) * 100 * time.Millisecond
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled: %w", ctx.Err())
		case <-time.After(backoffDelay):
		}
	}
	return nil
}

// for all gateway operations including connection management and retry attempts
func (g *GatewayService) executeWithRetry(
	ctx context.Context,
	operationName string,
	operation func(conn *grpc.ClientConn) (interface{}, error),
) (interface{}, error) {
	var lastErr error
	for attempt := 0; attempt <= g.maxRetries; attempt++ {
		if err := g.waitForRetry(ctx, attempt, operationName); err != nil {
			return nil, err
		}

		log.Printf("Target Node: %+v", g.lbClient.Target())

		// Execute the operation with the load balancer client
		result, err := operation(g.lbClient)
		if err != nil {
			lastErr = fmt.Errorf("%s failed: %w", operationName, err)
			log.Printf("Attempt %d: %s failed: %v", attempt+1, operationName, err)
			if attempt == g.maxRetries {
				return nil, fmt.Errorf("%s failed after %d attempts: %w", operationName, g.maxRetries+1, lastErr)
			}
			continue
		}

		// Success - return the result
		if attempt > 0 {
			log.Printf("%s succeeded after %d attempts", operationName, attempt+1)
		}

		return result, nil
	}

	return nil, fmt.Errorf("%s failed after %d attempts, last error: %w", operationName, g.maxRetries+1, lastErr)
}

func (g *GatewayService) AddTx(ctx context.Context, signedTx *mmnpb.SignedTxMsg) (*mmnpb.AddTxResponse, error) {
	log.Printf("AddTx called with request: %+v", signedTx)

	operation := func(conn *grpc.ClientConn) (interface{}, error) {
		txClient := mmnpb.NewTxServiceClient(conn)
		res, err := txClient.AddTx(ctx, signedTx)
		if err != nil {
			return nil, err
		}

		if !res.Ok {
			return nil, fmt.Errorf("add-tx failed: %s", res.Error)
		}

		return res, nil
	}

	result, err := g.executeWithRetry(ctx, "AddTx", operation)
	if err != nil {
		return nil, err
	}

	return result.(*mmnpb.AddTxResponse), nil
}

func (g *GatewayService) Check(ctx context.Context, req *mmnpb.Empty) (*mmnpb.HealthCheckResponse, error) {
	log.Printf("CheckHealth called with request: %+v", req)

	operation := func(conn *grpc.ClientConn) (interface{}, error) {
		healthClient := mmnpb.NewHealthServiceClient(conn)
		return healthClient.Check(ctx, &mmnpb.Empty{})
	}

	result, err := g.executeWithRetry(ctx, "health check", operation)
	if err != nil {
		return nil, err
	}

	return result.(*mmnpb.HealthCheckResponse), nil
}

func (g *GatewayService) GetAccount(ctx context.Context, req *mmnpb.GetAccountRequest) (*mmnpb.GetAccountResponse, error) {
	log.Printf("GetAccount called with request: %+v", req)

	operation := func(conn *grpc.ClientConn) (interface{}, error) {
		accountClient := mmnpb.NewAccountServiceClient(conn)
		return accountClient.GetAccount(ctx, req)
	}

	result, err := g.executeWithRetry(ctx, "GetAccount", operation)
	if err != nil {
		return nil, err
	}

	return result.(*mmnpb.GetAccountResponse), nil
}

func (g *GatewayService) GetTxHistory(ctx context.Context, req *mmnpb.GetTxHistoryRequest) (*mmnpb.GetTxHistoryResponse, error) {
	log.Printf("GetTxHistory called with request: %+v", req)

	operation := func(conn *grpc.ClientConn) (interface{}, error) {
		accountClient := mmnpb.NewAccountServiceClient(conn)
		return accountClient.GetTxHistory(ctx, req)
	}

	result, err := g.executeWithRetry(ctx, "GetTxHistory", operation)
	if err != nil {
		return nil, err
	}

	return result.(*mmnpb.GetTxHistoryResponse), nil
}
