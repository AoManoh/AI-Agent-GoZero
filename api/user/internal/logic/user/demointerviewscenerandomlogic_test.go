package user

import (
	"context"
	"testing"
	"time"

	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/api/user/internal/types"
	"GoZero-AI/internal/sessionmode"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sashabaranov/go-openai"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func TestSelectDemoInterviewSceneWindowPrefersAssistantUserAssistant(t *testing.T) {
	now := time.Date(2026, 5, 8, 15, 0, 0, 0, time.UTC)
	rows := []demoInterviewSceneMessageRow{
		{Role: openai.ChatMessageRoleUser, Content: "我熟悉 Go runtime。", CreatedAt: now},
		{Role: openai.ChatMessageRoleAssistant, Content: "请解释 goroutine 调度器。", CreatedAt: now.Add(time.Minute)},
		{Role: openai.ChatMessageRoleUser, Content: "GMP 模型会把 G 映射到 M。", CreatedAt: now.Add(2 * time.Minute)},
		{Role: openai.ChatMessageRoleAssistant, Content: "那阻塞 syscall 怎么处理？", CreatedAt: now.Add(3 * time.Minute)},
	}

	messages := buildDemoInterviewSceneMessages(rows, 3)
	if len(messages) != 3 {
		t.Fatalf("len(messages) = %d, want 3", len(messages))
	}
	if messages[0].Role != openai.ChatMessageRoleAssistant || messages[0].Name != "AI 面试官" {
		t.Fatalf("first message = %+v, want assistant scene opener", messages[0])
	}
	if messages[1].Role != openai.ChatMessageRoleUser || messages[1].Name != "你" {
		t.Fatalf("second message = %+v, want user answer", messages[1])
	}
	if messages[2].Role != openai.ChatMessageRoleAssistant || messages[2].Name != "AI · 追问 #1" {
		t.Fatalf("third message = %+v, want follow-up assistant", messages[2])
	}
}

func TestDemoInterviewSceneRandomReturnsEmptyWhenNoDemoSession(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`select session_id, title, mode`).
		WithArgs(demoInterviewSceneUserID, sessionmode.KeyInterview, 4, openai.ChatMessageRoleAssistant, openai.ChatMessageRoleUser).
		WillReturnRows(sqlmock.NewRows([]string{"session_id", "title", "mode"}))

	logic := NewDemoInterviewSceneRandomLogic(context.Background(), &svc.ServiceContext{
		DB: sqlx.NewSqlConnFromDB(db),
	})
	resp, err := logic.DemoInterviewSceneRandom(&types.DemoInterviewSceneRandomReq{Limit: 3})
	if err != nil {
		t.Fatalf("DemoInterviewSceneRandom() error = %v", err)
	}
	if resp.Available {
		t.Fatalf("resp.Available = true, want false")
	}
	if len(resp.Messages) != 0 {
		t.Fatalf("len(resp.Messages) = %d, want 0", len(resp.Messages))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestDemoInterviewSceneRandomBuildsDemoScene(t *testing.T) {
	now := time.Date(2026, 5, 8, 15, 0, 0, 0, time.UTC)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`select session_id, title, mode`).
		WithArgs(demoInterviewSceneUserID, sessionmode.KeyInterview, 4, openai.ChatMessageRoleAssistant, openai.ChatMessageRoleUser).
		WillReturnRows(sqlmock.NewRows([]string{"session_id", "title", "mode"}).
			AddRow("sess-demo", "Go 后端面试", sessionmode.KeyInterview))
	mock.ExpectQuery(`select role, content, created_at`).
		WithArgs("sess-demo", demoInterviewSceneUserID, openai.ChatMessageRoleUser, openai.ChatMessageRoleAssistant, maxDemoInterviewSceneScanSize).
		WillReturnRows(sqlmock.NewRows([]string{"role", "content", "created_at"}).
			AddRow(openai.ChatMessageRoleUser, "我熟悉 Go runtime。", now).
			AddRow(openai.ChatMessageRoleAssistant, "请解释 goroutine 调度器。", now.Add(time.Minute)).
			AddRow(openai.ChatMessageRoleUser, "GMP 模型会把 G 映射到 M。", now.Add(2*time.Minute)).
			AddRow(openai.ChatMessageRoleAssistant, "那阻塞 syscall 怎么处理？", now.Add(3*time.Minute)))

	logic := NewDemoInterviewSceneRandomLogic(context.Background(), &svc.ServiceContext{
		DB: sqlx.NewSqlConnFromDB(db),
	})
	resp, err := logic.DemoInterviewSceneRandom(&types.DemoInterviewSceneRandomReq{Limit: 3})
	if err != nil {
		t.Fatalf("DemoInterviewSceneRandom() error = %v", err)
	}
	if !resp.Available {
		t.Fatalf("resp.Available = false, want true")
	}
	if resp.SessionId != "sess-demo" {
		t.Fatalf("resp.SessionId = %q, want sess-demo", resp.SessionId)
	}
	if resp.Mode != sessionmode.LabelInterview || resp.ModeKey != sessionmode.KeyInterview {
		t.Fatalf("mode = %q/%q, want %q/%q", resp.Mode, resp.ModeKey, sessionmode.LabelInterview, sessionmode.KeyInterview)
	}
	if len(resp.Messages) != 3 {
		t.Fatalf("len(resp.Messages) = %d, want 3", len(resp.Messages))
	}
	if resp.Messages[0].Role != openai.ChatMessageRoleAssistant {
		t.Fatalf("first role = %q, want assistant", resp.Messages[0].Role)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
