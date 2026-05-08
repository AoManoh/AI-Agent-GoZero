package llmclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sashabaranov/go-openai"
)

// utf8BOM 是 UTF-8 字节序标记 (Byte Order Mark)。
// Windows PowerShell 5.1 的 `Set-Content -Encoding UTF8` 与部分编辑器
// 在保存文件时会写入此 3 字节前缀；多数解析器（含 Go encoding/json）
// 不会把它当成空白符跳过，导致 JSON 解析或字段判定失败。
var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

// ProviderConfig 描述一组兼容 OpenAI API 的服务凭证。
type ProviderConfig struct {
	ApiKey        string `json:",optional"`
	ApiKeyEnv     string `json:",optional"`
	ApiKeyFile    string `json:",optional"`
	ApiKeyJSONKey string `json:",optional"`
	BaseURL       string `json:",optional"`
	Model         string `json:",optional"`
}

// Endpoint 是经过回退规则解析后的实际调用端点。
type Endpoint struct {
	ApiKey        string
	ApiKeyEnv     string
	ApiKeyFile    string
	ApiKeyJSONKey string
	BaseURL       string
	Model         string
}

// ResolveEndpoint 将主配置与默认配置合并，主配置只覆盖自己显式声明的字段。
func ResolveEndpoint(primary, fallback ProviderConfig, defaultModel string) Endpoint {
	cfg := primary
	if cfg.ApiKey == "" && cfg.ApiKeyEnv == "" && cfg.ApiKeyFile == "" {
		cfg.ApiKey = fallback.ApiKey
		cfg.ApiKeyEnv = fallback.ApiKeyEnv
		cfg.ApiKeyFile = fallback.ApiKeyFile
		cfg.ApiKeyJSONKey = fallback.ApiKeyJSONKey
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = fallback.BaseURL
	}
	if cfg.Model == "" {
		if defaultModel != "" {
			cfg.Model = defaultModel
		} else {
			cfg.Model = fallback.Model
		}
	}

	return Endpoint{
		ApiKey:        cfg.ApiKey,
		ApiKeyEnv:     cfg.ApiKeyEnv,
		ApiKeyFile:    cfg.ApiKeyFile,
		ApiKeyJSONKey: cfg.ApiKeyJSONKey,
		BaseURL:       cfg.BaseURL,
		Model:         cfg.Model,
	}
}

// NewClient 构造兼容 OpenAI API 的客户端。
func NewClient(endpoint Endpoint) (*openai.Client, error) {
	apiKey, err := ResolveAPIKey(endpoint)
	if err != nil {
		return nil, err
	}
	if apiKey == "" {
		return nil, fmt.Errorf("模型服务 API key 为空")
	}

	conf := openai.DefaultConfig(apiKey)
	if endpoint.BaseURL != "" {
		conf.BaseURL = endpoint.BaseURL
	}

	return openai.NewClientWithConfig(conf), nil
}

// ResolveAPIKey 按 env、文件、明文配置的顺序解析 API key。
func ResolveAPIKey(endpoint Endpoint) (string, error) {
	if endpoint.ApiKeyEnv != "" {
		value := strings.TrimSpace(os.Getenv(endpoint.ApiKeyEnv))
		if value == "" {
			return "", fmt.Errorf("环境变量 %s 未设置或为空", endpoint.ApiKeyEnv)
		}
		return value, nil
	}

	if endpoint.ApiKeyFile != "" {
		return readAPIKeyFile(endpoint.ApiKeyFile, endpoint.ApiKeyJSONKey)
	}

	return strings.TrimSpace(endpoint.ApiKey), nil
}

func readAPIKeyFile(path, jsonKey string) (string, error) {
	expanded, err := expandPath(path)
	if err != nil {
		return "", err
	}

	raw, err := os.ReadFile(expanded)
	if err != nil {
		return "", fmt.Errorf("读取 API key 文件失败: %w", err)
	}

	// 容错：去除 UTF-8 BOM 前缀。
	// 如果不剥离，TrimSpace 不会处理 BOM，导致下方 HasPrefix("{") 判定失败 →
	// 整段含 BOM 的 JSON 文本会被当成纯 key 返回；上游用作 Bearer token 时
	// 必然返回 401 INVALID_API_KEY，且报错信息无法直接指向 BOM 这个根因。
	raw = bytes.TrimPrefix(raw, utf8BOM)

	content := strings.TrimSpace(string(raw))
	if content == "" {
		return "", fmt.Errorf("API key 文件为空")
	}

	if strings.HasPrefix(content, "{") {
		var values map[string]any
		if err := json.Unmarshal(raw, &values); err != nil {
			return "", fmt.Errorf("解析 API key JSON 文件失败: %w", err)
		}
		for _, key := range candidateJSONKeys(jsonKey) {
			if value, ok := values[key]; ok {
				if str, ok := value.(string); ok && strings.TrimSpace(str) != "" {
					return strings.TrimSpace(str), nil
				}
			}
		}
		return "", fmt.Errorf("API key JSON 文件未包含可用字段")
	}

	return content, nil
}

func candidateJSONKeys(jsonKey string) []string {
	keys := []string{}
	if jsonKey != "" {
		keys = append(keys, jsonKey)
	}
	keys = append(keys, "OPENAI_API_KEY", "api_key", "apiKey", "key")
	return keys
}

func expandPath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("路径为空")
	}
	if path == "~" || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("获取用户主目录失败: %w", err)
		}
		if path == "~" {
			return home, nil
		}
		return filepath.Join(home, strings.TrimPrefix(path, "~/")), nil
	}
	return path, nil
}
