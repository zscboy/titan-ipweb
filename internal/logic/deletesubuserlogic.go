package logic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"titan-ipweb/internal/middleware"
	"titan-ipweb/internal/svc"
	"titan-ipweb/internal/types"
	"titan-ipweb/ippmclient"
	"titan-ipweb/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteSubUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除socks5用户
func NewDeleteSubUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteSubUserLogic {
	return &DeleteSubUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteSubUserLogic) DeleteSubUser(req *types.DeleteSubUserReq) error {
	logx.Debugf("DeleteUser %v", req)
	v := l.ctx.Value(middleware.AuthKey)
	autCtxValue, ok := v.(middleware.AuthCtxValue)
	if !ok {
		return fmt.Errorf("auth failed")
	}

	if err := l.deleteSubUser(req); err != nil {
		return err
	}

	if err := model.RemoveSubUser(l.svcCtx.Redis, autCtxValue.UserId, req.Username); err != nil {
		return err
	}

	return nil
}

func (l *DeleteSubUserLogic) deleteSubUser(req *types.DeleteSubUserReq) error {
	url := fmt.Sprintf("%s/user/delete", l.svcCtx.Config.IPPMServer)
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
