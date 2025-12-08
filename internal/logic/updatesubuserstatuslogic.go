package logic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"titan-ipweb/internal/svc"
	"titan-ipweb/internal/types"
	"titan-ipweb/ippmclient"
	"titan-ipweb/model"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	subUserStatusStart = "start"
	subUserStatusStop  = "stop"
)

type UpdateSubUserStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新子账户状态
func NewUpdateSubUserStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateSubUserStatusLogic {
	return &UpdateSubUserStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateSubUserStatusLogic) UpdateSubUserStatus(req *types.UpdateSubUserStatusReq) error {
	subUser, err := model.GetSubUser(l.svcCtx.Redis, req.Username)
	if err != nil {
		return err
	}
	if subUser == nil {
		return fmt.Errorf("user %s not exist", req.Username)
	}

	if req.Status != subUserStatusStart && req.Status != subUserStatusStop {
		return fmt.Errorf("user status %s is not %s or %s", req.Status, subUserStatusStart, subUserStatusStop)
	}

	if err := l.updateSubUserStatus(req); err != nil {
		return err
	}

	if req.Status == subUserStatusStart {
		subUser.Status = "active"
	} else if req.Status == subUserStatusStop {
		subUser.Status = "stop"
	}

	return model.SaveSubUser(l.svcCtx.Redis, subUser)
}

func (l *UpdateSubUserStatusLogic) updateSubUserStatus(req *types.UpdateSubUserStatusReq) error {
	url := fmt.Sprintf("%s/user/startorstop", l.svcCtx.Config.IPPMServer)
	startOrStopReq := ippmclient.StartOrStopUserReq{
		UserName: req.Username,
		Action:   req.Status,
	}

	buf, err := json.Marshal(startOrStopReq)
	if err != nil {
		return fmt.Errorf("marshal error %v", err)
	}

	client := &http.Client{}
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(buf))
	if err != nil {
		return fmt.Errorf("NewRequest error %v", err)
	}

	request.Header.Set("Authorization", "Bearer "+l.svcCtx.IPPMAcessToken)
	request.Header.Set("Content-Type", "application/json")

	httpResp, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("http do error %v", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(httpResp.Body)
		if len(data) != 0 {
			return fmt.Errorf(string(data))
		}
		return fmt.Errorf("ip manager server response status code %d", httpResp.StatusCode)
	}

	return nil
}
