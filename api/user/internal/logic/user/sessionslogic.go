package user

import (
	"context"

	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/api/user/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SessionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSessionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SessionsLogic {
	return &SessionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SessionsLogic) Sessions(_ *types.SessionsReq) (*types.SessionsResp, error) {
	userID, err := currentUserID(l.ctx)
	if err != nil {
		return nil, err
	}

	sessions, err := l.svcCtx.ChatSessionsModel.FindByUserID(l.ctx, userID)
	if err != nil {
		return nil, err
	}

	items := make([]types.SessionItem, 0, len(sessions))
	for _, session := range sessions {
		items = append(items, buildSessionItem(session))
	}

	return &types.SessionsResp{
		Sessions: items,
		Total:    int64(len(items)),
	}, nil
}
