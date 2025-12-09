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

type GetTotalQuotaLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取无效的子用户列表
func NewGetTotalQuotaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTotalQuotaLogic {
	return &GetTotalQuotaLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTotalQuotaLogic) GetTotalQuota() (resp *types.GetTotalQuotaResponse, err error) {
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

	subUserCount, err := model.SubUserCount(l.svcCtx.Redis, autCtxValue.UserId)
	if err != nil {
		return nil, err
	}

	return &types.GetTotalQuotaResponse{
			TotalBandwidthLimit:     user.MaxBandwidthLimit,
			TotalBandwidthAllocated: user.MaxBandwidthAllocated,
			TotalTrafficLimit:       user.TotalTrafficLimit,
			TotalTrafficAllocated:   user.TotalTrafficAllocated,
			SubUserCount:            int64(subUserCount),
		},
		nil
}
