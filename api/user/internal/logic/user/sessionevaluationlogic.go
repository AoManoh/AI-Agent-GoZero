package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"net/http"
	"strings"
	"time"

	"GoZero-AI/api/user/internal/evaluation"
	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/api/user/internal/types"
	"GoZero-AI/api/user/model"
	"GoZero-AI/internal/statuserr"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type SessionEvaluationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

type evaluationMessageRow struct {
	Role      string    `db:"role"`
	Content   string    `db:"content"`
	CreatedAt time.Time `db:"created_at"`
}

type heuristicEvaluation struct {
	Status        string
	Summary       string
	Dimensions    []types.EvaluationDimension
	OverallScore  float64
	RubricVersion string
	ScoreSource   string
	Strengths     []string
	Risks         []string
	Suggestions   []string
}

type evaluationMessageWatermarkRow struct {
	LastMessageID sql.NullInt64 `db:"last_message_id"`
	LastMessageAt sql.NullTime  `db:"last_message_at"`
}

func NewSessionEvaluationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SessionEvaluationLogic {
	return &SessionEvaluationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SessionEvaluationLogic) SessionEvaluation(req *types.SessionEvaluationReq) (*types.SessionEvaluationResp, error) {
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

	record, err := l.resolveEvaluationRecord(session, userID, false)
	if err != nil {
		return nil, err
	}

	resp, buildErr := buildResponseFromRecord(*session, record)
	if buildErr != nil {
		return nil, statuserr.Internal("评估结果暂不可用，请稍后重试")
	}
	return resp, nil
}

func (l *SessionEvaluationLogic) SessionEvaluationRefresh(req *types.SessionEvaluationRefreshReq) (*types.SessionEvaluationResp, error) {
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

	record, err := l.refreshEvaluationRecord(session, userID, req.Force)
	if err != nil {
		return nil, err
	}

	resp, buildErr := buildResponseFromRecord(*session, record)
	if buildErr != nil {
		return nil, statuserr.Internal("评估结果暂不可用，请稍后重试")
	}
	return resp, nil
}

func (l *SessionEvaluationLogic) getOrRefreshEvaluationRecord(session *model.ChatSession, userID int64) (*model.SessionEvaluation, error) {
	return l.refreshEvaluationRecord(session, userID, false)
}

func (l *SessionEvaluationLogic) resolveEvaluationRecord(session *model.ChatSession, userID int64, allowRefresh bool) (*model.SessionEvaluation, error) {
	if session == nil {
		return nil, statuserr.NotFound("会话不存在或已删除")
	}

	latestWatermark, err := l.loadLatestEvaluationMessageWatermark(session.SessionId, userID)
	if err != nil {
		return nil, statuserr.ServiceUnavailable("会话评估暂不可用，请稍后重试")
	}

	existing, err := l.svcCtx.SessionEvaluationsModel.FindOneBySessionID(l.ctx, userID, session.SessionId)
	if err == nil && !shouldRefreshEvaluation(existing, latestWatermark) {
		return existing, nil
	}
	if !allowRefresh {
		if err != nil {
			if errors.Is(err, model.ErrNotFound) {
				return nil, statuserr.ServiceUnavailable("评估快照暂不可用，请先刷新评估")
			}
			return nil, statuserr.ServiceUnavailable("评估快照暂不可用，请稍后重试")
		}
		return nil, statuserr.ServiceUnavailable("评估快照已过期，请刷新评估")
	}
	if err != nil && !errors.Is(err, model.ErrNotFound) {
		return nil, statuserr.ServiceUnavailable("会话评估暂不可用，请稍后重试")
	}

	if existing != nil || errors.Is(err, model.ErrNotFound) {
		return l.refreshEvaluationIfNeeded(session, userID, existing, latestWatermark, false)
	}

	return nil, statuserr.ServiceUnavailable("会话评估暂不可用，请稍后重试")
}

