package gateway

import (
	"context"
	"fmt"
	"time"

	mmnpb "github.com/mezonai/mmn/proto"

	"github.com/mezonai/mmn-api-gateway/internal/loadbalancer"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type BlockchainGateway struct {
	mmnpb.UnimplementedTxServiceServer
	mmnpb.UnimplementedHealthServiceServer
	mmnpb.UnimplementedAccountServiceServer
	loadBalancer *loadbalancer.LoadBalancer
	logger       *logrus.Logger
	timeout      time.Duration
	maxRetries   int
}

func NewBlockchainGateway(lb *loadbalancer.LoadBalancer, timeout time.Duration, maxRetries int, logger *logrus.Logger) *BlockchainGateway {
	return &BlockchainGateway{
		loadBalancer: lb,
		logger:       logger,
		timeout:      timeout,
		maxRetries:   maxRetries,
	}
}

func (g *BlockchainGateway) AddTx(ctx context.Context, signedTx *mmnpb.SignedTxMsg) (*mmnpb.AddTxResponse, error) {
	g.logger.Infof("AddTx called with request: %+v", signedTx)

	var lastErr error
	for attempt := 0; attempt <= g.maxRetries; attempt++ {
		if attempt > 0 {
			g.logger.Infof("Retrying AddTx, attempt %d/%d", attempt, g.maxRetries)
			// Add exponential backoff delay
			backoffDelay := time.Duration(attempt) * 100 * time.Millisecond
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
			case <-time.After(backoffDelay):
			}
		}

		// Get connection from load balancer
		conn, err := g.loadBalancer.GetConnection()
		if err != nil {
			lastErr = fmt.Errorf("failed to get connection: %w", err)
			g.logger.Errorf("Attempt %d: %v", attempt+1, lastErr)
			if attempt == g.maxRetries {
				return nil, status.Errorf(codes.Unavailable, "failed to get connection after %d attempts: %v", g.maxRetries+1, lastErr)
			}
			continue
		}

		g.logger.Infof("Target Node: %+v", conn.Target())
		txClient := mmnpb.NewTxServiceClient(conn)
		res, err := txClient.AddTx(ctx, signedTx)
		if err != nil {
			lastErr = fmt.Errorf("add transaction failed: %w", err)
			g.logger.Errorf("Attempt %d: AddTx failed: %v", attempt+1, err)
			if attempt == g.maxRetries {
				return nil, fmt.Errorf("add transaction failed after %d attempts: %w", g.maxRetries+1, lastErr)
			}
			continue
		}

		if !res.Ok {
			lastErr = fmt.Errorf("add-tx failed: %s", res.Error)
			g.logger.Errorf("Attempt %d: %v", attempt+1, lastErr)
			if attempt == g.maxRetries {
				return nil, fmt.Errorf("add transaction failed after %d attempts: %w", g.maxRetries+1, lastErr)
			}
			continue
		}

		// Success - return the result
		if attempt > 0 {
			g.logger.Infof("AddTx succeeded after %d attempts", attempt+1)
		}
		return res, nil
	}

	return nil, fmt.Errorf("add transaction failed after %d attempts, last error: %w", g.maxRetries+1, lastErr)
}

func (g *BlockchainGateway) Check(ctx context.Context, req *mmnpb.Empty) (*mmnpb.HealthCheckResponse, error) {
	g.logger.Infof("CheckHealth called with request: %+v", req)

	var lastErr error
	for attempt := 0; attempt <= g.maxRetries; attempt++ {
		if attempt > 0 {
			g.logger.Infof("Retrying health check, attempt %d/%d", attempt, g.maxRetries)
			// Add exponential backoff delay
			backoffDelay := time.Duration(attempt) * 100 * time.Millisecond
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
			case <-time.After(backoffDelay):
			}
		}

		// Get connection from load balancer
		conn, err := g.loadBalancer.GetConnection()
		if err != nil {
			lastErr = fmt.Errorf("failed to get connection: %w", err)
			g.logger.Errorf("Attempt %d: %v", attempt+1, lastErr)
			if attempt == g.maxRetries {
				return nil, status.Errorf(codes.Unavailable, "failed to get connection after %d attempts: %v", g.maxRetries+1, lastErr)
			}
			continue
		}

		g.logger.Infof("Target Node: %+v", conn.Target())
		healthClient := mmnpb.NewHealthServiceClient(conn)
		res, err := healthClient.Check(ctx, &mmnpb.Empty{})
		if err != nil {
			lastErr = fmt.Errorf("health check failed: %w", err)
			g.logger.Errorf("Attempt %d: Health check failed: %v", attempt+1, err)
			if attempt == g.maxRetries {
				return nil, fmt.Errorf("health check failed after %d attempts: %w", g.maxRetries+1, lastErr)
			}
			continue
		}

		// Success - return the result
		if attempt > 0 {
			g.logger.Infof("Health check succeeded after %d attempts", attempt+1)
		}
		return res, nil
	}

	return nil, fmt.Errorf("health check failed after %d attempts, last error: %w", g.maxRetries+1, lastErr)
}

