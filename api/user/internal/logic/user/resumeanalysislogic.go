package user

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"sort"
	"strings"

	"GoZero-AI/api/user/internal/resumeevaluation"
	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/api/user/internal/types"
	"GoZero-AI/api/user/model"
	"GoZero-AI/internal/statuserr"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ResumeArtifactAnalysisLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

type resumeSignalDefinition struct {
	key      string
	label    string
	keywords []string
}

var resumeSkillDefinitions = []resumeSignalDefinition{
	{key: "go", label: "Go 语言", keywords: []string{"go", "golang", "goroutine", "channel", "gmp", "context", "mutex"}},
	{key: "go_zero", label: "go-zero 微服务", keywords: []string{"go-zero", "gozero", "微服务", "rpc", "etcd", "服务发现"}},
	{key: "database", label: "数据库与 SQL", keywords: []string{"postgresql", "postgres", "mysql", "sql", "索引", "事务", "慢查询", "连接池"}},
	{key: "pgvector_rag", label: "pgvector / RAG", keywords: []string{"pgvector", "rag", "embedding", "向量", "召回", "知识库"}},
	{key: "redis_cache", label: "Redis / 缓存", keywords: []string{"redis", "缓存", "cache", "限流", "分布式锁"}},
	{key: "observability", label: "可观测性", keywords: []string{"日志", "监控", "指标", "链路", "trace", "告警", "复盘"}},
	{key: "deployment", label: "部署与工程化", keywords: []string{"docker", "k8s", "kubernetes", "ci", "cd", "测试", "部署", "脚本"}},
}

var resumeFocusDefinitions = map[string][]string{
	"concurrency":   {"goroutine", "channel", "并发", "context", "gmp", "mutex", "锁", "调度"},
	"database":      {"postgresql", "postgres", "mysql", "sql", "索引", "事务", "pgvector", "数据库", "连接池"},
	"system_design": {"架构", "微服务", "分布式", "高并发", "缓存", "限流", "降级", "容量", "一致性"},
	"engineering":   {"go-zero", "gozero", "部署", "docker", "测试", "ci", "监控", "日志", "故障", "告警"},
	"network":       {"http", "rpc", "grpc", "tcp", "网络", "超时", "重试"},
	"performance":   {"性能", "优化", "p95", "qps", "延迟", "吞吐", "压测"},
	"algorithm":     {"算法", "复杂度", "数据结构", "topk", "lru"},
	"communication": {"复盘", "协作", "沟通", "文档", "推进"},
	"frontend_arch": {"vue", "vite", "组件", "前端", "浏览器"},
	"observability": {"日志", "监控", "指标", "链路", "trace", "告警", "可观测"},
}

func NewResumeArtifactAnalysisLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResumeArtifactAnalysisLogic {
	return &ResumeArtifactAnalysisLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ResumeArtifactAnalysisLogic) ResumeArtifactAnalysis(req *types.ResumeArtifactAnalysisReq) (*types.ResumeArtifactAnalysisResp, error) {
	userID, err := currentUserID(l.ctx)
	if err != nil {
		return nil, err
	}
	if err := validateInterviewPlanLimit(req.Limit); err != nil {
		return nil, err
	}

	artifact, rows, err := l.loadResumeAnalysisSource(userID, req.Id)
	if err != nil {
		return nil, err
	}

	resp, err := buildResumeArtifactAnalysis(artifact, rows, req.DirectionKey, req.Limit)
	if err != nil {
		return nil, err
	}
	if l.svcCtx.ResumeEvaluationsModel == nil {
		return &resp, nil
	}
	record, err := l.svcCtx.ResumeEvaluationsModel.FindOneByArtifactID(l.ctx, userID, artifact.ArtifactId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return &resp, nil
		}
		return nil, statuserr.ServiceUnavailable("简历评估暂不可用，请稍后重试")
	}
	directionKey, err := resolveResumeEvaluationDirection(req.DirectionKey, resp.FocusMatches)
	if err != nil {
		return nil, err
	}
	return buildResumeAnalysisResponseFromRecord(resp, record, resumeEvaluationRecordStale(record, artifact, directionKey)), nil
}

