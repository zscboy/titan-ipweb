package handler

import (
	"net/http"

	"titan-ipweb/internal/handler/utils"
	"titan-ipweb/internal/logic"
	"titan-ipweb/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取子账号资源使用情况
func GetSubUserUsageHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewGetSubUserUsageLogic(r.Context(), svcCtx)
		resp, err := l.GetSubUserUsage()
		if err != nil {
			httpx.OkJsonCtx(r.Context(), w, utils.Error(err))
		} else {
			httpx.OkJsonCtx(r.Context(), w, utils.Success(resp))
		}
	}
}