func (l *SessionEvaluationLogic) refreshEvaluationRecord(session *model.ChatSession, userID int64, force bool) (*model.SessionEvaluation, error) {
	if session == nil {
		return nil, statuserr.NotFound("会话不存在或已删除")
	}

	latestWatermark, err := l.loadLatestEvaluationMessageWatermark(session.SessionId, userID)
	if err != nil {
		return nil, statuserr.ServiceUnavailable("会话评估暂不可用，请稍后重试")
	}

	existing, err := l.svcCtx.SessionEvaluationsModel.FindOneBySessionID(l.ctx, userID, session.SessionId)
	if err == nil && !force && !shouldRefreshEvaluation(existing, latestWatermark) {
		return existing, nil
	}
	if err != nil && !errors.Is(err, model.ErrNotFound) {
		return nil, statuserr.ServiceUnavailable("会话评估暂不可用，请稍后重试")
	}
	if errors.Is(err, model.ErrNotFound) {
		existing = nil
	}

	return l.refreshEvaluationIfNeeded(session, userID, existing, latestWatermark, force)
}

func (l *SessionEvaluationLogic) refreshEvaluationIfNeeded(session *model.ChatSession, userID int64, existing *model.SessionEvaluation, latestWatermark evaluationMessageWatermarkRow, force bool) (*model.SessionEvaluation, error) {
	if existing != nil && !force && !shouldRefreshEvaluation(existing, latestWatermark) {
		return existing, nil
	}

	var refreshed *model.SessionEvaluation
	err := withEvaluationRefreshLock(l.ctx, l.svcCtx.DB, userID, session.SessionId, func() error {
		currentLatestWatermark, err := l.loadLatestEvaluationMessageWatermark(session.SessionId, userID)
		if err != nil {
			return err
		}

		currentExisting, err := l.svcCtx.SessionEvaluationsModel.FindOneBySessionID(l.ctx, userID, session.SessionId)
		switch {
		case err == nil && !force && !shouldRefreshEvaluation(currentExisting, currentLatestWatermark):
			refreshed = currentExisting
			return nil
		case err != nil && !errors.Is(err, model.ErrNotFound):
			return err
		}

		rows, err := loadEvaluationRowsWithReader(l.ctx, l.svcCtx.DB, session.SessionId, userID, currentLatestWatermark)
		if err != nil {
			return err
		}

		record, err := l.buildEvaluationRecordForMessages(currentExisting, session.SessionId, userID, currentLatestWatermark, rows)
		if err != nil {
			return err
		}
		if err := l.svcCtx.SessionEvaluationsModel.Upsert(l.ctx, record); err != nil {
			return err
		}

		refreshed = record
		return nil
	})
	if err != nil {
		if _, ok := statuserr.StatusCode(err); ok {
			return nil, err
		}
		return nil, statuserr.ServiceUnavailable("会话评估暂不可用，请稍后重试")
	}
	return refreshed, nil
}

