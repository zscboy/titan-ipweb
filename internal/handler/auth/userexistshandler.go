package auth

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"titan-ipweb/internal/logic/auth"
	"titan-ipweb/internal/svc"
	"titan-ipweb/internal/types"
)

// 账号是否注册
func UserExistsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UserExistsReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := auth.NewUserExistsLogic(r.Context(), svcCtx)
		resp, err := l.UserExists(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
