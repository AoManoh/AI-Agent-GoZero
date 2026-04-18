package user

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"GoZero-AI/api/user/internal/evaluation"
	"GoZero-AI/api/user/internal/types"
	"GoZero-AI/api/user/model"
)

func TestShouldRefreshEvaluation(t *testing.T) {
	now := time.Date(2026, 4, 13, 20, 0, 0, 0, time.UTC)
	completeRecord := &model.SessionEvaluation{
		Status:        "ready",
		OverallScore:  88,
		RubricVersion: evaluation.RubricVersion,
		ScoreSource:   "llm",
		Dimensions:    mustJSON(t, []types.EvaluationDimension{{Key: "technical_depth", Label: "技术深度", Score: 4, MaxScore: 5, Summary: "ok"}, {Key: "engineering_practice", Label: "工程实践", Score: 4, MaxScore: 5, Summary: "ok"}, {Key: "architecture_sense", Label: "架构意识", Score: 4, MaxScore: 5, Summary: "ok"}, {Key: "communication", Label: "表达与沟通", Score: 4, MaxScore: 5, Summary: "ok"}}),
		Suggestions:   mustJSON(t, []string{"next"}),
		Strengths:     mustJSON(t, []string{"strength"}),
		Risks:         mustJSON(t, []string{"risk"}),
		Evidence:      mustJSON(t, []types.EvaluationEvidence{{Role: "user", Content: "hello"}}),
		SourceLastMessageID: sql.NullInt64{
			Int64: 42,
			Valid: true,
		},
		SourceLastMessageAt: sql.NullTime{
			Time:  now.Add(-30 * time.Minute),
			Valid: true,
		},
		FirstGeneratedAt: now.Add(-2 * time.Hour),
		GeneratedAt:      now.Add(-1 * time.Hour),
		UpdatedAt:        now.Add(-30 * time.Minute),
	}

	tests := []struct {
		name            string
		existing        *model.SessionEvaluation
		latestMessageAt evaluationMessageWatermarkRow
		want            bool
	}{
		{
			name:            "nil record refreshes",
			existing:        nil,
			latestMessageAt: evaluationMessageWatermarkRow{},
			want:            true,
		},
		{
			name:            "fresh complete record does not refresh",
			existing:        completeRecord,
			latestMessageAt: evaluationMessageWatermarkRow{LastMessageID: sql.NullInt64{Int64: 41, Valid: true}},
			want:            false,
		},
		{
			name:            "newer message refreshes",
			existing:        completeRecord,
			latestMessageAt: evaluationMessageWatermarkRow{LastMessageID: sql.NullInt64{Int64: 43, Valid: true}},
			want:            true,
		},
		{
			name:            "no latest message keeps cache",
			existing:        completeRecord,
			latestMessageAt: evaluationMessageWatermarkRow{},
			want:            false,
		},
		{
			name: "incomplete record refreshes",
			existing: &model.SessionEvaluation{
				Status:        "ready",
				RubricVersion: evaluation.RubricVersion,
				ScoreSource:   "llm",
				Dimensions:    []byte(`invalid-json`),
				Suggestions:   mustJSON(t, []string{"next"}),
				Strengths:     mustJSON(t, []string{"strength"}),
				Risks:         mustJSON(t, []string{"risk"}),
				Evidence:      mustJSON(t, []types.EvaluationEvidence{{Role: "user", Content: "hello"}}),
				SourceLastMessageID: sql.NullInt64{
					Int64: 42,
					Valid: true,
				},
				SourceLastMessageAt: sql.NullTime{
					Time:  now.Add(-30 * time.Minute),
					Valid: true,
				},
				FirstGeneratedAt: now.Add(-2 * time.Hour),
				GeneratedAt:      now.Add(-1 * time.Hour),
				UpdatedAt:        now.Add(-30 * time.Minute),
			},
			latestMessageAt: evaluationMessageWatermarkRow{},
			want:            true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldRefreshEvaluation(tt.existing, tt.latestMessageAt); got != tt.want {
				t.Fatalf("shouldRefreshEvaluation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvaluationRecordIncompleteDetectsCorruptedSupplementalJSON(t *testing.T) {
	now := time.Date(2026, 4, 13, 20, 0, 0, 0, time.UTC)
	record := &model.SessionEvaluation{
		Status:        "ready",
		OverallScore:  88,
		RubricVersion: evaluation.RubricVersion,
		ScoreSource:   "llm",
		Dimensions:    mustJSON(t, []types.EvaluationDimension{{Key: "technical_depth", Label: "技术深度", Score: 4, MaxScore: 5, Summary: "ok"}, {Key: "engineering_practice", Label: "工程实践", Score: 4, MaxScore: 5, Summary: "ok"}, {Key: "architecture_sense", Label: "架构意识", Score: 4, MaxScore: 5, Summary: "ok"}, {Key: "communication", Label: "表达与沟通", Score: 4, MaxScore: 5, Summary: "ok"}}),
		Suggestions:   mustJSON(t, []string{"next"}),
		Strengths:     []byte(`broken`),
		Risks:         mustJSON(t, []string{"risk"}),
		Evidence:      mustJSON(t, []types.EvaluationEvidence{{Role: "user", Content: "hello"}}),
		SourceLastMessageID: sql.NullInt64{
			Int64: 42,
			Valid: true,
		},
		SourceLastMessageAt: sql.NullTime{
			Time:  now.Add(-30 * time.Minute),
			Valid: true,
		},
		FirstGeneratedAt: now.Add(-2 * time.Hour),
		GeneratedAt:      now.Add(-1 * time.Hour),
		UpdatedAt:        now.Add(-30 * time.Minute),
	}

	if !evaluationRecordIncomplete(record) {
		t.Fatal("evaluationRecordIncomplete() = false, want true")
	}
}

func TestChooseLatestMessageTime(t *testing.T) {
	older := sql.NullTime{Time: time.Date(2026, 4, 13, 19, 0, 0, 0, time.UTC), Valid: true}
	newer := sql.NullTime{Time: time.Date(2026, 4, 13, 20, 0, 0, 0, time.UTC), Valid: true}

	if got := chooseLatestMessageTime(older, newer); !got.Time.Equal(newer.Time) {
		t.Fatalf("chooseLatestMessageTime() = %v, want %v", got.Time, newer.Time)
	}
	if got := chooseLatestMessageTime(older, sql.NullTime{}); !got.Time.Equal(older.Time) {
		t.Fatalf("chooseLatestMessageTime() = %v, want %v", got.Time, older.Time)
	}
}

func mustJSON(t *testing.T, value any) []byte {
	t.Helper()

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	return data
}