func (l *SessionEvaluationLogic) buildEvaluationRecordForMessages(existing *model.SessionEvaluation, sessionID string, userID int64, latestWatermark evaluationMessageWatermarkRow, rows []evaluationMessageRow) (*model.SessionEvaluation, error) {
	userTurns, assistantTurns, totalUserChars, evidence := summarizeEvaluationRows(rows)
	heuristic := buildHeuristicEvaluation(rows, userTurns, assistantTurns, totalUserChars)
	status := heuristic.Status
	summary := heuristic.Summary
	dimensions := heuristic.Dimensions
	overallScore := heuristic.OverallScore
	rubricVersion := heuristic.RubricVersion
	scoreSource := heuristic.ScoreSource
	strengths := heuristic.Strengths
	risks := heuristic.Risks
	suggestions := heuristic.Suggestions

	if generated, err := l.generateEvaluation(rows); err == nil {
		status = mergeEvaluationStatus(status, generated)
		summary = coalesceString(generated.Summary, summary)
		dimensions = evaluation.NormalizeDimensions(generated.Dimensions, dimensions)
		rubricVersion = evaluation.RubricVersion
		scoreSource = resolveGeneratedScoreSource(generated)
		strengths = mergeStringSlice(generated.Strengths, strengths)
		risks = mergeStringSlice(generated.Risks, risks)
		suggestions = mergeStringSlice(generated.Suggestions, suggestions)
	} else {
		scoreSource = "heuristic_fallback"
		risks = append([]string{"评估模型暂不可用，当前结果已降级为启发式评估。"}, risks...)
		suggestions = append([]string{"建议稍后重新刷新评估，以获得模型增强结果。"}, suggestions...)
		l.Logger.Errorf("evaluation generator fallback to heuristic: %v", err)
	}
	dimensions, overallScore = normalizeEvaluationScoring(status, dimensions)

	record, err := buildEvaluationRecord(existing, sessionID, userID, latestWatermark, status, summary, userTurns, assistantTurns, overallScore, rubricVersion, scoreSource, dimensions, strengths, risks, suggestions, evidence)
	if err != nil {
		return nil, statuserr.New(http.StatusInternalServerError, "评估结果序列化失败，请稍后重试")
	}
	return record, nil
}

func summarizeEvaluationRows(rows []evaluationMessageRow) (int64, int64, int, []types.EvaluationEvidence) {
	var userTurns int64
	var assistantTurns int64
	totalUserChars := 0
	evidence := make([]types.EvaluationEvidence, 0, min(len(rows), 3))

	for _, row := range rows {
		switch row.Role {
		case "user":
			userTurns++
			totalUserChars += len([]rune(strings.TrimSpace(row.Content)))
		case "assistant":
			assistantTurns++
		}

		if len(evidence) < 3 {
			evidence = append(evidence, types.EvaluationEvidence{
				Role:      row.Role,
				Content:   truncateEvaluationContent(row.Content, 120),
				CreatedAt: row.CreatedAt.Format(timeLayout),
			})
		}
	}

	return userTurns, assistantTurns, totalUserChars, evidence
}

func buildEvaluationStatus(userTurns, assistantTurns int64) (string, string) {
	switch {
	case userTurns == 0:
		return "insufficient_data", "当前会话还没有形成有效的用户回答，暂时无法输出结构化评估。"
	case assistantTurns == 0:
		return "draft", "当前会话只有用户侧输入，缺少系统追问和反馈，暂时只能给出草稿级评估。"
	default:
		return "ready", "当前会话已经形成基础问答链路，可返回第一版结构化评估结果。"
	}
}

func buildHeuristicEvaluation(rows []evaluationMessageRow, userTurns, assistantTurns int64, totalUserChars int) heuristicEvaluation {
	status, summary := buildEvaluationStatus(userTurns, assistantTurns)
	dimensions := evaluation.NormalizeDimensions(buildEvaluationDimensions(rows, userTurns, assistantTurns, totalUserChars), nil)
	dimensions, overallScore := normalizeEvaluationScoring(status, dimensions)
	strengths, risks := buildEvaluationHighlights(userTurns, assistantTurns, totalUserChars)
	return heuristicEvaluation{
		Status:        status,
		Summary:       summary,
		Dimensions:    dimensions,
		OverallScore:  overallScore,
		RubricVersion: evaluation.RubricVersion,
		ScoreSource:   "heuristic",
		Strengths:     strengths,
		Risks:         risks,
		Suggestions:   buildEvaluationSuggestions(risks),
	}
}

