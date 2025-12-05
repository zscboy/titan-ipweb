package auth

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"titan-ipweb/internal/logic/auth"
	"titan-ipweb/internal/svc"
)

// test
func TestHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := auth.NewTestLogic(r.Context(), svcCtx)
		resp, err := l.Test()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
