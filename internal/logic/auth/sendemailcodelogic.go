package auth

import (
	"context"

	"titan-ipweb/internal/svc"
	"titan-ipweb/internal/types"
	"titan-ipweb/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendEmailCodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 发送验证码
func NewSendEmailCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendEmailCodeLogic {
	return &SendEmailCodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SendEmailCodeLogic) SendEmailCode(req *types.SendEmailCodeRequest) (resp *types.SendEmailCodeResponse, err error) {
	_, err = l.svcCtx.UserRpc.SendEmailVerificationCode(l.ctx, &user.SendEmailCodeRequest{
		Email:        req.Email,
		Purpose:      user.CodeType(req.Purpose),
		PointJson:    req.PointJson,
		CheckCaptcha: true,
	})

	if err != nil {
		logx.Errorf("send email code err: %w", err)
		return nil, err
	}
	return &types.SendEmailCodeResponse{}, nil
}
