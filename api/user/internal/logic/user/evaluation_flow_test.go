package user

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"GoZero-AI/api/user/internal/evaluation"
	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/api/user/internal/types"
	"GoZero-AI/api/user/model"
	"GoZero-AI/internal/statuserr"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func TestSessionEvaluationFreshCacheIsPureRead(t *testing.T) {
	now := time.Date(2026, 4, 16, 21, 0, 0, 0, time.UTC)
	session := &model.ChatSession{
		SessionId:    "sess-fresh",
		UserId:       7,
		Title:        "fresh",
		Mode:         "Interview",
		IsActive:     true,
		MessageCount: 2,
		CreatedAt:    now.Add(-2 * time.Hour),
		UpdatedAt:    now.Add(-1 * time.Hour),
		LastMessageAt: sql.NullTime{
			Time:  now.Add(-40 * time.Minute),
			Valid: true,
		},
	}
	record := completeEvaluationRecordForTest(t, session.SessionId, now)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`select id as last_message_id, created_at as last_message_at`).
		WithArgs(session.SessionId, session.UserId).
		WillReturnRows(sqlmock.NewRows([]string{"last_message_id", "last_message_at"}).AddRow(int64(41), now.Add(-50*time.Minute)))

	logic := NewSessionEvaluationLogic(withUserID(context.Background(), session.UserId), &svc.ServiceContext{
		DB:                      sqlx.NewSqlConnFromDB(db),
		ChatSessionsModel:       &stubChatSessionsModel{session: session},
		SessionEvaluationsModel: &stubSessionEvaluationsModel{record: record},
	})

	resp, err := logic.SessionEvaluation(&types.SessionEvaluationReq{Id: session.SessionId})
	if err != nil {
		t.Fatalf("SessionEvaluation() error = %v", err)
	}
	if resp.Status != record.Status {
		t.Fatalf("resp.Status = %q, want %q", resp.Status, record.Status)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestSessionReportSummaryFreshCacheIsPureRead(t *testing.T) {
	now := time.Date(2026, 4, 16, 21, 0, 0, 0, time.UTC)
	session := &model.ChatSession{
		SessionId:    "sess-report",
		UserId:       7,
		Title:        "report",
		Mode:         "Interview",
		IsActive:     true,
		MessageCount: 2,
		CreatedAt:    now.Add(-2 * time.Hour),
		UpdatedAt:    now.Add(-1 * time.Hour),
		LastMessageAt: sql.NullTime{
			Time:  now.Add(-40 * time.Minute),
			Valid: true,
		},
	}
	record := completeEvaluationRecordForTest(t, session.SessionId, now)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`select id as last_message_id, created_at as last_message_at`).
		WithArgs(session.SessionId, session.UserId).
		WillReturnRows(sqlmock.NewRows([]string{"last_message_id", "last_message_at"}).AddRow(int64(41), now.Add(-50*time.Minute)))
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
		WillReturnRows(sqlmock.NewRows([]string{"resume_chunks", "resume_updated_at"}).AddRow(1, now.Add(-30*time.Minute)))

	logic := NewSessionReportSummaryLogic(withUserID(context.Background(), session.UserId), &svc.ServiceContext{
		DB:                      sqlx.NewSqlConnFromDB(db),
		ChatSessionsModel:       &stubChatSessionsModel{session: session},
		SessionEvaluationsModel: &stubSessionEvaluationsModel{record: record},
	})

	resp, err := logic.SessionReportSummary(&types.SessionReportSummaryReq{Id: session.SessionId})
	if err != nil {
		t.Fatalf("SessionReportSummary() error = %v", err)
	}
	if resp.Evaluation.Status != record.Status {
		t.Fatalf("resp.Evaluation.Status = %q, want %q", resp.Evaluation.Status, record.Status)
	}
	if resp.Conversation.LatestUserMessage == "" || resp.Conversation.LatestAssistantMessage == "" {
		t.Fatalf("conversation summary missing latest messages: %+v", resp.Conversation)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestSessionReportSummaryStaleCacheReturnsUnavailable(t *testing.T) {
	now := time.Date(2026, 4, 16, 21, 0, 0, 0, time.UTC)
	session := &model.ChatSession{
		SessionId: "sess-report-stale",
		UserId:    7,
		Title:     "report stale",
		Mode:      "Interview",
		IsActive:  true,
		LastMessageAt: sql.NullTime{
			Time:  now.Add(-5 * time.Minute),
			Valid: true,
		},
	}
	record := completeEvaluationRecordForTest(t, session.SessionId, now)
	record.SourceLastMessageID = sql.NullInt64{Int64: 40, Valid: true}
	record.SourceLastMessageAt = sql.NullTime{Time: now.Add(-30 * time.Minute), Valid: true}

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`select id as last_message_id, created_at as last_message_at`).
		WithArgs(session.SessionId, session.UserId).
		WillReturnRows(sqlmock.NewRows([]string{"last_message_id", "last_message_at"}).AddRow(int64(43), now.Add(-5*time.Minute)))

	logic := NewSessionReportSummaryLogic(withUserID(context.Background(), session.UserId), &svc.ServiceContext{
		DB:                      sqlx.NewSqlConnFromDB(db),
		ChatSessionsModel:       &stubChatSessionsModel{session: session},
		SessionEvaluationsModel: &stubSessionEvaluationsModel{record: record},
	})

	_, err = logic.SessionReportSummary(&types.SessionReportSummaryReq{Id: session.SessionId})
	if err == nil {
		t.Fatal("SessionReportSummary() error = nil, want unavailable error")
	}
	code, ok := statuserr.StatusCode(err)
	if !ok || code != 503 {
		t.Fatalf("status code = %d, ok=%v, want 503/true", code, ok)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestLoadLatestEvaluationMessageWatermarkUsesLatestRow(t *testing.T) {
	now := time.Date(2026, 4, 16, 21, 0, 0, 0, time.UTC)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`select id as last_message_id, created_at as last_message_at`).
		WithArgs("sess-watermark", int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"last_message_id", "last_message_at"}).AddRow(int64(43), now))

	watermark, err := loadLatestEvaluationMessageWatermarkWithReader(context.Background(), sqlx.NewSqlConnFromDB(db), "sess-watermark", 7)
	if err != nil {
		t.Fatalf("loadLatestEvaluationMessageWatermarkWithReader() error = %v", err)
	}
	if !watermark.LastMessageID.Valid || watermark.LastMessageID.Int64 != 43 {
		t.Fatalf("LastMessageID = %+v, want 43", watermark.LastMessageID)
	}
	if !watermark.LastMessageAt.Valid || !watermark.LastMessageAt.Time.Equal(now) {
		t.Fatalf("LastMessageAt = %+v, want %v", watermark.LastMessageAt, now)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

type stubChatSessionsModel struct {
	session *model.ChatSession
	err     error
}

func (s *stubChatSessionsModel) Create(ctx context.Context, userID int64, sessionID, title, mode string) (*model.ChatSession, error) {
	panic("unexpected Create call")
}

func (s *stubChatSessionsModel) CreateWithConfig(ctx context.Context, userID int64, sessionID, title, mode string, config model.SessionCreateConfig) (*model.ChatSession, error) {
	panic("unexpected CreateWithConfig call")
}

func (s *stubChatSessionsModel) Complete(ctx context.Context, userID int64, sessionID string) (*model.ChatSession, error) {
	panic("unexpected Complete call")
}

func (s *stubChatSessionsModel) Deactivate(ctx context.Context, userID int64, sessionID string) error {
	panic("unexpected Deactivate call")
}

func (s *stubChatSessionsModel) FindByUserID(ctx context.Context, userID int64) ([]model.ChatSession, error) {
	panic("unexpected FindByUserID call")
}

func (s *stubChatSessionsModel) FindOneBySessionID(ctx context.Context, userID int64, sessionID string) (*model.ChatSession, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.session, nil
}

type stubSessionEvaluationsModel struct {
	record      *model.SessionEvaluation
	err         error
	records     []*model.SessionEvaluation
	errors      []error
	findCalls   int
	upsertCalls int
}

func (s *stubSessionEvaluationsModel) FindOneBySessionID(ctx context.Context, userID int64, sessionID string) (*model.SessionEvaluation, error) {
	if len(s.records) > 0 || len(s.errors) > 0 {
		index := s.findCalls
		s.findCalls++
		if index < len(s.errors) && s.errors[index] != nil {
			return nil, s.errors[index]
		}
		if index < len(s.records) {
			return s.records[index], nil
		}
		if len(s.errors) > 0 && s.errors[len(s.errors)-1] != nil {
			return nil, s.errors[len(s.errors)-1]
		}
		if len(s.records) > 0 {
			return s.records[len(s.records)-1], nil
		}
	}
	if s.err != nil {
		return nil, s.err
	}
	return s.record, nil
}

func (s *stubSessionEvaluationsModel) Upsert(ctx context.Context, data *model.SessionEvaluation) error {
	s.upsertCalls++
	s.record = data
	return s.err
}

func withUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, "userId", userID)
}

