package user

import (
	"net/http"

	logic "GoZero-AI/api/user/internal/logic/user"
	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/api/user/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func ProfileHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ProfileReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewProfileLogic(r.Context(), svcCtx)
		resp, err := l.Profile(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
