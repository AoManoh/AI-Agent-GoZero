package user

import (
	"testing"

	"GoZero-AI/internal/sessionmode"
)

func TestBootstrapDefaultMode(t *testing.T) {
	modeKey, err := resolveReportCenterModeFilter("", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if modeKey != "" {
		t.Fatalf("modeKey = %q, want empty before bootstrap default", modeKey)
	}
	if sessionmode.DefaultKey != sessionmode.KeyInterview {
		t.Fatalf("default key = %q, want %q", sessionmode.DefaultKey, sessionmode.KeyInterview)
	}
}