func completeEvaluationRecordForTest(t *testing.T, sessionID string, now time.Time) *model.SessionEvaluation {
	t.Helper()

	return &model.SessionEvaluation{
		SessionId:      sessionID,
		UserId:         7,
		Status:         "ready",
		Summary:        "summary",
		UserTurns:      2,
		AssistantTurns: 2,
		OverallScore:   88,
		RubricVersion:  evaluation.RubricVersion,
		ScoreSource:    "llm",
		Dimensions:     mustJSON(t, []types.EvaluationDimension{{Key: "technical_depth", Label: "技术深度", Score: 4, MaxScore: 5, Summary: "ok"}, {Key: "engineering_practice", Label: "工程实践", Score: 4, MaxScore: 5, Summary: "ok"}, {Key: "architecture_sense", Label: "架构意识", Score: 4, MaxScore: 5, Summary: "ok"}, {Key: "communication", Label: "表达与沟通", Score: 4, MaxScore: 5, Summary: "ok"}}),
		Strengths:      mustJSON(t, []string{"strength"}),
		Risks:          mustJSON(t, []string{"risk"}),
		Suggestions:    mustJSON(t, []string{"next"}),
		Evidence:       mustJSON(t, []types.EvaluationEvidence{{Role: "user", Content: "hello"}}),
		SourceLastMessageID: sql.NullInt64{
			Int64: 42,
			Valid: true,
		},
		SourceLastMessageAt: sql.NullTime{
			Time:  now.Add(-30 * time.Minute),
			Valid: true,
		},
		FirstGeneratedAt: now.Add(-2 * time.Hour),
		GeneratedAt:      now.Add(-1 * time.Hour),
		UpdatedAt:        now.Add(-30 * time.Minute),
	}
}

