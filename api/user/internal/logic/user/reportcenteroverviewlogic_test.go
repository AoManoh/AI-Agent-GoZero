package user

import (
	"database/sql"
	"testing"
	"time"

	"GoZero-AI/internal/sessionmode"
)

func TestBuildReportCenterOverview(t *testing.T) {
	now := time.Date(2026, 4, 9, 15, 0, 0, 0, time.FixedZone("CST", 8*3600))
	rows := []reportCenterOverviewRow{
		{
			SessionId:       "sess-memory",
			Title:           "memory",
			Mode:            "Memory",
			CreatedAt:       now.Add(-2 * time.Hour),
			UpdatedAt:       now.Add(-1 * time.Hour),
			LastMessageAt:   sql.NullTime{Time: now.Add(-30 * time.Minute), Valid: true},
			MessageCount:    2,
			IsActive:        true,
			Status:          sql.NullString{String: "ready", Valid: true},
			Summary:         sql.NullString{String: "memory summary", Valid: true},
			OverallScore:    sql.NullFloat64{Float64: 90, Valid: true},
			GeneratedAt:     sql.NullTime{Time: now.Add(-20 * time.Minute), Valid: true},
			Suggestions:     []byte(`["keep memory updated"]`),
			ResumeChunks:    3,
			ResumeUpdatedAt: sql.NullTime{Time: now.Add(-10 * time.Minute), Valid: true},
		},
		{
			SessionId:     "sess-research",
			Title:         "research",
			Mode:          "Research Desk",
			CreatedAt:     now.Add(-3 * time.Hour),
			UpdatedAt:     now.Add(-90 * time.Minute),
			MessageCount:  0,
			IsActive:      true,
			Status:        sql.NullString{String: "insufficient_data", Valid: true},
			OverallScore:  sql.NullFloat64{Float64: 0, Valid: true},
			Suggestions:   []byte(`["ask more"]`),
			ResumeChunks:  0,
			LastMessageAt: sql.NullTime{},
		},
	}

	totals, modes, recentReports := buildReportCenterOverview(rows)

	if totals.TotalSessions != 2 {
		t.Fatalf("totals.TotalSessions = %d, want 2", totals.TotalSessions)
	}
	if totals.EvaluatedSessions != 2 {
		t.Fatalf("totals.EvaluatedSessions = %d, want 2", totals.EvaluatedSessions)
	}
	if totals.ReadySessions != 1 {
		t.Fatalf("totals.ReadySessions = %d, want 1", totals.ReadySessions)
	}
	if totals.InsufficientDataSessions != 1 {
		t.Fatalf("totals.InsufficientDataSessions = %d, want 1", totals.InsufficientDataSessions)
	}
	if totals.ResumeBackedSessions != 1 {
		t.Fatalf("totals.ResumeBackedSessions = %d, want 1", totals.ResumeBackedSessions)
	}
	if totals.AverageScore != 90 {
		t.Fatalf("totals.AverageScore = %v, want 90", totals.AverageScore)
	}
	if totals.LastActivityAt == "" {
		t.Fatalf("totals.LastActivityAt should not be empty")
	}

	if len(modes) != len(sessionmode.AllKeys()) {
		t.Fatalf("len(modes) = %d, want %d", len(modes), len(sessionmode.AllKeys()))
	}
	if modes[1].ModeKey != sessionmode.KeyResearch {
		t.Fatalf("modes[1].ModeKey = %q, want %q", modes[1].ModeKey, sessionmode.KeyResearch)
	}
	if modes[1].SessionCount != 1 {
		t.Fatalf("research session count = %d, want 1", modes[1].SessionCount)
	}
	if modes[1].InsufficientDataSessions != 1 {
		t.Fatalf("research insufficient count = %d, want 1", modes[1].InsufficientDataSessions)
	}
	if modes[2].Mode != sessionmode.LabelMemory {
		t.Fatalf("memory mode label = %q, want %q", modes[2].Mode, sessionmode.LabelMemory)
	}
	if modes[2].ResumeBackedSessions != 1 {
		t.Fatalf("memory resume count = %d, want 1", modes[2].ResumeBackedSessions)
	}

	if len(recentReports) != 2 {
		t.Fatalf("len(recentReports) = %d, want 2", len(recentReports))
	}
	if recentReports[0].Session.ModeKey != sessionmode.KeyMemory {
		t.Fatalf("recentReports[0].Session.ModeKey = %q, want %q", recentReports[0].Session.ModeKey, sessionmode.KeyMemory)
	}
	if recentReports[0].NextAction != "keep memory updated" {
		t.Fatalf("recentReports[0].NextAction = %q", recentReports[0].NextAction)
	}
}

func TestDecodeFirstSuggestion(t *testing.T) {
	if got := decodeFirstSuggestion([]byte(`["first","second"]`)); got != "first" {
		t.Fatalf("decodeFirstSuggestion valid = %q, want %q", got, "first")
	}
	if got := decodeFirstSuggestion([]byte(`not-json`)); got != "" {
		t.Fatalf("decodeFirstSuggestion invalid = %q, want empty", got)
	}
}
