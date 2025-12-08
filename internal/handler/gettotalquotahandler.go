package handler

import (
	"net/http"

	"titan-ipweb/internal/handler/utils"
	"titan-ipweb/internal/logic"
	"titan-ipweb/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取无效的子用户列表
func GetTotalQuotaHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewGetTotalQuotaLogic(r.Context(), svcCtx)
		resp, err := l.GetTotalQuota()
		if err != nil {
			httpx.OkJsonCtx(r.Context(), w, utils.Error(err))
		} else {
			httpx.OkJsonCtx(r.Context(), w, utils.Success(resp))
		}
	}
}
