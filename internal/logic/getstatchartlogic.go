package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"titan-ipweb/internal/middleware"
	"titan-ipweb/internal/svc"
	"titan-ipweb/internal/types"
	"titan-ipweb/ippmclient"
	"titan-ipweb/model"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	chartTypeMinute = "minute"
	charTypeHour    = "hour"
	chartTypeDay    = "day"
)

type GetStatChartLogic struct {
	logx.Logger
	ctx         context.Context
	svcCtx      *svc.ServiceContext
	fiveMinutes int32
	oneHour     int32
	oneDay      int32
}

// 获取趋势图
func NewGetStatChartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetStatChartLogic {
	return &GetStatChartLogic{
		Logger:      logx.WithContext(ctx),
		ctx:         ctx,
		svcCtx:      svcCtx,
		fiveMinutes: 5 * 60,
		oneHour:     1 * 60 * 60,
		oneDay:      24 * 60 * 60,
	}
}

func (l *GetStatChartLogic) GetStatChart(req *types.StatChartReq) (resp *types.StatChartResponse, err error) {
	logx.Debugf("GetAllStatsPer5Min %#v", req)
	v := l.ctx.Value(middleware.AuthKey)
	autCtxValue, ok := v.(middleware.AuthCtxValue)
	if !ok {
		return nil, fmt.Errorf("auth failed")
	}

	if req.Username != "" {
		subUser, err := model.GetSubUser(l.svcCtx.Redis, req.Username)
		if err != nil {
			return nil, err
		}

		if subUser == nil {
			return nil, fmt.Errorf("username %s not exist", req.Username)
		}

		if subUser.UserID != autCtxValue.UserId {
			return nil, fmt.Errorf("subuser username %s not exist for user %s", req.Username, autCtxValue.Email)
		}
		return l.getStatChartForSingleUser(req, req.Username)
	}

	usernames, err := model.GetAllSubUsername(l.svcCtx.Redis, autCtxValue.UserId)
	if err != nil {
		return nil, err
	}

	if len(usernames) == 0 {
		return l.emptyReply(req)
	}

	var (
		wg         sync.WaitGroup
		mu         sync.Mutex
		firstError error
		count      int
	)

	statsMap := make(map[string]*types.StatChartResponse)

	wg.Add(len(usernames))
	for _, username := range usernames {
		uname := username // 避免 goroutine 捕获错误变量
		go func() {
			defer wg.Done()

			statsResp, err := l.getStatChartForSingleUser(req, username)
			if err != nil {
				// logx.Error("getAllStatsPer5Min failed:%v", err.Error())
				// 只记录第一个错误
				mu.Lock()
				if firstError == nil {
					firstError = fmt.Errorf("get user %s stats per 5min: %w", uname, err)
				}
				mu.Unlock()
				return
			}

			mu.Lock()
			statsMap[uname] = statsResp
			count = len(statsResp.Stats)
			mu.Unlock()
		}()
	}

	// 等待全部 goroutine 结束
	wg.Wait()

	if firstError != nil {
		return nil, firstError
	}

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

	return &types.StatChartResponse{Stats: stats}, nil

}

func (l *GetStatChartLogic) getStatChartForSingleUser(req *types.StatChartReq, username string) (resp *types.StatChartResponse, err error) {
	url := fmt.Sprintf("%s/user/stats/chart?type=%s&username=%s&start_time=%d&end_time=%d", l.svcCtx.Config.IPPMServer.URL, req.Type, username, req.StartTime, req.EndTime)

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

	return &types.StatChartResponse{Stats: stats}, nil
}

func (l *GetStatChartLogic) emptyReply(req *types.StatChartReq) (resp *types.StatChartResponse, err error) {
	switch req.Type {
	case chartTypeMinute:
		return l.replyEmtpyForMinute(req)
	case charTypeHour:
		return l.replyEmtpyForHour(req)
	case chartTypeDay:
		return l.replyEmtpyForDay(req)
	}

	return nil, fmt.Errorf("invalid type %s", req.Type)
}

func (l *GetStatChartLogic) replyEmtpyForMinute(req *types.StatChartReq) (resp *types.StatChartResponse, err error) {
	start := req.StartTime / int64(l.fiveMinutes)
	end := req.EndTime / int64(l.fiveMinutes)

	count := int32(req.EndTime-req.StartTime) / int32(l.fiveMinutes)
	stats := make([]*types.StatPoint, 0, count)
	for i := start; i <= end; i++ {
		ts := i * int64(l.fiveMinutes)
		stat := &types.StatPoint{Timestamp: ts}
		stats = append(stats, stat)
	}
	return &types.StatChartResponse{Stats: stats}, nil
}

func (l *GetStatChartLogic) replyEmtpyForHour(req *types.StatChartReq) (resp *types.StatChartResponse, err error) {
	start := req.StartTime / int64(l.oneHour)
	end := req.EndTime / int64(l.oneHour)

	count := int32(req.EndTime-req.StartTime) / int32(l.oneHour)
	stats := make([]*types.StatPoint, 0, count)
	for i := start; i <= end; i++ {
		ts := i * int64(l.oneHour)
		stat := &types.StatPoint{Timestamp: ts}
		stats = append(stats, stat)
	}
	return &types.StatChartResponse{Stats: stats}, nil
}

func (l *GetStatChartLogic) replyEmtpyForDay(req *types.StatChartReq) (resp *types.StatChartResponse, err error) {
	start := req.StartTime / int64(l.oneDay)
	end := req.EndTime / int64(l.oneDay)

	count := int32(req.EndTime-req.StartTime) / int32(l.oneDay)
	stats := make([]*types.StatPoint, 0, count)
	for i := start; i <= end; i++ {
		ts := i * int64(l.oneDay)
		stat := &types.StatPoint{Timestamp: ts}
		stats = append(stats, stat)
	}
	return &types.StatChartResponse{Stats: stats}, nil
}
