package user

import (
	"context"
	"errors"
	"strings"

	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/api/user/internal/types"
	"GoZero-AI/api/user/model"
	"GoZero-AI/internal/sessionruntime"
	"GoZero-AI/internal/statuserr"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type CreateSessionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateSessionLogic {
	return &CreateSessionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateSessionLogic) CreateSession(req *types.CreateSessionReq) (*types.CreateSessionResp, error) {
	userID, err := currentUserID(l.ctx)
	if err != nil {
		return nil, err
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = "新对话"
	}
	mode := normalizeSessionMode(req.Mode)
	config, configResp, err := buildSessionCreateConfig(req)
	if err != nil {
		return nil, err
	}
	var resumeBinding *types.ResumeBindingSummary
	resumeArtifactID := strings.TrimSpace(req.ResumeArtifactId)
	if resumeArtifactID != "" {
		artifact, err := loadResumeArtifactItem(l.ctx, l.svcCtx.DB, userID, resumeArtifactID)
		if err != nil {
			if errors.Is(err, model.ErrNotFound) || errors.Is(err, sqlx.ErrNotFound) {
				return nil, statuserr.NotFound("简历资料不存在或已删除")
			}
			return nil, err
		}
		if artifact.Status != "" && artifact.Status != "ready" {
			return nil, statuserr.Conflict("简历尚未解析完成，请稍后再创建面试")
		}
		config.ResumeArtifactId = artifact.ArtifactId
		resumeBinding = &types.ResumeBindingSummary{
			ArtifactId: artifact.ArtifactId,
			Title:      artifact.Title,
			Version:    artifact.Version,
			Status:     artifact.Status,
		}
	}
	var selectedQuestion *model.InterviewQuestion
	questionKey := strings.TrimSpace(req.QuestionKey)
	if questionKey != "" {
		question, _, err := l.svcCtx.InterviewQuestionsModel.FindOne(l.ctx, questionKey)
		if err != nil {
			if errors.Is(err, model.ErrNotFound) {
				return nil, statuserr.NotFound("题目不存在或已下线")
			}
			return nil, err
		}
		selectedQuestion = question
	}
	var generatedQuestion *types.InterviewPlanQuestion
	if selectedQuestion == nil && config.ResumeArtifactId != "" {
		question, err := l.loadResumeSuggestedQuestion(userID, config.ResumeArtifactId)
		if err != nil {
			return nil, err
		}
		generatedQuestion = question
	}
	applySessionRuntimeContext(&config, selectedQuestion, generatedQuestion)

	if title == "新对话" && selectedQuestion != nil {
		title = selectedQuestion.Title
	} else if title == "新对话" {
		title = configResp.DirectionLabel + "面试"
	}

	sessionID := uuid.NewString()
	var session *model.ChatSession
	if selectedQuestion != nil || generatedQuestion != nil {
		err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, tx sqlx.Session) error {
			created, err := model.CreateChatSessionWithConfigTx(ctx, tx, userID, sessionID, title, mode, config)
			if err != nil {
				return err
			}
			if selectedQuestion != nil {
				if err := l.svcCtx.InterviewQuestionsModel.AttachToSession(ctx, tx, userID, sessionID, *selectedQuestion); err != nil {
					return err
				}
			} else if generatedQuestion != nil && l.svcCtx.InterviewQuestionsModel != nil {
				if err := l.svcCtx.InterviewQuestionsModel.AttachGeneratedToSession(ctx, tx, userID, sessionID, generatedQuestion.Key, generatedQuestion.Prompt); err != nil {
					return err
				}
			}
			session = created
			return nil
		})
		if err != nil {
			return nil, err
		}
		session, err = l.svcCtx.ChatSessionsModel.FindOneBySessionID(l.ctx, userID, sessionID)
		if err != nil {
			return nil, err
		}
	} else {
		session, err = l.svcCtx.ChatSessionsModel.CreateWithConfig(l.ctx, userID, sessionID, title, mode, config)
		if err != nil {
			return nil, err
		}
	}

	return &types.CreateSessionResp{
		Session:       buildSessionItem(*session),
		Config:        buildSessionConfigSnapshot(*session),
		ResumeBinding: resumeBinding,
	}, nil
}

func applySessionRuntimeContext(config *model.SessionCreateConfig, selectedQuestion *model.InterviewQuestion, generatedQuestion *types.InterviewPlanQuestion) {
	if config == nil {
		return
	}
	config.ScenarioType = sessionruntime.ScenarioFormalInterview
	config.StarterSource = sessionruntime.StarterNone
	config.StarterQuestionKey = ""
	if selectedQuestion != nil {
		config.ScenarioType = sessionruntime.ScenarioQuestionPractice
		config.StarterSource = sessionruntime.StarterBank
		config.StarterQuestionKey = strings.TrimSpace(selectedQuestion.QuestionKey)
		return
	}
	if generatedQuestion != nil {
		config.StarterSource = sessionruntime.StarterResumePlan
		config.StarterQuestionKey = strings.TrimSpace(generatedQuestion.Key)
	}
}

func (l *CreateSessionLogic) loadResumeSuggestedQuestion(userID int64, artifactID string) (*types.InterviewPlanQuestion, error) {
	if l.svcCtx.ResumeEvaluationsModel == nil {
		return nil, nil
	}
	record, err := l.svcCtx.ResumeEvaluationsModel.FindOneByArtifactID(l.ctx, userID, artifactID)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	status := strings.TrimSpace(record.Status)
	if status != resumeEvaluationStatusReady && status != resumeEvaluationStatusStale {
		return nil, nil
	}
	var questions []types.InterviewPlanQuestion
	unmarshalJSONOrDefault(record.SuggestedQuestions, &questions)
	for _, question := range questions {
		question.Prompt = strings.TrimSpace(question.Prompt)
		question.Title = strings.TrimSpace(question.Title)
		if question.Prompt == "" {
			question.Prompt = question.Title
		}
		if question.Title == "" {
			question.Title = question.Prompt
		}
		if question.Prompt == "" {
			continue
		}
		question.Key = strings.TrimSpace(question.Key)
		if question.Key == "" {
			question.Key = "resume:" + artifactID + ":q1"
		}
		return &question, nil
	}
	return nil, nil
}
