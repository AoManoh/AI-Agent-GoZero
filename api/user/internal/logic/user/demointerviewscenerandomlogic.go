package user

import (
	"context"
	"strconv"
	"strings"
	"time"

	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/api/user/internal/types"
	"GoZero-AI/internal/sessionmode"

	"github.com/sashabaranov/go-openai"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	demoInterviewSceneUserID      int64 = 1
	defaultDemoInterviewSceneSize       = 3
	maxDemoInterviewSceneSize           = 6
	maxDemoInterviewSceneScanSize       = 64
)

type DemoInterviewSceneRandomLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

type demoInterviewSceneSessionRow struct {
	SessionId string `db:"session_id"`
	Title     string `db:"title"`
	Mode      string `db:"mode"`
}

type demoInterviewSceneMessageRow struct {
	Role      string    `db:"role"`
	Content   string    `db:"content"`
	CreatedAt time.Time `db:"created_at"`
}

func NewDemoInterviewSceneRandomLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DemoInterviewSceneRandomLogic {
	return &DemoInterviewSceneRandomLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DemoInterviewSceneRandomLogic) DemoInterviewSceneRandom(req *types.DemoInterviewSceneRandomReq) (*types.DemoInterviewSceneRandomResp, error) {
	limit := normalizeDemoInterviewSceneLimit(req)
	minMessageCount := limit + 1

	var sessions []demoInterviewSceneSessionRow
	err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &sessions, `select session_id, title, mode
from "public"."chat_sessions" cs
where cs.user_id = $1
  and cs.is_active = true
  and cs.mode = $2
  and cs.message_count >= $3
  and exists (
    select 1 from "public"."vector_store"
    where chat_id = cs.session_id
      and user_id = cs.user_id
      and doc_type = 'message'
      and role = $4
      and btrim(content) <> ''
  )
  and exists (
    select 1 from "public"."vector_store"
    where chat_id = cs.session_id
      and user_id = cs.user_id
      and doc_type = 'message'
      and role = $5
      and btrim(content) <> ''
  )
order by random()
limit 1`, demoInterviewSceneUserID, sessionmode.KeyInterview, minMessageCount, openai.ChatMessageRoleAssistant, openai.ChatMessageRoleUser)
	if err != nil && err != sqlx.ErrNotFound {
		return nil, err
	}
	if len(sessions) == 0 {
		return emptyDemoInterviewSceneResp(), nil
	}

	session := sessions[0]
	var rows []demoInterviewSceneMessageRow
	err = l.svcCtx.DB.QueryRowsCtx(l.ctx, &rows, `select role, content, created_at
from "public"."vector_store"
where chat_id = $1
  and user_id = $2
  and doc_type = 'message'
  and role in ($3, $4)
  and btrim(content) <> ''
order by created_at asc
limit $5`, session.SessionId, demoInterviewSceneUserID, openai.ChatMessageRoleUser, openai.ChatMessageRoleAssistant, maxDemoInterviewSceneScanSize)
	if err != nil && err != sqlx.ErrNotFound {
		return nil, err
	}

	messages := buildDemoInterviewSceneMessages(rows, limit)
	if len(messages) == 0 {
		return emptyDemoInterviewSceneResp(), nil
	}

	modeKey := normalizeSessionMode(session.Mode)
	return &types.DemoInterviewSceneRandomResp{
		Available: true,
		SessionId: session.SessionId,
		Title:     session.Title,
		Mode:      sessionmode.DisplayName(modeKey),
		ModeKey:   modeKey,
		Messages:  messages,
	}, nil
}

func normalizeDemoInterviewSceneLimit(req *types.DemoInterviewSceneRandomReq) int {
	if req == nil || req.Limit <= 0 {
		return defaultDemoInterviewSceneSize
	}
	if req.Limit > maxDemoInterviewSceneSize {
		return maxDemoInterviewSceneSize
	}
	if req.Limit < defaultDemoInterviewSceneSize {
		return defaultDemoInterviewSceneSize
	}
	return int(req.Limit)
}

func emptyDemoInterviewSceneResp() *types.DemoInterviewSceneRandomResp {
	return &types.DemoInterviewSceneRandomResp{
		Available: false,
		Messages:  []types.DemoInterviewSceneMessage{},
	}
}

func buildDemoInterviewSceneMessages(rows []demoInterviewSceneMessageRow, limit int) []types.DemoInterviewSceneMessage {
	window := selectDemoInterviewSceneWindow(rows, limit)
	messages := make([]types.DemoInterviewSceneMessage, 0, len(window))
	assistantTurns := 0

	for _, row := range window {
		role := normalizeDemoInterviewSceneRole(row.Role)
		content := strings.TrimSpace(row.Content)
		if role == "" || content == "" {
			continue
		}

		name := "你"
		if role == openai.ChatMessageRoleAssistant {
			if assistantTurns == 0 {
				name = "AI 面试官"
			} else {
				name = "AI · 追问 #" + strconv.Itoa(assistantTurns)
			}
			assistantTurns++
		}

		messages = append(messages, types.DemoInterviewSceneMessage{
			Role:      role,
			Name:      name,
			Content:   content,
			CreatedAt: row.CreatedAt.Format(timeLayout),
		})
	}

	return messages
}

func selectDemoInterviewSceneWindow(rows []demoInterviewSceneMessageRow, limit int) []demoInterviewSceneMessageRow {
	if limit <= 0 || len(rows) == 0 {
		return nil
	}
	if len(rows) <= limit {
		return append([]demoInterviewSceneMessageRow(nil), rows...)
	}

	for i := 0; i+limit <= len(rows); i++ {
		if normalizeDemoInterviewSceneRole(rows[i].Role) != openai.ChatMessageRoleAssistant {
			continue
		}
		window := rows[i : i+limit]
		if demoInterviewSceneWindowHasRole(window, openai.ChatMessageRoleUser) &&
			demoInterviewSceneWindowHasRole(window, openai.ChatMessageRoleAssistant) {
			return append([]demoInterviewSceneMessageRow(nil), window...)
		}
	}

	return append([]demoInterviewSceneMessageRow(nil), rows[:limit]...)
}

func demoInterviewSceneWindowHasRole(rows []demoInterviewSceneMessageRow, role string) bool {
	for _, row := range rows {
		if normalizeDemoInterviewSceneRole(row.Role) == role {
			return true
		}
	}
	return false
}

func normalizeDemoInterviewSceneRole(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case openai.ChatMessageRoleAssistant:
		return openai.ChatMessageRoleAssistant
	case openai.ChatMessageRoleUser:
		return openai.ChatMessageRoleUser
	default:
		return ""
	}
}
