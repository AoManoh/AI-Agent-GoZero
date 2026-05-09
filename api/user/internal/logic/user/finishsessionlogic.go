package user

import (
	"context"
	"errors"

	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/api/user/internal/types"
	"GoZero-AI/api/user/model"
	"GoZero-AI/internal/statuserr"

	"github.com/zeromicro/go-zero/core/logx"
)

type FinishSessionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFinishSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FinishSessionLogic {
	return &FinishSessionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FinishSessionLogic) FinishSession(req *types.FinishSessionReq) (*types.FinishSessionResp, error) {
	userID, err := currentUserID(l.ctx)
	if err != nil {
		return nil, err
	}

	session, err := l.svcCtx.ChatSessionsModel.Complete(l.ctx, userID, req.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, statuserr.NotFound("会话不存在或已删除")
		}
		return nil, err
	}

	return &types.FinishSessionResp{
		Session: buildSessionItem(*session),
		Config:  buildSessionConfigSnapshot(*session),
	}, nil
}
