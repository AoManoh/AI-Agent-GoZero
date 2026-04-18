package config

import (
	"time"

	"github.com/zeromicro/go-zero/rest"
)

const defaultRefreshExpire = 7 * 24 * time.Hour

type Config struct {
	rest.RestConf
	Postgres struct {
		DataSource string
	}
	OpenAI struct {
		ApiKey              string
		BaseURL             string
		EmbeddingModel      string
		EvaluationModel     string  `json:",optional"`
		EvaluationTemp      float32 `json:",optional"`
		MaxCompletionTokens int     `json:",optional"`
	}
	MCP struct {
		Endpoint string
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

func (c Config) EvaluationModel() string {
	if c.OpenAI.EvaluationModel != "" {
		return c.OpenAI.EvaluationModel
	}
	return "qwen-plus"
}

func (c Config) EvaluationTemperature() float32 {
	if c.OpenAI.EvaluationTemp <= 0 {
		return 0.2
	}
	return c.OpenAI.EvaluationTemp
}

func (c Config) EvaluationMaxTokens() int {
	if c.OpenAI.MaxCompletionTokens <= 0 {
		return 1200
	}
	return c.OpenAI.MaxCompletionTokens
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
