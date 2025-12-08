package logic

import (
	"context"

	"titan-ipweb/internal/svc"
	"titan-ipweb/internal/types"

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
	return &types.GetTotalQuotaResponse{}, nil
}
