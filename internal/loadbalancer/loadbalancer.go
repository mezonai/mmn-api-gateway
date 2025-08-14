package loadbalancer

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	mmnpb "github.com/mezonai/mmn/proto"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Node struct {
	Address    string
	Weight     int
	Conn       *grpc.ClientConn
	LastHealth time.Time
	Failures   int32
	IsHealthy  bool
	mu         sync.RWMutex
}

type LoadBalancer struct {
	nodes               []*Node
	strategy            string
	current             int32
	mu                  sync.RWMutex
	healthCheckDuration time.Duration
	maxFailures         int
	logger              *logrus.Logger
	stopHealthCheck     chan struct{}
	wg                  sync.WaitGroup
}

type LoadBalancerStrategy interface {
	SelectNode(nodes []*Node) *Node
}

type RoundRobinStrategy struct {
	current *int32
	logger  *logrus.Logger
}

type LeastConnectionsStrategy struct {
	logger *logrus.Logger
}

type HealthCheckStrategy struct {
	logger *logrus.Logger
}

func NewLoadBalancer(strategy string, healthCheckDuration time.Duration, maxFailures int, logger *logrus.Logger) *LoadBalancer {
	return &LoadBalancer{
		strategy:            strategy,
		healthCheckDuration: healthCheckDuration,
		maxFailures:         maxFailures,
		logger:              logger,
		stopHealthCheck:     make(chan struct{}),
	}
}

func (lb *LoadBalancer) AddNode(address string, weight int) error {
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		return fmt.Errorf("failed to connect to node %s: %w", address, err)
	}

	node := &Node{
		Address:    address,
		Weight:     weight,
		Conn:       conn,
		LastHealth: time.Now(),
		IsHealthy:  true,
	}

	lb.mu.Lock()
	lb.nodes = append(lb.nodes, node)
	lb.mu.Unlock()

	lb.logger.Infof("Added node: %s with weight: %d", address, weight)
	return nil
}

func (lb *LoadBalancer) RemoveNode(address string) error {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	for i, node := range lb.nodes {
		if node.Address == address {
			if node.Conn != nil {
				node.Conn.Close()
			}
			lb.nodes = append(lb.nodes[:i], lb.nodes[i+1:]...)
			lb.logger.Infof("Removed node: %s", address)
			return nil
		}
	}
	return fmt.Errorf("node %s not found", address)
}

func (lb *LoadBalancer) GetConnection() (*grpc.ClientConn, error) {
	lb.mu.RLock()
	if len(lb.nodes) == 0 {
		lb.mu.RUnlock()
		return nil, fmt.Errorf("no available nodes")
	}

	// Copy nodes slice to avoid race condition
	nodes := make([]*Node, len(lb.nodes))
	copy(nodes, lb.nodes)
	lb.mu.RUnlock()

	var strategy LoadBalancerStrategy
	switch lb.strategy {
	case "round_robin":
		currentVal := atomic.LoadInt32(&lb.current)
		lb.logger.Infof("Creating RoundRobinStrategy with current index: %d", currentVal)
		strategy = &RoundRobinStrategy{current: &lb.current, logger: lb.logger}
	case "least_connections":
		strategy = &LeastConnectionsStrategy{logger: lb.logger}
	case "health_check":
		strategy = &HealthCheckStrategy{logger: lb.logger}
	default:
		currentVal := atomic.LoadInt32(&lb.current)
		lb.logger.Infof("Creating default RoundRobinStrategy with current index: %d", currentVal)
		strategy = &RoundRobinStrategy{current: &lb.current, logger: lb.logger}
	}

	node := strategy.SelectNode(nodes)
	if node == nil {
		return nil, fmt.Errorf("no healthy nodes available")
	}

	lb.logger.Infof("Selected node: %s", node.Address)
	return node.Conn, nil
}

func (lb *LoadBalancer) StartHealthCheck() {
	ticker := time.NewTicker(lb.healthCheckDuration)
	lb.wg.Add(1)
	go func() {
		defer lb.wg.Done()
		for {
			select {
			case <-ticker.C:
				lb.checkNodesHealth()
			case <-lb.stopHealthCheck:
				ticker.Stop()
				return
			}
		}
	}()
}

func (lb *LoadBalancer) StopHealthCheck() {
	close(lb.stopHealthCheck)
	lb.wg.Wait()
}

func (lb *LoadBalancer) checkNodesHealth() {
	lb.mu.RLock()
	nodes := make([]*Node, len(lb.nodes))
	copy(nodes, lb.nodes)
	lb.mu.RUnlock()

	for _, node := range nodes {
		lb.wg.Add(1)
		go func(n *Node) {
			defer lb.wg.Done()
			defer func() {
				if r := recover(); r != nil {
					lb.logger.Errorf("Panic in health check for node %s: %v", n.Address, r)
				}
			}()
			lb.checkNodeHealth(n)
		}(node)
	}
}

