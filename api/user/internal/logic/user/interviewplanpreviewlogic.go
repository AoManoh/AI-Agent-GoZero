package user

import (
	"context"

	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/api/user/internal/types"
	"GoZero-AI/api/user/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type InterviewPlanPreviewLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewInterviewPlanPreviewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InterviewPlanPreviewLogic {
	return &InterviewPlanPreviewLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *InterviewPlanPreviewLogic) InterviewPlanPreview(req *types.InterviewPlanPreviewReq) (*types.InterviewPlanResp, error) {
	if _, err := currentUserID(l.ctx); err != nil {
		return nil, err
	}
	if err := validateInterviewPlanLimit(req.Limit); err != nil {
		return nil, err
	}

	config, err := buildInterviewPlanPreviewConfig(req)
	if err != nil {
		return nil, err
	}

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
			l.Errorf("load interview question bank plan failed: %v", err)
		}
	}

	resp := buildInterviewPlanResp(config, req.Limit)
	return &resp, nil
}

func interviewPlanFocusKeys(config types.SessionConfigSnapshot) []string {
	keys := make([]string, 0, len(config.FocusAreas))
	for _, focus := range config.FocusAreas {
		if focus.Key != "" {
			keys = append(keys, focus.Key)
		}
	}
	return keys
}