func loadResumeArtifactChunks(ctx context.Context, db sqlx.SqlConn, userID int64, artifactID string) ([]resumeArtifactChunkRow, error) {
	var rows []resumeArtifactChunkRow
	err := db.QueryRowsCtx(ctx, &rows, `select content, created_at
from "public"."vector_store"
where user_id = $1 and chat_id = $2 and doc_type = 'resume'
order by created_at asc, id asc`, userID, artifactID)
	if err != nil && err != sql.ErrNoRows && err != sqlx.ErrNotFound {
		return nil, err
	}
	return rows, nil
}

func buildResumeArtifactAnalysis(artifact types.ResumeArtifactItem, chunks []resumeArtifactChunkRow, directionKey string, limit int64) (types.ResumeArtifactAnalysisResp, error) {
	contents := resumeChunkContents(chunks)
	if len(contents) == 0 {
		return types.ResumeArtifactAnalysisResp{
			Artifact:         artifact,
			EvaluationStatus: resumeEvaluationStatusNoData,
			Level:            "high_risk",
			Summary:          "当前简历暂无可分析分块，请重新上传可解析的文本型 PDF。",
			AnalysisMeta: types.ReportMeta{
				SchemaVersion: resumeAnalysisSchemaVersion,
				Available:     false,
			},
			EvaluationMeta: types.ResumeEvaluationMeta{
				SchemaVersion: resumeEvaluationSchemaVersion,
				Available:     false,
				RubricVersion: resumeevaluation.RubricVersion,
				ScoreSource:   resumeEvaluationScoreHeuristic,
			},
		}, nil
	}

	skills := buildResumeSkillSignals(contents)
	projects := buildResumeProjectHighlights(contents)
	risks := buildResumeRiskSignals(contents, skills, projects)
	focusMatches := buildResumeFocusMatches(contents)
	config, err := buildResumeSuggestedQuestionConfig(directionKey, focusMatches)
	if err != nil {
		return types.ResumeArtifactAnalysisResp{}, err
	}
	questions := selectInterviewPlanQuestions(config, normalizeInterviewPlanLimit(limit))

	return types.ResumeArtifactAnalysisResp{
		Artifact:           artifact,
		EvaluationStatus:   resumeEvaluationStatusMissing,
		Summary:            buildResumeAnalysisSummary(artifact, skills, projects, risks),
		Skills:             skills,
		Projects:           projects,
		Risks:              risks,
		FocusMatches:       focusMatches,
		SuggestedQuestions: questions,
		AnalysisMeta: types.ReportMeta{
			SchemaVersion: resumeAnalysisSchemaVersion,
			Available:     true,
		},
		EvaluationMeta: types.ResumeEvaluationMeta{
			SchemaVersion: resumeEvaluationSchemaVersion,
			Available:     false,
			RubricVersion: resumeevaluation.RubricVersion,
			ScoreSource:   resumeEvaluationScoreHeuristic,
		},
	}, nil
}

func resumeChunkContents(chunks []resumeArtifactChunkRow) []string {
	contents := make([]string, 0, len(chunks))
	for _, chunk := range chunks {
		content := strings.TrimSpace(chunk.Content)
		if content != "" {
			contents = append(contents, content)
		}
	}
	return contents
}

func buildResumeSkillSignals(contents []string) []types.ResumeSkillSignal {
	signals := make([]types.ResumeSkillSignal, 0, len(resumeSkillDefinitions))
	for _, definition := range resumeSkillDefinitions {
		evidence := collectResumeEvidence(contents, definition.keywords, 3)
		if len(evidence) == 0 {
			continue
		}
		signals = append(signals, types.ResumeSkillSignal{
			Key:      definition.key,
			Label:    definition.label,
			Score:    min64(100, int64(len(evidence))*30+20),
			Evidence: evidence,
		})
	}
	sort.SliceStable(signals, func(i, j int) bool {
		return signals[i].Score > signals[j].Score
	})
	return signals
}

