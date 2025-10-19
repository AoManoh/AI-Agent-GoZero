package main

import (
	"flag"
	"fmt"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"

	"GoZero-AI/api/chat/internal/config"
	"GoZero-AI/api/chat/internal/handler"
	"GoZero-AI/api/chat/internal/svc"
)

var configFile = flag.String("f", "etc/chat.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 为所有响应附加通配 CORS 头，解决前端题型跨域失败错误提示
	// "*" 在生产环境中存在安全风险，部署时应替换为具体的前端域名。
	server := rest.MustNewServer(c.RestConf, rest.WithCors("*"))
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