func (lb *LoadBalancer) checkNodeHealth(node *Node) {
	status, err := node.HealthCheck()
	if err != nil {
		lb.markNodeUnhealthy(node)
		lb.logger.Errorf("Error checking health for node %s: %v", node.Address, err)
	} else {
		lb.logger.Infof("Server State for node %s: %v", node.Address, status)
		if status == mmnpb.HealthCheckResponse_SERVING.String() {
			lb.markNodeHealthy(node)
		} else {
			lb.markNodeUnhealthy(node)
			if atomic.LoadInt32(&node.Failures) >= int32(lb.maxFailures) {
				lb.logger.Warnf("Node %s exceeded max failures, marking as unhealthy", node.Address)
			}
		}
	}
}

// Strategy implementations
func (rr *RoundRobinStrategy) SelectNode(nodes []*Node) *Node {
	if len(nodes) == 0 {
		return nil
	}

	start := atomic.LoadInt32(rr.current)
	rr.logger.Infof("RoundRobin: Starting selection from index %d (current: %d)", start, start)

	for i := 0; i < len(nodes); i++ {
		idx := int((start + int32(i)) % int32(len(nodes)))
		node := nodes[idx]

		node.mu.RLock()
		healthy := node.IsHealthy
		node.mu.RUnlock()

		rr.logger.Infof("RoundRobin: Checking node %d (%s) - Healthy: %v", idx, node.Address, healthy)

		if healthy {
			nextIndex := int32((idx + 1) % len(nodes))
			atomic.StoreInt32(rr.current, nextIndex)
			rr.logger.Infof("RoundRobin: Selected node %d (%s), next index will be %d", idx, node.Address, nextIndex)
			return node
		}
	}

	rr.logger.Warnf("RoundRobin: No healthy nodes found after checking all %d nodes", len(nodes))
	return nil
}

// Todo
func (lc *LeastConnectionsStrategy) SelectNode(nodes []*Node) *Node {
	var bestNode *Node
	var minWeight int = -1

	for _, node := range nodes {
		node.mu.RLock()
		if !node.IsHealthy {
			node.mu.RUnlock()
			continue
		}

		// Use weight as a simple heuristic for least connections
		// Lower weight means fewer connections
		if bestNode == nil || node.Weight < minWeight {
			bestNode = node
			minWeight = node.Weight
		}
		node.mu.RUnlock()
	}

	if bestNode != nil {
		lc.logger.Infof("LeastConnections: Selected node %s with weight %d", bestNode.Address, bestNode.Weight)
	}

	return bestNode
}

// Todo
func (hc *HealthCheckStrategy) SelectNode(nodes []*Node) *Node {
	var bestNode *Node
	var bestHealth time.Time

	for _, node := range nodes {
		node.mu.RLock()
		if !node.IsHealthy {
			node.mu.RUnlock()
			continue
		}

		if bestNode == nil || node.LastHealth.After(bestHealth) {
			bestNode = node
			bestHealth = node.LastHealth
		}
		node.mu.RUnlock()
	}

	if bestNode != nil {
		hc.logger.Infof("HealthCheck: Selected node %s with last health check at %v", bestNode.Address, bestNode.LastHealth)
	}

	return bestNode
}

func (lb *LoadBalancer) GetStats() (int, int, string, []map[string]interface{}) {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	totalNodes := len(lb.nodes)
	healthyCount := 0
	nodesInfo := make([]map[string]interface{}, 0, totalNodes)

	for _, node := range lb.nodes {
		node.mu.RLock()
		nodeInfo := map[string]interface{}{
			"address":     node.Address,
			"weight":      node.Weight,
			"is_healthy":  node.IsHealthy,
			"failures":    int(node.Failures),
			"last_health": node.LastHealth.Format(time.RFC3339),
		}
		node.mu.RUnlock()

		if nodeInfo["is_healthy"].(bool) {
			healthyCount++
		}

		nodesInfo = append(nodesInfo, nodeInfo)
	}

	return totalNodes, healthyCount, lb.strategy, nodesInfo
}

func (lb *LoadBalancer) markNodeUnhealthy(node *Node) {
	node.mu.Lock()
	node.IsHealthy = false
	node.Failures++
	node.mu.Unlock()
}

func (lb *LoadBalancer) markNodeHealthy(node *Node) {
	node.mu.Lock()
	node.IsHealthy = true
	node.Failures = 0
	node.LastHealth = time.Now()
	node.mu.Unlock()
}

// Close closes all connections and stops health checks
func (lb *LoadBalancer) Close() error {
	lb.StopHealthCheck()

	lb.mu.Lock()
	defer lb.mu.Unlock()

	for _, node := range lb.nodes {
		if node.Conn != nil {
			node.Conn.Close()
		}
	}

	lb.logger.Info("LoadBalancer closed successfully")
	return nil
}

func (n *Node) HealthCheck() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	healthClient := mmnpb.NewHealthServiceClient(n.Conn)
	response, err := healthClient.Check(ctx, &mmnpb.Empty{})
	if err != nil || response == nil {
		return "", err
	}

	return response.Status.String(), nil
}
