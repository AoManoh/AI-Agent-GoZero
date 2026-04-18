package chatflow

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"GoZero-AI/internal/sessionmode"

	"github.com/redis/go-redis/v9"
)

const (
	SnapshotVersion       = "chat-flow-v1"
	StateKeyPrefix        = "chat_flow_state:v1:"
	EventKeyPrefix        = "chat_flow_events:v1:"
	LegacyStateKeyPrefix  = "chat_state:"
	StateTTL              = 24 * time.Hour
	DefaultEventLimit     = int64(50)
	mutateSnapshotRetries = 32

	InterviewStateStart = "start"

	LifecycleActive    = "active"
	LifecycleCompleted = "completed"
	LifecycleErrored   = "errored"

	ExecutionIdle       = "idle"
	ExecutionRetrieving = "retrieving"
	ExecutionGenerating = "generating"
	ExecutionPersisting = "persisting"
	ExecutionFailed     = "failed"

	OwnerScopeAnonymous = "anon"
	MemoryScopeSession  = "session"
	MemoryScopeUser     = "user"
)

type ContextKey struct {
	OwnerScope string `json:"ownerScope"`
	SessionID  string `json:"sessionId"`
	Lane       string `json:"lane"`
}

type Snapshot struct {
	Version        string `json:"version"`
	OwnerScope     string `json:"ownerScope"`
	SessionID      string `json:"sessionId"`
	Lane           string `json:"lane"`
	MemoryScope    string `json:"memoryScope"`
	InterviewState string `json:"interviewState"`
	LifecycleState string `json:"lifecycleState"`
	ExecutionState string `json:"executionState"`
	TurnCount      int64  `json:"turnCount"`
	LastEvent      string `json:"lastEvent"`
	LastReason     string `json:"lastReason"`
	UpdatedAt      string `json:"updatedAt"`
}

type Event struct {
	Type   string `json:"type"`
	From   string `json:"from,omitempty"`
	To     string `json:"to,omitempty"`
	Role   string `json:"role,omitempty"`
	Reason string `json:"reason,omitempty"`
	At     string `json:"at"`
}

type snapshotGetter interface {
	Get(ctx context.Context, key string) *redis.StringCmd
}

func BuildContextKey(sessionID string, userID *int64, mode string) ContextKey {
	return ContextKey{
		OwnerScope: ownerScope(userID),
		SessionID:  strings.TrimSpace(sessionID),
		Lane:       NormalizeLane(mode),
	}
}

func NormalizeLane(mode string) string {
	switch sessionmode.NormalizeKey(mode) {
	case sessionmode.KeyResearch:
		return "research"
	case sessionmode.KeyMemory:
		return "memory"
	case sessionmode.KeyCoach:
		return "coach"
	default:
		return "interview"
	}
}

func StateRedisKey(key ContextKey) string {
	return fmt.Sprintf("%s%s:%s:%s", StateKeyPrefix, key.OwnerScope, key.Lane, key.SessionID)
}

func EventRedisKey(key ContextKey) string {
	return fmt.Sprintf("%s%s:%s:%s", EventKeyPrefix, key.OwnerScope, key.Lane, key.SessionID)
}

func LegacyStateRedisKey(sessionID string) string {
	return LegacyStateKeyPrefix + strings.TrimSpace(sessionID)
}

func DefaultSnapshot(key ContextKey, initialState string) Snapshot {
	now := time.Now().Format(time.RFC3339)
	return Snapshot{
		Version:        SnapshotVersion,
		OwnerScope:     key.OwnerScope,
		SessionID:      key.SessionID,
		Lane:           key.Lane,
		MemoryScope:    DefaultMemoryScope(key.OwnerScope),
		InterviewState: coalesceState(initialState),
		LifecycleState: LifecycleActive,
		ExecutionState: ExecutionIdle,
		LastEvent:      "session.init",
		LastReason:     "default_initialized",
		UpdatedAt:      now,
	}
}

func ReadSnapshot(ctx context.Context, client *redis.Client, key ContextKey) (Snapshot, bool, error) {
	return readSnapshot(ctx, client, key)
}

func readSnapshot(ctx context.Context, getter snapshotGetter, key ContextKey) (Snapshot, bool, error) {
	raw, err := getter.Get(ctx, StateRedisKey(key)).Result()
	if err != nil {
		if err == redis.Nil {
			return Snapshot{}, false, nil
		}
		return Snapshot{}, false, err
	}

	var snapshot Snapshot
	if err := json.Unmarshal([]byte(raw), &snapshot); err != nil {
		return Snapshot{}, false, err
	}
	snapshot.OwnerScope = key.OwnerScope
	snapshot.SessionID = key.SessionID
	snapshot.Lane = key.Lane
	if snapshot.Version == "" {
		snapshot.Version = SnapshotVersion
	}
	if snapshot.MemoryScope == "" {
		snapshot.MemoryScope = DefaultMemoryScope(key.OwnerScope)
	}
	if snapshot.InterviewState == "" {
		snapshot.InterviewState = InterviewStateStart
	}
	if snapshot.LifecycleState == "" {
		snapshot.LifecycleState = LifecycleActive
	}
	if snapshot.ExecutionState == "" {
		snapshot.ExecutionState = ExecutionIdle
	}
	if snapshot.UpdatedAt == "" {
		snapshot.UpdatedAt = time.Now().Format(time.RFC3339)
	}
	return snapshot, true, nil
}

func LoadSnapshot(ctx context.Context, client *redis.Client, key ContextKey, initialState string) (Snapshot, error) {
	snapshot, found, err := ReadSnapshot(ctx, client, key)
	if err != nil {
		return Snapshot{}, err
	}
	if found {
		return snapshot, nil
	}
	return DefaultSnapshot(key, initialState), nil
}

