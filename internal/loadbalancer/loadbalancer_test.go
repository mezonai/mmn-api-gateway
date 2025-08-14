package loadbalancer

import (
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestNewLoadBalancer(t *testing.T) {
	logger := logrus.New()
	lb := NewLoadBalancer("round_robin", 30*time.Second, 3, logger)

	if lb == nil {
		t.Fatal("LoadBalancer should not be nil")
	}

	if lb.strategy != "round_robin" {
		t.Errorf("Expected strategy 'round_robin', got '%s'", lb.strategy)
	}

	if lb.healthCheckDuration != 30*time.Second {
		t.Errorf("Expected health check interval 30s, got %v", lb.healthCheckDuration)
	}

	if lb.maxFailures != 3 {
		t.Errorf("Expected max failures 3, got %d", lb.maxFailures)
	}
}

func TestLoadBalancerAddNode(t *testing.T) {
	logger := logrus.New()
	lb := NewLoadBalancer("round_robin", 30*time.Second, 3, logger)

	// Test adding a valid node
	err := lb.AddNode("localhost:50051", 1)
	if err != nil {
		t.Errorf("Failed to add node: %v", err)
	}

	if len(lb.nodes) != 1 {
		t.Errorf("Expected 1 node, got %d", len(lb.nodes))
	}

	// Test adding another node
	err = lb.AddNode("localhost:50052", 2)
	if err != nil {
		t.Errorf("Failed to add second node: %v", err)
	}

	if len(lb.nodes) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(lb.nodes))
	}
}

func TestLoadBalancerRemoveNode(t *testing.T) {
	logger := logrus.New()
	lb := NewLoadBalancer("round_robin", 30*time.Second, 3, logger)

	// Add nodes first
	lb.AddNode("localhost:50051", 1)
	lb.AddNode("localhost:50052", 1)

	if len(lb.nodes) != 2 {
		t.Fatal("Failed to add nodes for testing")
	}

	// Remove a node
	lb.RemoveNode("localhost:50051")

	if len(lb.nodes) != 1 {
		t.Errorf("Expected 1 node after removal, got %d", len(lb.nodes))
	}

	// Remove the remaining node
	lb.RemoveNode("localhost:50052")

	if len(lb.nodes) != 0 {
		t.Errorf("Expected 0 nodes after removal, got %d", len(lb.nodes))
	}
}

func TestLoadBalancerGetStats(t *testing.T) {
	logger := logrus.New()
	lb := NewLoadBalancer("round_robin", 30*time.Second, 3, logger)

	// Add some nodes
	lb.AddNode("localhost:50051", 1)
	lb.AddNode("localhost:50052", 1)

	totalNodes, healthyNodes, strategy, nodesRaw := lb.GetStats()

	if totalNodes != 2 {
		t.Errorf("Expected total_nodes 2, got %v", totalNodes)
	}

	if strategy != "round_robin" {
		t.Errorf("Expected strategy 'round_robin', got %v", strategy)
	}

	if healthyNodes != 2 {
		t.Errorf("Expected healthy_nodes 2, got %v", healthyNodes)
	}

	if len(nodesRaw) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(nodesRaw))
	}
}

func TestLoadBalancerGetDetailedStats(t *testing.T) {
	logger := logrus.New()
	lb := NewLoadBalancer("round_robin", 30*time.Second, 3, logger)

	// Add some nodes
	lb.AddNode("localhost:50051", 1)
	lb.AddNode("localhost:50052", 2)

	totalNodes, healthyNodes, strategy, nodesRaw := lb.GetStats()

	if totalNodes != 2 {
		t.Errorf("Expected total_nodes 2, got %v", totalNodes)
	}

	if strategy != "round_robin" {
		t.Errorf("Expected strategy 'round_robin', got %v", strategy)
	}

	if healthyNodes != 2 {
		t.Errorf("Expected healthy_nodes 2, got %v", healthyNodes)
	}

	if len(nodesRaw) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(nodesRaw))
	}

	// Check first node details
	node1 := nodesRaw[0]
	if node1["address"] != "localhost:50051" {
		t.Errorf("Expected address localhost:50051, got %v", node1["address"])
	}
	if node1["weight"] != 1 {
		t.Errorf("Expected weight 1, got %v", node1["weight"])
	}
	if !node1["is_healthy"].(bool) {
		t.Error("Expected node1 to be healthy")
	}

	// Check second node details
	node2 := nodesRaw[1]
	if node2["address"] != "localhost:50052" {
		t.Errorf("Expected address localhost:50052, got %v", node2["address"])
	}
	if node2["weight"] != 2 {
		t.Errorf("Expected weight 2, got %v", node2["weight"])
	}
	if !node2["is_healthy"].(bool) {
		t.Error("Expected node2 to be healthy")
	}
}

