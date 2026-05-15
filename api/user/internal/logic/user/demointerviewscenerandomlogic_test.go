package user

import (
	"context"
	"testing"
	"time"

	"GoZero-AI/api/user/internal/auth"
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

	mock.ExpectQuery(`select\s+session_id`).
		WithArgs(demoInterviewSceneUserID, sessionmode.KeyInterview, 4, openai.ChatMessageRoleAssistant, openai.ChatMessageRoleUser).
		WillReturnRows(sqlmock.NewRows([]string{
			"session_id",
			"title",
			"mode",
			"direction_label",
			"difficulty_label",
			"interviewer_style_label",
			"follow_up_depth",
		}))

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

	mock.ExpectQuery(`select\s+session_id`).
		WithArgs(demoInterviewSceneUserID, sessionmode.KeyInterview, 4, openai.ChatMessageRoleAssistant, openai.ChatMessageRoleUser).
		WillReturnRows(sqlmock.NewRows([]string{
			"session_id",
			"title",
			"mode",
			"direction_label",
			"difficulty_label",
			"interviewer_style_label",
			"follow_up_depth",
		}).AddRow("sess-demo", "Go 后端面试", sessionmode.KeyInterview, "Go 后端", "资深", "压力面试官", "N+5"))
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
	if resp.Source != demoInterviewSceneSourceAdmin || resp.SourceLabel != "管理员演示" {
		t.Fatalf("source = %q/%q, want admin/管理员演示", resp.Source, resp.SourceLabel)
	}
	if resp.DirectionLabel != "Go 后端" || resp.DifficultyLabel != "资深" || resp.InterviewerStyleLabel != "压力面试官" || resp.FollowUpDepth != "N+5" {
		t.Fatalf("demo facts = %q/%q/%q/%q, want Go 后端/资深/压力面试官/N+5",
			resp.DirectionLabel, resp.DifficultyLabel, resp.InterviewerStyleLabel, resp.FollowUpDepth)
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

func TestDemoInterviewSceneRandomPrefersCurrentUserScene(t *testing.T) {
	now := time.Date(2026, 5, 10, 10, 0, 0, 0, time.UTC)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`select\s+session_id`).
		WithArgs(int64(7), sessionmode.KeyInterview, 4, openai.ChatMessageRoleAssistant, openai.ChatMessageRoleUser).
		WillReturnRows(sqlmock.NewRows(demoInterviewSceneSessionColumns()).
			AddRow("sess-user", "我的 Go 面试", sessionmode.KeyInterview, "Go 后端", "中级", "资深技术官", "N+3"))
	mock.ExpectQuery(`select role, content, created_at`).
		WithArgs("sess-user", int64(7), openai.ChatMessageRoleUser, openai.ChatMessageRoleAssistant, maxDemoInterviewSceneScanSize).
		WillReturnRows(sqlmock.NewRows([]string{"role", "content", "created_at"}).
			AddRow(openai.ChatMessageRoleAssistant, "讲讲你的 Go 项目。", now).
			AddRow(openai.ChatMessageRoleUser, "我做过 GoZero 微服务。", now.Add(time.Minute)).
			AddRow(openai.ChatMessageRoleAssistant, "服务发现怎么做？", now.Add(2*time.Minute)))

	logic := NewDemoInterviewSceneRandomLogic(withDemoAccessToken(t, context.Background(), 7), demoInterviewSceneSvcCtx(sqlx.NewSqlConnFromDB(db)))
	resp, err := logic.DemoInterviewSceneRandom(&types.DemoInterviewSceneRandomReq{Limit: 3})
	if err != nil {
		t.Fatalf("DemoInterviewSceneRandom() error = %v", err)
	}
	if !resp.Available || resp.SessionId != "sess-user" {
		t.Fatalf("resp = %+v, want current user scene", resp)
	}
	if resp.Source != demoInterviewSceneSourceUser || resp.SourceLabel != "我的面试记录" {
		t.Fatalf("source = %q/%q, want user/我的面试记录", resp.Source, resp.SourceLabel)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestDemoInterviewSceneRandomReturnsEmptyWhenCurrentUserHasNoScene(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`select\s+session_id`).
		WithArgs(int64(7), sessionmode.KeyInterview, 4, openai.ChatMessageRoleAssistant, openai.ChatMessageRoleUser).
		WillReturnRows(sqlmock.NewRows(demoInterviewSceneSessionColumns()))

	logic := NewDemoInterviewSceneRandomLogic(withDemoAccessToken(t, context.Background(), 7), demoInterviewSceneSvcCtx(sqlx.NewSqlConnFromDB(db)))
	resp, err := logic.DemoInterviewSceneRandom(&types.DemoInterviewSceneRandomReq{Limit: 3})
	if err != nil {
		t.Fatalf("DemoInterviewSceneRandom() error = %v", err)
	}
	if resp.Available || len(resp.Messages) != 0 {
		t.Fatalf("resp = %+v, want empty response for frontend fallback", resp)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func demoInterviewSceneSessionColumns() []string {
	return []string{
		"session_id",
		"title",
		"mode",
		"direction_label",
		"difficulty_label",
		"interviewer_style_label",
		"follow_up_depth",
	}
}

func withDemoAccessToken(t *testing.T, ctx context.Context, userID int64) context.Context {
	t.Helper()

	pair, err := auth.IssueTokenPair("demo-secret", time.Hour, 24*time.Hour, userID, "demo-user")
	if err != nil {
		t.Fatalf("IssueTokenPair() error = %v", err)
	}
	return WithAccessToken(ctx, pair.AccessToken)
}

func demoInterviewSceneSvcCtx(conn sqlx.SqlConn) *svc.ServiceContext {
	svcCtx := &svc.ServiceContext{DB: conn}
	svcCtx.Config.Auth.AccessSecret = "demo-secret"
	return svcCtx
}
