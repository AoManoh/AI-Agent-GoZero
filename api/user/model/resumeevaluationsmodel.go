package model

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type (
	ResumeEvaluationsModel interface {
		FindOneByArtifactID(ctx context.Context, userID int64, artifactID string) (*ResumeEvaluation, error)
		Upsert(ctx context.Context, data *ResumeEvaluation) error
	}

	defaultResumeEvaluationsModel struct {
		conn  sqlx.SqlConn
		table string
	}

	ResumeEvaluation struct {
		Id                  int64        `db:"id"`
		ArtifactId          string       `db:"artifact_id"`
		UserId              int64        `db:"user_id"`
		Status              string       `db:"status"`
		Summary             string       `db:"summary"`
		OverallScore        float64      `db:"overall_score"`
		Level               string       `db:"level"`
		RubricVersion       string       `db:"rubric_version"`
		ScoreSource         string       `db:"score_source"`
		DirectionKey        string       `db:"direction_key"`
		Dimensions          []byte       `db:"dimensions"`
		Strengths           []byte       `db:"strengths"`
		Risks               []byte       `db:"risks"`
		Suggestions         []byte       `db:"suggestions"`
		FocusMatches        []byte       `db:"focus_matches"`
		SuggestedQuestions  []byte       `db:"suggested_questions"`
		Evidence            []byte       `db:"evidence"`
		SourceResumeVersion int64        `db:"source_resume_version"`
		SourceChunkCount    int64        `db:"source_chunk_count"`
		SourceUpdatedAt     sql.NullTime `db:"source_updated_at"`
		FirstGeneratedAt    time.Time    `db:"first_generated_at"`
		GeneratedAt         time.Time    `db:"generated_at"`
		UpdatedAt           time.Time    `db:"updated_at"`
	}
)

func NewResumeEvaluationsModel(conn sqlx.SqlConn) ResumeEvaluationsModel {
	return &defaultResumeEvaluationsModel{
		conn:  conn,
		table: `"public"."resume_evaluations"`,
	}
}

func (m *defaultResumeEvaluationsModel) FindOneByArtifactID(ctx context.Context, userID int64, artifactID string) (*ResumeEvaluation, error) {
	query := fmt.Sprintf(`select id, artifact_id, user_id, status, summary, overall_score, level,
rubric_version, score_source, direction_key, dimensions, strengths, risks, suggestions,
focus_matches, suggested_questions, evidence, source_resume_version, source_chunk_count,
source_updated_at, first_generated_at, generated_at, updated_at
from %s
where user_id = $1 and artifact_id = $2
limit 1`, m.table)

	var resp ResumeEvaluation
	err := m.conn.QueryRowCtx(ctx, &resp, query, userID, artifactID)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound, sql.ErrNoRows:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultResumeEvaluationsModel) Upsert(ctx context.Context, data *ResumeEvaluation) error {
	query := fmt.Sprintf(`insert into %s
(artifact_id, user_id, status, summary, overall_score, level, rubric_version, score_source, direction_key,
dimensions, strengths, risks, suggestions, focus_matches, suggested_questions, evidence,
source_resume_version, source_chunk_count, source_updated_at, first_generated_at, generated_at, updated_at)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10::jsonb, $11::jsonb, $12::jsonb, $13::jsonb, $14::jsonb, $15::jsonb, $16::jsonb, $17, $18, $19, $20, $21, $22)
on conflict (artifact_id, user_id) do update set
status = excluded.status,
summary = excluded.summary,
overall_score = excluded.overall_score,
level = excluded.level,
rubric_version = excluded.rubric_version,
score_source = excluded.score_source,
direction_key = excluded.direction_key,
dimensions = excluded.dimensions,
strengths = excluded.strengths,
risks = excluded.risks,
suggestions = excluded.suggestions,
focus_matches = excluded.focus_matches,
suggested_questions = excluded.suggested_questions,
evidence = excluded.evidence,
source_resume_version = excluded.source_resume_version,
source_chunk_count = excluded.source_chunk_count,
source_updated_at = excluded.source_updated_at,
first_generated_at = %s.first_generated_at,
generated_at = excluded.generated_at,
updated_at = excluded.updated_at`, m.table, m.table)

	_, err := m.conn.ExecCtx(ctx, query,
		data.ArtifactId,
		data.UserId,
		data.Status,
		data.Summary,
		data.OverallScore,
		data.Level,
		data.RubricVersion,
		data.ScoreSource,
		data.DirectionKey,
		string(data.Dimensions),
		string(data.Strengths),
		string(data.Risks),
		string(data.Suggestions),
		string(data.FocusMatches),
		string(data.SuggestedQuestions),
		string(data.Evidence),
		data.SourceResumeVersion,
		data.SourceChunkCount,
		data.SourceUpdatedAt,
		data.FirstGeneratedAt,
		data.GeneratedAt,
		data.UpdatedAt,
	)
	return err
}
