package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	MMNRpc RpcConfig
}

type RpcConfig struct {
	zrpc.RpcServerConf
	Target        string
	ServiceConfig string
	MaxRetries    int
}
