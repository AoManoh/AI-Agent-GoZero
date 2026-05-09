package user

import (
	"context"
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

func TestSessionReportPrepareFreshEvaluationReturnsSummary(t *testing.T) {
	now := time.Date(2026, 5, 9, 16, 0, 0, 0, time.UTC)
	session := reportReadinessSession(now)
	record := completeEvaluationRecordForTest(t, session.SessionId, now)
	record.SourceLastMessageID.Valid = true
	record.SourceLastMessageID.Int64 = 42
	record.SourceLastMessageAt.Valid = true
	record.SourceLastMessageAt.Time = now.Add(-30 * time.Minute)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`select id as last_message_id, created_at as last_message_at`).
		WithArgs(session.SessionId, session.UserId).
		WillReturnRows(sqlmock.NewRows([]string{"last_message_id", "last_message_at"}).AddRow(int64(42), now.Add(-30*time.Minute)))
	mock.ExpectQuery(`select id as last_message_id, created_at as last_message_at`).
		WithArgs(session.SessionId, session.UserId).
		WillReturnRows(sqlmock.NewRows([]string{"last_message_id", "last_message_at"}).AddRow(int64(42), now.Add(-30*time.Minute)))
	mock.ExpectQuery(`select\s+min\(created_at\) as first_message_at,\s+max\(created_at\) as last_message_at`).
		WithArgs(session.SessionId, session.UserId, record.SourceLastMessageID.Int64, record.GeneratedAt).
		WillReturnRows(sqlmock.NewRows([]string{"first_message_at", "last_message_at"}).AddRow(now.Add(-2*time.Hour), now.Add(-45*time.Minute)))
	mock.ExpectQuery(`select role, content, created_at`).
		WithArgs(session.SessionId, session.UserId, record.SourceLastMessageID.Int64, record.GeneratedAt).
		WillReturnRows(sqlmock.NewRows([]string{"role", "content", "created_at"}).
			AddRow("assistant", "assistant latest", now.Add(-45*time.Minute)).
			AddRow("user", "user latest", now.Add(-46*time.Minute)))
	mock.ExpectQuery(`select\s+count\(\*\) as resume_chunks,\s+max\(created_at\) as resume_updated_at`).
		WithArgs(session.SessionId, session.UserId, record.GeneratedAt).
		WillReturnRows(sqlmock.NewRows([]string{"resume_chunks", "resume_updated_at"}).AddRow(2, now.Add(-40*time.Minute)))

	logic := NewSessionReportPrepareLogic(withUserID(context.Background(), session.UserId), &svc.ServiceContext{
		DB:                      sqlx.NewSqlConnFromDB(db),
		ChatSessionsModel:       &stubChatSessionsModel{session: session},
		SessionEvaluationsModel: &stubSessionEvaluationsModel{record: record},
	})

	resp, err := logic.SessionReportPrepare(&types.SessionReportPrepareReq{Id: session.SessionId})
	if err != nil {
		t.Fatalf("SessionReportPrepare() error = %v", err)
	}
	if !resp.PrepareMeta.Available || resp.PrepareMeta.SchemaVersion != "session-report-prepare-v1" {
		t.Fatalf("PrepareMeta = %+v, want available session-report-prepare-v1", resp.PrepareMeta)
	}
	if !resp.Readiness.CanReadReport || resp.Readiness.NextAction != "open_report" {
		t.Fatalf("Readiness = %+v, want readable/open_report", resp.Readiness)
	}
	if resp.ReportSummary == nil || resp.ReportSummary.Assets.ResumeChunks != 2 {
		t.Fatalf("ReportSummary = %+v, want resume chunks", resp.ReportSummary)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestSessionReportPrepareNotFound(t *testing.T) {
	logic := NewSessionReportPrepareLogic(withUserID(context.Background(), 7), &svc.ServiceContext{
		ChatSessionsModel: &stubChatSessionsModel{err: model.ErrNotFound},
	})

	_, err := logic.SessionReportPrepare(&types.SessionReportPrepareReq{Id: "missing"})
	if err == nil {
		t.Fatal("SessionReportPrepare() error = nil, want not found")
	}
	code, ok := statuserr.StatusCode(err)
	if !ok || code != http.StatusNotFound {
		t.Fatalf("status code = %d, ok=%v, want %d/true", code, ok, http.StatusNotFound)
	}
}
