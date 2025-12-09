package auth

import (
	"context"
	"fmt"
	"strings"

	"titan-ipweb/internal/svc"
	"titan-ipweb/internal/types"
	"titan-ipweb/model"
	"titan-ipweb/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 注册账号
func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.UserRegisterReq) (resp *types.UserRegisterResp, err error) {
	// lang := l.ctx.Value(types.LanguageKey).(string)
	// clientIP := l.ctx.Value(types.ClientIPKey).(string)
	req.Email = strings.TrimSpace(req.Email)

	existsRes, err := l.svcCtx.UserRpc.UserExists(l.ctx, &user.UserExistsRequest{
		Email: req.Email,
	})

	if err != nil {
		logx.Errorf("call user-rpc failed, original error: %v", err)
		return nil, err
	}

	if existsRes.Exists {
		return nil, fmt.Errorf("user %s already exist", req.Email)
	}

	res, err := l.svcCtx.UserRpc.RegisterByEmail(l.ctx, &user.EmailRegisterRequest{
		Email:            req.Email,
		Password:         req.Password,
		VerificationCode: req.VerifyCode,
	})

	if err != nil {
		logx.Errorf("call user-rpc failed, original error: %v", err)
		return nil, err
	}

	index, err := model.UserIndex(l.svcCtx.Redis)
	if err != nil {
		return nil, err
	}

	user := &model.User{UUID: res.UserUuid, Email: req.Email, Index: index, TotalBandwidthLimit: l.svcCtx.Config.Quota.TotalTrafficLimit}

	if err := model.SaveUser(l.svcCtx.Redis, user); err != nil {
		return nil, err
	}

	return &types.UserRegisterResp{
		AccessToken:  res.AuthToken,
		RefreshToken: res.RefreshToken,
		UserId:       res.UserUuid,
		ExpiresAt:    res.ExpiresAt,
	}, nil
}
