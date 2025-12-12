package auth

import (
	"context"
	"time"

	"titan-ipweb/internal/svc"
	"titan-ipweb/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type TestLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// test
func NewTestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TestLogic {
	return &TestLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TestLogic) Test() (resp string, err error) {
	// todo: add your logic here and delete this line
	uuid := "ebcdba5c-d1b1-11f0-9f28-afc4dc85d792"
	email := "zscboy@gmail.com"

	user, err := model.GetUser(l.svcCtx.Redis, uuid)
	if err != nil {
		return "", err
	}

	if user == nil {
		u := &model.User{UUID: uuid, Email: email, MaxBandwidthLimit: l.svcCtx.Config.Quota.MaxBandwidthLimit, TotalTrafficLimit: l.svcCtx.Config.Quota.TotalTrafficLimit}
		if err := model.SaveUser(l.svcCtx.Redis, u); err != nil {
			return "", err
		}
	}

	accessExpire := l.svcCtx.Config.TokenAuth.AccessExpire
	td, err := time.ParseDuration(accessExpire)
	if err != nil {
		td = 24 * time.Hour
	}
	token, err := generateToken(l.svcCtx.Config.TokenAuth.AccessSecret, uuid, email, td)
	if err != nil {
		return "", err
	}

	return token, nil
}
