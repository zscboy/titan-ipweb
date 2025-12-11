package handler

import (
	"net/http"

	"titan-ipweb/internal/handler/utils"
	"titan-ipweb/internal/logic"
	"titan-ipweb/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取统计概要
func GetStatSummaryHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewGetStatSummaryLogic(r.Context(), svcCtx)
		resp, err := l.GetStatSummary()
		if err != nil {
			httpx.OkJsonCtx(r.Context(), w, utils.Error(err))
		} else {
			httpx.OkJsonCtx(r.Context(), w, utils.Success(resp))
		}
	}
}
