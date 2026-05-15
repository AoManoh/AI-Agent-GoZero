package transport

import (
	"context"
	"errors"
	"net/http"

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

		statusCode := http.StatusBadRequest
		switch {
		case func() bool {
			code, ok := statuserr.StatusCode(err)
			if ok {
				statusCode = code
			}
			return ok
		}():
		case errors.Is(err, context.DeadlineExceeded):
			statusCode = http.StatusGatewayTimeout
		case errors.Is(err, context.Canceled):
			statusCode = 499
		}

		return statusCode, map[string]any{
			"message": err.Error(),
		}
	})
}