func buildEvaluationDimensions(rows []evaluationMessageRow, userTurns, assistantTurns int64, totalUserChars int) []types.EvaluationDimension {
	avgUserChars := 0
	if userTurns > 0 {
		avgUserChars = totalUserChars / int(userTurns)
	}

	transcript := strings.ToLower(joinEvaluationTranscript(rows))
	technicalScore := clampScore(keywordScore(transcript, []string{"gmp", "gc", "csp", "goroutine", "channel", "mutex"}, 2) + avgUserChars/70)
	practiceScore := clampScore(keywordScore(transcript, []string{"gozero", "goframe", "mysql", "redis", "docker", "k8s", "索引", "事务"}, 2) + int(userTurns))
	architectureScore := clampScore(keywordScore(transcript, []string{"微服务", "高并发", "强一致性", "分布式", "服务治理", "etcd", "pgvector", "rag"}, 2) + int(min64(userTurns, assistantTurns)))
	communicationScore := clampScore(avgUserChars/50 + int(userTurns))

	return []types.EvaluationDimension{
		{
			Key:      "technical_depth",
			Label:    "技术深度",
			Score:    int64(technicalScore),
			MaxScore: 5,
			Summary:  buildTechnicalSummary(transcript, avgUserChars),
		},
		{
			Key:      "engineering_practice",
			Label:    "工程实践",
			Score:    int64(practiceScore),
			MaxScore: 5,
			Summary:  buildPracticeSummary(transcript, userTurns),
		},
		{
			Key:      "architecture_sense",
			Label:    "架构意识",
			Score:    int64(architectureScore),
			MaxScore: 5,
			Summary:  buildArchitectureSummary(transcript, userTurns, assistantTurns),
		},
		{
			Key:      "communication",
			Label:    "表达与沟通",
			Score:    int64(communicationScore),
			MaxScore: 5,
			Summary:  buildCommunicationSummary(avgUserChars, userTurns, assistantTurns),
		},
	}
}

func buildEvaluationHighlights(userTurns, assistantTurns int64, totalUserChars int) ([]string, []string) {
	avgUserChars := 0
	if userTurns > 0 {
		avgUserChars = totalUserChars / int(userTurns)
	}

	strengths := make([]string, 0, 2)
	risks := make([]string, 0, 2)

	if userTurns >= 2 {
		strengths = append(strengths, "候选人在当前会话中已经进行多轮回应，具备基础连续作答能力。")
	}
	if avgUserChars >= 80 {
		strengths = append(strengths, "用户侧回答长度较充分，说明表达意愿和信息展开度较好。")
	}
	if len(strengths) == 0 {
		strengths = append(strengths, "当前会话已具备最小可评估样本，可继续积累更多问答提高评估稳定性。")
	}

	if userTurns < 2 {
		risks = append(risks, "当前用户回合数偏少，评估结果更适合作为草稿而不是最终结论。")
	}
	if assistantTurns == 0 {
		risks = append(risks, "当前缺少助手侧追问或反馈，无法体现更完整的面试过程结构。")
	}
	if avgUserChars < 50 && userTurns > 0 {
		risks = append(risks, "用户单次回答偏短，可能导致信息覆盖不足。")
	}
	if len(risks) == 0 {
		risks = append(risks, "当前评估仍基于消息统计与启发式规则，后续应替换为 rubric 驱动的正式评估。")
	}

	return strengths, risks
}

func buildEvaluationSuggestions(risks []string) []string {
	suggestions := make([]string, 0, len(risks))

	for _, risk := range risks {
		switch {
		case strings.Contains(risk, "回合数偏少"):
			suggestions = append(suggestions, "建议继续完成更多轮面试问答，再刷新下一版结构化评估。")
		case strings.Contains(risk, "缺少助手侧追问"):
			suggestions = append(suggestions, "建议增加 1 到 2 个技术追问，形成更完整的评估样本。")
		case strings.Contains(risk, "回答偏短"):
			suggestions = append(suggestions, "建议后续回答补足背景、方案、权衡与结果，提升信息覆盖度。")
		default:
			suggestions = append(suggestions, "建议继续补充更多问答内容，以提升评估可信度。")
		}
	}

	if len(suggestions) == 0 {
		suggestions = append(suggestions, "当前会话已具备基础评估输入，可继续扩展更多维度。")
	}

	return suggestions
}

