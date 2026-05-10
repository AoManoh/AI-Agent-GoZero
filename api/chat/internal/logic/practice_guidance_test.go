package logic

import (
	"context"
	"testing"

	"GoZero-AI/api/chat/internal/interviewer"
	"GoZero-AI/api/chat/internal/svc"
	"GoZero-AI/internal/sessionmode"

	miniredis "github.com/alicebob/miniredis/v2"
)

func TestPracticeGuidanceStateTransitions(t *testing.T) {
	redisServer, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis.Run() error = %v", err)
	}
	defer redisServer.Close()

	svcCtx := &svc.ServiceContext{
		RedisClient: redisClientForTest(redisServer.Addr()),
	}
	defer svcCtx.RedisClient.Close()

	userID := int64(77)
	scope := ConversationScope{
		ChatID: "sess-practice-guidance",
		UserID: &userID,
		Mode:   sessionmode.KeyInterview,
	}
	sm := NewStateManager(context.Background(), svcCtx)

	first, err := sm.UpdatePracticeGuidance(scope, "不知道")
	if err != nil {
		t.Fatalf("UpdatePracticeGuidance(first) error = %v", err)
	}
	if first.StuckCount != 1 || first.LastSignal != interviewer.CandidateSignalStuck || first.TeachingMode {
		t.Fatalf("first guidance = %#v, want stuck_count=1 stuck signal teaching=false", first)
	}

	second, err := sm.UpdatePracticeGuidance(scope, "还是不会")
	if err != nil {
		t.Fatalf("UpdatePracticeGuidance(second) error = %v", err)
	}
	if second.StuckCount != 2 || second.LastSignal != interviewer.CandidateSignalStuck || second.TeachingMode {
		t.Fatalf("second guidance = %#v, want stuck_count=2 stuck signal teaching=false", second)
	}

	third, err := sm.UpdatePracticeGuidance(scope, "完全不会，没思路")
	if err != nil {
		t.Fatalf("UpdatePracticeGuidance(third) error = %v", err)
	}
	if third.StuckCount != 3 || !third.HelpOffered || third.LastSignal != interviewer.CandidateSignalStuck || third.TeachingMode {
		t.Fatalf("third guidance = %#v, want stuck_count=3 help_offered=true stuck signal teaching=false", third)
	}

	teaching, err := sm.UpdatePracticeGuidance(scope, "可以，详细讲一下")
	if err != nil {
		t.Fatalf("UpdatePracticeGuidance(teaching) error = %v", err)
	}
	if !teaching.TeachingMode || teaching.LastSignal != interviewer.CandidateSignalTeachingRequested {
		t.Fatalf("teaching guidance = %#v, want teaching mode with teaching_requested signal", teaching)
	}
}

func TestPracticeGuidanceDoesNotTreatGenericConsentAsTeachingBeforeOffer(t *testing.T) {
	redisServer, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis.Run() error = %v", err)
	}
	defer redisServer.Close()

	svcCtx := &svc.ServiceContext{
		RedisClient: redisClientForTest(redisServer.Addr()),
	}
	defer svcCtx.RedisClient.Close()

	userID := int64(77)
	scope := ConversationScope{
		ChatID: "sess-practice-generic-consent",
		UserID: &userID,
		Mode:   sessionmode.KeyInterview,
	}
	sm := NewStateManager(context.Background(), svcCtx)

	guidance, err := sm.UpdatePracticeGuidance(scope, "可以")
	if err != nil {
		t.Fatalf("UpdatePracticeGuidance() error = %v", err)
	}
	if guidance.TeachingMode || guidance.LastSignal != interviewer.CandidateSignalNone {
		t.Fatalf("guidance = %#v, want no teaching mode before help offer", guidance)
	}
}

func TestPracticeGuidanceSubstantiveAnswerResetsStuckCountBeforeTeaching(t *testing.T) {
	redisServer, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis.Run() error = %v", err)
	}
	defer redisServer.Close()

	svcCtx := &svc.ServiceContext{
		RedisClient: redisClientForTest(redisServer.Addr()),
	}
	defer svcCtx.RedisClient.Close()

	userID := int64(77)
	scope := ConversationScope{
		ChatID: "sess-practice-reset",
		UserID: &userID,
		Mode:   sessionmode.KeyInterview,
	}
	sm := NewStateManager(context.Background(), svcCtx)

	if _, err := sm.UpdatePracticeGuidance(scope, "不知道"); err != nil {
		t.Fatalf("UpdatePracticeGuidance(stuck) error = %v", err)
	}
	reset, err := sm.UpdatePracticeGuidance(scope, "我会先新建一套索引，按 embedding 版本隔离检索流量")
	if err != nil {
		t.Fatalf("UpdatePracticeGuidance(answer) error = %v", err)
	}
	if reset.StuckCount != 0 || reset.LastSignal != interviewer.CandidateSignalSubstantiveAnswer {
		t.Fatalf("reset guidance = %#v, want stuck_count=0 substantive signal", reset)
	}
}
