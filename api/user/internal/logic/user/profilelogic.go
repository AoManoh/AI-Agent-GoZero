package user

import (
	"context"

	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/api/user/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ProfileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProfileLogic {
	return &ProfileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ProfileLogic) Profile(_ *types.ProfileReq) (*types.ProfileResp, error) {
	userID, err := currentUserID(l.ctx)
	if err != nil {
		return nil, err
	}

	userEntity, err := l.svcCtx.UsersModel.FindOne(l.ctx, userID)
	if err != nil {
		return nil, err
	}

	return &types.ProfileResp{
		UserId:    userEntity.Id,
		Username:  userEntity.Username,
		CreatedAt: userEntity.CreatedAt.Format(timeLayout),
	}, nil
}