func (l *SessionEvaluationLogic) generateEvaluation(rows []evaluationMessageRow) (*svc.GeneratedEvaluation, error) {
	if l.svcCtx.EvaluationGenerator == nil {
		return nil, errors.New("evaluation generator unavailable")
	}
	messages := make([]svc.EvaluationMessage, 0, len(rows))
	for _, row := range rows {
		content := strings.TrimSpace(row.Content)
		if content == "" {
			continue
		}
		messages = append(messages, svc.EvaluationMessage{
			Role:    row.Role,
			Content: truncateEvaluationContent(content, 400),
		})
	}
	if len(messages) == 0 {
		return nil, errors.New("empty evaluation transcript")
	}
	return l.svcCtx.EvaluationGenerator.Generate(l.ctx, messages)
}

func mergeEvaluationStatus(fallback string, generated *svc.GeneratedEvaluation) string {
	if generated == nil {
		return fallback
	}
	if len(generated.Dimensions) > 0 || generated.Summary != "" {
		if fallback == "insufficient_data" {
			return "draft"
		}
		return "ready"
	}
	return fallback
}

func coalesceString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}

func mergeStringSlice(primary, fallback []string) []string {
	if len(primary) == 0 {
		return fallback
	}
	return primary
}

func mergeDimensions(primary, fallback []types.EvaluationDimension) []types.EvaluationDimension {
	if len(primary) == 0 {
		return fallback
	}
	return primary
}

func resolveGeneratedScoreSource(generated *svc.GeneratedEvaluation) string {
	if generated == nil {
		return "heuristic"
	}
	if strings.TrimSpace(generated.Summary) != "" &&
		evaluation.HasCompleteDimensions(generated.Dimensions) &&
		len(generated.Strengths) > 0 &&
		len(generated.Risks) > 0 &&
		len(generated.Suggestions) > 0 {
		return "llm"
	}
	return "mixed"
}

func normalizeEvaluationScoring(status string, dimensions []types.EvaluationDimension) ([]types.EvaluationDimension, float64) {
	if status == "ready" {
		return dimensions, evaluation.ComputeOverallScore(dimensions)
	}

	normalized := make([]types.EvaluationDimension, 0, len(dimensions))
	for _, dimension := range dimensions {
		dimension.Score = 0
		normalized = append(normalized, dimension)
	}

	return normalized, 0
}

func buildTechnicalSummary(transcript string, avgUserChars int) string {
	switch {
	case containsAnyLower(transcript, []string{"gmp", "gc", "csp", "goroutine"}):
		return "当前会话已出现 Go 底层机制或并发模型相关信号，可支撑对技术深度的正向判断。"
	case avgUserChars >= 80:
		return "回答有一定展开度，但尚未出现足够多的底层原理细节。"
	default:
		return "当前会话中的技术原理信号偏少，技术深度判断仍较保守。"
	}
}

func buildPracticeSummary(transcript string, userTurns int64) string {
	if containsAnyLower(transcript, []string{"gozero", "goframe", "mysql", "redis", "事务", "索引"}) {
		return "当前会话已出现框架、存储或性能优化相关经验，具备工程实践判断基础。"
	}
	if userTurns > 0 {
		return "当前会话已有最小样本，但工程实践细节仍偏少。"
	}
	return "当前会话尚未出现足够的工程实践信号。"
}

func buildArchitectureSummary(transcript string, userTurns, assistantTurns int64) string {
	if containsAnyLower(transcript, []string{"微服务", "高并发", "强一致性", "分布式", "服务治理"}) {
		return "当前会话已出现系统边界或复杂业务场景信号，可进行基础架构意识判断。"
	}
	if userTurns >= 1 && assistantTurns >= 1 {
		return "当前会话形成了基础问答闭环，但架构层面的信息仍有限。"
	}
	return "当前会话样本不足，架构意识判断偏保守。"
}