func buildResumeProjectHighlights(contents []string) []types.ResumeProjectHighlight {
	highlights := make([]types.ResumeProjectHighlight, 0, 3)
	for _, content := range contents {
		if len(highlights) >= 3 {
			break
		}
		if !containsAnyKeyword(content, []string{"项目", "系统", "平台", "服务", "go-zero", "gozero", "微服务", "rag", "pgvector", "redis", "etcd", "qps", "p95", "优化", "重构"}) {
			continue
		}
		snippet := truncateEvaluationContent(content, 220)
		highlights = append(highlights, types.ResumeProjectHighlight{
			Title:    buildResumeProjectHighlightTitle(content, len(highlights)+1),
			Summary:  buildResumeProjectHighlightSummary(content),
			Evidence: []string{snippet},
		})
	}
	return highlights
}

func buildResumeRiskSignals(contents []string, skills []types.ResumeSkillSignal, projects []types.ResumeProjectHighlight) []types.ResumeRiskSignal {
	risks := make([]types.ResumeRiskSignal, 0, 4)
	fullText := strings.Join(contents, "\n")
	if len(skills) == 0 {
		risks = append(risks, types.ResumeRiskSignal{
			Key:        "skill_signal_missing",
			Label:      "技能信号不足",
			Severity:   "high",
			Suggestion: "补充明确的技术栈、框架、数据库和工程工具关键词。",
		})
	}
	if len(projects) == 0 {
		risks = append(risks, types.ResumeRiskSignal{
			Key:        "project_context_missing",
			Label:      "项目上下文不足",
			Severity:   "high",
			Suggestion: "补充项目目标、职责边界、技术方案和最终结果。",
		})
	}
	if !containsAnyKeyword(fullText, []string{"qps", "p95", "p99", "%", "ms", "提升", "降低", "优化", "召回率"}) {
		risks = append(risks, types.ResumeRiskSignal{
			Key:        "metric_missing",
			Label:      "量化指标不足",
			Severity:   "medium",
			Suggestion: "为核心项目补充延迟、吞吐、准确率、成本或稳定性指标。",
		})
	}
	if !containsAnyKeyword(fullText, []string{"故障", "排查", "复盘", "监控", "告警", "链路", "日志"}) {
		risks = append(risks, types.ResumeRiskSignal{
			Key:        "incident_story_missing",
			Label:      "故障复盘素材不足",
			Severity:   "medium",
			Suggestion: "准备一次线上问题定位、止损和复盘的完整经历。",
		})
	}
	if !containsAnyKeyword(fullText, []string{"架构", "微服务", "分布式", "一致性", "限流", "降级", "缓存", "容量"}) {
		risks = append(risks, types.ResumeRiskSignal{
			Key:        "architecture_signal_missing",
			Label:      "架构表达不足",
			Severity:   "medium",
			Suggestion: "补充系统拆分、容量估算、缓存一致性或降级策略等架构取舍。",
		})
	}
	return risks
}

func buildResumeFocusMatches(contents []string) []types.ResumeFocusMatch {
	matches := make([]types.ResumeFocusMatch, 0, len(resumeFocusDefinitions))
	for _, option := range interviewFocusOptions {
		keywords := resumeFocusDefinitions[option.Key]
		evidence := collectResumeEvidence(contents, keywords, 2)
		if len(evidence) == 0 {
			continue
		}
		matches = append(matches, types.ResumeFocusMatch{
			Key:             option.Key,
			Label:           option.Label,
			MatchScore:      min64(100, int64(len(evidence))*35+15),
			Evidence:        evidence,
			PlannedQuestion: 0,
		})
	}
	sort.SliceStable(matches, func(i, j int) bool {
		return matches[i].MatchScore > matches[j].MatchScore
	})
	for i := range matches {
		matches[i].PlannedQuestion = int64(i + 1)
	}
	return matches
}

