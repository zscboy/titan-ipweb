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

	return &types.ListSubUserResponse{Users: users, Total: total}, nil
}
