package handler

import (
	"net/http"

	"titan-ipweb/internal/handler/utils"
	"titan-ipweb/internal/logic"
	"titan-ipweb/internal/svc"
	"titan-ipweb/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取废弃的子用户列表
func ListDeprecatedSubUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListDeprecatedSubUserReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewListDeprecatedSubUserLogic(r.Context(), svcCtx)
		resp, err := l.ListDeprecatedSubUser(&req)
		if err != nil {
			httpx.OkJsonCtx(r.Context(), w, utils.Error(err))
		} else {
			httpx.OkJsonCtx(r.Context(), w, utils.Success(resp))
		}
	}
}
