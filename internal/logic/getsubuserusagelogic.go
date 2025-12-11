package logic

import (
	"context"

	"titan-ipweb/internal/svc"
	"titan-ipweb/internal/types"

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
	// todo: add your logic here and delete this line

	return &types.GetSubUserUsageResponse{}, nil
}
