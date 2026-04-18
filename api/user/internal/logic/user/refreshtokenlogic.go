package user

import (
	"context"
	"errors"
	"strings"

	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/api/user/internal/types"
	"GoZero-AI/api/user/internal/auth"
	"GoZero-AI/api/user/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type RefreshTokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 刷新访问令牌
func NewRefreshTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshTokenLogic {
	return &RefreshTokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RefreshTokenLogic) RefreshToken(req *types.RefreshTokenReq) (resp *types.RefreshTokenResp, err error) {
	refreshToken := strings.TrimSpace(req.RefreshToken)
	if refreshToken == "" {
		return nil, errors.New("refresh token 不能为空")
	}

	claims, err := auth.ParseTokenWithType(l.svcCtx.Config.Auth.AccessSecret, refreshToken, auth.TokenTypeRefresh)
	if err != nil {
		return nil, errors.New("refresh token 无效或已过期")
	}

	if err := validateRefreshToken(l.ctx, l.svcCtx, claims.UserID, claims.ID); err != nil {
		return nil, err
	}

	userEntity, err := l.svcCtx.UsersModel.FindOne(l.ctx, claims.UserID)
	if errors.Is(err, model.ErrNotFound) {
		return nil, errors.New("用户不存在")
	}
	if err != nil {
		return nil, err
	}

	tokenPair, err := auth.IssueTokenPair(
		l.svcCtx.Config.Auth.AccessSecret,
		l.svcCtx.Config.AccessTokenTTL(),
		l.svcCtx.RefreshTokenTTL,
		userEntity.Id,
		userEntity.Username,
	)
	if err != nil {
		return nil, err
	}

	if err := rotateRefreshToken(l.ctx, l.svcCtx, userEntity.Id, claims.ID, tokenPair.RefreshTokenJTI); err != nil {
		return nil, err
	}

	return &types.RefreshTokenResp{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpireIn:     tokenPair.ExpireIn,
	}, nil
}