func ReadLegacyState(ctx context.Context, client *redis.Client, sessionID string) (string, bool, error) {
	legacyState, err := client.Get(ctx, LegacyStateRedisKey(sessionID)).Result()
	switch err {
	case nil:
		return legacyState, true, nil
	case redis.Nil:
		return "", false, nil
	default:
		return "", false, err
	}
}

func SaveSnapshot(ctx context.Context, client *redis.Client, key ContextKey, snapshot Snapshot) error {
	_, payload, err := normalizeSnapshotPayload(key, snapshot)
	if err != nil {
		return err
	}

	pipe := client.TxPipeline()
	pipe.Set(ctx, StateRedisKey(key), payload, StateTTL)
	_, err = pipe.Exec(ctx)
	return err
}

func MutateSnapshot(ctx context.Context, client *redis.Client, key ContextKey, initialState string, maxEvents int64,
	mutate func(Snapshot) (Snapshot, *Event, error),
) (Snapshot, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if maxEvents <= 0 {
		maxEvents = DefaultEventLimit
	}

	var finalSnapshot Snapshot
	for attempt := 0; attempt < mutateSnapshotRetries; attempt++ {
		err := client.Watch(ctx, func(tx *redis.Tx) error {
			snapshot, found, err := readSnapshot(ctx, tx, key)
			if err != nil {
				return err
			}
			if !found {
				snapshot = DefaultSnapshot(key, initialState)
			}

			nextSnapshot, event, err := mutate(snapshot)
			if err != nil {
				return err
			}

			normalizedSnapshot, payload, err := normalizeSnapshotPayload(key, nextSnapshot)
			if err != nil {
				return err
			}

			_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
				pipe.Set(ctx, StateRedisKey(key), payload, StateTTL)
				if event != nil {
					if event.At == "" {
						event.At = normalizedSnapshot.UpdatedAt
					}
					eventPayload, err := json.Marshal(event)
					if err != nil {
						return err
					}
					pipe.LPush(ctx, EventRedisKey(key), eventPayload)
					pipe.LTrim(ctx, EventRedisKey(key), 0, maxEvents-1)
					pipe.Expire(ctx, EventRedisKey(key), StateTTL)
				}
				return nil
			})
			if err != nil {
				return err
			}

			finalSnapshot = normalizedSnapshot
			return nil
		}, StateRedisKey(key))
		if err == nil {
			return finalSnapshot, nil
		}
		if err != redis.TxFailedErr {
			return Snapshot{}, err
		}
		select {
		case <-ctx.Done():
			return Snapshot{}, ctx.Err()
		case <-time.After(time.Millisecond):
		}
	}

	return Snapshot{}, redis.TxFailedErr
}

func AppendEvent(ctx context.Context, client *redis.Client, key ContextKey, event Event, maxEvents int64) error {
	if event.At == "" {
		event.At = time.Now().Format(time.RFC3339)
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	if maxEvents <= 0 {
		maxEvents = DefaultEventLimit
	}

	pipe := client.TxPipeline()
	pipe.LPush(ctx, EventRedisKey(key), payload)
	pipe.LTrim(ctx, EventRedisKey(key), 0, maxEvents-1)
	pipe.Expire(ctx, EventRedisKey(key), StateTTL)
	_, err = pipe.Exec(ctx)
	return err
}

func normalizeSnapshotPayload(key ContextKey, snapshot Snapshot) (Snapshot, []byte, error) {
	snapshot.OwnerScope = key.OwnerScope
	snapshot.SessionID = key.SessionID
	snapshot.Lane = key.Lane
	if snapshot.Version == "" {
		snapshot.Version = SnapshotVersion
	}
	if snapshot.MemoryScope == "" {
		snapshot.MemoryScope = DefaultMemoryScope(key.OwnerScope)
	}
	if snapshot.InterviewState == "" {
		snapshot.InterviewState = InterviewStateStart
	}
	if snapshot.LifecycleState == "" {
		snapshot.LifecycleState = LifecycleActive
	}
	if snapshot.ExecutionState == "" {
		snapshot.ExecutionState = ExecutionIdle
	}
	if snapshot.UpdatedAt == "" {
		snapshot.UpdatedAt = time.Now().Format(time.RFC3339)
	}

	payload, err := json.Marshal(snapshot)
	if err != nil {
		return Snapshot{}, nil, err
	}
	return snapshot, payload, nil
}

func LoadEvents(ctx context.Context, client *redis.Client, key ContextKey, limit int64) ([]Event, error) {
	if limit <= 0 {
		limit = DefaultEventLimit
	}
	values, err := client.LRange(ctx, EventRedisKey(key), 0, limit-1).Result()
	if err != nil {
		if err == redis.Nil {
			return []Event{}, nil
		}
		return nil, err
	}

	events := make([]Event, 0, len(values))
	for i := len(values) - 1; i >= 0; i-- {
		var event Event
		if err := json.Unmarshal([]byte(values[i]), &event); err != nil {
			continue
		}
		events = append(events, event)
	}
	return events, nil
}

func DefaultMemoryScope(ownerScope string) string {
	if ownerScope == OwnerScopeAnonymous {
		return MemoryScopeSession
	}
	return MemoryScopeUser
}

func ownerScope(userID *int64) string {
	if userID == nil {
		return OwnerScopeAnonymous
	}
	return fmt.Sprintf("user:%d", *userID)
}

func coalesceState(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return InterviewStateStart
	}
	return trimmed
}
