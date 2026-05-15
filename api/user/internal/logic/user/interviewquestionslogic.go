package user

import (
	"context"
	"strings"

	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/api/user/internal/types"
	"GoZero-AI/api/user/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type InterviewQuestionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewInterviewQuestionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InterviewQuestionsLogic {
	return &InterviewQuestionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *InterviewQuestionsLogic) InterviewQuestions(req *types.InterviewQuestionsReq) (*types.InterviewQuestionsResp, error) {
	if _, err := currentUserID(l.ctx); err != nil {
		return nil, err
	}

	limit := normalizeQuestionListLimit(req.Limit)
	offset := normalizeQuestionListOffset(req.Offset)
	rows, total, err := l.svcCtx.InterviewQuestionsModel.List(l.ctx, model.InterviewQuestionListOptions{
		DirectionKey: strings.TrimSpace(req.DirectionKey),
		Difficulty:   req.Difficulty,
		FocusKeys:    parseInterviewPlanFocusKeys(req.FocusKeys),
		Keyword:      strings.TrimSpace(req.Keyword),
		Limit:        limit,
		Offset:       offset,
		Sort:         strings.TrimSpace(req.Sort),
	})
	if err != nil {
		return nil, err
	}

	questions := make([]types.InterviewQuestionItem, 0, len(rows))
	for _, row := range rows {
		questions = append(questions, buildInterviewQuestionItem(row))
	}

	return &types.InterviewQuestionsResp{
		Questions: questions,
		Total:     total,
		Limit:     limit,
		Offset:    offset,
		HasMore:   offset+int64(len(questions)) < total,
		QuestionMeta: types.ReportMeta{
			SchemaVersion: interviewQuestionSchemaVersion,
			Available:     len(questions) > 0,
		},
	}, nil
}
