package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"time"

	"GoZero-AI/api/user/internal/config"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/zeromicro/go-zero/core/conf"
)

func main() {
	configFile := flag.String("f", "api/user/etc/user-api.yaml", "the config file")
	sqlFile := flag.String("sql", "", "the sql file to apply")
	flag.Parse()

	if *sqlFile == "" {
		fmt.Fprintln(os.Stderr, "usage: go run ./api/user/cmd/dbapply -f <config> -sql <sql-file>")
		os.Exit(2)
	}

	content, err := os.ReadFile(*sqlFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	var c config.Config
	conf.MustLoad(*configFile, &c)

	db, err := sql.Open("pgx", c.Postgres.DataSource)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if _, err := db.ExecContext(ctx, string(content)); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	fmt.Println("ok")
}
