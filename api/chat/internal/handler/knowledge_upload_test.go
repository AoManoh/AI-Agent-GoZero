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

func TestRequirePublicKnowledgeAdmin(t *testing.T) {
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
	}{
		{
			name:       "missing token",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "non admin",
			token:      signAccessToken(t, secret, 2),
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "admin",
			token:      signAccessToken(t, secret, publicKnowledgeAdminUserID),
			wantStatus: http.StatusOK,
			wantOK:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/ai/knowledge/upload", nil)
			if tt.token != "" {
				request.Header.Set("Authorization", "Bearer "+tt.token)
			}
			recorder := httptest.NewRecorder()

			_, ok := requirePublicKnowledgeAdmin(recorder, request, svcCtx)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if !tt.wantOK && recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}
			if tt.wantOK && recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, tt.wantStatus)
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
