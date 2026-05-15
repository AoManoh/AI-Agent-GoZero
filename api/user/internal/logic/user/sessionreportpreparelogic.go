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

type SessionReportPrepareLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSessionReportPrepareLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SessionReportPrepareLogic {
	return &SessionReportPrepareLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SessionReportPrepareLogic) SessionReportPrepare(req *types.SessionReportPrepareReq) (*types.SessionReportPrepareResp, error) {
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

	record, err := NewSessionEvaluationLogic(l.ctx, l.svcCtx).refreshEvaluationRecord(session, userID, req.Force)
	if err != nil {
		return nil, err
	}

	watermark := evaluationMessageWatermarkRow{
		LastMessageID: record.SourceLastMessageID,
		LastMessageAt: record.SourceLastMessageAt,
	}
	readiness := buildReportReadiness(*session, watermark, record)

	summary, err := NewSessionReportSummaryLogic(l.ctx, l.svcCtx).SessionReportSummary(&types.SessionReportSummaryReq{Id: req.Id})
	if err != nil {
		if _, ok := statuserr.StatusCode(err); ok {
			return nil, err
		}
		return nil, statuserr.ServiceUnavailable("报告准备失败，请稍后重试")
	}

	return &types.SessionReportPrepareResp{
		Readiness:     *readiness,
		ReportSummary: summary,
		PrepareMeta: types.ReportMeta{
			SchemaVersion: "session-report-prepare-v1",
			Available:     readiness.CanReadReport && summary != nil,
		},
	}, nil
}
