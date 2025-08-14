package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Gateway      GatewayConfig      `yaml:"gateway"`
	LoadBalancer LoadBalancerConfig `yaml:"loadbalancer"`
	Nodes        []NodeConfig       `yaml:"nodes"`
}

type GatewayConfig struct {
	GRPCPort   int           `yaml:"grpc_port"`
	HTTPPort   int           `yaml:"http_port"`
	EnableHTTP bool          `yaml:"enable_http"`
	EnableGRPC bool          `yaml:"enable_grpc"`
	Timeout    time.Duration `yaml:"timeout"`
	MaxRetries int           `yaml:"max_retries"`
	LogLevel   string        `yaml:"log_level"`
}

type LoadBalancerConfig struct {
	Strategy    string        `yaml:"strategy"`
	HealthCheck time.Duration `yaml:"health_check_interval"`
	MaxFailures int           `yaml:"max_failures"`
}

type NodeConfig struct {
	Address string `yaml:"address"`
	Weight  int    `yaml:"weight"`
}

func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %v", err)
	}

	return &config, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate Gateway configuration
	if err := c.Gateway.Validate(); err != nil {
		return fmt.Errorf("gateway config: %v", err)
	}

	// Validate LoadBalancer configuration
	if err := c.LoadBalancer.Validate(); err != nil {
		return fmt.Errorf("loadbalancer config: %v", err)
	}

	// Validate Nodes configuration
	if err := c.validateNodes(); err != nil {
		return fmt.Errorf("nodes config: %v", err)
	}

	return nil
}

// Validate checks if the GatewayConfig is valid
func (gc *GatewayConfig) Validate() error {
	if gc.GRPCPort < 1 || gc.GRPCPort > 65535 {
		return fmt.Errorf("invalid gRPC port: %d (must be 1-65535)", gc.GRPCPort)
	}

	if gc.HTTPPort < 1 || gc.HTTPPort > 65535 {
		return fmt.Errorf("invalid HTTP port: %d (must be 1-65535)", gc.HTTPPort)
	}

	if gc.GRPCPort == gc.HTTPPort {
		return fmt.Errorf("gRPC and HTTP ports cannot be the same: %d", gc.GRPCPort)
	}

	if gc.Timeout < 1*time.Second {
		return fmt.Errorf("timeout too short: %v (minimum 1s)", gc.Timeout)
	}

	if gc.Timeout > 60*time.Second {
		return fmt.Errorf("timeout too long: %v (maximum 60s)", gc.Timeout)
	}

	if gc.MaxRetries < 0 || gc.MaxRetries > 10 {
		return fmt.Errorf("invalid max retries: %d (must be 0-10)", gc.MaxRetries)
	}

	// Validate log level
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[gc.LogLevel] {
		return fmt.Errorf("invalid log level: %s (must be debug, info, warn, or error)", gc.LogLevel)
	}

	return nil
}

// Validate checks if the LoadBalancerConfig is valid
func (lbc *LoadBalancerConfig) Validate() error {
	validStrategies := map[string]bool{
		"round_robin":       true,
		"least_connections": true,
		"health_check":      true,
	}
	if !validStrategies[lbc.Strategy] {
		return fmt.Errorf("invalid load balancing strategy: %s (must be round_robin, least_connections, or health_check)", lbc.Strategy)
	}

	if lbc.HealthCheck < 5*time.Second {
		return fmt.Errorf("health check interval too short: %v (minimum 5s)", lbc.HealthCheck)
	}

	if lbc.HealthCheck > 300*time.Second {
		return fmt.Errorf("health check interval too long: %v (maximum 5m)", lbc.HealthCheck)
	}

	if lbc.MaxFailures < 1 || lbc.MaxFailures > 10 {
		return fmt.Errorf("invalid max failures: %d (must be 1-10)", lbc.MaxFailures)
	}

	return nil
}

// validateNodes checks if the nodes configuration is valid
func (c *Config) validateNodes() error {
	if len(c.Nodes) == 0 {
		return fmt.Errorf("at least one node must be configured")
	}

	if len(c.Nodes) > 100 {
		return fmt.Errorf("too many nodes: %d (maximum 100)", len(c.Nodes))
	}

	// Check for duplicate addresses
	addresses := make(map[string]bool)
	for i, node := range c.Nodes {
		if err := node.Validate(); err != nil {
			return fmt.Errorf("node %d: %v", i+1, err)
		}

		if addresses[node.Address] {
			return fmt.Errorf("duplicate node address: %s", node.Address)
		}
		addresses[node.Address] = true
	}

	return nil
}

// Validate checks if the NodeConfig is valid
func (nc *NodeConfig) Validate() error {
	if nc.Address == "" {
		return fmt.Errorf("address cannot be empty")
	}

	if nc.Weight < 1 || nc.Weight > 100 {
		return fmt.Errorf("invalid weight: %d (must be 1-100)", nc.Weight)
	}

	return nil
}

// GetNodes returns the list of nodes from config
func (c *Config) GetNodes() []NodeConfig {
	return c.Nodes
}

// GetNodeCount returns the number of configured nodes
func (c *Config) GetNodeCount() int {
	return len(c.Nodes)
}
