package config

import (
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf
	Auth struct {
		AccessSecret string `json:",optional"`
	}
	OpenAI struct {
		ApiKey              string
		BaseURL             string
		Model               string
		MaxCompletionTokens int // 注意 MaxTokens 字段已经废弃
		Temperature         float32
	}
	VectorDB VectorDBConfig // 新增向量数据库配置
	// MCPConfig MCP 服务配置
	MCP struct {
		Endpoint string // MCP 服务地址
	}
	Redis Redis // 新增：Redis 配置
}

// VectorDBConfig 向量数据库配置
type VectorDBConfig struct {
	Host           string    // 本地主机
	Port           int       // 之前 postgres 数据库的端口号
	DBName         string    // 之前数据库创建的名字
	User           string    // 数据库用户名
	Password       string    // 数据库账户密码
	Table          string    // 表名
	MaxConn        int       // 最大连接数
	EmbeddingModel string    // 最关键的，大模型的嵌入模型
	Knowledge      Knowledge // 新增：本地 RAG 知识库配置
	TimeZone       string    // 新增：时区配置
}

// Knowledge RAG 本地知识库配置结构体
// 新增：本地 RAG 知识库配置
// 用于配置知识库文档的分块、检索和上下文注入策略
type Knowledge struct {
	MaxChunkSize     int // 知识文档分块的最大字符数，用于将大文档切分为可处理的小块
	TopK             int // 向量检索时返回的最相似知识片段数量，影响检索结果的丰富度
	MaxContextLength int // 注入到 AI 上下文中的知识内容最大长度，防止超出模型 token 限制
}

// Redis Redis 数据库配置结构体
// 定义 Redis 连接所需的配置参数
type Redis struct {
	Host     string
	Port     int
	Password string
	DB       int
}
