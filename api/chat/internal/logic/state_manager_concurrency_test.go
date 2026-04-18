package logic

import (
	"context"
	"sync"
	"testing"

	"GoZero-AI/api/chat/internal/svc"
	"GoZero-AI/internal/chatflow"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestRecordTurnConcurrentDoesNotLoseUpdates(t *testing.T) {
	server, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis.Run() error = %v", err)
	}
	defer server.Close()

	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	defer client.Close()

	manager := NewStateManager(context.Background(), &svc.ServiceContext{
		RedisClient: client,
	})
	scope := ConversationScope{
		ChatID: "sess-concurrent",
		Mode:   "Interview",
	}

	const turns = 24
	var wg sync.WaitGroup
	errCh := make(chan error, turns)
	for i := 0; i < turns; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := manager.RecordTurn(scope, "user", "concurrent_test"); err != nil {
				errCh <- err
			}
		}()
	}
	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			t.Fatalf("RecordTurn() error = %v", err)
		}
	}

	snapshot, err := manager.GetFlowState(scope)
	if err != nil {
		t.Fatalf("GetFlowState() error = %v", err)
	}
	if snapshot.TurnCount != turns {
		t.Fatalf("snapshot.TurnCount = %d, want %d", snapshot.TurnCount, turns)
	}

	key := chatflow.BuildContextKey(scope.ChatID, scope.UserID, scope.Mode)
	events, err := chatflow.LoadEvents(context.Background(), client, key, turns)
	if err != nil {
		t.Fatalf("LoadEvents() error = %v", err)
	}
	if len(events) != turns {
		t.Fatalf("len(events) = %d, want %d", len(events), turns)
	}
}
