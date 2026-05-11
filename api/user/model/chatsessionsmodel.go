package model

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type (
	ChatSessionsModel interface {
		Create(ctx context.Context, userID int64, sessionID, title, mode string) (*ChatSession, error)
		CreateWithConfig(ctx context.Context, userID int64, sessionID, title, mode string, config SessionCreateConfig) (*ChatSession, error)
		Complete(ctx context.Context, userID int64, sessionID string) (*ChatSession, error)
		Deactivate(ctx context.Context, userID int64, sessionID string) error
		FindByUserID(ctx context.Context, userID int64) ([]ChatSession, error)
		FindOneBySessionID(ctx context.Context, userID int64, sessionID string) (*ChatSession, error)
	}

	defaultChatSessionsModel struct {
		conn  sqlx.SqlConn
		table string
	}

	ChatSession struct {
		Id                    int64        `db:"id"`
		SessionId             string       `db:"session_id"`
		UserId                int64        `db:"user_id"`
		Title                 string       `db:"title"`
		Mode                  string       `db:"mode"`
		DirectionKey          string       `db:"direction_key"`
		DirectionLabel        string       `db:"direction_label"`
		DifficultyLevel       int64        `db:"difficulty_level"`
		DifficultyLabel       string       `db:"difficulty_label"`
		InterviewerStyle      string       `db:"interviewer_style"`
		InterviewerStyleLabel string       `db:"interviewer_style_label"`
		FocusAreas            []byte       `db:"focus_areas"`
		FollowUpDepth         string       `db:"follow_up_depth"`
		EstimatedMinutes      int64        `db:"estimated_minutes"`
		ProgressPercent       int64        `db:"progress_percent"`
		CreatedAt             time.Time    `db:"created_at"`
		UpdatedAt             time.Time    `db:"updated_at"`
		LastMessageAt         sql.NullTime `db:"last_message_at"`
		StartedAt             sql.NullTime `db:"started_at"`
		CompletedAt           sql.NullTime `db:"completed_at"`
		DurationSeconds       int64        `db:"duration_seconds"`
		ResumeArtifactId      string       `db:"resume_artifact_id"`
		MessageCount          int64        `db:"message_count"`
		IsActive              bool         `db:"is_active"`
	}

	SessionCreateConfig struct {
		DirectionKey          string
		DirectionLabel        string
		DifficultyLevel       int64
		DifficultyLabel       string
		InterviewerStyle      string
		InterviewerStyleLabel string
		FocusAreas            any
		FollowUpDepth         string
		EstimatedMinutes      int64
		ResumeArtifactId      string
	}
)

type chatSessionRunner interface {
	QueryRowCtx(ctx context.Context, v any, query string, args ...any) error
}

const chatSessionSelectFields = `id, session_id, user_id, title, mode,
direction_key, direction_label, difficulty_level, difficulty_label,
	interviewer_style, interviewer_style_label, focus_areas, follow_up_depth,
	estimated_minutes, progress_percent, created_at, updated_at, last_message_at,
	started_at, completed_at, duration_seconds, resume_artifact_id, message_count, is_active`

func NewChatSessionsModel(conn sqlx.SqlConn) ChatSessionsModel {
	return &defaultChatSessionsModel{
		conn:  conn,
		table: `"public"."chat_sessions"`,
	}
}

func (m *defaultChatSessionsModel) Create(ctx context.Context, userID int64, sessionID, title, mode string) (*ChatSession, error) {
	return m.CreateWithConfig(ctx, userID, sessionID, title, mode, SessionCreateConfig{})
}

func (m *defaultChatSessionsModel) CreateWithConfig(ctx context.Context, userID int64, sessionID, title, mode string, config SessionCreateConfig) (*ChatSession, error) {
	return m.createWithConfig(ctx, m.conn, userID, sessionID, title, mode, config)
}

