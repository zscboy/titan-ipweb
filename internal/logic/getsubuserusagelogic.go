package logic

import (
	"context"
	"fmt"

	"titan-ipweb/internal/middleware"
	"titan-ipweb/internal/svc"
	"titan-ipweb/internal/types"
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

	subUsers, err := model.GetSubUsers(l.ctx, l.svcCtx.Redis, autCtxValue.UserId, 0, -1)
	if err != nil {
		return nil, err
	}

	logx.Debugf("subUsers len:%d", len(subUsers))

	activeCount := 0
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

		sUsers = append(sUsers, user)
	}

	// TODO:获取用户当前的流量

	subUserCount := &types.SubUserCount{Active: activeCount, Stop: len(subUsers) - activeCount, Deprecated: deprecatedCount}

	return &types.GetSubUserUsageResponse{
		SubUsers: sUsers,
		Count:    subUserCount,
	}, nil
}