func (g *BlockchainGateway) GetAccount(ctx context.Context, req *mmnpb.GetAccountRequest) (*mmnpb.GetAccountResponse, error) {
	g.logger.Infof("GetAccount called with request: %+v", req)

	var lastErr error
	for attempt := 0; attempt <= g.maxRetries; attempt++ {
		if attempt > 0 {
			g.logger.Infof("Retrying GetAccount, attempt %d/%d", attempt, g.maxRetries)
			// Add exponential backoff delay
			backoffDelay := time.Duration(attempt) * 100 * time.Millisecond
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
			case <-time.After(backoffDelay):
			}
		}

		// Get connection from load balancer
		conn, err := g.loadBalancer.GetConnection()
		if err != nil {
			lastErr = fmt.Errorf("failed to get connection: %w", err)
			g.logger.Errorf("Attempt %d: %v", attempt+1, lastErr)
			if attempt == g.maxRetries {
				return nil, status.Errorf(codes.Unavailable, "failed to get connection after %d attempts: %v", g.maxRetries+1, lastErr)
			}
			continue
		}

		g.logger.Infof("Target Node: %+v", conn.Target())
		accountClient := mmnpb.NewAccountServiceClient(conn)
		res, err := accountClient.GetAccount(ctx, req)
		if err != nil {
			lastErr = fmt.Errorf("get account failed: %w", err)
			g.logger.Errorf("Attempt %d: GetAccount failed: %v", attempt+1, err)
			if attempt == g.maxRetries {
				return nil, fmt.Errorf("get account failed after %d attempts: %w", g.maxRetries+1, lastErr)
			}
			continue
		}

		// Success - return the result
		if attempt > 0 {
			g.logger.Infof("GetAccount succeeded after %d attempts", attempt+1)
		}
		return res, nil
	}

	return nil, fmt.Errorf("get account failed after %d attempts, last error: %w", g.maxRetries+1, lastErr)
}

func (g *BlockchainGateway) GetTxHistory(ctx context.Context, req *mmnpb.GetTxHistoryRequest) (*mmnpb.GetTxHistoryResponse, error) {
	g.logger.Infof("GetTxHistory called with request: %+v", req)

	var lastErr error
	for attempt := 0; attempt <= g.maxRetries; attempt++ {
		if attempt > 0 {
			g.logger.Infof("Retrying GetTxHistory, attempt %d/%d", attempt, g.maxRetries)
			// Add exponential backoff delay
			backoffDelay := time.Duration(attempt) * 100 * time.Millisecond
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
			case <-time.After(backoffDelay):
			}
		}

		// Get connection from load balancer
		conn, err := g.loadBalancer.GetConnection()
		if err != nil {
			lastErr = fmt.Errorf("failed to get connection: %w", err)
			g.logger.Errorf("Attempt %d: %v", attempt+1, lastErr)
			if attempt == g.maxRetries {
				return nil, status.Errorf(codes.Unavailable, "failed to get connection after %d attempts: %v", g.maxRetries+1, lastErr)
			}
			continue
		}

		g.logger.Infof("Target Node: %+v", conn.Target())
		accountClient := mmnpb.NewAccountServiceClient(conn)
		res, err := accountClient.GetTxHistory(ctx, req)
		if err != nil {
			lastErr = fmt.Errorf("get transaction history failed: %w", err)
			g.logger.Errorf("Attempt %d: GetTxHistory failed: %v", attempt+1, err)
			if attempt == g.maxRetries {
				return nil, fmt.Errorf("get transaction history failed after %d attempts: %w", g.maxRetries+1, lastErr)
			}
			continue
		}

		// Success - return the result
		if attempt > 0 {
			g.logger.Infof("GetTxHistory succeeded after %d attempts", attempt+1)
		}
		return res, nil
	}

	return nil, fmt.Errorf("get transaction history failed after %d attempts, last error: %w", g.maxRetries+1, lastErr)
}

// GetStats returns comprehensive health statistics for all nodes in the load balancer.
func (g *BlockchainGateway) GetStats(ctx context.Context) (*StatsResponse, error) {
	// Use typed method for better type safety
	totalNodes, healthyNodes, strategy, nodesRaw := g.loadBalancer.GetStats()

	// Convert raw nodes to typed NodeInfo
	nodes := make([]NodeInfo, 0, len(nodesRaw))
	for _, nodeRaw := range nodesRaw {
		nodeInfo, err := convertToNodeInfo(nodeRaw)
		if err != nil {
			g.logger.Warnf("Failed to convert node info: %v", err)
			continue // Skip invalid nodes but continue processing
		}
		nodes = append(nodes, nodeInfo)
	}

	// Create structured stats response
	response := &StatsResponse{
		TotalNodes:   totalNodes,
		HealthyNodes: healthyNodes,
		Strategy:     strategy,
		Nodes:        nodes,
	}

	return response, nil
}

// convertToNodeInfo safely converts a raw node map to NodeInfo
func convertToNodeInfo(nodeRaw map[string]interface{}) (NodeInfo, error) {
	address, ok := nodeRaw["address"].(string)
	if !ok {
		return NodeInfo{}, fmt.Errorf("invalid address type")
	}

	weight, ok := nodeRaw["weight"].(int)
	if !ok {
		return NodeInfo{}, fmt.Errorf("invalid weight type")
	}

	isHealthy, ok := nodeRaw["is_healthy"].(bool)
	if !ok {
		return NodeInfo{}, fmt.Errorf("invalid is_healthy type")
	}

	failures, ok := nodeRaw["failures"].(int)
	if !ok {
		return NodeInfo{}, fmt.Errorf("invalid failures type")
	}

	lastHealthStr, ok := nodeRaw["last_health"].(string)
	if !ok {
		return NodeInfo{}, fmt.Errorf("invalid last_health type")
	}

	lastHealth, err := time.Parse(time.RFC3339, lastHealthStr)
	if err != nil {
		return NodeInfo{}, fmt.Errorf("invalid last_health format: %v", err)
	}

	return NodeInfo{
		Address:    address,
		Weight:     weight,
		IsHealthy:  isHealthy,
		Failures:   failures,
		LastHealth: lastHealth,
	}, nil
}
