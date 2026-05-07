package svc

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"

	"GoZero-AI/api/chat/internal/config"
	"GoZero-AI/internal/llmclient"

	"github.com/sashabaranov/go-openai"
)

// ServiceContext 服务上下文结构体
// 集中管理所有服务级别的依赖，包括配置、OpenAI 客户端和会话存储
type ServiceContext struct {
	Config       config.Config  // 依赖1：配置
	OpenAIClient *openai.Client // 依赖2：OpenAI 客户端
	// SessionStore types.SessionStore // 依赖3：会话存储 -- v1版本，已淘汰
	VectorStore *VectorStore  // 依赖3：向量存储
	PdfClient   *PdfClient    // 依赖4：PDF 客户端
	RedisClient *redis.Client // 依赖5：Redis 客户端
}

// NewServiceContext 创建服务上下文实例
// 初始化所有服务依赖，包括 OpenAI 客户端配置和内存会话存储
func NewServiceContext(c config.Config) *ServiceContext {
	// 1. 初始化对话模型客户端
	openAIClient, err := llmclient.NewClient(c.ChatEndpoint())
	if err != nil {
		log.Fatalf("初始化对话模型客户端失败: %v", err)
	}

	// 2. 初始化向量模型客户端，允许 embedding 与对话模型使用不同凭证。
	embeddingClient, err := llmclient.NewClient(c.EmbeddingEndpoint())
	if err != nil {
		log.Fatalf("初始化向量模型客户端失败: %v", err)
	}

	// 3. 初始化向量存储，读取配置文件中的向量数据库配置和 OpenAI 客户端，并初始化向量存储
	vectorDB := c.VectorDB
	vectorDB.EmbeddingModel = c.EmbeddingModel()
	vectorStore, err := NewVectorStore(vectorDB, embeddingClient)
	if err != nil {
		log.Fatalf("初始化向量存储失败: %v", err)
	}

	// 4. 测试向量存储连接
	if err := vectorStore.TestConnection(); err != nil {
		log.Fatalf("向量存储连接测试失败: %v", err)
	} else {
		log.Println("向量存储连接测试成功")
	}

	// 新增. 初始化 Redis 客户端
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port),
		Password: c.Redis.Password,
		DB:       c.Redis.DB,
	})

	// 新增：测试 Redis 连接
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Redis 连接测试失败: %v", err)
	} else {
		log.Println("Redis 连接测试成功")
	}

	// 5. 返回服务上下文实例
	return &ServiceContext{
		Config:       c,                                                                    // 基础配置
		OpenAIClient: openAIClient,                                                         // OpenAI 客户端
		VectorStore:  vectorStore,                                                          // 向量存储
		PdfClient:    NewPdfClient(c.MCP.Endpoint, c.MCP.AuthToken, c.MCPMaxUploadBytes()), // PDF 客户端
		RedisClient:  redisClient,                                                          // Redis 客户端
	}
}
