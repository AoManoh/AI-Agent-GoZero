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

	"GoZero-AI/api/user/internal/resumeevaluation"
	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/api/user/internal/types"
	"GoZero-AI/api/user/model"
	"GoZero-AI/internal/statuserr"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	resumeAnalysisSchemaVersion    = "resume-analysis-v1"
	resumeEvaluationSchemaVersion  = "resume-evaluation-v1"
	resumeEvaluationStatusMissing  = "missing"
	resumeEvaluationStatusReady    = "ready"
	resumeEvaluationStatusStale    = "stale"
	resumeEvaluationStatusRunning  = "evaluating"
	resumeEvaluationStatusNoData   = "insufficient_data"
	resumeEvaluationScoreLLM       = "llm"
	resumeEvaluationScoreHeuristic = "heuristic"
	resumeEvaluationScoreFallback  = "heuristic_fallback"
)

func (l *ResumeArtifactAnalysisLogic) ResumeArtifactAnalysisPrepare(req *types.ResumeArtifactAnalysisPrepareReq) (*types.ResumeArtifactAnalysisResp, error) {
	userID, err := currentUserID(l.ctx)
	if err != nil {
		return nil, err
	}
	if err := validateInterviewPlanLimit(req.Limit); err != nil {
		return nil, err
	}
	if l.svcCtx.ResumeEvaluationsModel == nil {
		return nil, statuserr.ServiceUnavailable("简历评估暂不可用，请稍后重试")
	}

	artifact, rows, err := l.loadResumeAnalysisSource(userID, req.Id)
	if err != nil {
		return nil, err
	}
	base, err := buildResumeArtifactAnalysis(artifact, rows, req.DirectionKey, req.Limit)
	if err != nil {
		return nil, err
	}
	directionKey, err := resolveResumeEvaluationDirection(req.DirectionKey, base.FocusMatches)
	if err != nil {
		return nil, err
	}

	record, err := l.refreshResumeEvaluationRecord(userID, artifact, rows, base, directionKey, req.Force)
	if err != nil {
		return nil, err
	}
	return buildResumeAnalysisResponseFromRecord(base, record, resumeEvaluationRecordStale(record, artifact, directionKey)), nil
}

func (l *ResumeArtifactAnalysisLogic) loadResumeAnalysisSource(userID int64, artifactID string) (types.ResumeArtifactItem, []resumeArtifactChunkRow, error) {
	artifact, err := loadResumeArtifactItem(l.ctx, l.svcCtx.DB, userID, artifactID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, sqlx.ErrNotFound) {
			return types.ResumeArtifactItem{}, nil, statuserr.NotFound("简历资料不存在或已删除")
		}
		return types.ResumeArtifactItem{}, nil, statuserr.ServiceUnavailable("简历资料暂不可用，请稍后重试")
	}

	rows, err := loadResumeArtifactChunks(l.ctx, l.svcCtx.DB, userID, artifactID)
	if err != nil {
		return types.ResumeArtifactItem{}, nil, statuserr.ServiceUnavailable("简历分块暂不可用，请稍后重试")
	}
	return artifact, rows, nil
}

func (l *ResumeArtifactAnalysisLogic) refreshResumeEvaluationRecord(userID int64, artifact types.ResumeArtifactItem, rows []resumeArtifactChunkRow, base types.ResumeArtifactAnalysisResp, directionKey string, force bool) (*model.ResumeEvaluation, error) {
	existing, err := l.svcCtx.ResumeEvaluationsModel.FindOneByArtifactID(l.ctx, userID, artifact.ArtifactId)
	if err == nil && !force && !resumeEvaluationRecordStale(existing, artifact, directionKey) {
		return existing, nil
	}
	if err != nil && !errors.Is(err, model.ErrNotFound) {
		return nil, statuserr.ServiceUnavailable("简历评估暂不可用，请稍后重试")
	}
	if errors.Is(err, model.ErrNotFound) {
		existing = nil
	}

	var refreshed *model.ResumeEvaluation
	err = withResumeEvaluationRefreshLock(l.ctx, l.svcCtx.DB, userID, artifact.ArtifactId, func() error {
		currentExisting, err := l.svcCtx.ResumeEvaluationsModel.FindOneByArtifactID(l.ctx, userID, artifact.ArtifactId)
		switch {
		case err == nil && !force && !resumeEvaluationRecordStale(currentExisting, artifact, directionKey):
			refreshed = currentExisting
			return nil
		case err != nil && !errors.Is(err, model.ErrNotFound):
			return err
		case errors.Is(err, model.ErrNotFound):
			currentExisting = nil
		}

		running := buildResumeEvaluationStatusRecord(currentExisting, userID, artifact, directionKey, resumeEvaluationStatusRunning)
		if err := l.svcCtx.ResumeEvaluationsModel.Upsert(l.ctx, running); err != nil {
			return err
		}

		record := l.buildResumeEvaluationRecord(currentExisting, userID, artifact, rows, base, directionKey)
		if err := l.svcCtx.ResumeEvaluationsModel.Upsert(l.ctx, record); err != nil {
			return err
		}
		refreshed = record
		return nil
	})
	if err != nil {
		if _, ok := statuserr.StatusCode(err); ok {
			return nil, err
		}
		return nil, statuserr.ServiceUnavailable("简历评估暂不可用，请稍后重试")
	}
	return refreshed, nil
}

