package rpc

import (
	"log"

	"github.com/mezonai/mmn-api-gateway/internal/config"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

func StartServer(cfg *config.RpcConfig) (*zrpc.RpcServer, error) {
	gateway := NewGatewayService(cfg)
	server := zrpc.MustNewServer(cfg.RpcServerConf, func(grpcServer *grpc.Server) {
		Register(grpcServer, gateway)
	})

	log.Printf("Starting gRPC server on %s", cfg.ListenOn)

	// Start the server in a goroutine
	go func() {
		server.Start()
	}()

	return server, nil
}
