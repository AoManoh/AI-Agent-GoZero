package chatflow

import (
	"context"
	"testing"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestBuildContextKey(t *testing.T) {
	userID := int64(42)
	key := BuildContextKey("sess-1", &userID, "Research Desk")
	if key.OwnerScope != "user:42" {
		t.Fatalf("ownerScope = %q", key.OwnerScope)
	}
	if key.Lane != "research" {
		t.Fatalf("lane = %q", key.Lane)
	}
}

func TestDefaultSnapshot(t *testing.T) {
	key := BuildContextKey("sess-1", nil, "")
	snapshot := DefaultSnapshot(key, "")
	if snapshot.MemoryScope != MemoryScopeSession {
		t.Fatalf("memoryScope = %q", snapshot.MemoryScope)
	}
	if snapshot.InterviewState != InterviewStateStart {
		t.Fatalf("interviewState = %q", snapshot.InterviewState)
	}
}

func TestLoadSnapshotDoesNotFallbackToLegacyKey(t *testing.T) {
	client := newMiniRedisClient(t)
	defer client.Close()

	if err := client.Set(context.Background(), LegacyStateRedisKey("sess-1"), "question", StateTTL).Err(); err != nil {
		t.Fatalf("Set legacy state error = %v", err)
	}

	key := BuildContextKey("sess-1", nil, "")
	snapshot, err := LoadSnapshot(context.Background(), client, key, InterviewStateStart)
	if err != nil {
		t.Fatalf("LoadSnapshot() error = %v", err)
	}
	if snapshot.InterviewState != InterviewStateStart {
		t.Fatalf("snapshot.InterviewState = %q, want %q", snapshot.InterviewState, InterviewStateStart)
	}
}

func TestSaveSnapshotDoesNotWriteLegacyKey(t *testing.T) {
	client := newMiniRedisClient(t)
	defer client.Close()

	key := BuildContextKey("sess-1", nil, "")
	snapshot := DefaultSnapshot(key, "question")
	if err := SaveSnapshot(context.Background(), client, key, snapshot); err != nil {
		t.Fatalf("SaveSnapshot() error = %v", err)
	}

	_, found, err := ReadLegacyState(context.Background(), client, "sess-1")
	if err != nil {
		t.Fatalf("ReadLegacyState() error = %v", err)
	}
	if found {
		t.Fatal("legacy state should not be written by SaveSnapshot")
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