func (l *ResumeArtifactAnalysisLogic) buildResumeEvaluationRecord(existing *model.ResumeEvaluation, userID int64, artifact types.ResumeArtifactItem, rows []resumeArtifactChunkRow, base types.ResumeArtifactAnalysisResp, directionKey string) *model.ResumeEvaluation {
	contents := resumeChunkContents(rows)
	heuristic := buildHeuristicResumeEvaluation(base, rows, directionKey)
	status := resumeEvaluationStatusReady
	summary := heuristic.Summary
	dimensions := heuristic.Dimensions
	scoreSource := resumeEvaluationScoreHeuristic
	strengths := heuristic.Strengths
	risks := heuristic.Risks
	suggestions := heuristic.Suggestions

	if len(contents) == 0 {
		status = resumeEvaluationStatusNoData
		summary = "当前简历暂无可分析分块，请重新上传可解析的文本型 PDF。"
		scoreSource = resumeEvaluationScoreHeuristic
		dimensions = resumeevaluation.NormalizeDimensions(nil, nil)
		strengths = []string{}
		risks = []types.ResumeRiskSignal{{
			Key:        "resume_text_missing",
			Label:      "简历文本不足",
			Severity:   "high",
			Suggestion: "重新上传可解析的文本型 PDF，或确认 PDF 未被扫描图片化。",
		}}
		suggestions = []string{"重新上传可解析的文本型 PDF，确保简历正文能被提取。"}
	} else if l.svcCtx.ResumeEvaluationGenerator != nil {
		input := svc.ResumeEvaluationInput{
			Title:        artifact.Title,
			Filename:     artifact.Filename,
			DirectionKey: directionKey,
			Chunks:       contents,
		}
		if generated, err := l.svcCtx.ResumeEvaluationGenerator.Generate(l.ctx, input); err == nil {
			summary = coalesceString(generated.Summary, summary)
			dimensions = resumeevaluation.NormalizeDimensions(generated.Dimensions, dimensions)
			scoreSource = resolveResumeGeneratedScoreSource(generated)
			strengths = mergeStringSlice(generated.Strengths, strengths)
			risks = mergeResumeRiskSignals(generated.Risks, risks)
			suggestions = mergeStringSlice(generated.Suggestions, suggestions)
		} else {
			scoreSource = resumeEvaluationScoreFallback
			risks = append([]types.ResumeRiskSignal{{
				Key:        "model_unavailable",
				Label:      "模型评估降级",
				Severity:   "medium",
				Suggestion: "评估模型暂不可用，当前结果已降级为规则式评估，可稍后刷新。",
			}}, risks...)
			suggestions = append([]string{"评估模型暂不可用，建议稍后重新刷新以获得模型增强结果。"}, suggestions...)
			l.Logger.Errorf("resume evaluation generator fallback to heuristic: %v", err)
		}
	}

	dimensions = resumeevaluation.NormalizeDimensions(dimensions, heuristic.Dimensions)
	overallScore := resumeevaluation.ComputeOverallScore(dimensions)
	level := resumeevaluation.Level(overallScore)
	if status == resumeEvaluationStatusNoData {
		overallScore = 0
		level = resumeevaluation.Level(0)
	}

	return buildResumeEvaluationRecord(existing, userID, artifact, status, summary, overallScore, level, scoreSource, directionKey, dimensions, strengths, risks, suggestions, base.FocusMatches, base.SuggestedQuestions, buildResumeEvaluationEvidence(rows))
}

