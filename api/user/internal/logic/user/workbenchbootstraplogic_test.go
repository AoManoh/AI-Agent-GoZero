package user

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/api/user/internal/types"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func TestBuildWorkbenchActionsUsesWorkbenchRoutes(t *testing.T) {
	actions := buildWorkbenchActions(1, types.WorkbenchResumeSummary{}, types.WorkbenchKnowledgeSummary{})
	routes := make(map[string]string, len(actions))
	for _, action := range actions {
		routes[action.Key] = action.Route
	}

	want := map[string]string{
		"new_interview":    "/workbench/new",
		"upload_resume":    "/workbench/resume",
		"import_knowledge": "/workbench/knowledge",
		"review_report":    "/workbench",
	}
	for key, route := range want {
		if routes[key] != route {
			t.Fatalf("route[%s] = %q, want %q", key, routes[key], route)
		}
	}
}

func TestBuildWorkbenchResumeProjectsCountUsesResumeAnalysisSignals(t *testing.T) {
	chunks := []resumeArtifactChunkRow{
		{Content: "项目一：GoZero-AI 面试系统，使用 go-zero、PostgreSQL、pgvector 和 Redis 构建 RAG 对话链路，完成 p95 延迟优化。"},
		{Content: "项目二：日志平台服务，负责监控告警、链路追踪和故障复盘，支撑微服务稳定性治理。"},
		{Content: "普通经历：参与日常需求评审和文档维护。"},
	}

	if got := buildWorkbenchResumeProjectsCount(chunks); got != 2 {
		t.Fatalf("projectsCount = %d, want 2", got)
	}
}

func TestBuildWorkbenchResumeProjectsCountEmptyChunks(t *testing.T) {
	chunks := []resumeArtifactChunkRow{
		{Content: "   "},
	}

	if got := buildWorkbenchResumeProjectsCount(chunks); got != 0 {
		t.Fatalf("projectsCount = %d, want 0", got)
	}
}

func TestBuildResumeSummaryIncludesLatestProjectsCount(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	now := time.Now()
	mock.ExpectQuery(`from "public"\."resume_documents" d`).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows(resumeArtifactSummaryColumns()).AddRow(
			"resume-1",
			"",
			"Go 后端简历",
			int64(1),
			"resume.pdf",
			"ready",
			"ready",
			int64(100),
			int64(3),
			int64(0),
			"",
			"",
			false,
			int64(3),
			"Go 后端简历",
			float64(0),
			"",
			"",
			int64(0),
			sql.NullTime{},
			now,
			now,
		))
	mock.ExpectQuery(`from "public"\."vector_store" v`).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows(resumeArtifactSummaryColumns()))
	mock.ExpectQuery(`select content, created_at`).
		WithArgs(int64(7), "resume-1").
		WillReturnRows(sqlmock.NewRows([]string{"content", "created_at"}).
			AddRow("项目一：GoZero-AI 面试系统，使用 go-zero、PostgreSQL、pgvector 和 Redis 构建 RAG 对话链路，完成 p95 延迟优化。", now).
			AddRow("项目二：日志平台服务，负责监控告警、链路追踪和故障复盘，支撑微服务稳定性治理。", now).
			AddRow("普通经历：参与日常需求评审和文档维护。", now))

	logic := NewWorkbenchBootstrapLogic(context.Background(), &svc.ServiceContext{
		DB: sqlx.NewSqlConnFromDB(db),
	})

	summary, err := logic.buildResumeSummary(7)
	if err != nil {
		t.Fatalf("buildResumeSummary error = %v", err)
	}
	if summary.Total != 1 || summary.ChunkCount != 3 {
		t.Fatalf("summary aggregate = total %d chunks %d, want 1/3", summary.Total, summary.ChunkCount)
	}
	if summary.LatestSessionId != "resume-1" || summary.LatestTitle != "Go 后端简历" {
		t.Fatalf("latest resume = %q/%q, want resume-1/Go 后端简历", summary.LatestSessionId, summary.LatestTitle)
	}
	if summary.ProjectsCount != 2 {
		t.Fatalf("projectsCount = %d, want 2", summary.ProjectsCount)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func resumeArtifactSummaryColumns() []string {
	return []string{
		"artifact_id",
		"bound_session_id",
		"title",
		"version",
		"filename",
		"status",
		"parse_stage",
		"parse_progress",
		"processed_chunk_count",
		"failed_chunk_count",
		"parse_error_code",
		"parse_error_message",
		"parse_retryable",
		"chunk_count",
		"bound_session_name",
		"overall_score",
		"level",
		"evaluation_status",
		"risk_count",
		"latest_evaluation_at",
		"uploaded_at",
		"updated_at",
	}
}
