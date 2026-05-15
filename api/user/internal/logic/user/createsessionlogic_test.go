package user

import (
	"database/sql"
	"regexp"
	"testing"
	"time"

	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/api/user/internal/types"
	"GoZero-AI/api/user/model"
	"GoZero-AI/internal/sessionmode"
	"GoZero-AI/internal/sessionruntime"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func TestCreateSessionAttachesResumeSuggestedQuestion(t *testing.T) {
	now := time.Date(2026, 5, 12, 10, 30, 0, 0, time.UTC)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	const artifactID = "resume-artifact-1"
	const prompt = "请介绍 GoZero AI 项目中的 RAG 设计。"

	expectResumeDocumentItem(mock, artifactID, now)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`insert into "public"."chat_sessions"`)).
		WithArgs(
			sqlmock.AnyArg(),
			int64(7),
			"简历面试",
			sessionmode.KeyInterview,
			"go_backend",
			"Go 后端",
			int64(3),
			"中级",
			"senior",
			"资深技术官",
			sqlmock.AnyArg(),
			"N+3",
			int64(30),
			artifactID,
			"formal_interview",
			"resume_plan",
			"resume-follow-1",
		).
		WillReturnRows(newCreateSessionRows("sess-created", artifactID, now, 0))
	mock.ExpectExec(regexp.QuoteMeta(`insert into "public"."vector_store"`)).
		WithArgs(sqlmock.AnyArg(), int64(7), prompt, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta(`insert into "public"."session_question_events"`)).
		WithArgs(sqlmock.AnyArg(), int64(7), sqlmock.AnyArg(), "resume-follow-1", "generated", prompt).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta(`update "public"."chat_sessions"`)).
		WithArgs(sqlmock.AnyArg(), int64(7)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	mock.ExpectQuery(`from "public"\."chat_sessions"`).
		WithArgs(int64(7), sqlmock.AnyArg()).
		WillReturnRows(newCreateSessionRows("sess-created", artifactID, now, 1))

	logic := NewCreateSessionLogic(withUserID(t.Context(), 7), &svc.ServiceContext{
		DB:                sqlx.NewSqlConnFromDB(db),
		ChatSessionsModel: model.NewChatSessionsModel(sqlx.NewSqlConnFromDB(db)),
		InterviewQuestionsModel: model.NewInterviewQuestionsModel(
			sqlx.NewSqlConnFromDB(db),
		),
		ResumeEvaluationsModel: &stubResumeEvaluationsModel{record: &model.ResumeEvaluation{
			ArtifactId: artifactID,
			UserId:     7,
			Status:     resumeEvaluationStatusReady,
			SuggestedQuestions: marshalJSONOrEmptyArray([]types.InterviewPlanQuestion{
				{
					Key:    "resume-follow-1",
					Title:  "RAG 设计追问",
					Prompt: prompt,
				},
			}),
			FirstGeneratedAt: now,
			GeneratedAt:      now,
			UpdatedAt:        now,
		}},
	})

	resp, err := logic.CreateSession(&types.CreateSessionReq{
		Title:            "简历面试",
		Mode:             sessionmode.KeyInterview,
		ResumeArtifactId: artifactID,
	})
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	if resp.ResumeBinding == nil || resp.ResumeBinding.ArtifactId != artifactID {
		t.Fatalf("ResumeBinding = %+v, want artifact %q", resp.ResumeBinding, artifactID)
	}
	if resp.Config.ResumeArtifactId != artifactID {
		t.Fatalf("Config.ResumeArtifactId = %q, want %q", resp.Config.ResumeArtifactId, artifactID)
	}
	if resp.Session.MessageCount != 1 {
		t.Fatalf("Session.MessageCount = %d, want generated starter message", resp.Session.MessageCount)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestApplySessionRuntimeContextUsesExplicitStarter(t *testing.T) {
	t.Run("bank question is question practice", func(t *testing.T) {
		config := model.SessionCreateConfig{}
		applySessionRuntimeContext(&config, &model.InterviewQuestion{QuestionKey: "go-rag"}, nil)
		if config.ScenarioType != sessionruntime.ScenarioQuestionPractice ||
			config.StarterSource != sessionruntime.StarterBank ||
			config.StarterQuestionKey != "go-rag" {
			t.Fatalf("config = %+v, want bank question practice", config)
		}
	})

	t.Run("resume generated starter remains formal interview", func(t *testing.T) {
		config := model.SessionCreateConfig{ResumeArtifactId: "resume-1"}
		applySessionRuntimeContext(&config, nil, &types.InterviewPlanQuestion{Key: "resume:q1"})
		if config.ScenarioType != sessionruntime.ScenarioFormalInterview ||
			config.StarterSource != sessionruntime.StarterResumePlan ||
			config.StarterQuestionKey != "resume:q1" {
			t.Fatalf("config = %+v, want resume formal interview", config)
		}
	})
}

func newCreateSessionRows(sessionID, resumeArtifactID string, now time.Time, messageCount int64) *sqlmock.Rows {
	return sqlmock.NewRows(chatSessionModelColumns()).AddRow(
		int64(1),
		sessionID,
		int64(7),
		"简历面试",
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
		int64(0),
		now,
		now,
		sql.NullTime{},
		sql.NullTime{Time: now, Valid: true},
		sql.NullTime{},
		int64(0),
		resumeArtifactID,
		"formal_interview",
		"resume_plan",
		"resume-follow-1",
		messageCount,
		true,
	)
}
