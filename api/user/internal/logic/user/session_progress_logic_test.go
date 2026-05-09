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

func TestSessionProgressPartial(t *testing.T) {
	now := time.Date(2026, 5, 9, 18, 0, 0, 0, time.UTC)
	session := progressSession(now)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`select\s+coalesce\(sum\(case when role = 'user' then 1 else 0 end\), 0\) as user_turns`).
		WithArgs(session.SessionId, session.UserId).
		WillReturnRows(sqlmock.NewRows([]string{"user_turns", "assistant_turns"}).AddRow(int64(2), int64(2)))

	logic := NewSessionProgressLogic(withUserID(context.Background(), session.UserId), &svc.ServiceContext{
		DB:                sqlx.NewSqlConnFromDB(db),
		ChatSessionsModel: &stubChatSessionsModel{session: session},
	})

	resp, err := logic.SessionProgress(&types.SessionProgressReq{Id: session.SessionId, PlanLimit: 4})
	if err != nil {
		t.Fatalf("SessionProgress() error = %v", err)
	}
	if resp.TotalQuestions != 4 || resp.CompletedQuestions != 2 || resp.CurrentQuestionIndex != 3 || resp.ProgressPercent != 50 {
		t.Fatalf("progress = %+v, want total=4 completed=2 current=3 percent=50", resp)
	}
	if resp.NextQuestion == nil {
		t.Fatal("NextQuestion = nil, want next planned question")
	}
	if len(resp.FocusProgress) == 0 {
		t.Fatal("FocusProgress is empty")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestSessionProgressCompletedSession(t *testing.T) {
	now := time.Date(2026, 5, 9, 18, 0, 0, 0, time.UTC)
	session := progressSession(now)
	session.ProgressPercent = 100
	session.CompletedAt = sql.NullTime{Time: now, Valid: true}

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`select\s+coalesce\(sum\(case when role = 'user' then 1 else 0 end\), 0\) as user_turns`).
		WithArgs(session.SessionId, session.UserId).
		WillReturnRows(sqlmock.NewRows([]string{"user_turns", "assistant_turns"}).AddRow(int64(1), int64(1)))

	logic := NewSessionProgressLogic(withUserID(context.Background(), session.UserId), &svc.ServiceContext{
		DB:                sqlx.NewSqlConnFromDB(db),
		ChatSessionsModel: &stubChatSessionsModel{session: session},
	})

	resp, err := logic.SessionProgress(&types.SessionProgressReq{Id: session.SessionId, PlanLimit: 4})
	if err != nil {
		t.Fatalf("SessionProgress() error = %v", err)
	}
	if resp.ProgressPercent != 100 || resp.CompletedQuestions != resp.TotalQuestions || resp.NextQuestion != nil {
		t.Fatalf("progress = %+v, want completed session with no next question", resp)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestSessionProgressNotFound(t *testing.T) {
	logic := NewSessionProgressLogic(withUserID(context.Background(), 7), &svc.ServiceContext{
		ChatSessionsModel: &stubChatSessionsModel{err: model.ErrNotFound},
	})

	_, err := logic.SessionProgress(&types.SessionProgressReq{Id: "missing"})
	if err == nil {
		t.Fatal("SessionProgress() error = nil, want not found")
	}
	code, ok := statuserr.StatusCode(err)
	if !ok || code != http.StatusNotFound {
		t.Fatalf("status code = %d, ok=%v, want %d/true", code, ok, http.StatusNotFound)
	}
}

func progressSession(now time.Time) *model.ChatSession {
	return &model.ChatSession{
		SessionId:             "sess-progress",
		UserId:                7,
		Title:                 "Go 后端面试",
		Mode:                  "Interview",
		DirectionKey:          "go_backend",
		DirectionLabel:        "Go 后端",
		DifficultyLevel:       4,
		DifficultyLabel:       "资深",
		InterviewerStyle:      "senior",
		InterviewerStyleLabel: "资深技术官",
		FocusAreas:            []byte(`[{"key":"concurrency","label":"并发与调度"},{"key":"database","label":"数据库"}]`),
		FollowUpDepth:         "N+5",
		EstimatedMinutes:      45,
		ProgressPercent:       0,
		CreatedAt:             now.Add(-time.Hour),
		UpdatedAt:             now,
		StartedAt:             sql.NullTime{Time: now.Add(-time.Hour), Valid: true},
		IsActive:              true,
	}
}
