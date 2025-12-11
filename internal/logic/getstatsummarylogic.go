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

type GetStatSummaryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取统计概要
func NewGetStatSummaryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetStatSummaryLogic {
	return &GetStatSummaryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetStatSummaryLogic) GetStatSummary() (resp *types.GetStatSummaryResponse, err error) {
	v := l.ctx.Value(middleware.AuthKey)
	autCtxValue, ok := v.(middleware.AuthCtxValue)
	if !ok {
		return nil, fmt.Errorf("auth failed")
	}

	user, err := model.GetUser(l.svcCtx.Redis, autCtxValue.UserId)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, fmt.Errorf("user not exist, please login again")
	}

	deprecatedCount, err := model.DeprecatedSubUserCount(l.svcCtx.Redis, autCtxValue.UserId)
	if err != nil {
		return nil, err
	}

	// TODO: Batch retrieval
	subUsers, err := model.GetSubUsers(l.ctx, l.svcCtx.Redis, autCtxValue.UserId, 0, -1)
	if err != nil {
		return nil, err
	}

	activeCount := 0
	for _, subUser := range subUsers {
		if subUser.Status == subUserStatusActive {
			activeCount++
		}
	}
	stopCount := len(subUsers) - activeCount

	return &types.GetStatSummaryResponse{
		TotalTrafficLimit:     user.TotalTrafficLimit,
		TotalTrafficAllocated: user.TotalTrafficAllocated,
		MaxBandwidthLimit:     user.MaxBandwidthAllocated,
		MaxBandwidthAllocated: user.MaxBandwidthAllocated,
		SubUserCount: &types.SubUserCount{
			Active:     activeCount,
			Stop:       stopCount,
			Deprecated: deprecatedCount,
		},
	}, nil

}
