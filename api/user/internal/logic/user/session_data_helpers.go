package user

import (
	"context"
	"database/sql"
	"time"

	"GoZero-AI/api/user/internal/types"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type sessionDataMessageRow struct {
	Id        int64     `db:"id"`
	Role      string    `db:"role"`
	Content   string    `db:"content"`
	CreatedAt time.Time `db:"created_at"`
}

func loadSessionMessages(ctx context.Context, db sqlx.SqlConn, sessionID string, userID int64, limit int64) ([]types.SessionMessage, error) {
	rows, err := loadSessionMessageRows(ctx, db, sessionID, userID, limit)
	if err != nil {
		return nil, err
	}

	messages := make([]types.SessionMessage, 0, len(rows))
	for _, row := range rows {
		messages = append(messages, types.SessionMessage{
			Role:      row.Role,
			Content:   row.Content,
			CreatedAt: row.CreatedAt.Format(timeLayout),
		})
	}
	return messages, nil
}

func loadSessionMessageRows(ctx context.Context, db sqlx.SqlConn, sessionID string, userID int64, limit int64) ([]sessionDataMessageRow, error) {
	query := `select id, role, content, created_at
from "public"."vector_store"
where chat_id = $1 and user_id = $2 and doc_type = 'message'
order by created_at asc, id asc`
	args := []any{sessionID, userID}
	if limit > 0 {
		query += ` limit $3`
		args = append(args, limit)
	}

	var rows []sessionDataMessageRow
	err := db.QueryRowsCtx(ctx, &rows, query, args...)
	switch err {
	case nil:
		return rows, nil
	case sqlx.ErrNotFound, sql.ErrNoRows:
		return []sessionDataMessageRow{}, nil
	default:
		return nil, err
	}
}
