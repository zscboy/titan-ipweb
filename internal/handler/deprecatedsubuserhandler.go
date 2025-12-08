package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"titan-ipweb/internal/logic"
	"titan-ipweb/internal/svc"
	"titan-ipweb/internal/types"
)

// 废弃子用户
func DeprecatedSubUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DeprecatedSubUserReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewDeprecatedSubUserLogic(r.Context(), svcCtx)
		err := l.DeprecatedSubUser(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.Ok(w)
		}
	}
}
