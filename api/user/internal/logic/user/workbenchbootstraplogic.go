package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"math"
	"sort"
	"time"

	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/api/user/internal/types"
	"GoZero-AI/api/user/model"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type WorkbenchBootstrapLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

type workbenchResumeLatestRow struct {
	SessionId string       `db:"session_id"`
	Title     string       `db:"title"`
	UpdatedAt sql.NullTime `db:"updated_at"`
}

type workbenchResumeAggregateRow struct {
	Total      int64 `db:"total"`
	ChunkCount int64 `db:"chunk_count"`
}

type workbenchKnowledgeAggregateRow struct {
	Documents     int64        `db:"documents"`
	Chunks        int64        `db:"chunks"`
	LatestAddedAt sql.NullTime `db:"latest_added_at"`
}

type workbenchKnowledgeLatestRow struct {
	Title   string    `db:"title"`
	AddedAt time.Time `db:"added_at"`
}

type workbenchEvaluationDimensionRow struct {
	Dimensions []byte `db:"dimensions"`
}

func NewWorkbenchBootstrapLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WorkbenchBootstrapLogic {
	return &WorkbenchBootstrapLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WorkbenchBootstrapLogic) WorkbenchBootstrap(_ *types.WorkbenchBootstrapReq) (*types.WorkbenchBootstrapResp, error) {
	userID, err := currentUserID(l.ctx)
	if err != nil {
		return nil, err
	}

	userEntity, err := l.svcCtx.UsersModel.FindOne(l.ctx, userID)
	if err != nil {
		return nil, err
	}

	sessions, err := l.svcCtx.ChatSessionsModel.FindByUserID(l.ctx, userID)
	if err != nil {
		return nil, err
	}

	rows, err := fetchReportCenterOverviewRows(l.ctx, l.svcCtx, userID)
	if err != nil {
		return nil, err
	}
	totals, _, _ := buildReportCenterOverview(rows)

	resumeSummary, err := l.buildResumeSummary(userID)
	if err != nil {
		return nil, err
	}
	knowledgeSummary, err := l.buildKnowledgeSummary(userID)
	if err != nil {
		return nil, err
	}
	abilityRadar, err := l.buildAbilityRadar(userID)
	if err != nil {
		return nil, err
	}

	recentSessions := buildWorkbenchRecentSessions(sessions, 5)
	continueSession := buildWorkbenchContinueSession(sessions)

	return &types.WorkbenchBootstrapResp{
		User: types.ProfileResp{
			UserId:    userEntity.Id,
			Username:  userEntity.Username,
			CreatedAt: userEntity.CreatedAt.Format(timeLayout),
		},
		Stats: types.WorkbenchStats{
			CompletedInterviews: int64(countCompletedSessions(sessions)),
			AverageScore:        totals.AverageScore,
			LastPracticeAt:      totals.LastActivityAt,
		},
		ContinueSession:  continueSession,
		RecentSessions:   recentSessions,
		AbilityRadar:     abilityRadar,
		ResumeSummary:    resumeSummary,
		KnowledgeSummary: knowledgeSummary,
		NextActions:      buildWorkbenchActions(len(sessions), resumeSummary, knowledgeSummary),
		WorkbenchMeta: types.ReportMeta{
			SchemaVersion: "workbench-bootstrap-v1",
			Available:     true,
		},
	}, nil
}

func buildWorkbenchRecentSessions(sessions []model.ChatSession, limit int) []types.SessionItem {
	if limit <= 0 || len(sessions) == 0 {
		return []types.SessionItem{}
	}
	if len(sessions) < limit {
		limit = len(sessions)
	}
	items := make([]types.SessionItem, 0, limit)
	for _, session := range sessions[:limit] {
		items = append(items, buildSessionItem(session))
	}
	return items
}

func buildWorkbenchContinueSession(sessions []model.ChatSession) *types.WorkbenchContinueSession {
	for _, session := range sessions {
		if session.CompletedAt.Valid {
			continue
		}
		config := buildSessionConfigSnapshot(session)
		return &types.WorkbenchContinueSession{
			Session:  buildSessionItem(session),
			Config:   config,
			Progress: config.ProgressPercent,
			Depth:    config.FollowUpDepth,
		}
	}
	return nil
}

func countCompletedSessions(sessions []model.ChatSession) int {
	count := 0
	for _, session := range sessions {
		if session.CompletedAt.Valid {
			count++
		}
	}
	return count
}

func (l *WorkbenchBootstrapLogic) buildResumeSummary(userID int64) (types.WorkbenchResumeSummary, error) {
	var aggregate workbenchResumeAggregateRow
	err := l.svcCtx.DB.QueryRowCtx(l.ctx, &aggregate, `select
count(distinct chat_id) as total,
count(*) as chunk_count
from "public"."vector_store"
where user_id = $1 and doc_type = 'resume'`, userID)
	if err != nil && err != sql.ErrNoRows {
		return types.WorkbenchResumeSummary{}, err
	}

	summary := types.WorkbenchResumeSummary{
		Total:      aggregate.Total,
		ChunkCount: aggregate.ChunkCount,
	}

	var latest workbenchResumeLatestRow
	err = l.svcCtx.DB.QueryRowCtx(l.ctx, &latest, `select
v.chat_id as session_id,
coalesce(s.title, v.chat_id) as title,
max(v.created_at) as updated_at
from "public"."vector_store" v
left join "public"."chat_sessions" s
  on s.session_id = v.chat_id and s.user_id = v.user_id
where v.user_id = $1 and v.doc_type = 'resume'
group by v.chat_id, s.title
order by max(v.created_at) desc
limit 1`, userID)
	if err != nil && err != sqlx.ErrNotFound && err != sql.ErrNoRows {
		return types.WorkbenchResumeSummary{}, err
	}
	if err == nil {
		summary.LatestSessionId = latest.SessionId
		summary.LatestTitle = latest.Title
		if latest.UpdatedAt.Valid {
			summary.LatestUpdatedAt = latest.UpdatedAt.Time.Format(timeLayout)
		}
		chunks, err := loadResumeArtifactChunks(l.ctx, l.svcCtx.DB, userID, latest.SessionId)
		if err != nil {
			return types.WorkbenchResumeSummary{}, err
		}
		summary.ProjectsCount = buildWorkbenchResumeProjectsCount(chunks)
	}
	return summary, nil
}

