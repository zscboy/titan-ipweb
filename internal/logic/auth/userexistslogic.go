package auth

import (
	"context"

	"titan-ipweb/internal/svc"
	"titan-ipweb/internal/types"
	"titan-ipweb/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserExistsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 账号是否注册
func NewUserExistsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserExistsLogic {
	return &UserExistsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserExistsLogic) UserExists(req *types.UserExistsReq) (resp *types.UserExistsResp, err error) {
	res, err := l.svcCtx.UserRpc.UserExists(l.ctx, &user.UserExistsRequest{
		Email: req.Email,
	})

	if err != nil {
		logx.Errorf("call user-rpc failed, original error: %v", err)
		return nil, err
	}

	return &types.UserExistsResp{
		Exists: res.Exists,
	}, nil
}
