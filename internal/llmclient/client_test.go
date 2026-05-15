package llmclient

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveEndpointUsesPrimaryFields(t *testing.T) {
	got := ResolveEndpoint(
		ProviderConfig{
			ApiKeyFile:    "auth.json",
			ApiKeyJSONKey: "OPENAI_API_KEY",
			BaseURL:       "https://chat.example/v1",
			Model:         "gpt-5.4",
		},
		ProviderConfig{
			ApiKey:  "fallback-key",
			BaseURL: "https://fallback.example/v1",
			Model:   "fallback-model",
		},
		"default-model",
	)

	if got.ApiKeyFile != "auth.json" {
		t.Fatalf("ApiKeyFile = %q, want auth.json", got.ApiKeyFile)
	}
	if got.BaseURL != "https://chat.example/v1" {
		t.Fatalf("BaseURL = %q, want primary base url", got.BaseURL)
	}
	if got.Model != "gpt-5.4" {
		t.Fatalf("Model = %q, want gpt-5.4", got.Model)
	}
}

func TestResolveEndpointFallsBackToDefaultModel(t *testing.T) {
	got := ResolveEndpoint(
		ProviderConfig{},
		ProviderConfig{
			ApiKey:  "fallback-key",
			BaseURL: "https://fallback.example/v1",
			Model:   "fallback-model",
		},
		"text-embedding-v1",
	)

	if got.ApiKey != "fallback-key" {
		t.Fatalf("ApiKey fallback not applied")
	}
	if got.BaseURL != "https://fallback.example/v1" {
		t.Fatalf("BaseURL fallback not applied")
	}
	if got.Model != "text-embedding-v1" {
		t.Fatalf("Model = %q, want text-embedding-v1", got.Model)
	}
}

func TestResolveAPIKeyFromEnv(t *testing.T) {
	t.Setenv("GOZERO_AI_TEST_KEY", " env-key ")

	got, err := ResolveAPIKey(Endpoint{ApiKeyEnv: "GOZERO_AI_TEST_KEY"})
	if err != nil {
		t.Fatalf("ResolveAPIKey returned error: %v", err)
	}
	if got != "env-key" {
		t.Fatalf("ResolveAPIKey = %q, want env-key", got)
	}
}

func TestResolveAPIKeyFromJSONFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "auth.json")
	if err := os.WriteFile(path, []byte(`{"OPENAI_API_KEY":"file-key"}`), 0o600); err != nil {
		t.Fatalf("write auth file: %v", err)
	}

	got, err := ResolveAPIKey(Endpoint{ApiKeyFile: path, ApiKeyJSONKey: "OPENAI_API_KEY"})
	if err != nil {
		t.Fatalf("ResolveAPIKey returned error: %v", err)
	}
	if got != "file-key" {
		t.Fatalf("ResolveAPIKey = %q, want file-key", got)
	}
}

// TestResolveAPIKeyFromJSONFileWithUTF8BOM 防回退：
// PowerShell 5.1 `Set-Content -Encoding UTF8` 与 Windows 记事本另存 UTF-8
// 都会写入 UTF-8 BOM (EF BB BF)。readAPIKeyFile 必须能透明吃掉这 3 字节，
// 否则 strings.HasPrefix(content, "{") 判定失败 → 整段文件被当作 key 返回，
// 用作 Bearer token 时上游返回 401 INVALID_API_KEY。
func TestResolveAPIKeyFromJSONFileWithUTF8BOM(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "auth.json")
	bomPayload := append([]byte{0xEF, 0xBB, 0xBF}, []byte(`{"OPENAI_API_KEY":"bom-key"}`)...)
	if err := os.WriteFile(path, bomPayload, 0o600); err != nil {
		t.Fatalf("write auth file: %v", err)
	}

	got, err := ResolveAPIKey(Endpoint{ApiKeyFile: path, ApiKeyJSONKey: "OPENAI_API_KEY"})
	if err != nil {
		t.Fatalf("ResolveAPIKey returned error: %v", err)
	}
	if got != "bom-key" {
		t.Fatalf("ResolveAPIKey = %q, want bom-key (BOM should be stripped)", got)
	}
}

// TestResolveAPIKeyFromPlainTextFileWithUTF8BOM 覆盖纯文本 key 文件场景：
// 即便用户把 key 直接写入 .txt 而非 JSON，加上 BOM 也必须能正确读出。
func TestResolveAPIKeyFromPlainTextFileWithUTF8BOM(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "key.txt")
	bomPayload := append([]byte{0xEF, 0xBB, 0xBF}, []byte("plain-bom-key\n")...)
	if err := os.WriteFile(path, bomPayload, 0o600); err != nil {
		t.Fatalf("write auth file: %v", err)
	}

	got, err := ResolveAPIKey(Endpoint{ApiKeyFile: path})
	if err != nil {
		t.Fatalf("ResolveAPIKey returned error: %v", err)
	}
	if got != "plain-bom-key" {
		t.Fatalf("ResolveAPIKey = %q, want plain-bom-key", got)
	}
}
