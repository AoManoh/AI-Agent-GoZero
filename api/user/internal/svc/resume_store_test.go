package svc

import (
	"testing"

	"GoZero-AI/internal/sessionmode"
)

func TestResolveResumeSessionMode(t *testing.T) {
	tests := []struct {
		name          string
		storedMode    string
		requestedMode string
		want          string
	}{
		{
			name:          "new session uses requested mode",
			storedMode:    "",
			requestedMode: sessionmode.KeyResearch,
			want:          sessionmode.KeyResearch,
		},
		{
			name:          "blank request keeps existing non-default mode",
			storedMode:    sessionmode.KeyMemory,
			requestedMode: "",
			want:          sessionmode.KeyMemory,
		},
		{
			name:          "requested mode cannot overwrite existing default mode",
			storedMode:    sessionmode.KeyInterview,
			requestedMode: sessionmode.KeyResearch,
			want:          sessionmode.KeyInterview,
		},
		{
			name:          "blank stored mode adopts requested mode",
			storedMode:    " ",
			requestedMode: sessionmode.KeyCoach,
			want:          sessionmode.KeyCoach,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveResumeSessionMode(tt.storedMode, tt.requestedMode); got != tt.want {
				t.Fatalf("resolveResumeSessionMode(%q, %q) = %q, want %q", tt.storedMode, tt.requestedMode, got, tt.want)
			}
		})
	}
}
