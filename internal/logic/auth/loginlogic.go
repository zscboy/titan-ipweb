package auth

import (
	"context"
	"strings"
	"time"

	"titan-ipweb/internal/svc"
	"titan-ipweb/internal/types"
	"titan-ipweb/model"
	"titan-ipweb/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 登陆
func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginRequest) (resp *types.LoginResponse, err error) {
	req.UserId = strings.TrimSpace(req.UserId)

	res, err := l.svcCtx.UserRpc.LoginByEmail(l.ctx, &user.EmailLoginRequest{
		Email:            req.UserId,
		Password:         req.Password,
		VerificationCode: req.VerifyCode,
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
		if err := model.HSetUser(l.svcCtx.Redis, &model.User{UUID: res.UserUuid, Email: req.UserId}); err != nil {
			return nil, err
		}
	}

	accessExpire := l.svcCtx.Config.TokenAuth.AccessExpire
	td, err := time.ParseDuration(accessExpire)
	if err != nil {
		td = 24 * time.Hour
	}

	accessSecret := l.svcCtx.Config.TokenAuth.AccessSecret
	token, err := generateToken(accessSecret, res.UserUuid, req.UserId, td)

	expiresAt := time.Now().Add(td).Unix()

	return &types.LoginResponse{
		AccessToken:  token,
		RefreshToken: res.RefreshToken,
		UserId:       res.UserUuid,
		Email:        req.UserId,
		Role:         res.Role,
		// InviteCode:   userInfo.InviteCode,
		ExpiresAt: expiresAt,
	}, nil

}
