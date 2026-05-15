package user

import (
	"context"
	"errors"

	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/api/user/internal/types"
	"GoZero-AI/api/user/internal/auth"

	"github.com/zeromicro/go-zero/core/logx"
)

type LogoutLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 用户退出登录
func NewLogoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LogoutLogic {
	return &LogoutLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LogoutLogic) Logout(req *types.LogoutReq) (resp *types.LogoutResp, err error) {
	accessToken := accessTokenFromContext(l.ctx)
	if accessToken == "" {
		return nil, errors.New("缺少 access token")
	}

	claims, err := auth.ParseTokenWithType(l.svcCtx.Config.Auth.AccessSecret, accessToken, auth.TokenTypeAccess)
	if err != nil {
		return nil, errors.New("access token 无效或已过期")
	}

	if err := revokeAllRefreshTokens(l.ctx, l.svcCtx, claims.UserID); err != nil {
		return nil, err
	}

	return &types.LogoutResp{}, nil
}
