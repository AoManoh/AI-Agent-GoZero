package model

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type (
	ChatSessionsModel interface {
		Create(ctx context.Context, userID int64, sessionID, title, mode string) (*ChatSession, error)
		Deactivate(ctx context.Context, userID int64, sessionID string) error
		FindByUserID(ctx context.Context, userID int64) ([]ChatSession, error)
		FindOneBySessionID(ctx context.Context, userID int64, sessionID string) (*ChatSession, error)
	}

	defaultChatSessionsModel struct {
		conn  sqlx.SqlConn
		table string
	}

	ChatSession struct {
		Id            int64        `db:"id"`
		SessionId     string       `db:"session_id"`
		UserId        int64        `db:"user_id"`
		Title         string       `db:"title"`
		Mode          string       `db:"mode"`
		CreatedAt     time.Time    `db:"created_at"`
		UpdatedAt     time.Time    `db:"updated_at"`
		LastMessageAt sql.NullTime `db:"last_message_at"`
		MessageCount  int64        `db:"message_count"`
		IsActive      bool         `db:"is_active"`
	}
)

func NewChatSessionsModel(conn sqlx.SqlConn) ChatSessionsModel {
	return &defaultChatSessionsModel{
		conn:  conn,
		table: `"public"."chat_sessions"`,
	}
}

func (m *defaultChatSessionsModel) Create(ctx context.Context, userID int64, sessionID, title, mode string) (*ChatSession, error) {
	query := fmt.Sprintf(`insert into %s (session_id, user_id, title, mode, is_active)
values ($1, $2, $3, $4, true)
returning id, session_id, user_id, title, mode, created_at, updated_at, last_message_at, message_count, is_active`, m.table)

	var resp ChatSession
	if err := m.conn.QueryRowCtx(ctx, &resp, query, sessionID, userID, title, mode); err != nil {
		return nil, err
	}
	return &resp, nil
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
	query := fmt.Sprintf(`select id, session_id, user_id, title, mode, created_at, updated_at, last_message_at, message_count, is_active
from %s
where user_id = $1 and is_active = true
order by coalesce(last_message_at, created_at) desc, id desc`, m.table)

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
	query := fmt.Sprintf(`select id, session_id, user_id, title, mode, created_at, updated_at, last_message_at, message_count, is_active
from %s
where user_id = $1 and session_id = $2 and is_active = true
limit 1`, m.table)

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
