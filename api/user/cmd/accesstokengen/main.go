package main

import (
	"flag"
	"fmt"
	"os"

	"GoZero-AI/api/user/internal/auth"
	"GoZero-AI/api/user/internal/config"

	"github.com/zeromicro/go-zero/core/conf"
)

func main() {
	configFile := flag.String("f", "api/user/etc/user-api.yaml", "the config file")
	userID := flag.Int64("user", 0, "the user id")
	username := flag.String("username", "", "the username")
	flag.Parse()

	if *userID <= 0 || *username == "" {
		fmt.Fprintln(os.Stderr, "usage: go run ./api/user/cmd/accesstokengen -f <config> -user <userId> -username <username>")
		os.Exit(2)
	}

	var c config.Config
	conf.MustLoad(*configFile, &c)

	pair, err := auth.IssueTokenPair(
		c.Auth.AccessSecret,
		c.AccessTokenTTL(),
		c.RefreshTokenTTL(),
		*userID,
		*username,
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	fmt.Println(pair.AccessToken)
}