func buildHeuristicResumeEvaluation(base types.ResumeArtifactAnalysisResp, rows []resumeArtifactChunkRow, directionKey string) *svc.GeneratedResumeEvaluation {
	skillScore := min64(100, int64(len(base.Skills))*18+28)
	projectScore := min64(100, int64(len(base.Projects))*25+25)
	focusScore := min64(100, int64(len(base.FocusMatches))*22+30)
	metricScore := int64(76)
	engineeringScore := int64(72)
	for _, risk := range base.Risks {
		switch risk.Key {
		case "metric_missing":
			metricScore = 45
		case "engineering_signal_missing":
			engineeringScore = 48
		case "incident_story_missing":
			if engineeringScore > 58 {
				engineeringScore = 58
			}
		}
	}
	if len(base.Projects) == 0 {
		projectScore = 35
	}
	if len(base.Skills) == 0 {
		skillScore = 35
	}

	dimensions := []types.EvaluationDimension{
		{Key: "target_alignment", Label: "方向匹配度", Score: focusScore, MaxScore: 100, Summary: "根据简历中与目标方向匹配的侧重点和证据片段估算。"},
		{Key: "technical_relevance", Label: "技术相关性", Score: skillScore, MaxScore: 100, Summary: "根据技术栈、框架、数据库和工程工具信号估算。"},
		{Key: "project_depth", Label: "项目深度", Score: projectScore, MaxScore: 100, Summary: "根据项目数量、项目上下文和可展开程度估算。"},
		{Key: "impact_evidence", Label: "结果证据", Score: metricScore, MaxScore: 100, Summary: "根据是否包含延迟、吞吐、准确率、成本或稳定性指标估算。"},
		{Key: "engineering_practice", Label: "工程实践", Score: engineeringScore, MaxScore: 100, Summary: "根据测试、部署、监控、故障复盘和协作信号估算。"},
		{Key: "clarity_structure", Label: "表达结构", Score: buildResumeClarityScore(rows), MaxScore: 100, Summary: "根据简历文本结构和信息密度估算。"},
		{Key: "interview_readiness", Label: "可追问度", Score: min64(100, int64(len(base.SuggestedQuestions))*18+42), MaxScore: 100, Summary: "根据可转化为面试追问的项目和侧重点估算。"},
	}
	dimensions = resumeevaluation.NormalizeDimensions(dimensions, nil)
	strengths := buildResumeEvaluationStrengths(base)
	suggestions := buildResumeEvaluationSuggestions(base.Risks)
	return &svc.GeneratedResumeEvaluation{
		Summary:     buildResumeEvaluationSummary(base, directionKey),
		Dimensions:  dimensions,
		Strengths:   strengths,
		Risks:       base.Risks,
		Suggestions: suggestions,
	}
}

func buildResumeClarityScore(rows []resumeArtifactChunkRow) int64 {
	contents := resumeChunkContents(rows)
	if len(contents) == 0 {
		return 0
	}
	fullText := strings.Join(contents, "\n")
	score := int64(68)
	if containsAnyKeyword(fullText, []string{"项目", "经历", "教育", "技能", "工作", "实习"}) {
		score += 12
	}
	if containsAnyKeyword(fullText, []string{"负责", "主导", "参与", "设计", "优化", "落地"}) {
		score += 10
	}
	return min64(100, score)
}

func buildResumeEvaluationStrengths(base types.ResumeArtifactAnalysisResp) []string {
	strengths := make([]string, 0, 3)
	if len(base.Skills) > 0 {
		strengths = append(strengths, fmt.Sprintf("识别到 %d 类技术能力信号，可作为面试展开基础。", len(base.Skills)))
	}
	if len(base.Projects) > 0 {
		strengths = append(strengths, fmt.Sprintf("识别到 %d 个项目亮点，具备项目追问素材。", len(base.Projects)))
	}
	if len(base.FocusMatches) > 0 {
		strengths = append(strengths, "简历内容已匹配到可规划的面试侧重点。")
	}
	if len(strengths) == 0 {
		strengths = append(strengths, "已完成基础解析，但简历中的能力信号仍需补充。")
	}
	return strengths
}

func buildResumeEvaluationSuggestions(risks []types.ResumeRiskSignal) []string {
	if len(risks) == 0 {
		return []string{"保留当前项目主线，并为核心项目准备 2-3 个可量化结果。"}
	}
	suggestions := make([]string, 0, len(risks))
	seen := make(map[string]struct{}, len(risks))
	for _, risk := range risks {
		suggestion := strings.TrimSpace(risk.Suggestion)
		if suggestion == "" {
			continue
		}
		if _, ok := seen[suggestion]; ok {
			continue
		}
		seen[suggestion] = struct{}{}
		suggestions = append(suggestions, suggestion)
	}
	return suggestions
}

