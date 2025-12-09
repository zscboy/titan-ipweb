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

const RouteModeCustom = 4

type CreateSubUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func createSubUsername(index int64, username string) string {
	return fmt.Sprintf("%05d_%s", index, username)
}

// 创建socks5用户
func NewCreateSubUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateSubUserLogic {
	return &CreateSubUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateSubUserLogic) CreateSubUser(req *types.CreateSubUserReq) (resp *types.SubUser, err error) {
	logx.Debugf("CreateUser %#v", req)
	v := l.ctx.Value(middleware.AuthKey)
	autCtxValue, ok := v.(middleware.AuthCtxValue)
	if !ok {
		return nil, fmt.Errorf("auth failed")
	}

	user, err := model.GetUser(l.svcCtx.Redis, autCtxValue.UserId)
	if err != nil {
		return nil, err
	}

	req.Username = createSubUsername(user.Index, req.Username)

	sUser, err := model.GetSubUser(l.svcCtx.Redis, req.Username)
	if err != nil {
		return nil, err
	}
	if sUser != nil {
		return nil, fmt.Errorf("user %s already exist", req.Username)
	}

	createUserResp, err := l.createSubUser(req)
	if err != nil {
		return nil, err
	}

	subUser := &model.SubUser{
		Password:          createUserResp.Password,
		Username:          createUserResp.Username,
		ServerAddress:     createUserResp.ServerAddress,
		TotalTrafficLimit: createUserResp.TotalTrafficLimit,
		MaxBandwidthLimit: createUserResp.MaxBandwidthLimit,
		UploadRateLimit:   createUserResp.UploadRateLimit,
		DownloadRateLimit: createUserResp.DownloadRateLimit,
		CreateTime:        createUserResp.CreateTime,
		Status:            createUserResp.Status,
	}

	if err := model.SaveSubUser(l.svcCtx.Redis, subUser); err != nil {
		return nil, err
	}

	if err := model.AddSubUserToList(l.svcCtx.Redis, autCtxValue.UserId, subUser.Username); err != nil {
		return nil, err
	}

	// update user quota
	user.TotalBandwidthAllocated = createUserResp.MaxBandwidthLimit
	user.TotalTrafficAllocated = createUserResp.TotalTrafficLimit
	if err := model.SaveUser(l.svcCtx.Redis, user); err != nil {
		return nil, err
	}

	return createUserResp, nil
}

func (l *CreateSubUserLogic) createSubUser(req *types.CreateSubUserReq) (resp *types.SubUser, err error) {
	url := fmt.Sprintf("%s/user/create", l.svcCtx.Config.IPPMServer)
	createUserReq := ippmclient.CreateUserReq{
		UserName:          req.Username,
		Password:          req.Password,
		PopId:             req.PopId,
		Route:             &ippmclient.Route{Mode: RouteModeCustom},
		UploadRateLimit:   req.UploadRateLimit,
		DownloadRateLimit: req.DownloadRateLimit,
	}

	traffic := &ippmclient.TrafficLimit{
		StartTime:    time.Now().Unix(),
		EndTime:      time.Now().Add(30 * 24 * time.Hour).Unix(),
		TotalTraffic: req.TotalTrafficLimit,
	}
	createUserReq.TrafficLimit = traffic
	buf, err := json.Marshal(createUserReq)
	if err != nil {
		return nil, fmt.Errorf("marshal error %v", err)
	}

	client := &http.Client{}
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(buf))
	if err != nil {
		return nil, fmt.Errorf("NewRequest error %v", err)
	}

	request.Header.Set("Authorization", "Bearer "+l.svcCtx.IPPMAcessToken)
	request.Header.Set("Content-Type", "application/json")

	httpResp, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("http do error %v", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("http status code %d, error msg %s", httpResp.StatusCode, string(data))
	}

	data, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}

	createUserResp := &ippmclient.CreateUserResp{}
	err = json.Unmarshal(data, createUserResp)
	if err != nil {
		return nil, fmt.Errorf("unmarshal error %v", err)
	}

	subUser := &types.SubUser{
		Username:          createUserReq.UserName,
		Password:          createUserReq.Password,
		TotalTrafficLimit: createUserResp.TrafficLimit.TotalTraffic,
		MaxBandwidthLimit: req.MaxBandwidthLimit,
		ServerAddress:     l.getSocks5Addrss(createUserReq.PopId),
		UploadRateLimit:   createUserReq.UploadRateLimit,
		DownloadRateLimit: createUserReq.DownloadRateLimit,
		CreateTime:        time.Now().Unix(),
		Status:            "active",
		StartTime:         createUserResp.TrafficLimit.StartTime,
		EndTime:           createUserResp.TrafficLimit.EndTime,
	}

	return subUser, nil
}

func (l *CreateSubUserLogic) getSocks5Addrss(popID string) string {
	pops := l.svcCtx.Pops
	for _, pop := range pops {
		if pop.ID == popID {
			return pop.Socks5Server
		}
	}
	return ""
}
