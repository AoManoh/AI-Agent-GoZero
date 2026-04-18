package user

import (
	"context"
	"strings"

	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/api/user/internal/types"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
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

	session, err := l.svcCtx.ChatSessionsModel.Create(l.ctx, userID, uuid.NewString(), title, mode)
	if err != nil {
		return nil, err
	}

	return &types.CreateSessionResp{
		Session: buildSessionItem(*session),
	}, nil
}
