package user

import (
	"net/http"
	"testing"

	"GoZero-AI/api/user/internal/types"
	"GoZero-AI/internal/statuserr"
)

func TestBuildSessionCreateConfigRejectsFocusOutsideDirection(t *testing.T) {
	_, _, err := buildSessionCreateConfig(&types.CreateSessionReq{
		DirectionKey: "frontend_vue",
		FocusKeys:    []string{"concurrency"},
	})
	if err == nil {
		t.Fatal("buildSessionCreateConfig() error = nil, want bad request")
	}
	code, ok := statuserr.StatusCode(err)
	if !ok || code != http.StatusBadRequest {
		t.Fatalf("status = %d ok=%v, want 400/true; err=%v", code, ok, err)
	}
}

func TestBuildSessionCreateConfigAcceptsDirectionFocus(t *testing.T) {
	_, config, err := buildSessionCreateConfig(&types.CreateSessionReq{
		DirectionKey: "frontend_vue",
		FocusKeys:    []string{"frontend_arch", "performance"},
	})
	if err != nil {
		t.Fatalf("buildSessionCreateConfig() error = %v", err)
	}
	if config.DirectionKey != "frontend_vue" {
		t.Fatalf("DirectionKey = %q, want frontend_vue", config.DirectionKey)
	}
	if len(config.FocusAreas) != 2 ||
		config.FocusAreas[0].Key != "frontend_arch" ||
		config.FocusAreas[1].Key != "performance" {
		t.Fatalf("FocusAreas = %#v, want frontend_arch/performance", config.FocusAreas)
	}
}