func TestRoundRobinStrategy(t *testing.T) {
	logger := logrus.New()
	var current int32 = 0
	strategy := &RoundRobinStrategy{current: &current, logger: logger}

	// Create test nodes with proper initialization
	nodes := []*Node{
		{Address: "node1", IsHealthy: true, mu: sync.RWMutex{}},
		{Address: "node2", IsHealthy: true, mu: sync.RWMutex{}},
		{Address: "node3", IsHealthy: false, mu: sync.RWMutex{}}, // unhealthy
	}

	// Test selection
	selected := strategy.SelectNode(nodes)
	if selected == nil {
		t.Fatal("Should select a healthy node")
	}

	if selected.Address != "node1" {
		t.Errorf("Expected node1, got %s", selected.Address)
	}

	// Test next selection
	selected = strategy.SelectNode(nodes)
	if selected.Address != "node2" {
		t.Errorf("Expected node2, got %s", selected.Address)
	}
}

func TestHealthCheckStrategy(t *testing.T) {
	logger := logrus.New()
	strategy := &HealthCheckStrategy{logger: logger}

	now := time.Now()
	nodes := []*Node{
		{Address: "node1", IsHealthy: true, LastHealth: now.Add(-10 * time.Second), mu: sync.RWMutex{}},
		{Address: "node2", IsHealthy: true, LastHealth: now.Add(-5 * time.Second), mu: sync.RWMutex{}},
		{Address: "node3", IsHealthy: false, LastHealth: now.Add(-15 * time.Second), mu: sync.RWMutex{}},
	}

	selected := strategy.SelectNode(nodes)
	if selected == nil {
		t.Fatal("Should select a healthy node")
	}

	// Should select the most recently healthy node
	if selected.Address != "node2" {
		t.Errorf("Expected node2 (most recent), got %s", selected.Address)
	}
}

func TestLeastConnectionsStrategy(t *testing.T) {
	logger := logrus.New()
	strategy := &LeastConnectionsStrategy{logger: logger}

	nodes := []*Node{
		{Address: "node1", IsHealthy: true, mu: sync.RWMutex{}},
		{Address: "node2", IsHealthy: false, mu: sync.RWMutex{}}, // unhealthy
		{Address: "node3", IsHealthy: true, mu: sync.RWMutex{}},
	}

	selected := strategy.SelectNode(nodes)
	if selected == nil {
		t.Fatal("Should select a healthy node")
	}

	// Should select the first healthy node
	if selected.Address != "node1" {
		t.Errorf("Expected node1, got %s", selected.Address)
	}
}

func TestLoadBalancerGetConnection(t *testing.T) {
	logger := logrus.New()
	lb := NewLoadBalancer("round_robin", 30*time.Second, 3, logger)

	// Test with no nodes
	_, err := lb.GetConnection()
	if err == nil {
		t.Error("Expected error when no nodes available")
	}

	// Add a node
	err = lb.AddNode("localhost:50051", 1)
	if err != nil {
		t.Fatalf("Failed to add node: %v", err)
	}

	// Test getting connection
	client, err := lb.GetConnection()
	if err != nil {
		t.Errorf("Failed to get connection: %v", err)
	}
	if client == nil {
		t.Error("Expected client to be returned")
	}
}
