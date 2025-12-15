package logic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"titan-ipweb/internal/middleware"
	"titan-ipweb/internal/svc"
	"titan-ipweb/internal/types"
	"titan-ipweb/ippmclient"
	"titan-ipweb/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeprecatedSubUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 废弃子用户
func NewDeprecatedSubUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeprecatedSubUserLogic {
	return &DeprecatedSubUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeprecatedSubUserLogic) DeprecatedSubUser(req *types.DeprecatedSubUserReq) error {
	logx.Debugf("DeleteUser %v", req)
	v := l.ctx.Value(middleware.AuthKey)
	autCtxValue, ok := v.(middleware.AuthCtxValue)
	if !ok {
		return fmt.Errorf("auth failed")
	}

	subUser, err := model.GetSubUser(l.svcCtx.Redis, req.Username)
	if err != nil {
		return err
	}

	if subUser == nil {
		return fmt.Errorf("sub user %s not exist", req.Username)
	}

	if subUser.UserID != autCtxValue.UserId {
		return fmt.Errorf("sub user %s not exist", req.Username)
	}

	if subUser.Status == subUserStatusDeprecated {
		return fmt.Errorf("sub user already deprecated")
	}

	if err := l.deprecatedSubUser(req); err != nil {
		return err
	}

	if err := model.AddSubUserToDeprecatedList(l.svcCtx.Redis, autCtxValue.UserId, req.Username); err != nil {
		return err
	}

	subUser.Status = subUserStatusDeprecated
	subUser.DeprecatedTime = time.Now().Unix()
	if err := model.SaveSubUser(l.svcCtx.Redis, subUser); err != nil {
		return err
	}

	user, err := model.GetUser(l.svcCtx.Redis, autCtxValue.UserId)
	if err != nil {
		return err
	}

	user.MaxBandwidthAllocated -= subUser.MaxBandwidthLimit
	user.TotalTrafficAllocated -= subUser.TotalTrafficLimit

	return model.SaveUser(l.svcCtx.Redis, user)
}

func (l *DeprecatedSubUserLogic) deprecatedSubUser(req *types.DeprecatedSubUserReq) error {
	url := fmt.Sprintf("%s/user/delete", l.svcCtx.Config.IPPMServer.URL)
	deleteUserReq := ippmclient.DeleteUserReq{
		UserName: req.Username,
	}

	buf, err := json.Marshal(deleteUserReq)
	if err != nil {
		return err
	}

	client := &http.Client{}
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(buf))
	if err != nil {
		return err
	}

	request.Header.Set("Authorization", "Bearer "+l.svcCtx.IPPMAcessToken)
	request.Header.Set("Content-Type", "application/json")

	httpResp, err := client.Do(request)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(httpResp.Body)
		return fmt.Errorf("http status code %d, error msg %s", httpResp.StatusCode, string(data))
	}

	data, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return err
	}

	userOperationResp := &ippmclient.UserOperationResp{}
	err = json.Unmarshal(data, userOperationResp)
	if err != nil {
		return err
	}

	if !userOperationResp.Success {
		return fmt.Errorf(userOperationResp.ErrMsg)
	}

	return nil
}
