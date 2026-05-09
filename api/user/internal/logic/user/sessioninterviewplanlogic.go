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

	resp := buildInterviewPlanResp(buildSessionConfigSnapshot(*session), req.Limit)
	return &resp, nil
}
