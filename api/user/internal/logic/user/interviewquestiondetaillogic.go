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

type InterviewQuestionDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewInterviewQuestionDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InterviewQuestionDetailLogic {
	return &InterviewQuestionDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *InterviewQuestionDetailLogic) InterviewQuestionDetail(req *types.InterviewQuestionDetailReq) (*types.InterviewQuestionDetailResp, error) {
	if _, err := currentUserID(l.ctx); err != nil {
		return nil, err
	}

	question, sources, err := l.svcCtx.InterviewQuestionsModel.FindOne(l.ctx, req.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, statuserr.NotFound("题目不存在或已下线")
		}
		return nil, err
	}

	sourceItems := make([]types.InterviewQuestionSourceItem, 0, len(sources))
	for _, source := range sources {
		sourceItems = append(sourceItems, buildInterviewQuestionSourceItem(source))
	}

	return &types.InterviewQuestionDetailResp{
		Question: buildInterviewQuestionItem(*question),
		Sources:  sourceItems,
		QuestionMeta: types.ReportMeta{
			SchemaVersion: interviewQuestionSchemaVersion,
			Available:     true,
		},
	}, nil
}