func buildCommunicationSummary(avgUserChars int, userTurns, assistantTurns int64) string {
	switch {
	case avgUserChars >= 120:
		return "用户回答展开较充分，表达与沟通维度表现较好。"
	case avgUserChars >= 60:
		return "用户回答长度中等，已能支撑基础沟通判断。"
	case userTurns >= 1 && assistantTurns >= 1:
		return "当前会话已形成最小互动闭环，但表达仍偏短。"
	default:
		return "当前暂无足够样本评估表达与沟通。"
	}
}

func truncateEvaluationContent(content string, maxLen int) string {
	runes := []rune(strings.TrimSpace(content))
	if len(runes) <= maxLen {
		return string(runes)
	}
	return string(runes[:maxLen]) + "..."
}

func clampScore(score int) int {
	if score < 1 {
		return 1
	}
	if score > 5 {
		return 5
	}
	return score
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func min64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func shouldRefreshEvaluation(existing *model.SessionEvaluation, latestWatermark evaluationMessageWatermarkRow) bool {
	if existing == nil {
		return true
	}
	if evaluationRecordIncomplete(existing) {
		return true
	}
	if latestWatermark.LastMessageID.Valid {
		if !existing.SourceLastMessageID.Valid {
			return true
		}
		return existing.SourceLastMessageID.Int64 < latestWatermark.LastMessageID.Int64
	}
	if latestWatermark.LastMessageAt.Valid {
		if !existing.SourceLastMessageAt.Valid {
			return true
		}
		return existing.SourceLastMessageAt.Time.Before(latestWatermark.LastMessageAt.Time)
	}
	return false
}

func evaluationRecordIncomplete(existing *model.SessionEvaluation) bool {
	if existing.RubricVersion != evaluation.RubricVersion || existing.ScoreSource == "" {
		return true
	}
	if existing.FirstGeneratedAt.IsZero() || existing.GeneratedAt.IsZero() || existing.UpdatedAt.IsZero() {
		return true
	}
	if existing.Status == "ready" && existing.OverallScore <= 0 {
		return true
	}
	if existing.Status != "ready" && existing.OverallScore != 0 {
		return true
	}

	var dimensions []types.EvaluationDimension
	if err := json.Unmarshal(existing.Dimensions, &dimensions); err != nil || !evaluation.HasCompleteDimensions(dimensions) {
		return true
	}

	var suggestions []string
	if err := json.Unmarshal(existing.Suggestions, &suggestions); err != nil || len(suggestions) == 0 {
		return true
	}

	var strengths []string
	if err := json.Unmarshal(existing.Strengths, &strengths); err != nil {
		return true
	}

	var risks []string
	if err := json.Unmarshal(existing.Risks, &risks); err != nil {
		return true
	}

	var evidence []types.EvaluationEvidence
	if err := json.Unmarshal(existing.Evidence, &evidence); err != nil {
		return true
	}

	return false
}

func (l *SessionEvaluationLogic) loadLatestEvaluationMessageWatermark(sessionID string, userID int64) (evaluationMessageWatermarkRow, error) {
	return loadLatestEvaluationMessageWatermarkWithReader(l.ctx, l.svcCtx.DB, sessionID, userID)
}

func chooseLatestMessageTime(primary, secondary sql.NullTime) sql.NullTime {
	switch {
	case primary.Valid && secondary.Valid:
		if secondary.Time.After(primary.Time) {
			return secondary
		}
		return primary
	case secondary.Valid:
		return secondary
	default:
		return primary
	}
}

func buildEvaluationRecord(existing *model.SessionEvaluation, sessionID string, userID int64, sourceWatermark evaluationMessageWatermarkRow, status, summary string, userTurns, assistantTurns int64,
	overallScore float64, rubricVersion, scoreSource string,
	dimensions []types.EvaluationDimension, strengths, risks, suggestions []string, evidence []types.EvaluationEvidence) (*model.SessionEvaluation, error) {
	dimensionsJSON, err := json.Marshal(dimensions)
	if err != nil {
		return nil, err
	}
	strengthsJSON, err := json.Marshal(strengths)
	if err != nil {
		return nil, err
	}
	risksJSON, err := json.Marshal(risks)
	if err != nil {
		return nil, err
	}
	suggestionsJSON, err := json.Marshal(suggestions)
	if err != nil {
		return nil, err
	}
	evidenceJSON, err := json.Marshal(evidence)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	firstGeneratedAt := now
	if existing != nil {
		switch {
		case !existing.FirstGeneratedAt.IsZero():
			firstGeneratedAt = existing.FirstGeneratedAt
		case !existing.GeneratedAt.IsZero():
			firstGeneratedAt = existing.GeneratedAt
		case !existing.UpdatedAt.IsZero():
			firstGeneratedAt = existing.UpdatedAt
		}
	}

	return &model.SessionEvaluation{
		SessionId:           sessionID,
		UserId:              userID,
		Status:              status,
		Summary:             summary,
		UserTurns:           userTurns,
		AssistantTurns:      assistantTurns,
		OverallScore:        overallScore,
		RubricVersion:       rubricVersion,
		ScoreSource:         scoreSource,
		Dimensions:          dimensionsJSON,
		Strengths:           strengthsJSON,
		Risks:               risksJSON,
		Suggestions:         suggestionsJSON,
		Evidence:            evidenceJSON,
		SourceLastMessageID: sourceWatermark.LastMessageID,
		SourceLastMessageAt: sourceWatermark.LastMessageAt,
		FirstGeneratedAt:    firstGeneratedAt,
		GeneratedAt:         now,
		UpdatedAt:           now,
	}, nil
}

type queryRowReader interface {
	QueryRowCtx(ctx context.Context, v any, query string, args ...any) error
}

type queryRowsReader interface {
	QueryRowsCtx(ctx context.Context, v any, query string, args ...any) error
}

func loadLatestEvaluationMessageWatermarkWithReader(ctx context.Context, reader queryRowReader, sessionID string, userID int64) (evaluationMessageWatermarkRow, error) {
	var aggregate evaluationMessageWatermarkRow
	err := reader.QueryRowCtx(ctx, &aggregate, `select id as last_message_id, created_at as last_message_at
from "public"."vector_store"
where chat_id = $1 and user_id = $2 and doc_type = 'message'
order by id desc
limit 1`, sessionID, userID)
	if err != nil && err != sql.ErrNoRows && err != sqlx.ErrNotFound {
		return evaluationMessageWatermarkRow{}, err
	}
	if err == sql.ErrNoRows || err == sqlx.ErrNotFound {
		return evaluationMessageWatermarkRow{}, nil
	}
	return aggregate, nil
}

func loadEvaluationRowsWithReader(ctx context.Context, reader queryRowsReader, sessionID string, userID int64, watermark evaluationMessageWatermarkRow) ([]evaluationMessageRow, error) {
	var rows []evaluationMessageRow
	query := `select role, content, created_at
from "public"."vector_store"
where chat_id = $1 and user_id = $2 and doc_type = 'message'
  and ($3::bigint is null or id <= $3)
order by created_at asc`
	err := reader.QueryRowsCtx(ctx, &rows, query, sessionID, userID, nullableEvaluationWatermarkID(watermark))
	if err != nil && err != sqlx.ErrNotFound {
		return nil, err
	}
	return rows, nil
}

func withEvaluationRefreshLock(ctx context.Context, conn sqlx.SqlConn, userID int64, sessionID string, fn func() error) error {
	rawDB, err := conn.RawDB()
	if err != nil {
		return err
	}

	rawConn, err := rawDB.Conn(ctx)
	if err != nil {
		return err
	}
	defer rawConn.Close()

	key := evaluationRefreshLockKey(userID, sessionID)
	lockRows, err := rawConn.QueryContext(ctx, `select pg_advisory_lock($1)`, key)
	if err != nil {
		return err
	}
	lockRows.Close()
	defer func() {
		unlockRows, unlockErr := rawConn.QueryContext(context.Background(), `select pg_advisory_unlock($1)`, key)
		if unlockErr == nil {
			unlockRows.Close()
		}
	}()

	return fn()
}

func evaluationRefreshLockKey(userID int64, sessionID string) int64 {
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(fmt.Sprintf("session-evaluation:%d:%s", userID, sessionID)))
	return int64(hasher.Sum64())
}

