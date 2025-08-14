package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/mezonai/mmn-api-gateway/internal/config"
	"github.com/mezonai/mmn-api-gateway/internal/rpc"
	"github.com/zeromicro/go-zero/core/conf"
)

var configFile = flag.String("f", "etc/gateway.yaml", "the config file")

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	flag.Parse()
	var c config.Config
	conf.MustLoad(*configFile, &c)

	grpcServer, err := rpc.StartServer(&c.MMNRpc)
	if err != nil {
		return fmt.Errorf("failed to start gRPC server: %v", err)
	}

	defer grpcServer.Stop()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down servers...")
	if grpcServer != nil {
		grpcServer.Stop()
	}

	log.Println("Servers stopped successfully")
	return nil
}
