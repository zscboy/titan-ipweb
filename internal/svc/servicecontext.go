package svc

import (
	"time"
	"titan-ipweb/internal/config"
	"titan-ipweb/internal/middleware"
	"titan-ipweb/internal/pop"
	"titan-ipweb/user"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

const (
	tokenExpire = 100 * 24 * 60 * 60
	userIPweb   = "ipweb"
)

type ServiceContext struct {
	Config         config.Config
	Header         rest.Middleware
	UserAgent      rest.Middleware
	UserRpc        user.UserServiceClient
	Auth           rest.Middleware
	Redis          *redis.Redis
	IPPMAcessToken string
	PopManager     *pop.Manager
}

func NewServiceContext(c config.Config) *ServiceContext {
	authToken, err := generateJwtToken(c.IPPMServer.AccessSecret, tokenExpire, userIPweb)
	if err != nil {
		panic("get ippm access token error" + err.Error())
	}
	logx.Debugf("authToken:%s", string(authToken))

	popManager, err := pop.NewPopManager(c.IPPMServer.URL, string(authToken))
	if err != nil {
		panic("get ippm access token error" + err.Error())
	}

	return &ServiceContext{
		Config:         c,
		Header:         middleware.NewHeaderMiddleware().Handle,
		UserAgent:      middleware.NewUserAgentMiddleware().Handle,
		UserRpc:        user.NewUserServiceClient(zrpc.MustNewClient(c.UserRpc).Conn()),
		Auth:           middleware.NewAuthMiddleware(c.TokenAuth.AccessSecret).Handle,
		Redis:          redis.MustNewRedis(c.Redis),
		IPPMAcessToken: string(authToken),
		PopManager:     popManager,
		// Pops:           pops,
	}
}

func generateJwtToken(secret string, expire int64, user string) ([]byte, error) {
	claims := jwt.MapClaims{
		"user": user,
		"exp":  time.Now().Add(time.Second * time.Duration(expire)).Unix(),
		"iat":  time.Now().Add(-5 * time.Second).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		return nil, err
	}

	return []byte(tokenStr), nil
}
