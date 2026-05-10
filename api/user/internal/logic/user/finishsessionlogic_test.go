package user

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/api/user/internal/types"
	"GoZero-AI/api/user/model"
	"GoZero-AI/internal/chatflow"
	"GoZero-AI/internal/sessionmode"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func TestFinishSessionSyncsCompletedFlowState(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 5, 10, 10, 0, 0, 0, time.UTC)
	mock.ExpectQuery(regexp.QuoteMeta(`update "public"."chat_sessions"`)).
		WithArgs(int64(7), "sess-finish").
		WillReturnRows(sqlmock.NewRows(chatSessionModelColumns()).AddRow(
			int64(1),
			"sess-finish",
			int64(7),
			"Go 后端面试",
			sessionmode.KeyInterview,
			"go_backend",
			"Go 后端",
			int64(3),
			"中级",
			"senior",
			"资深技术官",
			[]byte("[]"),
			"N+3",
			int64(30),
			int64(100),
			now,
			now,
			sql.NullTime{},
			sql.NullTime{Time: now.Add(-20 * time.Minute), Valid: true},
			sql.NullTime{Time: now, Valid: true},
			int64(1200),
			int64(8),
			true,
		))

	conn := sqlx.NewSqlConnFromDB(db)
	redisClient := newMiniRedisClient(t)
	defer redisClient.Close()

	ctx := context.WithValue(context.Background(), "userId", int64(7))
	logic := NewFinishSessionLogic(ctx, &svc.ServiceContext{
		ChatSessionsModel: model.NewChatSessionsModel(conn),
		RedisClient:       redisClient,
	})
	resp, err := logic.FinishSession(&types.FinishSessionReq{Id: "sess-finish"})
	if err != nil {
		t.Fatalf("FinishSession() error = %v", err)
	}
	if resp.Session.IsActive {
		t.Fatal("resp.Session.IsActive = true, want false")
	}

	key := chatflow.BuildContextKey("sess-finish", ptrInt64(7), sessionmode.KeyInterview)
	snapshot, err := chatflow.LoadSnapshot(context.Background(), redisClient, key, chatflow.InterviewStateStart)
	if err != nil {
		t.Fatalf("LoadSnapshot() error = %v", err)
	}
	if snapshot.InterviewState != chatflow.InterviewStateEnd ||
		snapshot.LifecycleState != chatflow.LifecycleCompleted ||
		snapshot.ExecutionState != chatflow.ExecutionIdle ||
		snapshot.LastReason != "session_finished" {
		t.Fatalf("snapshot = %+v, want completed/end/idle/session_finished", snapshot)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func chatSessionModelColumns() []string {
	return []string{
		"id", "session_id", "user_id", "title", "mode",
		"direction_key", "direction_label", "difficulty_level", "difficulty_label",
		"interviewer_style", "interviewer_style_label", "focus_areas", "follow_up_depth",
		"estimated_minutes", "progress_percent", "created_at", "updated_at", "last_message_at",
		"started_at", "completed_at", "duration_seconds", "message_count", "is_active",
	}
}

func ptrInt64(value int64) *int64 {
	return &value
}
