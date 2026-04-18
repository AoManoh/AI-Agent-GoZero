package user

import (
	"net/http"
	"strings"

	"GoZero-AI/api/user/internal/logic/user"
	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/api/user/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 用户退出登录
func LogoutHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.LogoutReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		ctx := user.WithAccessToken(r.Context(), bearerTokenFromHeader(r.Header.Get("Authorization")))
		l := user.NewLogoutLogic(ctx, svcCtx)
		resp, err := l.Logout(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

func bearerTokenFromHeader(headerValue string) string {
	const prefix = "bearer "

	value := strings.TrimSpace(headerValue)
	if len(value) < len(prefix) || strings.ToLower(value[:len(prefix)]) != prefix {
		return ""
	}

	return strings.TrimSpace(value[len(prefix):])
}
