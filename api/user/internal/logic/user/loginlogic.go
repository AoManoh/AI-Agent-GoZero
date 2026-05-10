package user

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"GoZero-AI/api/user/internal/auth"
	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/api/user/internal/types"
	"GoZero-AI/api/user/model"
	"GoZero-AI/internal/statuserr"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewLoginLogic 用户登录
func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginReq) (resp *types.LoginResp, err error) {
	if req == nil {
		return nil, statuserr.New(http.StatusBadRequest, "登录参数不能为空")
	}

	username := strings.TrimSpace(req.Username)
	if username == "" || req.Password == "" {
		return nil, statuserr.New(http.StatusBadRequest, "用户名和密码不能为空")
	}

	userEntity, err := l.svcCtx.UsersModel.FindOneByUsername(l.ctx, username)
	if errors.Is(err, model.ErrNotFound) {
		return nil, statuserr.Unauthorized("用户名或密码错误")
	}
	if err != nil {
		return nil, err
	}

	if err := auth.ComparePassword(userEntity.PasswordHash, req.Password); err != nil {
		return nil, statuserr.Unauthorized("用户名或密码错误")
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

	if err := storeRefreshToken(l.ctx, l.svcCtx, userEntity.Id, tokenPair.RefreshTokenJTI); err != nil {
		return nil, err
	}

	return &types.LoginResp{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpireIn:     tokenPair.ExpireIn,
	}, nil
}
