package user

import (
	"context"
	"net/http"
	"testing"

	"GoZero-AI/internal/statuserr"
)

func TestCurrentUserIDMissingReturnsUnauthorized(t *testing.T) {
	_, err := currentUserID(context.Background())
	if err == nil {
		t.Fatal("currentUserID() error = nil, want unauthorized error")
	}

	code, ok := statuserr.StatusCode(err)
	if !ok {
		t.Fatal("currentUserID() error does not expose status code")
	}
	if code != http.StatusUnauthorized {
		t.Fatalf("status code = %d, want %d", code, http.StatusUnauthorized)
	}
}

func TestCurrentUserIDRejectsNegativeValue(t *testing.T) {
	ctx := context.WithValue(context.Background(), "userId", int64(-1))
	_, err := currentUserID(ctx)
	if err == nil {
		t.Fatal("currentUserID() error = nil, want unauthorized error")
	}

	code, ok := statuserr.StatusCode(err)
	if !ok || code != http.StatusUnauthorized {
		t.Fatalf("status code = %d, ok=%v, want %d/true", code, ok, http.StatusUnauthorized)
	}
}
