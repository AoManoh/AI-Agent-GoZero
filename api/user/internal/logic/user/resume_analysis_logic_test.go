package user

import (
	"context"
	"database/sql"
	"net/http"
	"testing"
	"time"

	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/api/user/internal/types"
	"GoZero-AI/internal/statuserr"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func TestResumeArtifactAnalysisExtractsSignals(t *testing.T) {
	now := time.Date(2026, 5, 9, 20, 0, 0, 0, time.UTC)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	expectResumeDocumentItem(mock, "sess-resume", now)
	mock.ExpectQuery(`from "public"\."vector_store"`).
		WithArgs(int64(7), "sess-resume").
		WillReturnRows(sqlmock.NewRows([]string{"content", "created_at"}).
			AddRow("GoZero AI 面试系统：使用 go-zero、ETCD、Redis、PostgreSQL 和 pgvector 实现 RAG 召回，P95 从 220ms 优化到 80ms。", now.Add(-time.Hour)).
			AddRow("负责 goroutine 并发任务、context 超时取消、数据库索引优化、监控告警和线上故障复盘。", now.Add(-30*time.Minute)))

	logic := NewResumeArtifactAnalysisLogic(withUserID(context.Background(), 7), &svc.ServiceContext{
		DB: sqlx.NewSqlConnFromDB(db),
	})

	resp, err := logic.ResumeArtifactAnalysis(&types.ResumeArtifactAnalysisReq{
		Id:           "sess-resume",
		DirectionKey: "go_backend",
		Limit:        4,
	})
	if err != nil {
		t.Fatalf("ResumeArtifactAnalysis() error = %v", err)
	}
	if !resp.AnalysisMeta.Available || resp.AnalysisMeta.SchemaVersion != "resume-analysis-v1" {
		t.Fatalf("AnalysisMeta = %+v, want available resume-analysis-v1", resp.AnalysisMeta)
	}
	if len(resp.Skills) == 0 {
		t.Fatal("Skills is empty")
	}
	if len(resp.Projects) == 0 {
		t.Fatal("Projects is empty")
	}
	if len(resp.FocusMatches) == 0 {
		t.Fatal("FocusMatches is empty")
	}
	if len(resp.SuggestedQuestions) == 0 || len(resp.SuggestedQuestions) > 4 {
		t.Fatalf("len(SuggestedQuestions) = %d, want 1..4", len(resp.SuggestedQuestions))
	}
	if resp.Summary == "" {
		t.Fatal("Summary is empty")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestResumeArtifactAnalysisEmptyChunks(t *testing.T) {
	now := time.Date(2026, 5, 9, 20, 0, 0, 0, time.UTC)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	expectResumeDocumentItem(mock, "sess-resume", now)
	mock.ExpectQuery(`from "public"\."vector_store"`).
		WithArgs(int64(7), "sess-resume").
		WillReturnRows(sqlmock.NewRows([]string{"content", "created_at"}))

	logic := NewResumeArtifactAnalysisLogic(withUserID(context.Background(), 7), &svc.ServiceContext{
		DB: sqlx.NewSqlConnFromDB(db),
	})

	resp, err := logic.ResumeArtifactAnalysis(&types.ResumeArtifactAnalysisReq{Id: "sess-resume"})
	if err != nil {
		t.Fatalf("ResumeArtifactAnalysis() error = %v", err)
	}
	if resp.AnalysisMeta.Available {
		t.Fatalf("AnalysisMeta.Available = true, want false")
	}
	if len(resp.Skills) != 0 || len(resp.SuggestedQuestions) != 0 {
		t.Fatalf("empty chunk analysis should not produce signals: %+v", resp)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestResumeArtifactAnalysisNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`from "public"\."resume_documents" d`).
		WithArgs(int64(7), "missing").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`from "public"\."vector_store" v`).
		WithArgs(int64(7), "missing").
		WillReturnError(sql.ErrNoRows)

	logic := NewResumeArtifactAnalysisLogic(withUserID(context.Background(), 7), &svc.ServiceContext{
		DB: sqlx.NewSqlConnFromDB(db),
	})

	_, err = logic.ResumeArtifactAnalysis(&types.ResumeArtifactAnalysisReq{Id: "missing"})
	if err == nil {
		t.Fatal("ResumeArtifactAnalysis() error = nil, want not found")
	}
	code, ok := statuserr.StatusCode(err)
	if !ok || code != http.StatusNotFound {
		t.Fatalf("status code = %d, ok=%v, want %d/true", code, ok, http.StatusNotFound)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func expectResumeDocumentItem(mock sqlmock.Sqlmock, artifactID string, now time.Time) {
	mock.ExpectQuery(`from "public"\."resume_documents" d`).
		WithArgs(int64(7), artifactID).
		WillReturnRows(sqlmock.NewRows([]string{
			"artifact_id",
			"title",
			"version",
			"filename",
			"status",
			"chunk_count",
			"bound_session_name",
			"uploaded_at",
			"updated_at",
		}).AddRow(
			artifactID,
			"GoZero AI 后端简历",
			int64(1),
			"resume.pdf",
			"ready",
			int64(2),
			"Go 后端面试",
			now.Add(-time.Hour),
			now,
		))
}
