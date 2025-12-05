package svc

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"titan-ipweb/internal/config"
	"titan-ipweb/internal/middleware"
	"titan-ipweb/internal/types"
	"titan-ipweb/ippmclient"
	"titan-ipweb/user"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config         config.Config
	Header         rest.Middleware
	UserAgent      rest.Middleware
	UserRpc        user.UserServiceClient
	Auth           rest.Middleware
	Redis          *redis.Redis
	IPPMAcessToken string
	Pops           []*types.Pop
}

func NewServiceContext(c config.Config) *ServiceContext {
	authToken, err := getIPPMAccessToken(c.IPPMServer)
	if err != nil {
		panic("get ippm access token error" + err.Error())
	}

	pops, err := getPops(c.IPPMServer, authToken.Token)
	if err != nil {
		panic("get pops error" + err.Error())
	}

	return &ServiceContext{
		Config:         c,
		Header:         middleware.NewHeaderMiddleware().Handle,
		UserAgent:      middleware.NewUserAgentMiddleware().Handle,
		UserRpc:        user.NewUserServiceClient(zrpc.MustNewClient(c.UserRpc).Conn()),
		Auth:           middleware.NewAuthMiddleware(c.TokenAuth.AccessSecret).Handle,
		Redis:          redis.MustNewRedis(c.Redis),
		IPPMAcessToken: authToken.Token,
		Pops:           pops,
	}
}

func getIPPMAccessToken(ippmServer string) (*ippmclient.GetAuthTokenResp, error) {
	url := fmt.Sprintf("%s/auth/token", ippmServer)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("http status code %d, error msg %s", resp.StatusCode, string(data))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	authToken := ippmclient.GetAuthTokenResp{}
	err = json.Unmarshal(data, &authToken)
	if err != nil {
		return nil, err
	}

	return &authToken, nil
}

func getPops(ippmServer, accessToken string) ([]*types.Pop, error) {
	url := fmt.Sprintf("%s/pops", ippmServer)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("http status code %d, error msg %s", resp.StatusCode, string(data))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	popsResp := &ippmclient.GetPopsResp{}
	err = json.Unmarshal(data, popsResp)
	if err != nil {
		return nil, err
	}

	pops := make([]*types.Pop, 0, len(popsResp.Pops))
	for _, p := range popsResp.Pops {
		pop := &types.Pop{ID: p.ID, Area: p.Area, Socks5Server: p.Socks5Addr}
		pops = append(pops, pop)
	}
	return pops, nil
}
