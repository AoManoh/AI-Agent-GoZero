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
