package user

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/api/user/internal/types"
	"GoZero-AI/api/user/model"
	"GoZero-AI/internal/statuserr"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type SessionReportSummaryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

type reportMessageAggregateRow struct {
	FirstMessageAt sql.NullTime `db:"first_message_at"`
	LastMessageAt  sql.NullTime `db:"last_message_at"`
}

type reportLatestMessageRow struct {
	Role      string    `db:"role"`
	Content   string    `db:"content"`
	CreatedAt time.Time `db:"created_at"`
}

type reportResumeAggregateRow struct {
	ResumeChunks    int64        `db:"resume_chunks"`
	ResumeUpdatedAt sql.NullTime `db:"resume_updated_at"`
}

func NewSessionReportSummaryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SessionReportSummaryLogic {
	return &SessionReportSummaryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SessionReportSummaryLogic) SessionReportSummary(req *types.SessionReportSummaryReq) (*types.SessionReportSummaryResp, error) {
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

	evaluationRecord, err := NewSessionEvaluationLogic(l.ctx, l.svcCtx).resolveEvaluationRecord(session, userID, false)
	if err != nil {
		return nil, err
	}
	evaluationResp, err := buildResponseFromRecord(*session, evaluationRecord)
	if err != nil {
		return nil, statuserr.Internal("报告摘要暂不可用，请稍后重试")
	}

	conversationSummary, err := l.buildConversationSummary(session.SessionId, userID, evaluationRecord)
	if err != nil {
		return nil, statuserr.ServiceUnavailable("报告摘要暂不可用，请稍后重试")
	}

	assetSummary, err := l.buildAssetSummary(session.SessionId, userID, evaluationRecord)
	if err != nil {
		return nil, statuserr.ServiceUnavailable("报告摘要暂不可用，请稍后重试")
	}

	return &types.SessionReportSummaryResp{
		Session: buildSessionItem(*session),
		Snapshot: types.ReportSnapshot{
			Title:      buildReportSnapshotTitle(evaluationResp.Status),
			Summary:    evaluationResp.Summary,
			Status:     evaluationResp.Status,
			NextAction: firstSuggestion(evaluationResp.Suggestions),
		},
		Conversation: conversationSummary,
		Assets:       assetSummary,
		Evaluation:   buildReportEvaluationSummary(evaluationResp),
		ReportMeta: types.ReportMeta{
			SchemaVersion: "report-summary-v1",
			Available:     true,
		},
	}, nil
}

func (l *SessionReportSummaryLogic) buildConversationSummary(sessionID string, userID int64, evaluationRecord *model.SessionEvaluation) (types.ReportConversationSummary, error) {
	var aggregate reportMessageAggregateRow
	if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &aggregate, `select
	min(created_at) as first_message_at,
	max(created_at) as last_message_at
from "public"."vector_store"
where chat_id = $1 and user_id = $2 and doc_type = 'message'
  and ($3::bigint is null or id <= $3)
  and ($4::timestamptz is null or created_at <= $4)`, sessionID, userID, nullableMessageWatermarkID(evaluationRecord), nullableEvaluationGeneratedAt(evaluationRecord)); err != nil && err != sql.ErrNoRows {
		return types.ReportConversationSummary{}, err
	}

	var latestRows []reportLatestMessageRow
	if err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &latestRows, `select role, content, created_at
from "public"."vector_store"
where chat_id = $1 and user_id = $2 and doc_type = 'message'
  and ($3::bigint is null or id <= $3)
  and ($4::timestamptz is null or created_at <= $4)
order by created_at desc
limit 8`, sessionID, userID, nullableMessageWatermarkID(evaluationRecord), nullableEvaluationGeneratedAt(evaluationRecord)); err != nil && err != sqlx.ErrNotFound && err != sql.ErrNoRows {
		return types.ReportConversationSummary{}, err
	}

	summary := types.ReportConversationSummary{
		MessageCount:   evaluationRecord.UserTurns + evaluationRecord.AssistantTurns,
		UserTurns:      evaluationRecord.UserTurns,
		AssistantTurns: evaluationRecord.AssistantTurns,
	}
	if aggregate.FirstMessageAt.Valid {
		summary.FirstMessageAt = aggregate.FirstMessageAt.Time.Format(timeLayout)
	}
	if aggregate.LastMessageAt.Valid {
		summary.LastMessageAt = aggregate.LastMessageAt.Time.Format(timeLayout)
	}

	for _, row := range latestRows {
		switch row.Role {
		case "user":
			if summary.LatestUserMessage == "" {
				summary.LatestUserMessage = truncateEvaluationContent(row.Content, 160)
			}
		case "assistant":
			if summary.LatestAssistantMessage == "" {
				summary.LatestAssistantMessage = truncateEvaluationContent(row.Content, 160)
			}
		}
		if summary.LatestUserMessage != "" && summary.LatestAssistantMessage != "" {
			break
		}
	}

	return summary, nil
}

func (l *SessionReportSummaryLogic) buildAssetSummary(sessionID string, userID int64, evaluationRecord *model.SessionEvaluation) (types.ReportAssetSummary, error) {
	var aggregate reportResumeAggregateRow
	if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &aggregate, `select
	count(*) as resume_chunks,
	max(created_at) as resume_updated_at
from "public"."vector_store"
where chat_id = $1 and user_id = $2 and doc_type = 'resume'
  and ($3::timestamptz is null or created_at <= $3)`, sessionID, userID, nullableEvaluationGeneratedAt(evaluationRecord)); err != nil && err != sql.ErrNoRows {
		return types.ReportAssetSummary{}, err
	}

	summary := types.ReportAssetSummary{
		HasResume:    aggregate.ResumeChunks > 0,
		ResumeChunks: aggregate.ResumeChunks,
	}
	if aggregate.ResumeUpdatedAt.Valid {
		summary.ResumeUpdatedAt = aggregate.ResumeUpdatedAt.Time.Format(timeLayout)
	}

	return summary, nil
}

func buildReportSnapshotTitle(status string) string {
	switch status {
	case "ready":
		return "Current Evaluation Snapshot"
	case "draft":
		return "Draft Evaluation Snapshot"
	default:
		return "Session Snapshot"
	}
}

func firstSuggestion(suggestions []string) string {
	if len(suggestions) == 0 {
		return ""
	}
	return suggestions[0]
}

func buildReportEvaluationSummary(resp *types.SessionEvaluationResp) types.ReportEvaluationSummary {
	return types.ReportEvaluationSummary{
		Status:           resp.Status,
		Summary:          resp.Summary,
		OverallScore:     resp.OverallScore,
		RubricVersion:    resp.RubricVersion,
		ScoreSource:      resp.ScoreSource,
		Dimensions:       resp.Dimensions,
		Strengths:        resp.Strengths,
		Risks:            resp.Risks,
		Suggestions:      resp.Suggestions,
		FirstGeneratedAt: resp.FirstGeneratedAt,
		LastRefreshedAt:  resp.LastRefreshedAt,
		GeneratedAt:      resp.GeneratedAt,
	}
}

func nullableMessageWatermarkID(record *model.SessionEvaluation) any {
	if record == nil || !record.SourceLastMessageID.Valid {
		return nil
	}
	return record.SourceLastMessageID.Int64
}

func nullableEvaluationGeneratedAt(record *model.SessionEvaluation) any {
	if record == nil || record.GeneratedAt.IsZero() {
		return nil
	}
	return record.GeneratedAt
}
