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

type GetUserStatsPerHourLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserStatsPerHourLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserStatsPerHourLogic {
	return &GetUserStatsPerHourLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserStatsPerHourLogic) GetUserStatsPerHour(req *types.UserStatsPerHourReq) (resp *types.StatsResp, err error) {
	return l.getUserStatsPerHour(req)
}

func (l *GetUserStatsPerHourLogic) getUserStatsPerHour(req *types.UserStatsPerHourReq) (resp *types.StatsResp, err error) {
	url := fmt.Sprintf("%s/user/stats/perhour?username=%s&hours=%d", l.svcCtx.Config.IPPMServer, req.Username, req.Hours)

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
