package user

import (
	"context"
	"database/sql"
	"net/http"
	"testing"
	"time"

	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/api/user/internal/types"
	"GoZero-AI/api/user/model"
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
	if resp.EvaluationStatus != resumeEvaluationStatusMissing {
		t.Fatalf("EvaluationStatus = %q, want missing", resp.EvaluationStatus)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestResumeArtifactAnalysisReturnsPersistedEvaluation(t *testing.T) {
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
			AddRow("GoZero AI 面试系统：使用 go-zero、Redis 和 PostgreSQL 实现面试 RAG，P95 从 220ms 优化到 80ms。", now.Add(-time.Hour)).
			AddRow("负责项目架构、数据库索引优化、监控告警和线上故障复盘。", now.Add(-30*time.Minute)))

	record := &model.ResumeEvaluation{
		ArtifactId:          "sess-resume",
		UserId:              7,
		Status:              resumeEvaluationStatusReady,
		Summary:             "持久化评估摘要",
		OverallScore:        88,
		Level:               "strong",
		RubricVersion:       "resume-rubric-v1",
		ScoreSource:         "llm",
		DirectionKey:        "go_backend",
		Dimensions:          marshalJSONOrEmptyArray([]types.EvaluationDimension{{Key: "target_alignment", Label: "方向匹配度", Score: 88, MaxScore: 100, Summary: "匹配"}}),
		Strengths:           marshalJSONOrEmptyArray([]string{"项目经历清晰"}),
		Risks:               marshalJSONOrEmptyArray([]types.ResumeRiskSignal{}),
		Suggestions:         marshalJSONOrEmptyArray([]string{"补充更多压测数据"}),
		FocusMatches:        marshalJSONOrEmptyArray([]types.ResumeFocusMatch{}),
		SuggestedQuestions:  marshalJSONOrEmptyArray([]types.InterviewPlanQuestion{}),
		Evidence:            marshalJSONOrEmptyArray([]types.EvaluationEvidence{}),
		SourceResumeVersion: 1,
		SourceChunkCount:    2,
		SourceUpdatedAt:     sql.NullTime{Time: now, Valid: true},
		FirstGeneratedAt:    now,
		GeneratedAt:         now,
		UpdatedAt:           now,
	}

	logic := NewResumeArtifactAnalysisLogic(withUserID(context.Background(), 7), &svc.ServiceContext{
		DB:                     sqlx.NewSqlConnFromDB(db),
		ResumeEvaluationsModel: &stubResumeEvaluationsModel{record: record},
	})

	resp, err := logic.ResumeArtifactAnalysis(&types.ResumeArtifactAnalysisReq{
		Id:           "sess-resume",
		DirectionKey: "go_backend",
	})
	if err != nil {
		t.Fatalf("ResumeArtifactAnalysis() error = %v", err)
	}
	if resp.EvaluationStatus != resumeEvaluationStatusReady || resp.OverallScore != 88 {
		t.Fatalf("evaluation = status %q score %.2f, want ready/88", resp.EvaluationStatus, resp.OverallScore)
	}
	if resp.Summary != "持久化评估摘要" {
		t.Fatalf("Summary = %q, want persisted summary", resp.Summary)
	}
	if !resp.EvaluationMeta.Available || resp.EvaluationMeta.ScoreSource != "llm" {
		t.Fatalf("EvaluationMeta = %+v, want available llm", resp.EvaluationMeta)
	}
	if len(resp.Dimensions) == 0 || len(resp.Strengths) == 0 || len(resp.Suggestions) == 0 {
		t.Fatalf("persisted evaluation fields missing: %+v", resp)
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
	if resp.EvaluationStatus != resumeEvaluationStatusNoData {
		t.Fatalf("EvaluationStatus = %q, want insufficient_data", resp.EvaluationStatus)
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
		}).AddRow(
			artifactID,
			artifactID,
			"GoZero AI 后端简历",
			int64(1),
			"resume.pdf",
			"ready",
			"ready",
			int64(100),
			int64(2),
			int64(0),
			"",
			"",
			false,
			int64(2),
			"Go 后端面试",
			float64(0),
			"",
			"",
			int64(0),
			sql.NullTime{},
			now.Add(-time.Hour),
			now,
		))
}

type stubResumeEvaluationsModel struct {
	record *model.ResumeEvaluation
	err    error
}

func (s *stubResumeEvaluationsModel) FindOneByArtifactID(_ context.Context, _ int64, _ string) (*model.ResumeEvaluation, error) {
	if s.err != nil {
		return nil, s.err
	}
	if s.record == nil {
		return nil, model.ErrNotFound
	}
	return s.record, nil
}

func (s *stubResumeEvaluationsModel) Upsert(_ context.Context, data *model.ResumeEvaluation) error {
	s.record = data
	return nil
}
