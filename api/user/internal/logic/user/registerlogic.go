package user

import (
	"context"
	"errors"
	"strings"

	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/api/user/internal/types"
	"GoZero-AI/api/user/model"
	"GoZero-AI/api/user/internal/auth"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewRegisterLogic 用户注册
func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterReq) (resp *types.RegisterResp, err error) {
	username := strings.TrimSpace(req.Username)
	if username == "" || req.Password == "" {
		return nil, errors.New("用户名和密码不能为空")
	}
	if req.Password != req.ConfirmPassword {
		return nil, errors.New("两次输入的密码不一致")
	}

	_, err = l.svcCtx.UsersModel.FindOneByUsername(l.ctx, username)
	if err == nil {
		return nil, errors.New("用户名已存在")
	}
	if err != nil && !errors.Is(err, model.ErrNotFound) {
		return nil, err
	}

	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	_, err = l.svcCtx.UsersModel.Insert(l.ctx, &model.Users{
		Username:     username,
		PasswordHash: passwordHash,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return nil, errors.New("用户名已存在")
		}
		return nil, err
	}

	return &types.RegisterResp{}, nil
}
