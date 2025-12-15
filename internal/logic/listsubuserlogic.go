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

type ListSubUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 拉取socks5用户列表
func NewListSubUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListSubUserLogic {
	return &ListSubUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListSubUserLogic) ListSubUser(req *types.ListSubUserReq) (resp *types.ListSubUserResponse, err error) {
	v := l.ctx.Value(middleware.AuthKey)
	autCtxValue, ok := v.(middleware.AuthCtxValue)
	if !ok {
		return nil, fmt.Errorf("auth failed")
	}

	total, err := model.SubUserCount(l.svcCtx.Redis, autCtxValue.UserId)
	if err != nil {
		return nil, err
	}

	subUsers, err := model.GetSubUsers(context.Background(), l.svcCtx.Redis, autCtxValue.UserId, req.Start, req.End)
	if err != nil {
		return nil, err
	}

	usernames := make([]string, 0, len(subUsers))
	users := make([]*types.SubUser, 0, len(subUsers))
	for _, subUser := range subUsers {
		user := &types.SubUser{
			Username:          subUser.Username,
			Password:          subUser.Password,
			ServerAddress:     subUser.ServerAddress,
			TotalTrafficLimit: subUser.TotalTrafficLimit,
			MaxBandwidthLimit: subUser.MaxBandwidthLimit,
			UploadRateLimit:   subUser.UploadRateLimit,
			DownloadRateLimit: subUser.DownloadRateLimit,
			CreateTime:        subUser.CreateTime,
			Status:            subUser.Status,
		}

		pop, err := l.svcCtx.PopManager.Get(subUser.PopID)
		if err == nil {
			user.AreaName = pop.Name
		} else {
			logx.Debugf("get pop %v", err.Error())
		}

		users = append(users, user)
		usernames = append(usernames, subUser.Username)
	}

	baseStatsMap, err := l.getBaseStatsForUsers(usernames)
	if err != nil {
		return nil, err
	}

	for _, subUser := range users {
		baseStats, ok := baseStatsMap[subUser.Username]
		if ok {
			subUser.CurrentTraffic = baseStats.TotalTraffic
		}
	}

	return &types.ListSubUserResponse{Users: users, Total: total}, nil
}

func (l *ListSubUserLogic) getBaseStatsForUsers(usernames []string) (map[string]*ippmclient.UserBaseStatsResp, error) {
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

func (l *ListSubUserLogic) getUserBaseStats(username string) (resp *ippmclient.UserBaseStatsResp, err error) {
	url := fmt.Sprintf("%s/user/stats/base?username=%s", l.svcCtx.Config.IPPMServer.URL, username)

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
