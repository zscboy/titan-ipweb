package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"titan-ipweb/internal/middleware"
	"titan-ipweb/internal/svc"
	"titan-ipweb/internal/types"
	"titan-ipweb/ippmclient"
	"titan-ipweb/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAllStatsPerHourLogic struct {
	logx.Logger
	ctx     context.Context
	svcCtx  *svc.ServiceContext
	oneHour int64
}

func NewGetAllStatsPerHourLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAllStatsPerHourLogic {
	return &GetAllStatsPerHourLogic{
		Logger:  logx.WithContext(ctx),
		ctx:     ctx,
		svcCtx:  svcCtx,
		oneHour: 1 * 60 * 60,
	}
}

func (l *GetAllStatsPerHourLogic) GetAllStatsPerHour(req *types.AllStatsPerHourReq) (resp *types.StatsResp, err error) {
	logx.Debugf("GetAllStatsPer5Min %#v", req)
	v := l.ctx.Value(middleware.AuthKey)
	autCtxValue, ok := v.(middleware.AuthCtxValue)
	if !ok {
		return nil, fmt.Errorf("auth failed")
	}

	usernames, err := model.GetAllSubUsername(l.svcCtx.Redis, autCtxValue.UserId)
	if err != nil {
		return nil, err
	}

	if len(usernames) == 0 {
		return l.emtpyReply(req)
	}

	var (
		wg         sync.WaitGroup
		mu         sync.Mutex
		statsMap   map[string]*types.StatsResp
		firstError error
	)

	wg.Add(len(usernames))
	for _, username := range usernames {
		uname := username // 避免 goroutine 捕获错误变量
		go func() {
			defer wg.Done()

			statsResp, err := l.getUserStatsPerHour(uname, req.Hours)
			if err != nil {
				// logx.Error("getAllStatsPer5Min failed:%v", err.Error())
				// 只记录第一个错误
				mu.Lock()
				if firstError == nil {
					firstError = fmt.Errorf("popID %s: %w", uname, err)
				}
				mu.Unlock()
				return
			}

			mu.Lock()
			statsMap[uname] = statsResp
			mu.Unlock()
		}()
	}

	// 等待全部 goroutine 结束
	wg.Wait()

	if firstError != nil {
		return nil, firstError
	}

	count := req.Hours * 60 * 60 / int32(l.oneHour)
	stats := make([]*types.StatPoint, 0, count)
	for i := 0; i < int(count); i++ {
		stat := &types.StatPoint{}
		for _, statsResp := range statsMap {
			s := statsResp.Stats[i]

			stat.Bandwidth += s.Bandwidth
			stat.Traffic += s.Traffic

			if s.Timestamp != 0 {
				stat.Timestamp = s.Timestamp
			}
		}

		stats = append(stats, stat)

	}

	return &types.StatsResp{Stats: stats}, nil
}

func (l *GetAllStatsPerHourLogic) getUserStatsPerHour(username string, days int32) (resp *types.StatsResp, err error) {
	url := fmt.Sprintf("%s/user/stats/perhour?username=%s&hours=%d", l.svcCtx.Config.IPPMServer, username, days)

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

func (l *GetAllStatsPerHourLogic) emtpyReply(req *types.AllStatsPerHourReq) (resp *types.StatsResp, err error) {
	start := time.Now().Add(-time.Hour*time.Duration(req.Hours)).Unix() / int64(l.oneHour)
	end := time.Now().Unix() / int64(l.oneHour)

	count := req.Hours * 60 * 60 / int32(l.oneHour)
	stats := make([]*types.StatPoint, 0, count)
	for i := start; i <= end; i++ {
		ts := i * l.oneHour
		stat := &types.StatPoint{Timestamp: ts}
		stats = append(stats, stat)
	}
	return &types.StatsResp{Stats: stats}, nil
}