func CreateChatSessionWithConfigTx(ctx context.Context, session sqlx.Session, userID int64, sessionID, title, mode string, config SessionCreateConfig) (*ChatSession, error) {
	return (&defaultChatSessionsModel{table: `"public"."chat_sessions"`}).createWithConfig(ctx, session, userID, sessionID, title, mode, config)
}

func (m *defaultChatSessionsModel) createWithConfig(ctx context.Context, runner chatSessionRunner, userID int64, sessionID, title, mode string, config SessionCreateConfig) (*ChatSession, error) {
	focusAreas, err := marshalFocusAreas(config.FocusAreas)
	if err != nil {
		return nil, err
	}
	query := fmt.Sprintf(`insert into %s (
session_id, user_id, title, mode,
direction_key, direction_label, difficulty_level, difficulty_label,
interviewer_style, interviewer_style_label, focus_areas, follow_up_depth,
estimated_minutes, progress_percent, started_at, is_active,
resume_artifact_id
)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11::jsonb, $12, $13, 0, now(), true, $14)
	returning %s`, m.table, chatSessionSelectFields)

	var resp ChatSession
	if err := runner.QueryRowCtx(ctx, &resp, query,
		sessionID,
		userID,
		title,
		mode,
		defaultString(config.DirectionKey, "go_backend"),
		defaultString(config.DirectionLabel, "Go 后端"),
		defaultInt64(config.DifficultyLevel, 3),
		defaultString(config.DifficultyLabel, "中级"),
		defaultString(config.InterviewerStyle, "senior"),
		defaultString(config.InterviewerStyleLabel, "资深技术官"),
		string(focusAreas),
		defaultString(config.FollowUpDepth, "N+3"),
		defaultInt64(config.EstimatedMinutes, 30),
		config.ResumeArtifactId,
	); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (m *defaultChatSessionsModel) Complete(ctx context.Context, userID int64, sessionID string) (*ChatSession, error) {
	query := fmt.Sprintf(`update %s
set completed_at = coalesce(completed_at, now()),
    progress_percent = 100,
    duration_seconds = case
        when duration_seconds > 0 then duration_seconds
        when started_at is not null then greatest(1, extract(epoch from (now() - started_at))::integer)
        else duration_seconds
    end,
    updated_at = now()
where user_id = $1 and session_id = $2 and is_active = true
returning %s`, m.table, chatSessionSelectFields)

	var resp ChatSession
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

func (m *defaultChatSessionsModel) Deactivate(ctx context.Context, userID int64, sessionID string) error {
	query := fmt.Sprintf(`update %s
set is_active = false,
    updated_at = now()
where user_id = $1 and session_id = $2`, m.table)

	result, err := m.conn.ExecCtx(ctx, query, userID, sessionID)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (m *defaultChatSessionsModel) FindByUserID(ctx context.Context, userID int64) ([]ChatSession, error) {
	query := fmt.Sprintf(`select %s
from %s
where user_id = $1 and is_active = true
order by coalesce(last_message_at, updated_at, created_at) desc, id desc`, chatSessionSelectFields, m.table)

	var resp []ChatSession
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userID)
	switch err {
	case nil:
		return resp, nil
	case sqlx.ErrNotFound, sql.ErrNoRows:
		return []ChatSession{}, nil
	default:
		return nil, err
	}
}

func (m *defaultChatSessionsModel) FindOneBySessionID(ctx context.Context, userID int64, sessionID string) (*ChatSession, error) {
	query := fmt.Sprintf(`select %s
from %s
where user_id = $1 and session_id = $2 and is_active = true
limit 1`, chatSessionSelectFields, m.table)

	var resp ChatSession
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

func marshalFocusAreas(value any) ([]byte, error) {
	if value == nil {
		return []byte("[]"), nil
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 || string(raw) == "null" {
		return []byte("[]"), nil
	}
	return raw, nil
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func defaultInt64(value, fallback int64) int64 {
	if value == 0 {
		return fallback
	}
	return value
}
