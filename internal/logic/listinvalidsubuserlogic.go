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

type ListInvalidSubUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取无效的子用户列表
func NewListInvalidSubUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListInvalidSubUserLogic {
	return &ListInvalidSubUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListInvalidSubUserLogic) ListInvalidSubUser() (resp *types.ListInvalidSubUserResponse, err error) {
	v := l.ctx.Value(middleware.AuthKey)
	autCtxValue, ok := v.(middleware.AuthCtxValue)
	if !ok {
		return nil, fmt.Errorf("auth failed")
	}

	subUsers, err := model.GetInvalidSubUsers(context.Background(), l.svcCtx.Redis, autCtxValue.UserId)
	if err != nil {
		return nil, err
	}

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
		users = append(users, user)
	}

	return &types.ListInvalidSubUserResponse{Users: users}, nil
}
