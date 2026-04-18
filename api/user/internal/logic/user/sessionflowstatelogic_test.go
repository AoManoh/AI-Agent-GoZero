package user

import (
	"context"
	"testing"
	"time"

	"GoZero-AI/internal/chatflow"
	"GoZero-AI/internal/sessionmode"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestResolveSessionFlowStateNoSnapshotIsNotAvailable(t *testing.T) {
	client := newMiniRedisClient(t)
	defer client.Close()

	userID := int64(7)
	requestedKey := chatflow.BuildContextKey("sess-1", &userID, sessionmode.KeyInterview)

	resolvedKey, snapshot, available, err := resolveSessionFlowState(context.Background(), client, requestedKey)
	if err != nil {
		t.Fatalf("resolveSessionFlowState() error = %v", err)
	}
	if available {
		t.Fatal("available = true, want false")
	}
	if resolvedKey != requestedKey {
		t.Fatalf("resolvedKey = %#v, want %#v", resolvedKey, requestedKey)
	}
	if snapshot.InterviewState != chatflow.InterviewStateStart {
		t.Fatalf("snapshot.InterviewState = %q, want %q", snapshot.InterviewState, chatflow.InterviewStateStart)
	}
}

func TestResolveSessionFlowStateUsesAlternateLaneSnapshot(t *testing.T) {
	client := newMiniRedisClient(t)
	defer client.Close()

	userID := int64(7)
	requestedKey := chatflow.BuildContextKey("sess-1", &userID, sessionmode.KeyInterview)
	researchKey := chatflow.BuildContextKey("sess-1", &userID, sessionmode.KeyResearch)
	researchSnapshot := chatflow.DefaultSnapshot(researchKey, "question")
	researchSnapshot.UpdatedAt = time.Date(2026, 4, 12, 23, 0, 0, 0, time.UTC).Format(time.RFC3339)
	if err := chatflow.SaveSnapshot(context.Background(), client, researchKey, researchSnapshot); err != nil {
		t.Fatalf("SaveSnapshot() error = %v", err)
	}

	resolvedKey, snapshot, available, err := resolveSessionFlowState(context.Background(), client, requestedKey)
	if err != nil {
		t.Fatalf("resolveSessionFlowState() error = %v", err)
	}
	if !available {
		t.Fatal("available = false, want true")
	}
	if resolvedKey.Lane != researchKey.Lane {
		t.Fatalf("resolvedKey.Lane = %q, want %q", resolvedKey.Lane, researchKey.Lane)
	}
	if snapshot.Lane != researchKey.Lane {
		t.Fatalf("snapshot.Lane = %q, want %q", snapshot.Lane, researchKey.Lane)
	}
}

func TestResolveSessionFlowStateUsesLegacyAsReadOnlyFallback(t *testing.T) {
	client := newMiniRedisClient(t)
	defer client.Close()

	userID := int64(7)
	requestedKey := chatflow.BuildContextKey("sess-1", &userID, sessionmode.KeyInterview)
	if err := client.Set(context.Background(), chatflow.LegacyStateRedisKey("sess-1"), "question", chatflow.StateTTL).Err(); err != nil {
		t.Fatalf("Set legacy state error = %v", err)
	}

	_, snapshot, available, err := resolveSessionFlowState(context.Background(), client, requestedKey)
	if err != nil {
		t.Fatalf("resolveSessionFlowState() error = %v", err)
	}
	if !available {
		t.Fatal("available = false, want true")
	}
	if snapshot.InterviewState != "question" {
		t.Fatalf("snapshot.InterviewState = %q, want question", snapshot.InterviewState)
	}
	if snapshot.LastReason != "legacy_read_only_fallback" {
		t.Fatalf("snapshot.LastReason = %q, want legacy_read_only_fallback", snapshot.LastReason)
	}
}

func newMiniRedisClient(t *testing.T) *redis.Client {
	t.Helper()

	server, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis.Run() error = %v", err)
	}
	t.Cleanup(server.Close)

	return redis.NewClient(&redis.Options{Addr: server.Addr()})
}
