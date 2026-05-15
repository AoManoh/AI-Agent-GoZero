package user

import (
	"context"
	"errors"
	"time"

	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/api/user/internal/types"
	"GoZero-AI/api/user/model"
	"GoZero-AI/internal/statuserr"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type SessionDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

type sessionMessageRow struct {
	Role      string    `db:"role"`
	Content   string    `db:"content"`
	CreatedAt time.Time `db:"created_at"`
}

func NewSessionDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SessionDetailLogic {
	return &SessionDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SessionDetailLogic) SessionDetail(req *types.SessionDetailReq) (*types.SessionDetailResp, error) {
	userID, err := currentUserID(l.ctx)
	if err != nil {
		return nil, err
	}

	session, err := l.svcCtx.ChatSessionsModel.FindOneBySessionID(l.ctx, userID, req.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, statuserr.NotFound("会话不存在或已删除")
		}
		return nil, err
	}

	var rows []sessionMessageRow
	query := `select role, content, created_at
from "public"."vector_store"
where chat_id = $1 and user_id = $2 and doc_type = 'message'
order by created_at asc`
	err = l.svcCtx.DB.QueryRowsCtx(l.ctx, &rows, query, session.SessionId, userID)
	if err != nil && err != sqlx.ErrNotFound {
		return nil, err
	}

	messages := make([]types.SessionMessage, 0, len(rows))
	for _, row := range rows {
		messages = append(messages, types.SessionMessage{
			Role:      row.Role,
			Content:   row.Content,
			CreatedAt: row.CreatedAt.Format(timeLayout),
		})
	}

	return &types.SessionDetailResp{
		Session:  buildSessionItem(*session),
		Messages: messages,
	}, nil
}