func TestSessionEvaluationStaleCacheIsReadOnlyUnavailable(t *testing.T) {
	now := time.Date(2026, 4, 16, 21, 0, 0, 0, time.UTC)
	session := &model.ChatSession{
		SessionId: "sess-stale-read",
		UserId:    7,
		Title:     "stale read",
		Mode:      "Interview",
		IsActive:  true,
		LastMessageAt: sql.NullTime{
			Time:  now.Add(-5 * time.Minute),
			Valid: true,
		},
	}
	staleRecord := completeEvaluationRecordForTest(t, session.SessionId, now)
	staleRecord.SourceLastMessageID = sql.NullInt64{Int64: 40, Valid: true}
	staleRecord.SourceLastMessageAt = sql.NullTime{Time: now.Add(-30 * time.Minute), Valid: true}

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`select id as last_message_id, created_at as last_message_at`).
		WithArgs(session.SessionId, session.UserId).
		WillReturnRows(sqlmock.NewRows([]string{"last_message_id", "last_message_at"}).AddRow(int64(43), now.Add(-5*time.Minute)))

	evalModel := &stubSessionEvaluationsModel{record: staleRecord}
	logic := NewSessionEvaluationLogic(withUserID(context.Background(), session.UserId), &svc.ServiceContext{
		DB:                      sqlx.NewSqlConnFromDB(db),
		ChatSessionsModel:       &stubChatSessionsModel{session: session},
		SessionEvaluationsModel: evalModel,
	})

	_, err = logic.SessionEvaluation(&types.SessionEvaluationReq{Id: session.SessionId})
	if err == nil {
		t.Fatal("SessionEvaluation() error = nil, want unavailable error")
	}
	code, ok := statuserr.StatusCode(err)
	if !ok || code != 503 {
		t.Fatalf("status code = %d, ok=%v, want 503/true", code, ok)
	}
	if evalModel.upsertCalls != 0 {
		t.Fatalf("upsertCalls = %d, want 0", evalModel.upsertCalls)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestSessionEvaluationRefreshStaleCacheRefreshesAndPersists(t *testing.T) {
	now := time.Date(2026, 4, 16, 21, 0, 0, 0, time.UTC)
	session := &model.ChatSession{
		SessionId: "sess-stale",
		UserId:    7,
		Title:     "stale",
		Mode:      "Interview",
		IsActive:  true,
		LastMessageAt: sql.NullTime{
			Time:  now.Add(-5 * time.Minute),
			Valid: true,
		},
	}
	staleRecord := completeEvaluationRecordForTest(t, session.SessionId, now)
	staleRecord.SourceLastMessageID = sql.NullInt64{Int64: 40, Valid: true}
	staleRecord.SourceLastMessageAt = sql.NullTime{Time: now.Add(-30 * time.Minute), Valid: true}

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`select id as last_message_id, created_at as last_message_at`).
		WithArgs(session.SessionId, session.UserId).
		WillReturnRows(sqlmock.NewRows([]string{"last_message_id", "last_message_at"}).AddRow(int64(43), now.Add(-5*time.Minute)))
	mock.ExpectQuery(`select pg_advisory_lock`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"pg_advisory_lock"}).AddRow(nil))
	mock.ExpectQuery(`select id as last_message_id, created_at as last_message_at`).
		WithArgs(session.SessionId, session.UserId).
		WillReturnRows(sqlmock.NewRows([]string{"last_message_id", "last_message_at"}).AddRow(int64(43), now.Add(-5*time.Minute)))
	mock.ExpectQuery(`select id, role, content, created_at`).
		WithArgs(session.SessionId, session.UserId, int64(43)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "role", "content", "created_at"}).
			AddRow(int64(41), "user", "我负责 GoZero 项目后端开发和 PostgreSQL 设计", now.Add(-20*time.Minute)).
			AddRow(int64(42), "assistant", "请继续说说你做过的并发优化", now.Add(-19*time.Minute)).
			AddRow(int64(43), "user", "我做过索引优化和服务拆分", now.Add(-18*time.Minute)))
	mock.ExpectBegin()
	mock.ExpectExec(`delete from "public"."session_evaluation_items"`).
		WithArgs(session.UserId, session.SessionId).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`insert into "public"."session_evaluation_items"`).
		WithArgs(session.SessionId, session.UserId, int64(1), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), int64(41), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`insert into "public"."session_evaluation_items"`).
		WithArgs(session.SessionId, session.UserId, int64(2), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), int64(43), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectCommit()
	mock.ExpectQuery(`select pg_advisory_unlock`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"pg_advisory_unlock"}).AddRow(true))

	evalModel := &stubSessionEvaluationsModel{
		records: []*model.SessionEvaluation{staleRecord, staleRecord},
		errors:  []error{nil, nil},
	}

	logic := NewSessionEvaluationLogic(withUserID(context.Background(), session.UserId), &svc.ServiceContext{
		DB:                      sqlx.NewSqlConnFromDB(db),
		ChatSessionsModel:       &stubChatSessionsModel{session: session},
		SessionEvaluationsModel: evalModel,
	})

	resp, err := logic.SessionEvaluationRefresh(&types.SessionEvaluationRefreshReq{Id: session.SessionId})
	if err != nil {
		t.Fatalf("SessionEvaluationRefresh() error = %v", err)
	}
	if resp.ScoreSource == "" {
		t.Fatalf("resp.ScoreSource should not be empty")
	}
	if evalModel.upsertCalls != 1 {
		t.Fatalf("upsertCalls = %d, want 1", evalModel.upsertCalls)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}
