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

type GetAllStatsPer5MinLogic struct {
	logx.Logger
	ctx         context.Context
	svcCtx      *svc.ServiceContext
	fiveMinutes int32
}

func NewGetAllStatsPer5MinLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAllStatsPer5MinLogic {
	return &GetAllStatsPer5MinLogic{
		Logger:      logx.WithContext(ctx),
		ctx:         ctx,
		svcCtx:      svcCtx,
		fiveMinutes: 5 * 60,
	}
}

func (l *GetAllStatsPer5MinLogic) GetAllStatsPer5Min(req *types.AllStatsPer5MinReq) (resp *types.StatsResp, err error) {
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
		return l.emptyReply(req)
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

			statsResp, err := l.getUserStatsPer5Min(uname, req.Minutes)
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
			mu.Unlock()
		}()
	}

	// 等待全部 goroutine 结束
	wg.Wait()

	if firstError != nil {
		return nil, firstError
	}

	count := req.Minutes * 60 / l.fiveMinutes
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

func (l *GetAllStatsPer5MinLogic) getUserStatsPer5Min(username string, minutes int32) (resp *types.StatsResp, err error) {
	url := fmt.Sprintf("%s/user/stats/per5min?username=%s&minutes=%d", l.svcCtx.Config.IPPMServer, username, minutes)

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

func (l *GetAllStatsPer5MinLogic) emptyReply(req *types.AllStatsPer5MinReq) (resp *types.StatsResp, err error) {
	start := time.Now().Add(-time.Minute*time.Duration(req.Minutes)).Unix() / int64(l.fiveMinutes)
	end := time.Now().Unix() / int64(l.fiveMinutes)

	count := req.Minutes * 60 / int32(l.fiveMinutes)
	stats := make([]*types.StatPoint, 0, count)
	for i := start; i <= end; i++ {
		ts := i * int64(l.fiveMinutes)
		stat := &types.StatPoint{Timestamp: ts}
		stats = append(stats, stat)
	}
	return &types.StatsResp{Stats: stats}, nil
}
