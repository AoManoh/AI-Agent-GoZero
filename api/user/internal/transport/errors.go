package transport

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"GoZero-AI/internal/statuserr"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func InstallErrorHandlers() {
	httpx.SetErrorHandlerCtx(func(_ context.Context, err error) (int, any) {
		if err == nil {
			return http.StatusBadRequest, map[string]any{
				"message": "请求处理失败",
			}
		}

		statusCode := http.StatusInternalServerError
		message := "服务暂不可用，请稍后重试"
		switch {
		case func() bool {
			code, ok := statuserr.StatusCode(err)
			if ok {
				statusCode = code
				message = err.Error()
			}
			return ok
		}():
		case isInfraUnavailable(err):
			statusCode = http.StatusServiceUnavailable
			message = "后端依赖暂不可用，请稍后重试"
		case errors.Is(err, context.DeadlineExceeded):
			statusCode = http.StatusGatewayTimeout
			message = "请求超时，请稍后重试"
		case errors.Is(err, context.Canceled):
			statusCode = 499
			message = "请求已取消"
		}

		return statusCode, map[string]any{
			"message": message,
		}
	})
}

func isInfraUnavailable(err error) bool {
	if err == nil {
		return false
	}

	lower := strings.ToLower(err.Error())
	return strings.Contains(lower, "failed to connect to") ||
		strings.Contains(lower, "dial tcp") ||
		strings.Contains(lower, "connection refused") ||
		strings.Contains(lower, "i/o timeout")
}
