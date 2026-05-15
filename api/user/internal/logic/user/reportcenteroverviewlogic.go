package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"sort"
	"time"

	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/api/user/internal/types"
	"GoZero-AI/api/user/model"
	"GoZero-AI/internal/sessionmode"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ReportCenterOverviewLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

type reportCenterOverviewRow struct {
	SessionId       string          `db:"session_id"`
	Title           string          `db:"title"`
	Mode            string          `db:"mode"`
	CreatedAt       time.Time       `db:"created_at"`
	UpdatedAt       time.Time       `db:"updated_at"`
	LastMessageAt   sql.NullTime    `db:"last_message_at"`
	CompletedAt     sql.NullTime    `db:"completed_at"`
	MessageCount    int64           `db:"message_count"`
	IsActive        bool            `db:"is_active"`
	Status          sql.NullString  `db:"status"`
	Summary         sql.NullString  `db:"summary"`
	OverallScore    sql.NullFloat64 `db:"overall_score"`
	GeneratedAt     sql.NullTime    `db:"generated_at"`
	Suggestions     []byte          `db:"suggestions"`
	ResumeChunks    int64           `db:"resume_chunks"`
	ResumeUpdatedAt sql.NullTime    `db:"resume_updated_at"`
}

type reportCenterAccumulator struct {
	ModeKey                  string
	SessionCount             int64
	MessageCount             int64
	EvaluatedSessions        int64
	ReadySessions            int64
	DraftSessions            int64
	InsufficientDataSessions int64
	ResumeBackedSessions     int64
	ScoreTotal               float64
	ScoreCount               int64
	LastActivityAt           time.Time
}

func NewReportCenterOverviewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReportCenterOverviewLogic {
	return &ReportCenterOverviewLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReportCenterOverviewLogic) ReportCenterOverview(_ *types.ReportCenterOverviewReq) (*types.ReportCenterOverviewResp, error) {
	userID, err := currentUserID(l.ctx)
	if err != nil {
		return nil, err
	}

	rows, err := fetchReportCenterOverviewRows(l.ctx, l.svcCtx, userID)
	if err != nil {
		return nil, err
	}

	totals, modes, recentReports := buildReportCenterOverview(rows)

	return &types.ReportCenterOverviewResp{
		Totals:        totals,
		Modes:         modes,
		RecentReports: recentReports,
		OverviewMeta: types.ReportMeta{
			SchemaVersion: "report-center-overview-v1",
			Available:     true,
		},
	}, nil
}

func fetchReportCenterOverviewRows(ctx context.Context, svcCtx *svc.ServiceContext, userID int64) ([]reportCenterOverviewRow, error) {
	var rows []reportCenterOverviewRow
	query := `select
    s.session_id,
    s.title,
    s.mode,
    s.created_at,
    s.updated_at,
    s.last_message_at,
    s.completed_at,
    s.message_count,
    s.is_active,
    e.status,
    e.summary,
    e.overall_score,
    e.generated_at,
    e.suggestions,
    coalesce(r.resume_chunks, 0) as resume_chunks,
    r.resume_updated_at
from "public"."chat_sessions" s
left join "public"."session_evaluations" e
  on e.session_id = s.session_id and e.user_id = s.user_id
left join (
    select chat_id, user_id, count(*) as resume_chunks, max(created_at) as resume_updated_at
    from "public"."vector_store"
    where doc_type = 'resume'
    group by chat_id, user_id
) r
  on r.chat_id = coalesce(nullif(s.resume_artifact_id, ''), s.session_id) and r.user_id = s.user_id
where s.user_id = $1 and s.is_active = true
order by coalesce(s.last_message_at, s.updated_at, s.created_at) desc, s.id desc`

	err := svcCtx.DB.QueryRowsCtx(ctx, &rows, query, userID)
	switch err {
	case nil:
		sortReportCenterOverviewRows(rows)
		return rows, nil
	case sqlx.ErrNotFound, sql.ErrNoRows:
		return []reportCenterOverviewRow{}, nil
	default:
		return nil, err
	}
}

func sortReportCenterOverviewRows(rows []reportCenterOverviewRow) {
	sort.Slice(rows, func(i, j int) bool {
		left := resolveReportCenterActivity(rows[i])
		right := resolveReportCenterActivity(rows[j])
		if !left.Equal(right) {
			return left.After(right)
		}
		if !rows[i].UpdatedAt.Equal(rows[j].UpdatedAt) {
			return rows[i].UpdatedAt.After(rows[j].UpdatedAt)
		}
		if !rows[i].CreatedAt.Equal(rows[j].CreatedAt) {
			return rows[i].CreatedAt.After(rows[j].CreatedAt)
		}
		return rows[i].SessionId > rows[j].SessionId
	})
}

