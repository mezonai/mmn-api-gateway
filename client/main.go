package main

import (
	"context"
	"fmt"
	"log"
	"time"

	mmnClient "github.com/mezonai/mmn/client"
	mmnpb "github.com/mezonai/mmn/proto"
)

const (
	defaultEndpoint = "localhost:9090"
	numTestCalls    = 10
	callDelay       = 100 * time.Millisecond
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Fatal error: %v", err)
	}
}

func run() error {
	fmt.Println("🚀 Gateway gRPC Health Check Test")
	fmt.Println("==================================")
	fmt.Printf("Target: %s\n", defaultEndpoint)

	client, err := mmnClient.NewClient(mmnClient.Config{
		Endpoint: defaultEndpoint,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to gateway: %w", err)
	}
	defer client.Close()

	fmt.Println("\n🔄 Testing Round-Robin Load Balancing...")
	fmt.Println("==========================================")

	return runHealthCheckTest(client)
}

func runHealthCheckTest(client *mmnClient.MmnClient) error {
	for i := 1; i <= numTestCalls; i++ {
		fmt.Printf("\n--- Call %d ---\n", i)

		if err := performHealthCheck(client, i); err != nil {
			log.Printf("❌ Call %d failed: %v", i, err)
			continue
		}

		time.Sleep(callDelay)
	}

	fmt.Println("\n🎯 Round-Robin Test Completed!")
	fmt.Println("Check the logs above to see if different nodes are being selected.")
	return nil
}

func performHealthCheck(client *mmnClient.MmnClient, callNum int) error {
	resp, err := client.CheckHealth(context.Background())
	if err != nil {
		return fmt.Errorf("CheckHealth() error: %w", err)
	}

	if resp == nil {
		return fmt.Errorf("CheckHealth() returned nil response")
	}

	if resp.ErrorMessage != "" {
		return fmt.Errorf("error message: %s", resp.ErrorMessage)
	}

	// Validate response
	if err := validateHealthResponse(resp, callNum); err != nil {
		return err
	}

	// Log successful response
	logHealthResponse(resp, callNum)
	return nil
}

func validateHealthResponse(resp *mmnpb.HealthCheckResponse, callNum int) error {
	if resp.NodeId == "" {
		return fmt.Errorf("expected non-empty node ID")
	}

	if resp.Status == mmnpb.HealthCheckResponse_UNKNOWN {
		log.Printf("⚠️  Call %d: Warning: Node returned UNKNOWN status", callNum)
	}

	if resp.Status == mmnpb.HealthCheckResponse_NOT_SERVING {
		log.Printf("⚠️  Call %d: Warning: Node is NOT_SERVING", callNum)
	}

	if resp.CurrentSlot <= 0 {
		log.Printf("⚠️  Call %d: Warning: Node returned invalid current slot: %d", callNum, resp.CurrentSlot)
	}

	return nil
}

func logHealthResponse(resp *mmnpb.HealthCheckResponse, callNum int) {
	log.Printf("✅ Call %d - Status: %v, Node ID: %s, Slot: %d, Height: %d, Mempool: %d, Leader: %v, Follower: %v, Version: %s, Uptime: %ds",
		callNum, resp.Status, resp.NodeId, resp.CurrentSlot, resp.BlockHeight, resp.MempoolSize, resp.IsLeader, resp.IsFollower, resp.Version, resp.Uptime)
}
