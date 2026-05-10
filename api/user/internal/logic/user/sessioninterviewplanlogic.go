package user

import (
	"context"
	"errors"

	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/api/user/internal/types"
	"GoZero-AI/api/user/model"
	"GoZero-AI/internal/statuserr"

	"github.com/zeromicro/go-zero/core/logx"
)

type SessionInterviewPlanLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSessionInterviewPlanLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SessionInterviewPlanLogic {
	return &SessionInterviewPlanLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SessionInterviewPlanLogic) SessionInterviewPlan(req *types.SessionInterviewPlanReq) (*types.InterviewPlanResp, error) {
	userID, err := currentUserID(l.ctx)
	if err != nil {
		return nil, err
	}
	if err := validateInterviewPlanLimit(req.Limit); err != nil {
		return nil, err
	}

	session, err := l.svcCtx.ChatSessionsModel.FindOneBySessionID(l.ctx, userID, req.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, statuserr.NotFound("会话不存在或已删除")
		}
		return nil, err
	}

	config := buildSessionConfigSnapshot(*session)
	if l.svcCtx.InterviewQuestionsModel != nil {
		if rows, _, err := l.svcCtx.InterviewQuestionsModel.List(l.ctx, model.InterviewQuestionListOptions{
			DirectionKey: config.DirectionKey,
			Difficulty:   config.DifficultyLevel,
			FocusKeys:    interviewPlanFocusKeys(config),
			Limit:        normalizeInterviewPlanLimit(req.Limit),
			Sort:         "hot",
		}); err == nil && len(rows) > 0 {
			questions := make([]types.InterviewPlanQuestion, 0, len(rows))
			for _, row := range rows {
				questions = append(questions, buildInterviewPlanQuestionFromBank(row))
			}
			return &types.InterviewPlanResp{
				Config:     config,
				Questions:  questions,
				Milestones: buildInterviewPlanMilestones(int64(len(questions))),
				PlanMeta: types.ReportMeta{
					SchemaVersion: "interview-plan-v1",
					Available:     true,
				},
			}, nil
		} else if err != nil {
			l.Errorf("load session interview question bank plan failed: %v", err)
		}
	}

	resp := buildInterviewPlanResp(config, req.Limit)
	return &resp, nil
}
