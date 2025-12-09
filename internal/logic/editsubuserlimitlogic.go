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

type EditSubUserLimitLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 编辑用户的流量配额与带宽限制
func NewEditSubUserLimitLogic(ctx context.Context, svcCtx *svc.ServiceContext) *EditSubUserLimitLogic {
	return &EditSubUserLimitLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *EditSubUserLimitLogic) EditSubUserLimit(req *types.EditSubUserLimitReq) error {
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
		return fmt.Errorf("user %s not exist", req.Username)
	}

	if subUser.UserID != autCtxValue.UserId {
		return fmt.Errorf("sub user %s not exist", req.Username)
	}

	if err := l.editSubUserLimit(req, subUser); err != nil {
		return err
	}

	if req.MaxBandwidthLimit != nil {
		subUser.MaxBandwidthLimit = *req.MaxBandwidthLimit
	}

	if req.TotalTrafficLimit != nil {
		subUser.TotalTrafficLimit = *req.TotalTrafficLimit
	}

	return model.SaveSubUser(l.svcCtx.Redis, subUser)
}

func (l *EditSubUserLimitLogic) editSubUserLimit(req *types.EditSubUserLimitReq, subUser *model.SubUser) error {
	url := fmt.Sprintf("%s/user/modify", l.svcCtx.Config.IPPMServer)
	modifyUserReq := ippmclient.ModifyUserReq{
		UserName: req.Username,
	}

	if req.TotalTrafficLimit != nil {
		trafficLimit := ippmclient.TrafficLimit{}
		trafficLimit.TotalTraffic = *req.TotalTrafficLimit
		trafficLimit.StartTime = subUser.StartTime
		trafficLimit.EndTime = subUser.EndTime
		modifyUserReq.TrafficLimit = &trafficLimit
	}

	buf, err := json.Marshal(modifyUserReq)
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