func buildWorkbenchResumeProjectsCount(chunks []resumeArtifactChunkRow) int64 {
	contents := resumeChunkContents(chunks)
	if len(contents) == 0 {
		return 0
	}
	return int64(len(buildResumeProjectHighlights(contents)))
}

func (l *WorkbenchBootstrapLogic) buildKnowledgeSummary(userID int64) (types.WorkbenchKnowledgeSummary, error) {
	var aggregate workbenchKnowledgeAggregateRow
	err := l.svcCtx.DB.QueryRowCtx(l.ctx, &aggregate, `select
count(distinct title) as documents,
count(*) as chunks,
max(created_at) as latest_added_at
from "public"."knowledge_base"
where user_id = $1 or user_id = 1`, userID)
	if err != nil && err != sql.ErrNoRows {
		return types.WorkbenchKnowledgeSummary{}, err
	}

	summary := types.WorkbenchKnowledgeSummary{
		Documents: aggregate.Documents,
		Chunks:    aggregate.Chunks,
	}
	if aggregate.LatestAddedAt.Valid {
		summary.LatestAddedAt = aggregate.LatestAddedAt.Time.Format(timeLayout)
	}

	var latest workbenchKnowledgeLatestRow
	err = l.svcCtx.DB.QueryRowCtx(l.ctx, &latest, `select title, created_at as added_at
from "public"."knowledge_base"
where user_id = $1 or user_id = 1
order by created_at desc, id desc
limit 1`, userID)
	if err != nil && err != sqlx.ErrNotFound && err != sql.ErrNoRows {
		return types.WorkbenchKnowledgeSummary{}, err
	}
	if err == nil {
		summary.LatestTitle = latest.Title
	}
	return summary, nil
}

func (l *WorkbenchBootstrapLogic) buildAbilityRadar(userID int64) ([]types.AbilityRadarPoint, error) {
	var rows []workbenchEvaluationDimensionRow
	err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &rows, `select dimensions
from "public"."session_evaluations"
where user_id = $1 and status = 'ready'
order by generated_at desc
limit 30`, userID)
	if err != nil && err != sqlx.ErrNotFound && err != sql.ErrNoRows {
		return nil, err
	}
	if len(rows) == 0 {
		return defaultAbilityRadar(), nil
	}

	type acc struct {
		label string
		total float64
		count int64
	}
	accs := make(map[string]*acc)
	for _, row := range rows {
		var dimensions []types.EvaluationDimension
		if err := json.Unmarshal(row.Dimensions, &dimensions); err != nil {
			continue
		}
		for _, dimension := range dimensions {
			if dimension.MaxScore <= 0 {
				continue
			}
			current := accs[dimension.Key]
			if current == nil {
				current = &acc{label: dimension.Label}
				accs[dimension.Key] = current
			}
			current.total += float64(dimension.Score) / float64(dimension.MaxScore) * 100
			current.count++
		}
	}
	if len(accs) == 0 {
		return defaultAbilityRadar(), nil
	}

	points := make([]types.AbilityRadarPoint, 0, len(accs))
	for key, current := range accs {
		score := int64(math.Round(current.total / float64(current.count)))
		points = append(points, types.AbilityRadarPoint{
			Key:      key,
			Label:    current.label,
			Score:    score,
			MaxScore: 100,
		})
	}
	sort.Slice(points, func(i, j int) bool {
		return points[i].Key < points[j].Key
	})
	return points, nil
}

func defaultAbilityRadar() []types.AbilityRadarPoint {
	return []types.AbilityRadarPoint{
		{Key: "technical_depth", Label: "技术深度", Score: 0, MaxScore: 100},
		{Key: "engineering_practice", Label: "工程实践", Score: 0, MaxScore: 100},
		{Key: "architecture_sense", Label: "架构意识", Score: 0, MaxScore: 100},
		{Key: "communication", Label: "表达与沟通", Score: 0, MaxScore: 100},
	}
}

func buildWorkbenchActions(sessionCount int, resume types.WorkbenchResumeSummary, knowledge types.WorkbenchKnowledgeSummary) []types.WorkbenchAction {
	actions := []types.WorkbenchAction{
		{Key: "new_interview", Label: "新建面试", Description: "选择方向、难度和侧重点，开始一场新的模拟面试。", Route: "/workbench/new"},
	}
	if resume.Total == 0 {
		actions = append(actions, types.WorkbenchAction{Key: "upload_resume", Label: "上传简历", Description: "上传简历后，AI 会围绕你的项目经历追问。", Route: "/workbench/resume"})
	}
	if knowledge.Chunks == 0 {
		actions = append(actions, types.WorkbenchAction{Key: "import_knowledge", Label: "导入知识", Description: "导入题库或技术资料，提高 RAG 追问质量。", Route: "/workbench/knowledge"})
	}
	if sessionCount > 0 {
		actions = append(actions, types.WorkbenchAction{Key: "review_report", Label: "复盘报告", Description: "查看最近面试的评分、证据和改进建议。", Route: "/workbench"})
	}
	return actions
}
