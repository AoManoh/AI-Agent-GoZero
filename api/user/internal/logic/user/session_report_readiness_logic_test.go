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

func TestSessionReportReadinessFresh(t *testing.T) {
	now := time.Date(2026, 5, 9, 14, 0, 0, 0, time.UTC)
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

	logic := NewSessionReportReadinessLogic(withUserID(context.Background(), session.UserId), &svc.ServiceContext{
		DB:                      sqlx.NewSqlConnFromDB(db),
		ChatSessionsModel:       &stubChatSessionsModel{session: session},
		SessionEvaluationsModel: &stubSessionEvaluationsModel{record: record},
	})

	resp, err := logic.SessionReportReadiness(&types.SessionReportReadinessReq{Id: session.SessionId})
	if err != nil {
		t.Fatalf("SessionReportReadiness() error = %v", err)
	}
	if !resp.CanReadReport || resp.NeedsRefresh || resp.NextAction != "open_report" {
		t.Fatalf("readiness = %+v, want readable/open_report", resp)
	}
	if resp.LastMessageAt == "" || resp.LastRefreshedAt == "" {
		t.Fatalf("readiness missing timestamps: %+v", resp)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestSessionReportReadinessMissingEvaluation(t *testing.T) {
	now := time.Date(2026, 5, 9, 14, 0, 0, 0, time.UTC)
	session := reportReadinessSession(now)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`select id as last_message_id, created_at as last_message_at`).
		WithArgs(session.SessionId, session.UserId).
		WillReturnRows(sqlmock.NewRows([]string{"last_message_id", "last_message_at"}).AddRow(int64(42), now.Add(-30*time.Minute)))

	logic := NewSessionReportReadinessLogic(withUserID(context.Background(), session.UserId), &svc.ServiceContext{
		DB:                      sqlx.NewSqlConnFromDB(db),
		ChatSessionsModel:       &stubChatSessionsModel{session: session},
		SessionEvaluationsModel: &stubSessionEvaluationsModel{err: model.ErrNotFound},
	})

	resp, err := logic.SessionReportReadiness(&types.SessionReportReadinessReq{Id: session.SessionId})
	if err != nil {
		t.Fatalf("SessionReportReadiness() error = %v", err)
	}
	if resp.CanReadReport || !resp.NeedsRefresh || resp.ReportStatus != "missing" || resp.NextAction != "refresh_evaluation" {
		t.Fatalf("readiness = %+v, want missing/refresh_evaluation", resp)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestSessionReportReadinessStaleEvaluation(t *testing.T) {
	now := time.Date(2026, 5, 9, 14, 0, 0, 0, time.UTC)
	session := reportReadinessSession(now)
	record := completeEvaluationRecordForTest(t, session.SessionId, now)
	record.SourceLastMessageID.Valid = true
	record.SourceLastMessageID.Int64 = 40

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`select id as last_message_id, created_at as last_message_at`).
		WithArgs(session.SessionId, session.UserId).
		WillReturnRows(sqlmock.NewRows([]string{"last_message_id", "last_message_at"}).AddRow(int64(42), now.Add(-30*time.Minute)))

	logic := NewSessionReportReadinessLogic(withUserID(context.Background(), session.UserId), &svc.ServiceContext{
		DB:                      sqlx.NewSqlConnFromDB(db),
		ChatSessionsModel:       &stubChatSessionsModel{session: session},
		SessionEvaluationsModel: &stubSessionEvaluationsModel{record: record},
	})

	resp, err := logic.SessionReportReadiness(&types.SessionReportReadinessReq{Id: session.SessionId})
	if err != nil {
		t.Fatalf("SessionReportReadiness() error = %v", err)
	}
	if resp.CanReadReport || !resp.NeedsRefresh || resp.ReportStatus != "stale" || resp.NextAction != "refresh_evaluation" {
		t.Fatalf("readiness = %+v, want stale/refresh_evaluation", resp)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestSessionReportReadinessNotFound(t *testing.T) {
	logic := NewSessionReportReadinessLogic(withUserID(context.Background(), 7), &svc.ServiceContext{
		ChatSessionsModel: &stubChatSessionsModel{err: model.ErrNotFound},
	})

	_, err := logic.SessionReportReadiness(&types.SessionReportReadinessReq{Id: "missing"})
	if err == nil {
		t.Fatal("SessionReportReadiness() error = nil, want not found")
	}
	code, ok := statuserr.StatusCode(err)
	if !ok || code != http.StatusNotFound {
		t.Fatalf("status code = %d, ok=%v, want %d/true", code, ok, http.StatusNotFound)
	}
}

func reportReadinessSession(now time.Time) *model.ChatSession {
	return &model.ChatSession{
		SessionId:    "sess-readiness",
		UserId:       7,
		Title:        "Go 后端面试",
		Mode:         "Interview",
		IsActive:     true,
		MessageCount: 4,
		CreatedAt:    now.Add(-time.Hour),
		UpdatedAt:    now,
	}
}
