package user

import (
	"context"

	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/api/user/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type InterviewPresetsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewInterviewPresetsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InterviewPresetsLogic {
	return &InterviewPresetsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *InterviewPresetsLogic) InterviewPresets(_ *types.InterviewPresetsReq) (*types.InterviewPresetsResp, error) {
	directions := append([]types.InterviewDirectionPreset(nil), interviewDirections...)
	if l.svcCtx.InterviewQuestionsModel != nil {
		if counts, err := l.svcCtx.InterviewQuestionsModel.DirectionCounts(l.ctx); err == nil && len(counts) > 0 {
			for idx := range directions {
				if count, ok := counts[directions[idx].Key]; ok {
					directions[idx].QuestionCount = count
				}
			}
		}
	}

	return &types.InterviewPresetsResp{
		Directions:        directions,
		Difficulties:      append([]types.InterviewDifficultyPreset(nil), interviewDifficulties...),
		FocusOptions:      append([]types.InterviewFocusOption(nil), interviewFocusOptions...),
		InterviewerStyles: append([]types.InterviewStyleOption(nil), interviewStyles...),
		DefaultConfig:     defaultSessionConfigSnapshot(),
	}, nil
}
