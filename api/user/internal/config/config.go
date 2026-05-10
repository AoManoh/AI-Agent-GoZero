package config

import (
	"strings"
	"time"

	"GoZero-AI/internal/llmclient"

	"github.com/zeromicro/go-zero/rest"
)

const defaultRefreshExpire = 7 * 24 * time.Hour

type Config struct {
	rest.RestConf
	Postgres struct {
		DataSource string
	}
	OpenAI     OpenAIConfig
	Embedding  llmclient.ProviderConfig `json:",optional"`
	Evaluation llmclient.ProviderConfig `json:",optional"`
	MCP        struct {
		Endpoint       string
		AuthToken      string `json:",optional"`
		MaxUploadBytes int64  `json:",optional"`
	}
	Resume struct {
		MaxChunkSize int `json:",optional"`
	}
	Redis struct {
		Host           string
		Pass           string `json:",optional"`
		Password       string `json:",optional"`
		DB             int    `json:",optional"`
		Type           string `json:",optional"`
		DialTimeoutMs  int    `json:",optional"`
		ReadTimeoutMs  int    `json:",optional"`
		WriteTimeoutMs int    `json:",optional"`
	}
	Auth struct {
		AccessSecret  string
		AccessExpire  int64
		RefreshExpire int64 `json:",optional"`
	}
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
	EmbeddingModel      string
	EvaluationModel     string  `json:",optional"`
	EvaluationTemp      float32 `json:",optional"`
	MaxCompletionTokens int     `json:",optional"`
}

func (c OpenAIConfig) ProviderConfig() llmclient.ProviderConfig {
	return llmclient.ProviderConfig{
		ApiKey:        c.ApiKey,
		ApiKeyEnv:     c.ApiKeyEnv,
		ApiKeyFile:    c.ApiKeyFile,
		ApiKeyJSONKey: c.ApiKeyJSONKey,
		BaseURL:       c.BaseURL,
	}
}

func (c Config) RefreshTokenTTL() time.Duration {
	if c.Auth.RefreshExpire <= 0 {
		return defaultRefreshExpire
	}

	return time.Duration(c.Auth.RefreshExpire) * time.Second
}

func (c Config) AccessTokenTTL() time.Duration {
	if c.Auth.AccessExpire <= 0 {
		return 2 * time.Hour
	}

	return time.Duration(c.Auth.AccessExpire) * time.Second
}

func (c Config) RedisPassword() string {
	if c.Redis.Password != "" {
		return c.Redis.Password
	}

	return c.Redis.Pass
}

func (c Config) ResumeChunkSize() int {
	if c.Resume.MaxChunkSize <= 0 {
		return 1000
	}

	return c.Resume.MaxChunkSize
}

func (c Config) EmbeddingModel() string {
	if c.Embedding.Model != "" {
		return c.Embedding.Model
	}
	return c.OpenAI.EmbeddingModel
}

func (c Config) EvaluationModel() string {
	if c.Evaluation.Model != "" {
		return c.Evaluation.Model
	}
	if c.OpenAI.EvaluationModel != "" {
		return c.OpenAI.EvaluationModel
	}
	return "qwen-plus"
}

func (c Config) EvaluationTemperature() float32 {
	if modelRequiresUnitTemperature(c.EvaluationModel()) {
		return 1
	}
	if c.OpenAI.EvaluationTemp <= 0 {
		return 0.2
	}
	return c.OpenAI.EvaluationTemp
}

func modelRequiresUnitTemperature(model string) bool {
	normalized := strings.ToLower(strings.TrimSpace(model))
	return strings.HasPrefix(normalized, "gpt-5")
}

func (c Config) EvaluationMaxTokens() int {
	if c.OpenAI.MaxCompletionTokens <= 0 {
		return 1200
	}
	return c.OpenAI.MaxCompletionTokens
}

func (c Config) EmbeddingEndpoint() llmclient.Endpoint {
	return llmclient.ResolveEndpoint(c.Embedding, c.OpenAI.ProviderConfig(), c.EmbeddingModel())
}

func (c Config) EvaluationEndpoint() llmclient.Endpoint {
	fallback := c.OpenAI.ProviderConfig()
	fallback.Model = c.OpenAI.EvaluationModel
	return llmclient.ResolveEndpoint(c.Evaluation, fallback, c.EvaluationModel())
}

func (c Config) RedisDialTimeout() time.Duration {
	if c.Redis.DialTimeoutMs <= 0 {
		return time.Second
	}
	return time.Duration(c.Redis.DialTimeoutMs) * time.Millisecond
}

func (c Config) RedisReadTimeout() time.Duration {
	if c.Redis.ReadTimeoutMs <= 0 {
		return time.Second
	}
	return time.Duration(c.Redis.ReadTimeoutMs) * time.Millisecond
}

func (c Config) RedisWriteTimeout() time.Duration {
	if c.Redis.WriteTimeoutMs <= 0 {
		return time.Second
	}
	return time.Duration(c.Redis.WriteTimeoutMs) * time.Millisecond
}
