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

type SessionReportReadinessLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSessionReportReadinessLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SessionReportReadinessLogic {
	return &SessionReportReadinessLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SessionReportReadinessLogic) SessionReportReadiness(req *types.SessionReportReadinessReq) (*types.SessionReportReadinessResp, error) {
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

	watermark, err := loadLatestEvaluationMessageWatermarkWithReader(l.ctx, l.svcCtx.DB, session.SessionId, userID)
	if err != nil {
		return nil, statuserr.ServiceUnavailable("报告准备状态暂不可用，请稍后重试")
	}

	record, err := l.svcCtx.SessionEvaluationsModel.FindOneBySessionID(l.ctx, userID, session.SessionId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return buildMissingReportReadiness(*session, watermark), nil
		}
		return nil, statuserr.ServiceUnavailable("报告准备状态暂不可用，请稍后重试")
	}

	return buildReportReadiness(*session, watermark, record), nil
}

func buildMissingReportReadiness(session model.ChatSession, watermark evaluationMessageWatermarkRow) *types.SessionReportReadinessResp {
	resp := baseReportReadinessResp(session, watermark)
	resp.EvaluationStatus = "missing"
	resp.ReportStatus = "missing"
	resp.CanReadReport = false
	resp.NeedsRefresh = true
	resp.Reason = "评估快照不存在，请先刷新评估。"
	resp.NextAction = "refresh_evaluation"
	return resp
}

func buildReportReadiness(session model.ChatSession, watermark evaluationMessageWatermarkRow, record *model.SessionEvaluation) *types.SessionReportReadinessResp {
	resp := baseReportReadinessResp(session, watermark)
	resp.EvaluationStatus = defaultStringValue(record.Status, "draft")
	if !record.GeneratedAt.IsZero() {
		resp.LastRefreshedAt = record.GeneratedAt.Format(timeLayout)
	}

	if shouldRefreshEvaluation(record, watermark) {
		resp.ReportStatus = "stale"
		resp.CanReadReport = false
		resp.NeedsRefresh = true
		resp.Reason = "评估快照已过期，请刷新评估后再查看报告。"
		resp.NextAction = "refresh_evaluation"
		return resp
	}

	resp.ReportStatus = "readable"
	resp.CanReadReport = true
	resp.NeedsRefresh = false
	resp.Reason = "评估快照有效，报告可以直接读取。"
	resp.NextAction = "open_report"
	return resp
}

func baseReportReadinessResp(session model.ChatSession, watermark evaluationMessageWatermarkRow) *types.SessionReportReadinessResp {
	resp := &types.SessionReportReadinessResp{
		Session: buildSessionItem(session),
		ReadinessMeta: types.ReportMeta{
			SchemaVersion: "session-report-readiness-v1",
			Available:     true,
		},
	}
	if watermark.LastMessageAt.Valid {
		resp.LastMessageAt = watermark.LastMessageAt.Time.Format(timeLayout)
	}
	return resp
}