func buildReportCenterOverview(rows []reportCenterOverviewRow) (types.ReportCenterTotals, []types.ReportCenterModeSummary, []types.ReportCenterRecentReport) {
	totals := types.ReportCenterTotals{
		TotalSessions: int64(len(rows)),
	}
	modeAccumulators := make(map[string]*reportCenterAccumulator, len(sessionmode.AllKeys()))
	for _, key := range sessionmode.AllKeys() {
		modeAccumulators[key] = &reportCenterAccumulator{ModeKey: key}
	}

	recentReports := make([]types.ReportCenterRecentReport, 0, len(rows))
	var lastActivity time.Time
	for _, row := range rows {
		modeKey := sessionmode.NormalizeKey(row.Mode)
		activityAt := resolveReportCenterActivity(row)
		if activityAt.After(lastActivity) {
			lastActivity = activityAt
		}

		if row.ResumeChunks > 0 {
			totals.ResumeBackedSessions++
		}

		status := normalizeOverviewStatus(row.Status)
		switch status {
		case "ready":
			totals.ReadySessions++
		case "insufficient_data":
			totals.InsufficientDataSessions++
		default:
			totals.DraftSessions++
		}

		if row.Status.Valid {
			totals.EvaluatedSessions++
		}
		if status == "ready" && row.OverallScore.Valid {
			totals.AverageScore += row.OverallScore.Float64
		}

		acc := modeAccumulators[modeKey]
		acc.SessionCount++
		acc.MessageCount += row.MessageCount
		if row.Status.Valid {
			acc.EvaluatedSessions++
		}
		if status == "ready" {
			acc.ReadySessions++
		} else if status == "insufficient_data" {
			acc.InsufficientDataSessions++
		} else {
			acc.DraftSessions++
		}
		if row.ResumeChunks > 0 {
			acc.ResumeBackedSessions++
		}
		if status == "ready" && row.OverallScore.Valid {
			acc.ScoreTotal += row.OverallScore.Float64
			acc.ScoreCount++
		}
		if activityAt.After(acc.LastActivityAt) {
			acc.LastActivityAt = activityAt
		}

		recentReports = append(recentReports, buildReportCenterRecentReport(row))
	}

	if totals.ReadySessions > 0 {
		totals.AverageScore = totals.AverageScore / float64(totals.ReadySessions)
	}
	totals.LastActivityAt = formatTimeIfSet(lastActivity)

	sort.Slice(recentReports, func(i, j int) bool {
		return recentReports[i].LastActivityAt > recentReports[j].LastActivityAt
	})
	if len(recentReports) > 6 {
		recentReports = recentReports[:6]
	}

	modes := make([]types.ReportCenterModeSummary, 0, len(modeAccumulators))
	for _, key := range sessionmode.AllKeys() {
		acc := modeAccumulators[key]
		avgScore := 0.0
		if acc.ScoreCount > 0 {
			avgScore = acc.ScoreTotal / float64(acc.ScoreCount)
		}
		modes = append(modes, types.ReportCenterModeSummary{
			Mode:                     sessionmode.DisplayName(key),
			ModeKey:                  key,
			SessionCount:             acc.SessionCount,
			MessageCount:             acc.MessageCount,
			EvaluatedSessions:        acc.EvaluatedSessions,
			ReadySessions:            acc.ReadySessions,
			DraftSessions:            acc.DraftSessions,
			InsufficientDataSessions: acc.InsufficientDataSessions,
			ResumeBackedSessions:     acc.ResumeBackedSessions,
			AverageScore:             avgScore,
			LastActivityAt:           formatTimeIfSet(acc.LastActivityAt),
		})
	}

	return totals, modes, recentReports
}

func buildReportCenterModeCards(rows []reportCenterOverviewRow) []types.ReportCenterModeCard {
	modeAccumulators := make(map[string]*reportCenterAccumulator, len(sessionmode.AllKeys()))
	spotlights := make(map[string]*types.ReportCenterRecentReport, len(sessionmode.AllKeys()))
	for _, key := range sessionmode.AllKeys() {
		modeAccumulators[key] = &reportCenterAccumulator{ModeKey: key}
	}

	for _, row := range rows {
		modeKey := sessionmode.NormalizeKey(row.Mode)
		acc := modeAccumulators[modeKey]
		status := normalizeOverviewStatus(row.Status)
		acc.SessionCount++
		acc.MessageCount += row.MessageCount
		if row.Status.Valid {
			acc.EvaluatedSessions++
		}
		if status == "ready" {
			acc.ReadySessions++
		} else if status == "insufficient_data" {
			acc.InsufficientDataSessions++
		} else {
			acc.DraftSessions++
		}
		if row.ResumeChunks > 0 {
			acc.ResumeBackedSessions++
		}
		if status == "ready" && row.OverallScore.Valid {
			acc.ScoreTotal += row.OverallScore.Float64
			acc.ScoreCount++
		}
		activityAt := resolveReportCenterActivity(row)
		if activityAt.After(acc.LastActivityAt) {
			acc.LastActivityAt = activityAt
			report := buildReportCenterRecentReport(row)
			spotlights[modeKey] = &report
		}
	}

	cards := make([]types.ReportCenterModeCard, 0, len(sessionmode.AllKeys()))
	for _, key := range sessionmode.AllKeys() {
		acc := modeAccumulators[key]
		avgScore := 0.0
		if acc.ScoreCount > 0 {
			avgScore = acc.ScoreTotal / float64(acc.ScoreCount)
		}
		cards = append(cards, types.ReportCenterModeCard{
			Mode:                     sessionmode.DisplayName(key),
			ModeKey:                  key,
			SessionCount:             acc.SessionCount,
			MessageCount:             acc.MessageCount,
			EvaluatedSessions:        acc.EvaluatedSessions,
			ReadySessions:            acc.ReadySessions,
			DraftSessions:            acc.DraftSessions,
			InsufficientDataSessions: acc.InsufficientDataSessions,
			ResumeBackedSessions:     acc.ResumeBackedSessions,
			AverageScore:             avgScore,
			HasSessions:              acc.SessionCount > 0,
			HasReadyReport:           acc.ReadySessions > 0,
			HasResumeContext:         acc.ResumeBackedSessions > 0,
			AttentionState:           buildModeAttentionState(acc),
			LastActivityAt:           formatTimeIfSet(acc.LastActivityAt),
			Spotlight:                spotlights[key],
		})
	}

	return cards
}

