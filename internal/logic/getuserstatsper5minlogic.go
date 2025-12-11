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

type GetUserStatsPer5MinLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserStatsPer5MinLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserStatsPer5MinLogic {
	return &GetUserStatsPer5MinLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserStatsPer5MinLogic) GetUserStatsPer5Min(req *types.UserStats5PerMinReq) (resp *types.StatsResp, err error) {
	return l.getUserStatsPer5Min(req)
}

func (l *GetUserStatsPer5MinLogic) getUserStatsPer5Min(req *types.UserStats5PerMinReq) (resp *types.StatsResp, err error) {
	url := fmt.Sprintf("%s/user/stats/per5min?username=%s&minutes=%d", l.svcCtx.Config.IPPMServer, req.Username, req.Minutes)

	client := &http.Client{}
	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+l.svcCtx.IPPMAcessToken)
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(httpResp.Body)
		if len(data) > 0 {
			return nil, fmt.Errorf("%s", string(data))
		}
		return nil, fmt.Errorf("http status code %d", httpResp.StatusCode)
	}

	data, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}

	statsResp := &ippmclient.StatsResp{}
	err = json.Unmarshal(data, statsResp)
	if err != nil {
		return nil, err
	}

	stats := make([]*types.StatPoint, 0, len(statsResp.Stats))
	for _, statPonint := range statsResp.Stats {
		stat := &types.StatPoint{Timestamp: statPonint.Timestamp, Bandwidth: statPonint.Bandwidth, Traffic: statPonint.Traffic}
		stats = append(stats, stat)
	}

	return &types.StatsResp{Stats: stats}, nil
}
