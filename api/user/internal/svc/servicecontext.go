package svc

import (
	"context"
	"log"
	"net/url"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"GoZero-AI/api/user/internal/config"
	"GoZero-AI/api/user/model"
	"GoZero-AI/internal/llmclient"
)

type ServiceContext struct {
	Config                  config.Config
	DB                      sqlx.SqlConn
	UsersModel              model.UsersModel
	ChatSessionsModel       model.ChatSessionsModel
	SessionEvaluationsModel model.SessionEvaluationsModel
	EvaluationGenerator     *EvaluationGenerator
	ResumeStore             *ResumeStore
	PdfClient               *PdfClient
	RedisClient             *redis.Client
	RefreshTokenTTL         time.Duration
}

func NewServiceContext(c config.Config) *ServiceContext {
	sqlConn := sqlx.NewSqlConn("pgx", withPostgresConnectTimeout(c.Postgres.DataSource))
	evaluationClient, err := llmclient.NewClient(c.EvaluationEndpoint())
	if err != nil {
		log.Fatalf("初始化评估模型客户端失败: %v", err)
	}
	embeddingClient, err := llmclient.NewClient(c.EmbeddingEndpoint())
	if err != nil {
		log.Fatalf("初始化向量模型客户端失败: %v", err)
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr:         c.Redis.Host,
		Password:     c.RedisPassword(),
		DB:           c.Redis.DB,
		DialTimeout:  c.RedisDialTimeout(),
		ReadTimeout:  c.RedisReadTimeout(),
		WriteTimeout: c.RedisWriteTimeout(),
	})

	if c.Redis.Host != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := redisClient.Ping(ctx).Err(); err != nil {
			logx.Errorf("user service redis ping failed: %v", err)
		}
	}

	return &ServiceContext{
		Config:                  c,
		DB:                      sqlConn,
		UsersModel:              model.NewUsersModel(sqlConn),
		ChatSessionsModel:       model.NewChatSessionsModel(sqlConn),
		SessionEvaluationsModel: model.NewSessionEvaluationsModel(sqlConn),
		EvaluationGenerator:     NewEvaluationGenerator(evaluationClient, c),
		ResumeStore:             NewResumeStore(sqlConn, embeddingClient, c.EmbeddingModel()),
		PdfClient:               NewPdfClient(c.MCP.Endpoint, c.MCP.AuthToken, c.MCPMaxUploadBytes()),
		RedisClient:             redisClient,
		RefreshTokenTTL:         c.RefreshTokenTTL(),
	}
}

func withPostgresConnectTimeout(dataSource string) string {
	if dataSource == "" {
		return dataSource
	}

	parsed, err := url.Parse(dataSource)
	if err != nil {
		return dataSource
	}

	query := parsed.Query()
	if query.Get("connect_timeout") == "" {
		query.Set("connect_timeout", "3")
	}
	parsed.RawQuery = query.Encode()

	return parsed.String()
}
