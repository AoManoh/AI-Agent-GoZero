package user

import (
	"context"
	"errors"
	"time"

	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/api/user/internal/types"
	"GoZero-AI/api/user/model"
	"GoZero-AI/internal/chatflow"
	"GoZero-AI/internal/sessionmode"
	"GoZero-AI/internal/statuserr"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

type SessionFlowStateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSessionFlowStateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SessionFlowStateLogic {
	return &SessionFlowStateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SessionFlowStateLogic) SessionFlowState(req *types.SessionFlowStateReq) (*types.SessionFlowStateResp, error) {
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

	key := chatflow.BuildContextKey(session.SessionId, &userID, session.Mode)
	resolvedKey, snapshot, available, err := resolveSessionFlowState(l.ctx, l.svcCtx.RedisClient, key)
	if err != nil {
		return nil, err
	}
	events, err := chatflow.LoadEvents(l.ctx, l.svcCtx.RedisClient, resolvedKey, 20)
	if err != nil {
		return nil, err
	}

	flowEvents := make([]types.FlowStateEvent, 0, len(events))
	for _, event := range events {
		flowEvents = append(flowEvents, types.FlowStateEvent{
			Type:   event.Type,
			From:   event.From,
			To:     event.To,
			Role:   event.Role,
			Reason: event.Reason,
			At:     event.At,
		})
	}

	return &types.SessionFlowStateResp{
		Session:        buildSessionItem(*session),
		OwnerScope:     snapshot.OwnerScope,
		Lane:           snapshot.Lane,
		MemoryScope:    snapshot.MemoryScope,
		InterviewState: snapshot.InterviewState,
		LifecycleState: snapshot.LifecycleState,
		ExecutionState: snapshot.ExecutionState,
		TurnCount:      snapshot.TurnCount,
		LastEvent:      snapshot.LastEvent,
		LastReason:     snapshot.LastReason,
		UpdatedAt:      snapshot.UpdatedAt,
		Events:         flowEvents,
		StateMeta: types.ReportMeta{
			SchemaVersion: chatflow.SnapshotVersion,
			Available:     available,
		},
	}, nil
}

func resolveSessionFlowState(ctx context.Context, client *redis.Client, requestedKey chatflow.ContextKey) (chatflow.ContextKey, chatflow.Snapshot, bool, error) {
	snapshot, found, err := chatflow.ReadSnapshot(ctx, client, requestedKey)
	if err != nil {
		return requestedKey, chatflow.Snapshot{}, false, err
	}
	if found {
		return requestedKey, snapshot, true, nil
	}

	resolvedKey, alternateSnapshot, alternateFound, err := findAlternateLaneSnapshot(ctx, client, requestedKey)
	if err != nil {
		return requestedKey, chatflow.Snapshot{}, false, err
	}
	if alternateFound {
		return resolvedKey, alternateSnapshot, true, nil
	}

	legacyState, legacyFound, err := chatflow.ReadLegacyState(ctx, client, requestedKey.SessionID)
	if err != nil {
		return requestedKey, chatflow.Snapshot{}, false, err
	}
	if legacyFound {
		snapshot = chatflow.DefaultSnapshot(requestedKey, legacyState)
		snapshot.LastEvent = "session.legacy_fallback"
		snapshot.LastReason = "legacy_read_only_fallback"
		return requestedKey, snapshot, true, nil
	}

	return requestedKey, chatflow.DefaultSnapshot(requestedKey, chatflow.InterviewStateStart), false, nil
}

func findAlternateLaneSnapshot(ctx context.Context, client *redis.Client, requestedKey chatflow.ContextKey) (chatflow.ContextKey, chatflow.Snapshot, bool, error) {
	var (
		bestKey      chatflow.ContextKey
		bestSnapshot chatflow.Snapshot
		bestTime     time.Time
		found        bool
	)

	for _, modeKey := range sessionmode.AllKeys() {
		lane := chatflow.NormalizeLane(modeKey)
		if lane == requestedKey.Lane {
			continue
		}

		candidateKey := requestedKey
		candidateKey.Lane = lane

		snapshot, snapshotFound, err := chatflow.ReadSnapshot(ctx, client, candidateKey)
		if err != nil {
			return requestedKey, chatflow.Snapshot{}, false, err
		}
		if !snapshotFound {
			continue
		}

		candidateTime := parseFlowSnapshotTime(snapshot.UpdatedAt)
		if !found || candidateTime.After(bestTime) {
			bestKey = candidateKey
			bestSnapshot = snapshot
			bestTime = candidateTime
			found = true
		}
	}

	return bestKey, bestSnapshot, found, nil
}

func parseFlowSnapshotTime(value string) time.Time {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}
	}
	return parsed
}
