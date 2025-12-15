package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	UserRpc   zrpc.RpcClientConf
	TokenAuth TokenAuth
	Redis     redis.RedisConf
	// IP pop manager server
	IPPMServer IPPMServer
	Quota      Quota
	RunMode    string `json:",default=prod"` // dev / test / prod
}

type TokenAuth struct {
	AccessSecret string
	AccessExpire string `json:",default='24h'"`
}

type Quota struct {
	MaxBandwidthLimit int64
	TotalTrafficLimit int64
}

type IPPMServer struct {
	URL          string
	AccessSecret string
}
