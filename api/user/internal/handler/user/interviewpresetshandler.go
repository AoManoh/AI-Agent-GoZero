package user

import (
	"net/http"

	logic "GoZero-AI/api/user/internal/logic/user"
	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/api/user/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func InterviewPresetsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.InterviewPresetsReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewInterviewPresetsLogic(r.Context(), svcCtx)
		resp, err := l.InterviewPresets(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
