package auth

import (
	"context"
	"fmt"
	"time"

	"titan-ipweb/internal/svc"
	"titan-ipweb/internal/types"
	"titan-ipweb/model"
	"titan-ipweb/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type RefreshTokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 刷新令牌
func NewRefreshTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshTokenLogic {
	return &RefreshTokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RefreshTokenLogic) RefreshToken(req *types.RefreshTokenRequest) (resp *types.RefreshTokenResponse, err error) {
	res, err := l.svcCtx.UserRpc.RefreshToken(l.ctx, &user.RefreshTokenRequest{
		RefreshToken: req.RefreshToken,
	})

	if err != nil {
		logx.Errorf("call user-rpc failed, original error: %v", err)
		return nil, err
	}

	user, err := model.HGetUser(l.svcCtx.Redis, res.UserUuid)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, fmt.Errorf("user %s not exist", res.UserUuid)
	}

	accessExpire := l.svcCtx.Config.TokenAuth.AccessExpire
	td, err := time.ParseDuration(accessExpire)
	if err != nil {
		td = 24 * time.Hour
	}

	accessSecret := l.svcCtx.Config.TokenAuth.AccessSecret
	token, err := generateToken(accessSecret, res.UserUuid, user.Email, td)

	expiresAt := time.Now().Add(td).Unix()

	return &types.RefreshTokenResponse{
		AccessToken:  token,
		RefreshToken: res.RefreshToken,
		UserId:       res.UserUuid,
		Role:         res.Role,
		ExpiresAt:    expiresAt,
	}, nil
}
