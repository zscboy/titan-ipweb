package auth

import (
	"context"
	"time"

	"titan-ipweb/internal/svc"
	"titan-ipweb/internal/types"
	"titan-ipweb/model"
	"titan-ipweb/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginByGoogleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 谷歌登陆
func NewLoginByGoogleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginByGoogleLogic {
	return &LoginByGoogleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginByGoogleLogic) LoginByGoogle(req *types.LoginByGoogleRequest) (resp *types.LoginByGoogleResponse, err error) {
	res, err := l.svcCtx.UserRpc.LoginByGoogle(l.ctx, &user.LoginByGoogleRequest{
		Credential:  req.Credential,
		AccessToken: req.AccessToken,
	})

	if err != nil {
		logx.Errorf("login by google: %w", err)
		return nil, err
	}

	user, err := model.HGetUser(l.svcCtx.Redis, res.UserUuid)
	if err != nil {
		return nil, err
	}

	if user == nil {
		if err := model.HSetUser(l.svcCtx.Redis, &model.User{UUID: res.UserUuid, Email: res.Email}); err != nil {
			return nil, err
		}
	}

	accessExpire := l.svcCtx.Config.TokenAuth.AccessExpire
	td, err := time.ParseDuration(accessExpire)
	if err != nil {
		td = 24 * time.Hour
	}

	accessSecret := l.svcCtx.Config.TokenAuth.AccessSecret
	token, err := generateToken(accessSecret, res.UserUuid, res.Email, td)
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(td).Unix()

	return &types.LoginByGoogleResponse{
		AccessToken:  token,
		RefreshToken: res.RefreshToken,
		UserId:       res.UserUuid,
		Email:        res.Email,
		Role:         res.Role,
		// InviteCode:   userInfo.InviteCode,
		ExpiresAt: expiresAt,
	}, nil
}
