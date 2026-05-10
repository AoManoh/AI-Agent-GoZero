package user

import (
	"context"
	"errors"
	"time"

	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/api/user/internal/types"
	"GoZero-AI/api/user/model"
	"GoZero-AI/internal/chatflow"
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

	if err := l.syncFinishedFlowState(userID, *session); err != nil {
		l.Logger.Errorf("同步结束 flow state 失败: %v", err)
	}

	return &types.FinishSessionResp{
		Session: buildSessionItem(*session),
		Config:  buildSessionConfigSnapshot(*session),
	}, nil
}

func (l *FinishSessionLogic) syncFinishedFlowState(userID int64, session model.ChatSession) error {
	if l.svcCtx == nil || l.svcCtx.RedisClient == nil {
		return nil
	}

	key := chatflow.BuildContextKey(session.SessionId, &userID, session.Mode)
	_, err := chatflow.MutateSnapshot(l.ctx, l.svcCtx.RedisClient, key, chatflow.InterviewStateStart, 50, func(snapshot chatflow.Snapshot) (chatflow.Snapshot, *chatflow.Event, error) {
		from := snapshot.InterviewState
		now := time.Now().Format(time.RFC3339)
		snapshot.InterviewState = chatflow.InterviewStateEnd
		snapshot.LifecycleState = chatflow.LifecycleCompleted
		snapshot.ExecutionState = chatflow.ExecutionIdle
		if from == chatflow.InterviewStateEnd {
			snapshot.LastEvent = "state.stable"
		} else {
			snapshot.LastEvent = "state.transition"
		}
		snapshot.LastReason = "session_finished"
		snapshot.UpdatedAt = now
		return snapshot, &chatflow.Event{
			Type:   "state",
			From:   from,
			To:     chatflow.InterviewStateEnd,
			Reason: "session_finished",
			At:     now,
		}, nil
	})
	return err
}
