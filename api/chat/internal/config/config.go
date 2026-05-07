package config

import (
	"GoZero-AI/internal/llmclient"

	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf
	Auth struct {
		AccessSecret string `json:",optional"`
	}
	OpenAI    OpenAIConfig
	Embedding llmclient.ProviderConfig `json:",optional"`
	VectorDB  VectorDBConfig           // 新增向量数据库配置
	// MCPConfig MCP 服务配置
	MCP struct {
		Endpoint       string // MCP 服务地址
		AuthToken      string `json:",optional"`
		MaxUploadBytes int64  `json:",optional"`
	}
	Redis Redis // 新增：Redis 配置
}

func (c Config) MCPMaxUploadBytes() int64 {
	if c.MCP.MaxUploadBytes <= 0 {
		return 50 * 1024 * 1024
	}
	return c.MCP.MaxUploadBytes
}

type OpenAIConfig struct {
	ApiKey              string `json:",optional"`
	ApiKeyEnv           string `json:",optional"`
	ApiKeyFile          string `json:",optional"`
	ApiKeyJSONKey       string `json:",optional"`
	BaseURL             string
	Model               string
	MaxCompletionTokens int // 注意 MaxTokens 字段已经废弃
	Temperature         float32
}

func (c OpenAIConfig) ProviderConfig() llmclient.ProviderConfig {
	return llmclient.ProviderConfig{
		ApiKey:        c.ApiKey,
		ApiKeyEnv:     c.ApiKeyEnv,
		ApiKeyFile:    c.ApiKeyFile,
		ApiKeyJSONKey: c.ApiKeyJSONKey,
		BaseURL:       c.BaseURL,
		Model:         c.Model,
	}
}

func (c Config) ChatEndpoint() llmclient.Endpoint {
	return llmclient.ResolveEndpoint(c.OpenAI.ProviderConfig(), llmclient.ProviderConfig{}, c.OpenAI.Model)
}

func (c Config) EmbeddingEndpoint() llmclient.Endpoint {
	fallback := c.OpenAI.ProviderConfig()
	if c.Embedding.BaseURL != "" && c.Embedding.ApiKey == "" && c.Embedding.ApiKeyEnv == "" && c.Embedding.ApiKeyFile == "" && c.OpenAI.ApiKey != "" {
		// 对话模型可能通过 ApiKeyFile 指向 Codex 网关；embedding 指定独立 BaseURL 时，
		// 优先沿用旧 OpenAI.ApiKey，避免把向量请求误打到对话模型凭证。
		fallback.ApiKeyEnv = ""
		fallback.ApiKeyFile = ""
		fallback.ApiKeyJSONKey = ""
	}
	return llmclient.ResolveEndpoint(c.Embedding, fallback, c.EmbeddingModel())
}

func (c Config) EmbeddingModel() string {
	if c.Embedding.Model != "" {
		return c.Embedding.Model
	}
	return c.VectorDB.EmbeddingModel
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
