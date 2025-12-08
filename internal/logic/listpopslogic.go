package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"titan-ipweb/internal/svc"
	"titan-ipweb/internal/types"
	"titan-ipweb/ippmclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListPopsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 拉取pops列表
func NewListPopsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPopsLogic {
	return &ListPopsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListPopsLogic) ListPops() (resp *types.ListPopsResponse, err error) {
	return l.listPops()
}

func (l *ListPopsLogic) listPops() (resp *types.ListPopsResponse, err error) {
	url := fmt.Sprintf("%s/pops", l.svcCtx.Config.IPPMServer)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+l.svcCtx.IPPMAcessToken)
	req.Header.Set("Content-Type", "application/json")

	httpResp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("http status code %d, error msg %s", httpResp.StatusCode, string(data))
	}

	data, err := io.ReadAll(httpResp.Body)
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
		pop := &types.Pop{ID: p.ID, Name: p.Name, Area: p.Area, Socks5Server: p.Socks5Addr}
		pops = append(pops, pop)
	}
	return &types.ListPopsResponse{Pops: pops}, nil
}
