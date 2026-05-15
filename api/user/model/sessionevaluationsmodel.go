package model

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type (
	SessionEvaluationsModel interface {
		FindOneBySessionID(ctx context.Context, userID int64, sessionID string) (*SessionEvaluation, error)
		Upsert(ctx context.Context, data *SessionEvaluation) error
	}

	defaultSessionEvaluationsModel struct {
		conn  sqlx.SqlConn
		table string
	}

	SessionEvaluation struct {
		Id                  int64         `db:"id"`
		SessionId           string        `db:"session_id"`
		UserId              int64         `db:"user_id"`
		Status              string        `db:"status"`
		Summary             string        `db:"summary"`
		UserTurns           int64         `db:"user_turns"`
		AssistantTurns      int64         `db:"assistant_turns"`
		OverallScore        float64       `db:"overall_score"`
		RubricVersion       string        `db:"rubric_version"`
		ScoreSource         string        `db:"score_source"`
		Dimensions          []byte        `db:"dimensions"`
		Strengths           []byte        `db:"strengths"`
		Risks               []byte        `db:"risks"`
		Suggestions         []byte        `db:"suggestions"`
		Evidence            []byte        `db:"evidence"`
		SourceLastMessageID sql.NullInt64 `db:"source_last_message_id"`
		SourceLastMessageAt sql.NullTime  `db:"source_last_message_at"`
		FirstGeneratedAt    time.Time     `db:"first_generated_at"`
		GeneratedAt         time.Time     `db:"generated_at"`
		UpdatedAt           time.Time     `db:"updated_at"`
	}
)

func NewSessionEvaluationsModel(conn sqlx.SqlConn) SessionEvaluationsModel {
	return &defaultSessionEvaluationsModel{
		conn:  conn,
		table: `"public"."session_evaluations"`,
	}
}

func (m *defaultSessionEvaluationsModel) FindOneBySessionID(ctx context.Context, userID int64, sessionID string) (*SessionEvaluation, error) {
	query := fmt.Sprintf(`select id, session_id, user_id, status, summary, user_turns, assistant_turns,
overall_score, rubric_version, score_source, dimensions, strengths, risks, suggestions, evidence, source_last_message_id, source_last_message_at, first_generated_at, generated_at, updated_at
from %s
where user_id = $1 and session_id = $2
limit 1`, m.table)

	var resp SessionEvaluation
	err := m.conn.QueryRowCtx(ctx, &resp, query, userID, sessionID)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound, sql.ErrNoRows:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultSessionEvaluationsModel) Upsert(ctx context.Context, data *SessionEvaluation) error {
	query := fmt.Sprintf(`insert into %s
(session_id, user_id, status, summary, user_turns, assistant_turns, overall_score, rubric_version, score_source, dimensions, strengths, risks, suggestions, evidence, source_last_message_id, source_last_message_at, first_generated_at, generated_at, updated_at)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10::jsonb, $11::jsonb, $12::jsonb, $13::jsonb, $14::jsonb, $15, $16, $17, $18, $19)
on conflict (session_id, user_id) do update set
status = excluded.status,
summary = excluded.summary,
user_turns = excluded.user_turns,
assistant_turns = excluded.assistant_turns,
overall_score = excluded.overall_score,
rubric_version = excluded.rubric_version,
score_source = excluded.score_source,
dimensions = excluded.dimensions,
strengths = excluded.strengths,
risks = excluded.risks,
suggestions = excluded.suggestions,
evidence = excluded.evidence,
source_last_message_id = excluded.source_last_message_id,
source_last_message_at = excluded.source_last_message_at,
first_generated_at = %s.first_generated_at,
generated_at = excluded.generated_at,
updated_at = excluded.updated_at`, m.table, m.table)

	_, err := m.conn.ExecCtx(ctx, query,
		data.SessionId,
		data.UserId,
		data.Status,
		data.Summary,
		data.UserTurns,
		data.AssistantTurns,
		data.OverallScore,
		data.RubricVersion,
		data.ScoreSource,
		string(data.Dimensions),
		string(data.Strengths),
		string(data.Risks),
		string(data.Suggestions),
		string(data.Evidence),
		data.SourceLastMessageID,
		data.SourceLastMessageAt,
		data.FirstGeneratedAt,
		data.GeneratedAt,
		data.UpdatedAt,
	)
	return err
}
