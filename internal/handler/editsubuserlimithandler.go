package handler

import (
	"net/http"

	"titan-ipweb/internal/handler/utils"
	"titan-ipweb/internal/logic"
	"titan-ipweb/internal/svc"
	"titan-ipweb/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 编辑用户的流量配额与带宽限制
func EditSubUserLimitHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.EditSubUserLimitReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewEditSubUserLimitLogic(r.Context(), svcCtx)
		err := l.EditSubUserLimit(&req)
		if err != nil {
			httpx.OkJsonCtx(r.Context(), w, utils.Error(err))
		} else {
			httpx.OkJsonCtx(r.Context(), w, utils.Success(nil))
		}
	}
}
