package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"GoZero-AI/api/chat/internal/svc"
)

func TestBuildToolKnowledgeContent(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		content string
		want    string
	}{
		{
			name:    "content only",
			content: "  MCP 服务可以作为旁路资料源。  ",
			want:    "MCP 服务可以作为旁路资料源。",
		},
		{
			name:    "source and content",
			source:  "  Grok Search MCP  ",
			content: "\n资料包正文\n",
			want:    "资料来源: Grok Search MCP\n\n资料包正文",
		},
		{
			name:    "source without content",
			source:  "Grok Search MCP",
			content: "  ",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildToolKnowledgeContent(tt.source, tt.content); got != tt.want {
				t.Fatalf("buildToolKnowledgeContent() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestKnowledgeTextUploadHandlerRequiresAdminToken(t *testing.T) {
	svcCtx := &svc.ServiceContext{}
	handler := KnowledgeTextUploadHandler(svcCtx)
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/ai/knowledge/text",
		strings.NewReader(`{"title":"Grok Search MCP","content":"context pack"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}
