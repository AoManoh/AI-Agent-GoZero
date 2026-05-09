package user

import (
	"context"
	"errors"
	"math"
	"strings"

	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/api/user/internal/types"
	"GoZero-AI/api/user/model"
	"GoZero-AI/internal/statuserr"

	"github.com/zeromicro/go-zero/core/logx"
)

type SessionReportLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSessionReportLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SessionReportLogic {
	return &SessionReportLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SessionReportLogic) SessionReport(req *types.SessionReportReq) (*types.SessionReportResp, error) {
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

	record, err := NewSessionEvaluationLogic(l.ctx, l.svcCtx).resolveEvaluationRecord(session, userID, false)
	if err != nil {
		return nil, err
	}
	evaluationResp, err := buildResponseFromRecord(*session, record)
	if err != nil {
		return nil, statuserr.Internal("报告详情暂不可用，请稍后重试")
	}

	rows, err := loadSessionMessageRows(l.ctx, l.svcCtx.DB, session.SessionId, userID, 0)
	if err != nil {
		return nil, statuserr.ServiceUnavailable("报告消息暂不可用，请稍后重试")
	}

	questionCards := buildSessionReportQuestionCards(rows, evaluationResp)

	return &types.SessionReportResp{
		Session: buildSessionItem(*session),
		Config:  buildSessionConfigSnapshot(*session),
		Overview: types.SessionReportOverview{
			Score:           evaluationResp.OverallScore,
			MaxScore:        100,
			Summary:         evaluationResp.Summary,
			Status:          evaluationResp.Status,
			DurationSeconds: resolveReportDurationSeconds(*session, rows),
			QuestionCount:   int64(len(questionCards)),
			UserTurns:       evaluationResp.UserTurns,
			AssistantTurns:  evaluationResp.AssistantTurns,
		},
		Radar:         buildRadarFromEvaluation(evaluationResp.Dimensions),
		QuestionCards: questionCards,
		Strengths:     evaluationResp.Strengths,
		Risks:         evaluationResp.Risks,
		Suggestions:   evaluationResp.Suggestions,
		ReportMeta: types.ReportMeta{
			SchemaVersion: "session-report-v1",
			Available:     evaluationResp.Status == "ready" || len(questionCards) > 0,
		},
	}, nil
}

func buildRadarFromEvaluation(dimensions []types.EvaluationDimension) []types.AbilityRadarPoint {
	if len(dimensions) == 0 {
		return defaultAbilityRadar()
	}
	points := make([]types.AbilityRadarPoint, 0, len(dimensions))
	for _, dimension := range dimensions {
		score := int64(0)
		if dimension.MaxScore > 0 {
			score = int64(math.Round(float64(dimension.Score) / float64(dimension.MaxScore) * 100))
		}
		points = append(points, types.AbilityRadarPoint{
			Key:      dimension.Key,
			Label:    dimension.Label,
			Score:    score,
			MaxScore: 100,
		})
	}
	return points
}

func buildSessionReportQuestionCards(rows []sessionDataMessageRow, evaluationResp *types.SessionEvaluationResp) []types.SessionReportQuestionCard {
	cards := make([]types.SessionReportQuestionCard, 0)
	for idx, row := range rows {
		if row.Role != "user" {
			continue
		}

		question := nearestAssistantBefore(rows, idx)
		if question == "" {
			question = "请展开说明你的技术方案与取舍。"
		}
		comment := nearestAssistantAfter(rows, idx)
		if comment == "" {
			comment = buildFallbackQuestionComment(row.Content, evaluationResp)
		}

		score := estimateQuestionScore(row.Content, evaluationResp)
		cards = append(cards, types.SessionReportQuestionCard{
			TurnIndex: int64(len(cards) + 1),
			Depth:     buildQuestionDepthLabel(len(cards)+1, evaluationResp.Session.MessageCount),
			Question:  truncateEvaluationContent(question, 220),
			Answer:    truncateEvaluationContent(row.Content, 360),
			AiComment: truncateEvaluationContent(comment, 220),
			Score:     score,
			MaxScore:  5,
			Tags:      buildQuestionTags(score, row.Content),
		})
	}
	return cards
}

func nearestAssistantBefore(rows []sessionDataMessageRow, index int) string {
	for i := index - 1; i >= 0; i-- {
		if rows[i].Role == "assistant" {
			return rows[i].Content
		}
	}
	return ""
}

func nearestAssistantAfter(rows []sessionDataMessageRow, index int) string {
	for i := index + 1; i < len(rows); i++ {
		if rows[i].Role == "assistant" {
			return rows[i].Content
		}
		if rows[i].Role == "user" {
			return ""
		}
	}
	return ""
}

func estimateQuestionScore(answer string, evaluationResp *types.SessionEvaluationResp) int64 {
	if evaluationResp.Status != "ready" {
		return 0
	}
	base := int64(math.Round(evaluationResp.OverallScore / 20))
	if base < 1 {
		base = 1
	}
	if base > 5 {
		base = 5
	}
	answerLen := len([]rune(strings.TrimSpace(answer)))
	if answerLen >= 160 && base < 5 {
		base++
	}
	if answerLen < 40 && base > 1 {
		base--
	}
	return base
}

func buildQuestionTags(score int64, answer string) []types.SessionReportTag {
	tags := make([]types.SessionReportTag, 0, 3)
	switch {
	case score >= 4:
		tags = append(tags, types.SessionReportTag{Key: "concept_clear", Label: "概念清晰", Level: "positive"})
	case score >= 2:
		tags = append(tags, types.SessionReportTag{Key: "basic_complete", Label: "基本完整", Level: "neutral"})
	default:
		tags = append(tags, types.SessionReportTag{Key: "needs_detail", Label: "缺关键细节", Level: "risk"})
	}
	if len([]rune(strings.TrimSpace(answer))) >= 120 {
		tags = append(tags, types.SessionReportTag{Key: "expression_full", Label: "表达充分", Level: "positive"})
	} else {
		tags = append(tags, types.SessionReportTag{Key: "expression_short", Label: "展开不足", Level: "risk"})
	}
	return tags
}

func buildFallbackQuestionComment(answer string, evaluationResp *types.SessionEvaluationResp) string {
	if evaluationResp.Status != "ready" {
		return "当前回答样本仍偏少，建议继续补充更多技术细节后刷新报告。"
	}
	if len([]rune(strings.TrimSpace(answer))) >= 120 {
		return "回答有一定展开度，可继续补充边界条件、权衡和故障处理细节。"
	}
	return "回答较短，建议按照背景、方案、取舍、结果的结构补足信息。"
}

func buildQuestionDepthLabel(index int, messageCount int64) string {
	if index <= 1 {
		return "N+1"
	}
	if messageCount >= 8 {
		return "N+3"
	}
	return "N+2"
}

func resolveReportDurationSeconds(session model.ChatSession, rows []sessionDataMessageRow) int64 {
	if session.DurationSeconds > 0 {
		return session.DurationSeconds
	}
	if session.StartedAt.Valid && session.CompletedAt.Valid {
		return int64(session.CompletedAt.Time.Sub(session.StartedAt.Time).Seconds())
	}
	if len(rows) >= 2 {
		duration := rows[len(rows)-1].CreatedAt.Sub(rows[0].CreatedAt)
		if duration > 0 {
			return int64(duration.Seconds())
		}
	}
	return 0
}
