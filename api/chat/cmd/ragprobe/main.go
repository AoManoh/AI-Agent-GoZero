package main

import (
	"fmt"
	"log"

	chatcfg "GoZero-AI/api/chat/internal/config"
	chatsvc "GoZero-AI/api/chat/internal/svc"

	"github.com/sashabaranov/go-openai"
)

func main() {
	cfg := chatcfg.Config{}
	cfg.OpenAI.ApiKey = "sk-f4eac47cba374fe6b18f7231c40c543b"
	cfg.OpenAI.BaseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
	cfg.VectorDB.Host = "127.0.0.1"
	cfg.VectorDB.Port = 5432
	cfg.VectorDB.DBName = "gozero_ai_agent"
	cfg.VectorDB.User = "root"
	cfg.VectorDB.Password = "020926ouhao."
	cfg.VectorDB.Table = "vector_store"
	cfg.VectorDB.MaxConn = 20
	cfg.VectorDB.EmbeddingModel = "text-embedding-v1"
	cfg.VectorDB.TimeZone = "Asia/Shanghai"

	openCfg := openai.DefaultConfig(cfg.OpenAI.ApiKey)
	openCfg.BaseURL = cfg.OpenAI.BaseURL
	client := openai.NewClientWithConfig(openCfg)

	vs, err := chatsvc.NewVectorStore(cfg.VectorDB, client)
	if err != nil {
		log.Fatal(err)
	}
	defer vs.Pool.Close()

	queries := []string{
		"欧豪 熟悉的 Go 微服务框架",
		"候选人的求职意向是什么",
		"欧豪 的 简历 内容",
	}

	for _, q := range queries {
		fmt.Printf("=== QUERY: %s ===\n", q)
		chunks, err := vs.RetrieveKnowledge(q, 3)
		if err != nil {
			fmt.Printf("ERR: %v\n\n", err)
			continue
		}
		for i, c := range chunks {
			fmt.Printf("[%d] title=%s\n%s\n\n", i+1, c.Title, c.Content)
		}
		if len(chunks) == 0 {
			fmt.Println("(no chunks)")
			fmt.Println()
		}
	}
}