func buildResumeEvaluationSummary(base types.ResumeArtifactAnalysisResp, directionKey string) string {
	direction := strings.TrimSpace(directionKey)
	if direction == "" {
		direction = "默认方向"
	}
	return fmt.Sprintf("已基于 %s 评估简历：识别到 %d 类技能信号、%d 个项目亮点、%d 个可优化风险点。", direction, len(base.Skills), len(base.Projects), len(base.Risks))
}

func buildResumeEvaluationEvidence(rows []resumeArtifactChunkRow) []types.EvaluationEvidence {
	evidence := make([]types.EvaluationEvidence, 0, min(len(rows), 4))
	for _, row := range rows {
		content := strings.TrimSpace(row.Content)
		if content == "" {
			continue
		}
		evidence = append(evidence, types.EvaluationEvidence{
			Role:      "resume",
			Content:   truncateEvaluationContent(content, 180),
			CreatedAt: row.CreatedAt.Format(timeLayout),
		})
		if len(evidence) >= 4 {
			break
		}
	}
	return evidence
}

func buildResumeEvaluationRecord(existing *model.ResumeEvaluation, userID int64, artifact types.ResumeArtifactItem, status, summary string, overallScore float64, level, scoreSource, directionKey string,
	dimensions []types.EvaluationDimension, strengths []string, risks []types.ResumeRiskSignal, suggestions []string, focusMatches []types.ResumeFocusMatch, questions []types.InterviewPlanQuestion, evidence []types.EvaluationEvidence) *model.ResumeEvaluation {
	now := time.Now().UTC()
	firstGeneratedAt := now
	if existing != nil && !existing.FirstGeneratedAt.IsZero() {
		firstGeneratedAt = existing.FirstGeneratedAt
	}
	return &model.ResumeEvaluation{
		ArtifactId:          artifact.ArtifactId,
		UserId:              userID,
		Status:              status,
		Summary:             strings.TrimSpace(summary),
		OverallScore:        overallScore,
		Level:               level,
		RubricVersion:       resumeevaluation.RubricVersion,
		ScoreSource:         scoreSource,
		DirectionKey:        directionKey,
		Dimensions:          marshalJSONOrEmptyArray(dimensions),
		Strengths:           marshalJSONOrEmptyArray(strengths),
		Risks:               marshalJSONOrEmptyArray(risks),
		Suggestions:         marshalJSONOrEmptyArray(suggestions),
		FocusMatches:        marshalJSONOrEmptyArray(focusMatches),
		SuggestedQuestions:  marshalJSONOrEmptyArray(questions),
		Evidence:            marshalJSONOrEmptyArray(evidence),
		SourceResumeVersion: artifact.Version,
		SourceChunkCount:    artifact.ChunkCount,
		SourceUpdatedAt:     resumeArtifactUpdatedAt(artifact),
		FirstGeneratedAt:    firstGeneratedAt,
		GeneratedAt:         now,
		UpdatedAt:           now,
	}
}

func buildResumeEvaluationStatusRecord(existing *model.ResumeEvaluation, userID int64, artifact types.ResumeArtifactItem, directionKey, status string) *model.ResumeEvaluation {
	return buildResumeEvaluationRecord(existing, userID, artifact, status, "", 0, "", resumeEvaluationScoreHeuristic, directionKey, nil, nil, nil, nil, nil, nil, nil)
}

func buildResumeAnalysisResponseFromRecord(base types.ResumeArtifactAnalysisResp, record *model.ResumeEvaluation, stale bool) *types.ResumeArtifactAnalysisResp {
	resp := base
	status := strings.TrimSpace(record.Status)
	if status == "" {
		status = resumeEvaluationStatusMissing
	}
	if stale && status == resumeEvaluationStatusReady {
		status = resumeEvaluationStatusStale
	}

	resp.EvaluationStatus = status
	resp.OverallScore = record.OverallScore
	resp.Level = record.Level
	resp.Summary = coalesceString(record.Summary, base.Summary)
	unmarshalJSONOrDefault(record.Dimensions, &resp.Dimensions)
	unmarshalJSONOrDefault(record.Strengths, &resp.Strengths)
	unmarshalJSONOrDefault(record.Risks, &resp.Risks)
	unmarshalJSONOrDefault(record.Suggestions, &resp.Suggestions)
	unmarshalJSONOrDefault(record.FocusMatches, &resp.FocusMatches)
	unmarshalJSONOrDefault(record.SuggestedQuestions, &resp.SuggestedQuestions)
	unmarshalJSONOrDefault(record.Evidence, &resp.Evidence)
	resp.EvaluationMeta = buildResumeEvaluationMeta(record, status)
	return &resp
}

