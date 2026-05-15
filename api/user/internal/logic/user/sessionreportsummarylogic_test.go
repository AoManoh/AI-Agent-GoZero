package user

import "testing"

func TestBuildReportSnapshotTitle(t *testing.T) {
	tests := []struct {
		status string
		want   string
	}{
		{status: "ready", want: "Current Evaluation Snapshot"},
		{status: "draft", want: "Draft Evaluation Snapshot"},
		{status: "insufficient_data", want: "Session Snapshot"},
	}

	for _, tt := range tests {
		if got := buildReportSnapshotTitle(tt.status); got != tt.want {
			t.Fatalf("buildReportSnapshotTitle(%q) = %q, want %q", tt.status, got, tt.want)
		}
	}
}

func TestFirstSuggestion(t *testing.T) {
	if got := firstSuggestion(nil); got != "" {
		t.Fatalf("firstSuggestion(nil) = %q, want empty", got)
	}

	if got := firstSuggestion([]string{"first", "second"}); got != "first" {
		t.Fatalf("firstSuggestion returned %q, want %q", got, "first")
	}
}
