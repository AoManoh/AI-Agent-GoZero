package user

import (
	"context"
	"strconv"
	"strings"
	"time"

	"GoZero-AI/api/user/internal/auth"
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
	demoInterviewSceneSourceUser        = "user"
	demoInterviewSceneSourceAdmin       = "admin"
)

type DemoInterviewSceneRandomLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

type demoInterviewSceneSessionRow struct {
	SessionId             string `db:"session_id"`
	Title                 string `db:"title"`
	Mode                  string `db:"mode"`
	DirectionLabel        string `db:"direction_label"`
	DifficultyLabel       string `db:"difficulty_label"`
	InterviewerStyleLabel string `db:"interviewer_style_label"`
	FollowUpDepth         string `db:"follow_up_depth"`
}

type demoInterviewSceneMessageRow struct {
	Role      string    `db:"role"`
	Content   string    `db:"content"`
	CreatedAt time.Time `db:"created_at"`
}

type demoInterviewSceneCandidate struct {
	UserID      int64
	Source      string
	SourceLabel string
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

	for _, candidate := range l.demoInterviewSceneCandidates() {
		resp, ok, err := l.loadDemoInterviewScene(candidate, limit, minMessageCount)
		if err != nil {
			return nil, err
		}
		if ok {
			return resp, nil
		}
	}

	return emptyDemoInterviewSceneResp(), nil
}

func (l *DemoInterviewSceneRandomLogic) loadDemoInterviewScene(candidate demoInterviewSceneCandidate, limit, minMessageCount int) (*types.DemoInterviewSceneRandomResp, bool, error) {
	var sessions []demoInterviewSceneSessionRow
	err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &sessions, `select
  session_id,
  title,
  mode,
  coalesce(nullif(btrim(direction_label), ''), 'Go 后端') as direction_label,
  coalesce(nullif(btrim(difficulty_label), ''), '中级') as difficulty_label,
  coalesce(nullif(btrim(interviewer_style_label), ''), '资深技术官') as interviewer_style_label,
  coalesce(nullif(btrim(follow_up_depth), ''), 'N+3') as follow_up_depth
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
limit 1`, candidate.UserID, sessionmode.KeyInterview, minMessageCount, openai.ChatMessageRoleAssistant, openai.ChatMessageRoleUser)
	if err != nil && err != sqlx.ErrNotFound {
		return nil, false, err
	}
	if len(sessions) == 0 {
		return nil, false, nil
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
limit $5`, session.SessionId, candidate.UserID, openai.ChatMessageRoleUser, openai.ChatMessageRoleAssistant, maxDemoInterviewSceneScanSize)
	if err != nil && err != sqlx.ErrNotFound {
		return nil, false, err
	}

	messages := buildDemoInterviewSceneMessages(rows, limit)
	if len(messages) == 0 {
		return nil, false, nil
	}

	modeKey := normalizeSessionMode(session.Mode)
	return &types.DemoInterviewSceneRandomResp{
		Available:             true,
		SessionId:             session.SessionId,
		Title:                 session.Title,
		Mode:                  sessionmode.DisplayName(modeKey),
		ModeKey:               modeKey,
		Source:                candidate.Source,
		SourceLabel:           candidate.SourceLabel,
		DirectionLabel:        defaultStringValue(session.DirectionLabel, "Go 后端"),
		DifficultyLabel:       defaultStringValue(session.DifficultyLabel, "中级"),
		InterviewerStyleLabel: defaultStringValue(session.InterviewerStyleLabel, "资深技术官"),
		FollowUpDepth:         defaultStringValue(session.FollowUpDepth, "N+3"),
		Messages:              messages,
	}, true, nil
}

func (l *DemoInterviewSceneRandomLogic) demoInterviewSceneCandidates() []demoInterviewSceneCandidate {
	if userID, ok := l.optionalCurrentUserID(); ok {
		return []demoInterviewSceneCandidate{
			{
				UserID:      userID,
				Source:      demoInterviewSceneSourceUser,
				SourceLabel: "我的面试记录",
			},
		}
	}

	return []demoInterviewSceneCandidate{
		{
			UserID:      demoInterviewSceneUserID,
			Source:      demoInterviewSceneSourceAdmin,
			SourceLabel: "管理员演示",
		},
	}
}

func (l *DemoInterviewSceneRandomLogic) optionalCurrentUserID() (int64, bool) {
	token := accessTokenFromContext(l.ctx)
	if token == "" || l.svcCtx == nil || l.svcCtx.Config.Auth.AccessSecret == "" {
		return 0, false
	}

	claims, err := auth.ParseTokenWithType(l.svcCtx.Config.Auth.AccessSecret, token, auth.TokenTypeAccess)
	if err != nil || claims.UserID <= 0 {
		return 0, false
	}

	return claims.UserID, true
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
