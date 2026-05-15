package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"GoZero-AI/api/chat/internal/config"
	"GoZero-AI/api/chat/internal/svc"

	"github.com/golang-jwt/jwt/v4"
)

// TestRequireKnowledgeUploaderUserID 验证 PDF 上传的鉴权 helper 行为（2026-05-12 Q7=B 决策）。
//
// 覆盖场景:
//   - 匿名（无 Authorization header）→ 401
//   - 无效 access token → 401
//   - admin（user_id == publicKnowledgeAdminUserID）登录 → 200，userID 正确
//   - 普通 user 登录 → 200，userID 正确（不再 reject 非 admin，是 Q7=B 与原 admin gate 的核心差别）
func TestRequireKnowledgeUploaderUserID(t *testing.T) {
	const secret = "test-secret"
	svcCtx := &svc.ServiceContext{
		Config: config.Config{},
	}
	svcCtx.Config.Auth.AccessSecret = secret

	tests := []struct {
		name       string
		token      string
		wantStatus int
		wantOK     bool
		wantUserID int64
	}{
		{
			name:       "missing token",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid token",
			token:      "not-a-jwt",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "admin login allowed",
			token:      signAccessToken(t, secret, publicKnowledgeAdminUserID),
			wantStatus: http.StatusOK,
			wantOK:     true,
			wantUserID: publicKnowledgeAdminUserID,
		},
		{
			name:       "non admin login allowed",
			token:      signAccessToken(t, secret, 42),
			wantStatus: http.StatusOK,
			wantOK:     true,
			wantUserID: 42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/ai/knowledge/upload", nil)
			if tt.token != "" {
				request.Header.Set("Authorization", "Bearer "+tt.token)
			}
			recorder := httptest.NewRecorder()

			gotUserID, ok := requireKnowledgeUploaderUserID(recorder, request, svcCtx)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if !tt.wantOK && recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}
			if tt.wantOK && gotUserID != tt.wantUserID {
				t.Fatalf("userID = %d, want %d", gotUserID, tt.wantUserID)
			}
		})
	}
}

func signAccessToken(t *testing.T, secret string, userID int64) string {
	t.Helper()

	claims := jwt.MapClaims{
		"userId":    userID,
		"tokenType": "access",
		"exp":       time.Now().Add(time.Hour).Unix(),
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return token
}
