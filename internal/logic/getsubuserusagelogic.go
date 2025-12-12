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

type GetSubUserUsageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取子账号资源使用情况
func NewGetSubUserUsageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSubUserUsageLogic {
	return &GetSubUserUsageLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSubUserUsageLogic) GetSubUserUsage() (resp *types.GetSubUserUsageResponse, err error) {
	v := l.ctx.Value(middleware.AuthKey)
	autCtxValue, ok := v.(middleware.AuthCtxValue)
	if !ok {
		return nil, fmt.Errorf("auth failed")
	}

	deprecatedCount, err := model.DeprecatedSubUserCount(l.svcCtx.Redis, autCtxValue.UserId)
	if err != nil {
		return nil, err
	}

	// TODO: need to limit subuser number
	subUsers, err := model.GetSubUsers(l.ctx, l.svcCtx.Redis, autCtxValue.UserId, 0, -1)
	if err != nil {
		return nil, err
	}

	logx.Debugf("subUsers len:%d", len(subUsers))

	activeCount := 0
	usernames := make([]string, 0, len(subUsers))
	sUsers := make([]*types.SubUserUsage, 0, len(subUsers))
	for _, subUser := range subUsers {
		if subUser.Status == subUserStatusActive {
			activeCount++
		}

		user := &types.SubUserUsage{
			Username:          subUser.Username,
			MaxBandwidth:      subUser.MaxBandwidthLimit,
			TotalTrafficLimit: subUser.TotalTrafficLimit,
			Status:            subUser.Status,
		}

		usernames = append(usernames, subUser.Username)
		sUsers = append(sUsers, user)
	}

	// 获取所有用户当前的流量
	baseStatsRespMap, err := l.getBaseStatsForUsers(usernames)
	if err != nil {
		return nil, err
	}

	totalTraffic := int64(0)
	totalCurrentBandwidth := int64(0)

	for _, user := range sUsers {
		baseStatsResp, ok := baseStatsRespMap[user.Username]
		if ok {
			user.CurrentBandwidth = baseStatsResp.CurrentBandwidth
			user.TrafficUsed = baseStatsResp.TotalTraffic
			totalCurrentBandwidth += user.CurrentBandwidth
			totalTraffic += user.TrafficUsed
		}
	}

	subUserCount := &types.SubUserCount{Active: activeCount, Stop: len(subUsers) - activeCount, Deprecated: deprecatedCount}

	return &types.GetSubUserUsageResponse{
		SubUsers:              sUsers,
		Count:                 subUserCount,
		TotalTrafficUsed:      totalTraffic,
		TotalCurrentBandwidth: totalCurrentBandwidth,
	}, nil
}

func (l *GetSubUserUsageLogic) getBaseStatsForUsers(usernames []string) (map[string]*ippmclient.UserBaseStatsResp, error) {
	// 用waitgroup 拉取所有用户的基础统计
	statsMap := make(map[string]*ippmclient.UserBaseStatsResp)
	mu := sync.Mutex{}
	wg := sync.WaitGroup{}
	wg.Add(len(usernames))
	for _, username := range usernames {
		uname := username // 避免 goroutine 捕获错误变量
		go func() {
			defer wg.Done()

			statsResp, err := l.getUserBaseStats(uname)
			if err != nil {
				return
			}

			mu.Lock()
			statsMap[uname] = statsResp
			mu.Unlock()
		}()
	}

	wg.Wait()
	return statsMap, nil
}

func (l *GetSubUserUsageLogic) getUserBaseStats(username string) (resp *ippmclient.UserBaseStatsResp, err error) {
	url := fmt.Sprintf("%s/user/stats/base?username=%s", l.svcCtx.Config.IPPMServer, username)

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

	baseStatsResp := &ippmclient.UserBaseStatsResp{}
	err = json.Unmarshal(data, baseStatsResp)
	if err != nil {
		return nil, err
	}

	return baseStatsResp, nil
}
