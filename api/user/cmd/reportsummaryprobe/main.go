package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"GoZero-AI/api/user/internal/config"
	logic "GoZero-AI/api/user/internal/logic/user"
	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/api/user/internal/types"

	"github.com/zeromicro/go-zero/core/conf"
)

func main() {
	configFile := flag.String("f", "api/user/etc/user-api.yaml", "the config file")
	sessionID := flag.String("session", "", "the session id")
	userID := flag.Int64("user", 0, "the user id")
	flag.Parse()

	if *sessionID == "" || *userID <= 0 {
		fmt.Fprintln(os.Stderr, "usage: go run ./api/user/cmd/reportsummaryprobe -f <config> -session <sessionId> -user <userId>")
		os.Exit(2)
	}

	var c config.Config
	conf.MustLoad(*configFile, &c)

	svcCtx := svc.NewServiceContext(c)
	ctx := context.WithValue(context.Background(), "userId", *userID)

	resp, err := logic.NewSessionReportSummaryLogic(ctx, svcCtx).SessionReportSummary(&types.SessionReportSummaryReq{
		Id: *sessionID,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	encoded, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	fmt.Println(string(encoded))
}
