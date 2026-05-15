package user

import (
	"database/sql"
	"testing"
	"time"

	"GoZero-AI/internal/sessionmode"
)

func TestBuildReportCenterModeCards(t *testing.T) {
	now := time.Date(2026, 4, 9, 16, 30, 0, 0, time.FixedZone("CST", 8*3600))
	rows := []reportCenterOverviewRow{
		{
			SessionId:       "sess-memory",
			Title:           "memory",
			Mode:            "Memory Atlas",
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

	cards := buildReportCenterModeCards(rows)
	if len(cards) != len(sessionmode.AllKeys()) {
		t.Fatalf("len(cards) = %d, want %d", len(cards), len(sessionmode.AllKeys()))
	}
	if cards[1].ModeKey != sessionmode.KeyResearch {
		t.Fatalf("cards[1].ModeKey = %q, want %q", cards[1].ModeKey, sessionmode.KeyResearch)
	}
	if cards[1].InsufficientDataSessions != 1 {
		t.Fatalf("research insufficient count = %d, want 1", cards[1].InsufficientDataSessions)
	}
	if cards[1].Spotlight.Session.SessionId != "sess-research" {
		t.Fatalf("research spotlight session = %q", cards[1].Spotlight.Session.SessionId)
	}
	if cards[2].Mode != sessionmode.LabelMemory {
		t.Fatalf("cards[2].Mode = %q, want %q", cards[2].Mode, sessionmode.LabelMemory)
	}
	if !cards[2].HasReadyReport {
		t.Fatalf("memory hasReadyReport should be true")
	}
	if cards[2].AttentionState != "ready" {
		t.Fatalf("memory attentionState = %q, want ready", cards[2].AttentionState)
	}
	if cards[2].Spotlight.NextAction != "keep memory updated" {
		t.Fatalf("memory spotlight nextAction = %q", cards[2].Spotlight.NextAction)
	}
	if cards[2].ResumeBackedSessions != 1 {
		t.Fatalf("memory resume count = %d, want 1", cards[2].ResumeBackedSessions)
	}
}
