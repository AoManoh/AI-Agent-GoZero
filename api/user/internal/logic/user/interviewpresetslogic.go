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
	return &types.InterviewPresetsResp{
		Directions:        append([]types.InterviewDirectionPreset(nil), interviewDirections...),
		Difficulties:      append([]types.InterviewDifficultyPreset(nil), interviewDifficulties...),
		FocusOptions:      append([]types.InterviewFocusOption(nil), interviewFocusOptions...),
		InterviewerStyles: append([]types.InterviewStyleOption(nil), interviewStyles...),
		DefaultConfig:     defaultSessionConfigSnapshot(),
	}, nil
}
