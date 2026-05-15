package user

import (
	"database/sql"
	"testing"
	"time"

	"GoZero-AI/api/user/internal/types"
	"GoZero-AI/internal/sessionmode"
)

func TestFindReportCenterModeCard(t *testing.T) {
	card := findReportCenterModeCard([]types.ReportCenterModeCard{
		{ModeKey: sessionmode.KeyResearch, Mode: sessionmode.LabelResearch},
	}, sessionmode.KeyResearch)

	if card.Mode != sessionmode.LabelResearch {
		t.Fatalf("card.Mode = %q, want %q", card.Mode, sessionmode.LabelResearch)
	}
}

func TestFindReportCenterModeCardFallback(t *testing.T) {
	card := findReportCenterModeCard(nil, sessionmode.KeyMemory)
	if card.Mode != sessionmode.LabelMemory {
		t.Fatalf("fallback mode = %q, want %q", card.Mode, sessionmode.LabelMemory)
	}
	if card.AttentionState != "empty" {
		t.Fatalf("fallback attentionState = %q, want empty", card.AttentionState)
	}
}

func TestFilterReportCenterReportsByMode(t *testing.T) {
	now := time.Date(2026, 4, 10, 10, 0, 0, 0, time.FixedZone("CST", 8*3600))
	rows := []reportCenterOverviewRow{
		{
			SessionId:     "sess-memory",
			Title:         "memory",
			Mode:          "Memory Atlas",
			CreatedAt:     now,
			UpdatedAt:     now,
			MessageCount:  1,
			IsActive:      true,
			Status:        sql.NullString{String: "draft", Valid: true},
			Summary:       sql.NullString{String: "draft summary", Valid: true},
			OverallScore:  sql.NullFloat64{Float64: 0, Valid: true},
			GeneratedAt:   sql.NullTime{},
			ResumeChunks:  0,
			LastMessageAt: sql.NullTime{},
		},
		{
			SessionId:     "sess-research",
			Title:         "research",
			Mode:          "Research",
			CreatedAt:     now,
			UpdatedAt:     now,
			MessageCount:  1,
			IsActive:      true,
			Status:        sql.NullString{String: "ready", Valid: true},
			Summary:       sql.NullString{String: "ready summary", Valid: true},
			OverallScore:  sql.NullFloat64{Float64: 88, Valid: true},
			GeneratedAt:   sql.NullTime{Time: now, Valid: true},
			ResumeChunks:  1,
			LastMessageAt: sql.NullTime{},
		},
	}

	reports := filterReportCenterReportsByMode(rows, sessionmode.KeyResearch)
	if len(reports) != 1 {
		t.Fatalf("len(reports) = %d, want 1", len(reports))
	}
	if reports[0].Session.ModeKey != sessionmode.KeyResearch {
		t.Fatalf("reports[0].Session.ModeKey = %q, want %q", reports[0].Session.ModeKey, sessionmode.KeyResearch)
	}
}
