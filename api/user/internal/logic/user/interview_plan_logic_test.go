package user

import (
	"context"
	"database/sql"
	"net/http"
	"testing"
	"time"

	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/api/user/internal/types"
	"GoZero-AI/api/user/model"
	"GoZero-AI/internal/statuserr"
)

func TestInterviewPlanPreviewUsesQueryConfig(t *testing.T) {
	logic := NewInterviewPlanPreviewLogic(withUserID(context.Background(), 7), &svc.ServiceContext{})

	resp, err := logic.InterviewPlanPreview(&types.InterviewPlanPreviewReq{
		DirectionKey: "go_backend",
		Difficulty:   4,
		FocusKeys:    "database,concurrency",
		Limit:        5,
	})
	if err != nil {
		t.Fatalf("InterviewPlanPreview() error = %v", err)
	}
	if resp.Config.DirectionKey != "go_backend" {
		t.Fatalf("DirectionKey = %q, want go_backend", resp.Config.DirectionKey)
	}
	if resp.Config.DifficultyLevel != 4 {
		t.Fatalf("DifficultyLevel = %d, want 4", resp.Config.DifficultyLevel)
	}
	if len(resp.Questions) != 5 {
		t.Fatalf("len(Questions) = %d, want 5", len(resp.Questions))
	}
	if !resp.PlanMeta.Available || resp.PlanMeta.SchemaVersion != "interview-plan-v1" {
		t.Fatalf("PlanMeta = %+v, want available interview-plan-v1", resp.PlanMeta)
	}

	allowedFocus := map[string]struct{}{
		"database":    {},
		"concurrency": {},
	}
	for _, question := range resp.Questions {
		if _, ok := allowedFocus[question.FocusKey]; !ok {
			t.Fatalf("question %s FocusKey = %q, want database/concurrency", question.Key, question.FocusKey)
		}
		if question.Prompt == "" || len(question.ExpectedSignals) == 0 || len(question.FollowUps) == 0 {
			t.Fatalf("question %s missing prompt/signals/followups: %+v", question.Key, question)
		}
	}
	if len(resp.Milestones) != 3 {
		t.Fatalf("len(Milestones) = %d, want 3", len(resp.Milestones))
	}
}

func TestInterviewPlanPreviewRejectsInvalidFocus(t *testing.T) {
	logic := NewInterviewPlanPreviewLogic(withUserID(context.Background(), 7), &svc.ServiceContext{})

	_, err := logic.InterviewPlanPreview(&types.InterviewPlanPreviewReq{
		DirectionKey: "go_backend",
		FocusKeys:    "missing_focus",
	})
	if err == nil {
		t.Fatal("InterviewPlanPreview() error = nil, want bad request")
	}
	code, ok := statuserr.StatusCode(err)
	if !ok || code != http.StatusBadRequest {
		t.Fatalf("status code = %d, ok=%v, want %d/true", code, ok, http.StatusBadRequest)
	}
}

func TestInterviewPlanPreviewRejectsNegativeLimit(t *testing.T) {
	logic := NewInterviewPlanPreviewLogic(withUserID(context.Background(), 7), &svc.ServiceContext{})

	_, err := logic.InterviewPlanPreview(&types.InterviewPlanPreviewReq{Limit: -1})
	if err == nil {
		t.Fatal("InterviewPlanPreview() error = nil, want bad request")
	}
	code, ok := statuserr.StatusCode(err)
	if !ok || code != http.StatusBadRequest {
		t.Fatalf("status code = %d, ok=%v, want %d/true", code, ok, http.StatusBadRequest)
	}
}

func TestSessionInterviewPlanUsesPersistedSessionConfig(t *testing.T) {
	now := time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC)
	session := &model.ChatSession{
		SessionId:             "sess-plan",
		UserId:                7,
		Title:                 "Go 后端面试",
		Mode:                  "Interview",
		DirectionKey:          "go_backend",
		DirectionLabel:        "Go 后端",
		DifficultyLevel:       4,
		DifficultyLabel:       "资深",
		InterviewerStyle:      "senior",
		InterviewerStyleLabel: "资深技术官",
		FocusAreas:            []byte(`[{"key":"system_design","label":"系统设计"}]`),
		FollowUpDepth:         "N+5",
		EstimatedMinutes:      45,
		CreatedAt:             now.Add(-time.Hour),
		UpdatedAt:             now,
		StartedAt:             sql.NullTime{Time: now.Add(-time.Hour), Valid: true},
		IsActive:              true,
	}
	logic := NewSessionInterviewPlanLogic(withUserID(context.Background(), 7), &svc.ServiceContext{
		ChatSessionsModel: &stubChatSessionsModel{session: session},
	})

	resp, err := logic.SessionInterviewPlan(&types.SessionInterviewPlanReq{Id: session.SessionId, Limit: 2})
	if err != nil {
		t.Fatalf("SessionInterviewPlan() error = %v", err)
	}
	if resp.Config.EstimatedMinutes != 45 || resp.Config.FollowUpDepth != "N+5" {
		t.Fatalf("Config = %+v, want persisted estimated minutes/follow up depth", resp.Config)
	}
	if len(resp.Questions) != 2 {
		t.Fatalf("len(Questions) = %d, want 2", len(resp.Questions))
	}
	for _, question := range resp.Questions {
		if question.FocusKey != "system_design" {
			t.Fatalf("question %s FocusKey = %q, want system_design", question.Key, question.FocusKey)
		}
	}
}

func TestSessionInterviewPlanNotFound(t *testing.T) {
	logic := NewSessionInterviewPlanLogic(withUserID(context.Background(), 7), &svc.ServiceContext{
		ChatSessionsModel: &stubChatSessionsModel{err: model.ErrNotFound},
	})

	_, err := logic.SessionInterviewPlan(&types.SessionInterviewPlanReq{Id: "missing"})
	if err == nil {
		t.Fatal("SessionInterviewPlan() error = nil, want not found")
	}
	code, ok := statuserr.StatusCode(err)
	if !ok || code != http.StatusNotFound {
		t.Fatalf("status code = %d, ok=%v, want %d/true", code, ok, http.StatusNotFound)
	}
}
