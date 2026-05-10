package user

import (
	"database/sql"
	"testing"
	"time"

	"GoZero-AI/api/user/model"
	"GoZero-AI/internal/sessionmode"
)

func TestNormalizeSessionMode(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty defaults to interview", input: "", want: sessionmode.KeyInterview},
		{name: "interview studio alias", input: "Interview Studio", want: sessionmode.KeyInterview},
		{name: "research desk alias", input: "Research Desk", want: sessionmode.KeyResearch},
		{name: "memory atlas alias", input: "Memory Atlas", want: sessionmode.KeyMemory},
		{name: "coach passthrough", input: "Coach", want: sessionmode.KeyCoach},
		{name: "unknown falls back", input: "Companion", want: sessionmode.KeyInterview},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeSessionMode(tt.input); got != tt.want {
				t.Fatalf("normalizeSessionMode(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestBuildSessionItemMarksCompletedSessionInactive(t *testing.T) {
	now := time.Date(2026, 5, 10, 10, 0, 0, 0, time.FixedZone("CST", 8*3600))
	item := buildSessionItem(model.ChatSession{
		SessionId:   "sess-finished",
		Title:       "Go 后端面试",
		Mode:        sessionmode.KeyInterview,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
		CompletedAt: sql.NullTime{Time: now, Valid: true},
	})

	if item.IsActive {
		t.Fatalf("item.IsActive = true, want false for completed session")
	}
	if item.CompletedAt == "" {
		t.Fatal("item.CompletedAt is empty, want finished timestamp")
	}
}

func TestBuildSessionItemNormalizesMode(t *testing.T) {
	now := time.Date(2026, 4, 9, 10, 0, 0, 0, time.FixedZone("CST", 8*3600))
	item := buildSessionItem(model.ChatSession{
		SessionId:    "sess-1",
		Title:        "title",
		Mode:         "Research Desk",
		MessageCount: 2,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	})

	if item.Mode != sessionmode.LabelResearch {
		t.Fatalf("item.Mode = %q, want %q", item.Mode, sessionmode.LabelResearch)
	}
	if item.ModeKey != sessionmode.KeyResearch {
		t.Fatalf("item.ModeKey = %q, want %q", item.ModeKey, sessionmode.KeyResearch)
	}
}