func buildResumeEvaluationMeta(record *model.ResumeEvaluation, status string) types.ResumeEvaluationMeta {
	available := status == resumeEvaluationStatusReady || status == resumeEvaluationStatusStale
	return types.ResumeEvaluationMeta{
		SchemaVersion:    resumeEvaluationSchemaVersion,
		Available:        available,
		RubricVersion:    record.RubricVersion,
		ScoreSource:      record.ScoreSource,
		GeneratedAt:      record.GeneratedAt.Format(timeLayout),
		FirstGeneratedAt: record.FirstGeneratedAt.Format(timeLayout),
		LastRefreshedAt:  record.UpdatedAt.Format(timeLayout),
	}
}

func resumeEvaluationRecordStale(record *model.ResumeEvaluation, artifact types.ResumeArtifactItem, directionKey string) bool {
	if record == nil {
		return true
	}
	if record.RubricVersion != resumeevaluation.RubricVersion {
		return true
	}
	if record.SourceResumeVersion != artifact.Version || record.SourceChunkCount != artifact.ChunkCount {
		return true
	}
	if strings.TrimSpace(directionKey) != "" && record.DirectionKey != directionKey {
		return true
	}
	updatedAt := resumeArtifactUpdatedAt(artifact)
	if updatedAt.Valid && record.SourceUpdatedAt.Valid && record.SourceUpdatedAt.Time.Before(updatedAt.Time) {
		return true
	}
	return false
}

func resolveResumeEvaluationDirection(directionKey string, matches []types.ResumeFocusMatch) (string, error) {
	normalizedDirection := strings.TrimSpace(directionKey)
	if normalizedDirection == "" {
		normalizedDirection = inferResumeDirection(matches)
	}
	if _, ok := findDirectionPreset(normalizedDirection); !ok {
		return "", statuserr.New(http.StatusBadRequest, "不支持的面试方向")
	}
	return normalizedDirection, nil
}

func resolveResumeGeneratedScoreSource(generated *svc.GeneratedResumeEvaluation) string {
	if generated != nil && resumeevaluation.HasCompleteDimensions(generated.Dimensions) {
		return resumeEvaluationScoreLLM
	}
	return "mixed"
}

func mergeResumeRiskSignals(primary, fallback []types.ResumeRiskSignal) []types.ResumeRiskSignal {
	if len(primary) == 0 {
		return fallback
	}
	seen := make(map[string]struct{}, len(primary)+len(fallback))
	result := make([]types.ResumeRiskSignal, 0, len(primary)+len(fallback))
	for _, risk := range append(primary, fallback...) {
		key := strings.TrimSpace(risk.Key)
		if key == "" {
			key = strings.TrimSpace(risk.Label)
		}
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		if !validResumeRiskSeverity(risk.Severity) {
			risk.Severity = "medium"
		}
		result = append(result, risk)
	}
	return result
}

func validResumeRiskSeverity(value string) bool {
	switch strings.TrimSpace(value) {
	case "low", "medium", "high":
		return true
	default:
		return false
	}
}

func resumeArtifactUpdatedAt(artifact types.ResumeArtifactItem) sql.NullTime {
	updatedAt := strings.TrimSpace(artifact.UpdatedAt)
	if updatedAt == "" {
		return sql.NullTime{}
	}
	parsed, err := time.Parse(timeLayout, updatedAt)
	if err != nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: parsed, Valid: true}
}

func marshalJSONOrEmptyArray(value any) []byte {
	if value == nil {
		return []byte("[]")
	}
	raw, err := json.Marshal(value)
	if err != nil || len(raw) == 0 || string(raw) == "null" {
		return []byte("[]")
	}
	return raw
}

func unmarshalJSONOrDefault(raw []byte, target any) {
	if len(raw) == 0 {
		return
	}
	_ = json.Unmarshal(raw, target)
}

func withResumeEvaluationRefreshLock(ctx context.Context, conn sqlx.SqlConn, userID int64, artifactID string, fn func() error) error {
	rawDB, err := conn.RawDB()
	if err != nil {
		return err
	}

	rawConn, err := rawDB.Conn(ctx)
	if err != nil {
		return err
	}
	defer rawConn.Close()

	key := resumeEvaluationRefreshLockKey(userID, artifactID)
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

func resumeEvaluationRefreshLockKey(userID int64, artifactID string) int64 {
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(fmt.Sprintf("resume-evaluation:%d:%s", userID, artifactID)))
	return int64(hasher.Sum64())
}