func nullableEvaluationWatermarkID(watermark evaluationMessageWatermarkRow) any {
	if !watermark.LastMessageID.Valid {
		return nil
	}
	return watermark.LastMessageID.Int64
}

func buildResponseFromRecord(session model.ChatSession, record *model.SessionEvaluation) (*types.SessionEvaluationResp, error) {
	var dimensions []types.EvaluationDimension
	if len(record.Dimensions) > 0 {
		if err := json.Unmarshal(record.Dimensions, &dimensions); err != nil {
			return nil, err
		}
	}

	var strengths []string
	if len(record.Strengths) > 0 {
		if err := json.Unmarshal(record.Strengths, &strengths); err != nil {
			return nil, err
		}
	}

	var risks []string
	if len(record.Risks) > 0 {
		if err := json.Unmarshal(record.Risks, &risks); err != nil {
			return nil, err
		}
	}

	var suggestions []string
	if len(record.Suggestions) > 0 {
		if err := json.Unmarshal(record.Suggestions, &suggestions); err != nil {
			return nil, err
		}
	}
	if len(suggestions) == 0 {
		suggestions = buildEvaluationSuggestions(risks)
	}

	var evidence []types.EvaluationEvidence
	if len(record.Evidence) > 0 {
		if err := json.Unmarshal(record.Evidence, &evidence); err != nil {
			return nil, err
		}
	}

	return &types.SessionEvaluationResp{
		Session:          buildSessionItem(session),
		Status:           record.Status,
		Summary:          record.Summary,
		UserTurns:        record.UserTurns,
		AssistantTurns:   record.AssistantTurns,
		OverallScore:     record.OverallScore,
		RubricVersion:    record.RubricVersion,
		ScoreSource:      record.ScoreSource,
		Dimensions:       dimensions,
		Strengths:        strengths,
		Risks:            risks,
		Suggestions:      suggestions,
		Evidence:         evidence,
		FirstGeneratedAt: firstGeneratedAtOrFallback(record).Format(timeLayout),
		LastRefreshedAt:  record.UpdatedAt.Format(timeLayout),
		GeneratedAt:      record.GeneratedAt.Format(timeLayout),
	}, nil
}

func firstGeneratedAtOrFallback(record *model.SessionEvaluation) time.Time {
	if record == nil {
		return time.Time{}
	}
	if !record.FirstGeneratedAt.IsZero() {
		return record.FirstGeneratedAt
	}
	if !record.GeneratedAt.IsZero() {
		return record.GeneratedAt
	}
	return record.UpdatedAt
}

func joinEvaluationTranscript(rows []evaluationMessageRow) string {
	var builder strings.Builder
	for _, row := range rows {
		builder.WriteString(row.Role)
		builder.WriteString(": ")
		builder.WriteString(row.Content)
		builder.WriteString("\n")
	}
	return builder.String()
}

func keywordScore(text string, keywords []string, bonus int) int {
	score := 1
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			score += bonus
			break
		}
	}
	return score
}

func containsAnyLower(text string, keywords []string) bool {
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}
