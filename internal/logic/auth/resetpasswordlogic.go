package auth

import (
	"context"

	"titan-ipweb/internal/svc"
	"titan-ipweb/internal/types"
	"titan-ipweb/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type ResetPasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 重设密码
func NewResetPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetPasswordLogic {
	return &ResetPasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ResetPasswordLogic) ResetPassword(req *types.ResetPasswordRequest) (resp *types.ResetPasswordResponse, err error) {
	res, err := l.svcCtx.UserRpc.ResetPassword(l.ctx, &user.ResetPasswordRequest{
		Email:      req.Email,
		Password:   req.Password,
		VerifyCode: req.VerifyCode,
	})

	if err != nil {
		logx.Errorf("call user-rpc failed, original error: %v", err)
		return nil, err
	}

	return &types.ResetPasswordResponse{
		AccessToken:  res.AuthToken,
		RefreshToken: res.RefreshToken,
		UserId:       res.UserUuid,
		Role:         res.Role,
		ExpiresAt:    res.ExpiresAt,
	}, nil
}
