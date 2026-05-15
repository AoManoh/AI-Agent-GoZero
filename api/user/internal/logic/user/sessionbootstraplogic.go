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

type SessionBootstrapLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSessionBootstrapLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SessionBootstrapLogic {
	return &SessionBootstrapLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SessionBootstrapLogic) SessionBootstrap(req *types.SessionBootstrapReq) (*types.SessionBootstrapResp, error) {
	userID, err := currentUserID(l.ctx)
	if err != nil {
		return nil, err
	}

	session, err := l.svcCtx.ChatSessionsModel.FindOneBySessionID(l.ctx, userID, req.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, statuserr.NotFound("会话不存在或已删除")
		}
		return nil, err
	}

	messages, err := loadSessionMessages(l.ctx, l.svcCtx.DB, session.SessionId, userID, 0)
	if err != nil {
		return nil, statuserr.ServiceUnavailable("会话消息暂不可用，请稍后重试")
	}

	var flowState *types.SessionFlowStateResp
	if l.svcCtx.RedisClient != nil {
		if resp, err := NewSessionFlowStateLogic(l.ctx, l.svcCtx).SessionFlowState(&types.SessionFlowStateReq{Id: req.Id}); err == nil {
			flowState = resp
		}
	}

	var reportSummary *types.SessionReportSummaryResp
	if resp, err := NewSessionReportSummaryLogic(l.ctx, l.svcCtx).SessionReportSummary(&types.SessionReportSummaryReq{Id: req.Id}); err == nil {
		reportSummary = resp
	}

	return &types.SessionBootstrapResp{
		Session:       buildSessionItem(*session),
		Config:        buildSessionConfigSnapshot(*session),
		Messages:      messages,
		FlowState:     flowState,
		ReportSummary: reportSummary,
		BootstrapMeta: types.ReportMeta{
			SchemaVersion: "session-bootstrap-v1",
			Available:     true,
		},
	}, nil
}