func buildResumeSuggestedQuestionConfig(directionKey string, matches []types.ResumeFocusMatch) (types.SessionConfigSnapshot, error) {
	normalizedDirection := strings.TrimSpace(directionKey)
	if normalizedDirection == "" {
		normalizedDirection = inferResumeDirection(matches)
	}
	if _, ok := findDirectionPreset(normalizedDirection); !ok {
		return types.SessionConfigSnapshot{}, statuserr.New(http.StatusBadRequest, "不支持的面试方向")
	}
	direction, _ := findDirectionPreset(normalizedDirection)
	allowedFocus := make(map[string]struct{}, len(direction.FocusKeys))
	for _, key := range direction.FocusKeys {
		allowedFocus[key] = struct{}{}
	}

	focusKeys := make([]string, 0, 3)
	for _, match := range matches {
		if len(focusKeys) >= 3 {
			break
		}
		if _, ok := allowedFocus[match.Key]; !ok {
			continue
		}
		focusKeys = append(focusKeys, match.Key)
	}
	createReq := types.CreateSessionReq{
		DirectionKey: normalizedDirection,
		Difficulty:   4,
		FocusKeys:    focusKeys,
	}
	_, config, err := buildSessionCreateConfig(&createReq)
	return config, err
}

func inferResumeDirection(matches []types.ResumeFocusMatch) string {
	for _, match := range matches {
		switch match.Key {
		case "frontend_arch":
			return "frontend_vue"
		case "algorithm":
			return "algorithm"
		case "system_design":
			return "system_design"
		}
	}
	return "go_backend"
}

func collectResumeEvidence(contents []string, keywords []string, limit int) []string {
	if limit <= 0 || len(keywords) == 0 {
		return []string{}
	}
	evidence := make([]string, 0, limit)
	seen := make(map[string]struct{}, limit)
	for _, content := range contents {
		if !containsAnyKeyword(content, keywords) {
			continue
		}
		snippet := truncateEvaluationContent(content, 180)
		if _, ok := seen[snippet]; ok {
			continue
		}
		seen[snippet] = struct{}{}
		evidence = append(evidence, snippet)
		if len(evidence) >= limit {
			break
		}
	}
	return evidence
}

func containsAnyKeyword(content string, keywords []string) bool {
	lowerContent := strings.ToLower(content)
	for _, keyword := range keywords {
		if keyword == "" {
			continue
		}
		if strings.Contains(lowerContent, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

func buildResumeProjectHighlightTitle(content string, index int) string {
	switch {
	case containsAnyKeyword(content, []string{"rag", "pgvector", "embedding", "知识库"}):
		return "RAG 与知识库项目"
	case containsAnyKeyword(content, []string{"go-zero", "gozero", "微服务", "etcd"}):
		return "GoZero 微服务项目"
	case containsAnyKeyword(content, []string{"redis", "缓存", "限流"}):
		return "缓存与高并发项目"
	default:
		return "项目亮点 " + string(rune('0'+index))
	}
}

func buildResumeProjectHighlightSummary(content string) string {
	switch {
	case containsAnyKeyword(content, []string{"qps", "p95", "p99", "%", "ms", "提升", "降低", "召回率"}):
		return "包含可量化的工程结果，适合作为面试中的项目成果展开。"
	case containsAnyKeyword(content, []string{"rag", "pgvector", "embedding", "知识库"}):
		return "体现了 AI 应用、向量检索和知识库工程能力。"
	case containsAnyKeyword(content, []string{"go-zero", "gozero", "微服务", "etcd"}):
		return "体现了 Go 微服务拆分、注册发现和工程治理经验。"
	case containsAnyKeyword(content, []string{"故障", "排查", "复盘", "监控", "告警"}):
		return "包含故障处理或稳定性建设素材，适合行为面试追问。"
	default:
		return "包含可展开的项目实践，需要进一步说明职责、方案和结果。"
	}
}

func buildResumeAnalysisSummary(artifact types.ResumeArtifactItem, skills []types.ResumeSkillSignal, projects []types.ResumeProjectHighlight, risks []types.ResumeRiskSignal) string {
	return "已分析简历《" + artifact.Title + "》：识别到 " +
		intToString(int64(len(skills))) + " 类技能信号、" +
		intToString(int64(len(projects))) + " 个项目亮点、" +
		intToString(int64(len(risks))) + " 个可优化风险点。"
}

func intToString(value int64) string {
	if value == 0 {
		return "0"
	}
	var digits [20]byte
	i := len(digits)
	for value > 0 {
		i--
		digits[i] = byte('0' + value%10)
		value /= 10
	}
	return string(digits[i:])
}
