package user

import (
	"context"

	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/api/user/internal/types"

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

	resp := buildInterviewPlanResp(config, req.Limit)
	return &resp, nil
}
