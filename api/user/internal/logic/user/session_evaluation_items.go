package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"GoZero-AI/api/user/internal/types"
	"GoZero-AI/api/user/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type sessionEvaluationItemRow struct {
	TurnIndex       int64         `db:"turn_index"`
	Question        string        `db:"question"`
	Answer          string        `db:"answer"`
	AiComment       string        `db:"ai_comment"`
	Score           float64       `db:"score"`
	MaxScore        float64       `db:"max_score"`
	Tags            []byte        `db:"tags"`
	SourceMessageID sql.NullInt64 `db:"source_message_id"`
	GeneratedAt     time.Time     `db:"generated_at"`
}

func replaceSessionEvaluationItems(ctx context.Context, db sqlx.SqlConn, session model.ChatSession, record *model.SessionEvaluation, rows []evaluationMessageRow) error {
	evaluationResp, err := buildResponseFromRecord(session, record)
	if err != nil {
		return err
	}

	messageRows := make([]sessionDataMessageRow, 0, len(rows))
	userRows := make([]sessionDataMessageRow, 0)
	for _, row := range rows {
		messageRow := sessionDataMessageRow{
			Id:        row.Id,
			Role:      row.Role,
			Content:   row.Content,
			CreatedAt: row.CreatedAt,
		}
		messageRows = append(messageRows, messageRow)
		if row.Role == "user" {
			userRows = append(userRows, messageRow)
		}
	}

	cards := buildSessionReportQuestionCards(messageRows, evaluationResp)
	return db.TransactCtx(ctx, func(ctx context.Context, tx sqlx.Session) error {
		if _, err := tx.ExecCtx(ctx, `delete from "public"."session_evaluation_items" where user_id = $1 and session_id = $2`, session.UserId, session.SessionId); err != nil {
			return err
		}

		for idx, card := range cards {
			tagsJSON, err := json.Marshal(card.Tags)
			if err != nil {
				return err
			}

			var (
				sourceID any
				sourceAt any
			)
			if idx < len(userRows) {
				sourceID = userRows[idx].Id
				sourceAt = userRows[idx].CreatedAt
			}

			if _, err := tx.ExecCtx(ctx, `insert into "public"."session_evaluation_items"
(session_id, user_id, turn_index, question, answer, ai_comment, score, max_score, tags, source_message_id, source_message_at, generated_at, updated_at)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb, $10, $11, $12, now())`,
				session.SessionId,
				session.UserId,
				card.TurnIndex,
				card.Question,
				card.Answer,
				card.AiComment,
				card.Score,
				card.MaxScore,
				string(tagsJSON),
				sourceID,
				sourceAt,
				record.GeneratedAt,
			); err != nil {
				return err
			}
		}

		return nil
	})
}

func loadPersistedSessionEvaluationItems(ctx context.Context, db sqlx.SqlConn, sessionID string, userID int64) ([]types.SessionReportQuestionCard, error) {
	var rows []sessionEvaluationItemRow
	err := db.QueryRowsCtx(ctx, &rows, `select
turn_index,
question,
answer,
ai_comment,
score,
max_score,
tags,
source_message_id,
generated_at
from "public"."session_evaluation_items"
where user_id = $1 and session_id = $2
order by turn_index asc`, userID, sessionID)
	if err != nil {
		if err == sql.ErrNoRows || err == sqlx.ErrNotFound {
			return []types.SessionReportQuestionCard{}, nil
		}
		return nil, err
	}

	cards := make([]types.SessionReportQuestionCard, 0, len(rows))
	for _, row := range rows {
		var tags []types.SessionReportTag
		if len(row.Tags) > 0 {
			if err := json.Unmarshal(row.Tags, &tags); err != nil {
				return nil, err
			}
		}

		cards = append(cards, types.SessionReportQuestionCard{
			TurnIndex: row.TurnIndex,
			Depth:     buildQuestionDepthLabel(int(row.TurnIndex), int64(len(rows)*2)),
			Question:  row.Question,
			Answer:    row.Answer,
			AiComment: row.AiComment,
			Score:     int64(row.Score),
			MaxScore:  int64(row.MaxScore),
			Tags:      tags,
		})
	}

	return cards, nil
}
