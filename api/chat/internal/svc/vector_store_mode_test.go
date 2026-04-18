package svc

import (
	"testing"

	"GoZero-AI/internal/sessionmode"
)

func TestEffectiveSessionMode(t *testing.T) {
	tests := []struct {
		name          string
		storedMode    string
		requestedMode string
		want          string
	}{
		{
			name:          "stored mode wins when request missing",
			storedMode:    sessionmode.KeyResearch,
			requestedMode: "",
			want:          sessionmode.KeyResearch,
		},
		{
			name:          "stored mode wins over conflicting request",
			storedMode:    sessionmode.KeyMemory,
			requestedMode: sessionmode.KeyInterview,
			want:          sessionmode.KeyMemory,
		},
		{
			name:          "request mode used when no stored mode",
			storedMode:    "",
			requestedMode: sessionmode.LabelResearch,
			want:          sessionmode.KeyResearch,
		},
		{
			name:          "empty values default to interview",
			storedMode:    "",
			requestedMode: "",
			want:          sessionmode.KeyInterview,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := effectiveSessionMode(tt.storedMode, tt.requestedMode); got != tt.want {
				t.Fatalf("effectiveSessionMode(%q, %q) = %q, want %q", tt.storedMode, tt.requestedMode, got, tt.want)
			}
		})
	}
}
