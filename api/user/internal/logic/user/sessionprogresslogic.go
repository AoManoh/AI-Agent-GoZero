package user

import (
	"context"
	"errors"

	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/api/user/internal/types"
	"GoZero-AI/api/user/model"
	"GoZero-AI/internal/statuserr"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type SessionProgressLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

type sessionProgressMessageAggregateRow struct {
	UserTurns      int64 `db:"user_turns"`
	AssistantTurns int64 `db:"assistant_turns"`
}

func NewSessionProgressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SessionProgressLogic {
	return &SessionProgressLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SessionProgressLogic) SessionProgress(req *types.SessionProgressReq) (*types.SessionProgressResp, error) {
	userID, err := currentUserID(l.ctx)
	if err != nil {
		return nil, err
	}
	if err := validateInterviewPlanLimit(req.PlanLimit); err != nil {
		return nil, err
	}

	session, err := l.svcCtx.ChatSessionsModel.FindOneBySessionID(l.ctx, userID, req.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, statuserr.NotFound("会话不存在或已删除")
		}
		return nil, err
	}

	aggregate, err := loadSessionProgressMessageAggregate(l.ctx, l.svcCtx.DB, session.SessionId, userID)
	if err != nil {
		return nil, statuserr.ServiceUnavailable("会话进度暂不可用，请稍后重试")
	}

	resp := buildSessionProgressResp(*session, aggregate, req.PlanLimit)
	return &resp, nil
}

func loadSessionProgressMessageAggregate(ctx context.Context, db sqlx.SqlConn, sessionID string, userID int64) (sessionProgressMessageAggregateRow, error) {
	var aggregate sessionProgressMessageAggregateRow
	err := db.QueryRowCtx(ctx, &aggregate, `select
coalesce(sum(case when role = 'user' then 1 else 0 end), 0) as user_turns,
coalesce(sum(case when role = 'assistant' then 1 else 0 end), 0) as assistant_turns
from "public"."vector_store"
where chat_id = $1 and user_id = $2 and doc_type = 'message'`, sessionID, userID)
	if err != nil && err != sqlx.ErrNotFound {
		return sessionProgressMessageAggregateRow{}, err
	}
	return aggregate, nil
}

func buildSessionProgressResp(session model.ChatSession, aggregate sessionProgressMessageAggregateRow, rawPlanLimit int64) types.SessionProgressResp {
	config := buildSessionConfigSnapshot(session)
	questions := selectInterviewPlanQuestions(config, normalizeInterviewPlanLimit(rawPlanLimit))
	totalQuestions := int64(len(questions))
	completedQuestions := min64(aggregate.UserTurns, totalQuestions)
	if session.CompletedAt.Valid || session.ProgressPercent >= 100 {
		completedQuestions = totalQuestions
	}

	currentQuestionIndex := int64(0)
	if totalQuestions > 0 {
		currentQuestionIndex = min64(completedQuestions+1, totalQuestions)
	}

	var nextQuestion *types.InterviewPlanQuestion
	if completedQuestions < totalQuestions {
		question := questions[completedQuestions]
		nextQuestion = &question
	}

	return types.SessionProgressResp{
		Session:              buildSessionItem(session),
		Config:               config,
		TotalQuestions:       totalQuestions,
		CompletedQuestions:   completedQuestions,
		CurrentQuestionIndex: currentQuestionIndex,
		ProgressPercent:      percentage(completedQuestions, totalQuestions),
		UserTurns:            aggregate.UserTurns,
		AssistantTurns:       aggregate.AssistantTurns,
		FocusProgress:        buildSessionFocusProgress(config.FocusAreas, questions, completedQuestions),
		NextQuestion:         nextQuestion,
		ProgressMeta: types.ReportMeta{
			SchemaVersion: "session-progress-v1",
			Available:     true,
		},
	}
}

func buildSessionFocusProgress(focusAreas []types.FocusAreaSelection, questions []types.InterviewPlanQuestion, completedQuestions int64) []types.SessionFocusProgress {
	type focusCounter struct {
		label     string
		planned   int64
		completed int64
	}

	order := make([]string, 0, len(focusAreas))
	counters := make(map[string]*focusCounter, len(focusAreas))
	for _, area := range focusAreas {
		if area.Key == "" {
			continue
		}
		if _, ok := counters[area.Key]; ok {
			continue
		}
		order = append(order, area.Key)
		counters[area.Key] = &focusCounter{label: area.Label}
	}

	for index, question := range questions {
		counter, ok := counters[question.FocusKey]
		if !ok {
			order = append(order, question.FocusKey)
			counter = &focusCounter{label: question.FocusLabel}
			counters[question.FocusKey] = counter
		}
		if counter.label == "" {
			counter.label = question.FocusLabel
		}
		counter.planned++
		if int64(index+1) <= completedQuestions {
			counter.completed++
		}
	}

	progress := make([]types.SessionFocusProgress, 0, len(order))
	for _, key := range order {
		counter := counters[key]
		progress = append(progress, types.SessionFocusProgress{
			Key:                key,
			Label:              counter.label,
			PlannedQuestions:   counter.planned,
			CompletedQuestions: counter.completed,
			ProgressPercent:    percentage(counter.completed, counter.planned),
		})
	}
	return progress
}

func percentage(value, total int64) int64 {
	if total <= 0 {
		return 0
	}
	if value <= 0 {
		return 0
	}
	if value >= total {
		return 100
	}
	return value * 100 / total
}
