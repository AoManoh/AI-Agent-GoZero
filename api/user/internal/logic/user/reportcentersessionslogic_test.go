package user

import "testing"

func TestResolveReportCenterModeFilter(t *testing.T) {
	tests := []struct {
		name    string
		mode    string
		modeKey string
		want    string
		wantErr bool
	}{
		{name: "empty", want: ""},
		{name: "display value", mode: "Research Desk", want: "Research"},
		{name: "key value", modeKey: "Memory", want: "Memory"},
		{name: "invalid alias", mode: "Companion", wantErr: true},
		{name: "invalid key", modeKey: "Memory Atlas", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveReportCenterModeFilter(tt.mode, tt.modeKey)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("resolveReportCenterModeFilter(%q,%q) = %q, want %q", tt.mode, tt.modeKey, got, tt.want)
			}
		})
	}
}

func TestNormalizeReportCenterStatusFilter(t *testing.T) {
	if got, err := normalizeReportCenterStatusFilter("ready"); err != nil || got != "ready" {
		t.Fatalf("normalizeReportCenterStatusFilter ready = %q, %v", got, err)
	}
	if _, err := normalizeReportCenterStatusFilter("invalid"); err == nil {
		t.Fatalf("expected invalid status error")
	}
}

func TestNormalizeReportCenterSessionsLimit(t *testing.T) {
	if got := normalizeReportCenterSessionsLimit(0); got != defaultReportCenterSessionsLimit {
		t.Fatalf("default limit = %d, want %d", got, defaultReportCenterSessionsLimit)
	}
	if got := normalizeReportCenterSessionsLimit(999); got != maxReportCenterSessionsLimit {
		t.Fatalf("max limit = %d, want %d", got, maxReportCenterSessionsLimit)
	}
}
