package user

import (
	"context"
	"errors"
	"strings"

	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/api/user/internal/types"
	"GoZero-AI/api/user/model"
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

	if title == "新对话" && selectedQuestion != nil {
		title = selectedQuestion.Title
	} else if title == "新对话" {
		title = configResp.DirectionLabel + "面试"
	}

	sessionID := uuid.NewString()
	var session *model.ChatSession
	if selectedQuestion != nil {
		err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, tx sqlx.Session) error {
			created, err := model.CreateChatSessionWithConfigTx(ctx, tx, userID, sessionID, title, mode, config)
			if err != nil {
				return err
			}
			if err := l.svcCtx.InterviewQuestionsModel.AttachToSession(ctx, tx, userID, sessionID, *selectedQuestion); err != nil {
				return err
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
		Session: buildSessionItem(*session),
		Config:  buildSessionConfigSnapshot(*session),
	}, nil
}