func buildModeAttentionState(acc *reportCenterAccumulator) string {
	switch {
	case acc.SessionCount == 0:
		return "empty"
	case acc.ResumeBackedSessions > 0 && acc.ReadySessions == 0:
		return "resume_only"
	case acc.ReadySessions > 0:
		return "ready"
	default:
		return "draft"
	}
}

func buildReportCenterSession(row reportCenterOverviewRow) types.SessionItem {
	session := model.ChatSession{
		SessionId:     row.SessionId,
		UserId:        0,
		Title:         row.Title,
		Mode:          row.Mode,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
		LastMessageAt: row.LastMessageAt,
		CompletedAt:   row.CompletedAt,
		MessageCount:  row.MessageCount,
		IsActive:      row.IsActive,
	}
	return buildSessionItem(session)
}

func buildReportCenterRecentReport(row reportCenterOverviewRow) types.ReportCenterRecentReport {
	status := normalizeOverviewStatus(row.Status)
	activityAt := resolveReportCenterActivity(row)
	return types.ReportCenterRecentReport{
		Session:        buildReportCenterSession(row),
		Status:         status,
		Summary:        nullStringOrFallback(row.Summary, buildOverviewFallbackSummary(status)),
		OverallScore:   nullFloatOrZero(row.OverallScore),
		HasResume:      row.ResumeChunks > 0,
		ResumeChunks:   row.ResumeChunks,
		NextAction:     decodeFirstSuggestion(row.Suggestions),
		GeneratedAt:    formatNullTime(row.GeneratedAt),
		LastActivityAt: formatTimeIfSet(activityAt),
	}
}

func resolveReportCenterActivity(row reportCenterOverviewRow) time.Time {
	candidates := []time.Time{row.CreatedAt, row.UpdatedAt}
	if row.LastMessageAt.Valid {
		candidates = append(candidates, row.LastMessageAt.Time)
	}
	if row.CompletedAt.Valid {
		candidates = append(candidates, row.CompletedAt.Time)
	}
	if row.GeneratedAt.Valid {
		candidates = append(candidates, row.GeneratedAt.Time)
	}
	if row.ResumeUpdatedAt.Valid {
		candidates = append(candidates, row.ResumeUpdatedAt.Time)
	}

	latest := candidates[0]
	for _, candidate := range candidates[1:] {
		if candidate.After(latest) {
			latest = candidate
		}
	}
	return latest
}

func normalizeOverviewStatus(value sql.NullString) string {
	if !value.Valid {
		return "draft"
	}
	switch value.String {
	case "ready", "insufficient_data", "draft":
		return value.String
	default:
		return "draft"
	}
}

func nullStringOrFallback(value sql.NullString, fallback string) string {
	if value.Valid && value.String != "" {
		return value.String
	}
	return fallback
}

func nullFloatOrZero(value sql.NullFloat64) float64 {
	if value.Valid {
		return value.Float64
	}
	return 0
}

func formatNullTime(value sql.NullTime) string {
	if value.Valid {
		return value.Time.Format(timeLayout)
	}
	return ""
}

func formatTimeIfSet(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format(timeLayout)
}

func decodeFirstSuggestion(raw []byte) string {
	if len(raw) == 0 {
		return ""
	}
	var suggestions []string
	if err := json.Unmarshal(raw, &suggestions); err != nil || len(suggestions) == 0 {
		return ""
	}
	return suggestions[0]
}

func buildOverviewFallbackSummary(status string) string {
	switch status {
	case "ready":
		return "当前会话已生成可用的结构化评估摘要。"
	case "insufficient_data":
		return "当前会话样本不足，尚未形成稳定报告结论。"
	default:
		return "当前会话已进入报告中心，但仍处于草稿或待补样本状态。"
	}
}
