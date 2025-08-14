package gateway

import "time"

// StatsResponse represents the response structure for GetStats
type StatsResponse struct {
	TotalNodes   int        `json:"total_nodes"`
	HealthyNodes int        `json:"healthy_nodes"`
	Strategy     string     `json:"strategy"`
	Nodes        []NodeInfo `json:"nodes"`
}

// NodeInfo represents detailed information about a single node
type NodeInfo struct {
	Address    string    `json:"address"`
	Weight     int       `json:"weight"`
	IsHealthy  bool      `json:"is_healthy"`
	Failures   int       `json:"failures"`
	LastHealth time.Time `json:"last_health"`
}
