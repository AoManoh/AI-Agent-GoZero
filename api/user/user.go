package main

import (
	"flag"
	"fmt"
	"net/http"

	"GoZero-AI/api/user/internal/config"
	"GoZero-AI/api/user/internal/handler"
	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/api/user/internal/transport"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

var configFile = flag.String("f", "etc/user-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	transport.InstallErrorHandlers()

	server := rest.MustNewServer(c.RestConf, rest.WithUnauthorizedCallback(func(w http.ResponseWriter, r *http.Request, err error) {
		message := "未授权访问"
		if err != nil {
			message = err.Error()
		}
		httpx.WriteJsonCtx(r.Context(), w, http.StatusUnauthorized, map[string]any{
			"message": message,
		})
	}))
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
