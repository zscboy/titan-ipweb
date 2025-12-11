package logic

import (
	"context"

	"titan-ipweb/internal/svc"
	"titan-ipweb/internal/types"

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
	// todo: add your logic here and delete this line

	return &types.GetStatSummaryResponse{}, nil
}
