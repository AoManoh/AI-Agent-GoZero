package logic

import (
	"context"
	"net/http"
	"testing"

	"GoZero-AI/api/chat/internal/svc"
	"GoZero-AI/api/chat/internal/types"
	"GoZero-AI/internal/chatflow"
	"GoZero-AI/internal/statuserr"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestTransitionStateDetailedFromStart(t *testing.T) {
	sm := &StateManager{}

	tests := []struct {
		name       string
		reply      string
		wantState  string
		wantReason string
	}{
		{
			name:       "welcome phrase transitions to question",
			reply:      "你好，欢迎来到今天的 Go 后端面试。",
			wantState:  types.StateQuestion,
			wantReason: "welcome_signal",
		},
		{
			name:       "opening question without welcome still transitions",
			reply:      "嗯，好的，我们来看看这个问题，你提到简历里深入理解 Go 的 G，先聊聊 GMP 调度模型。",
			wantState:  types.StateQuestion,
			wantReason: "opening_question_signal",
		},
		{
			name:       "opening interview prompt using explain phrase transitions",
			reply:      "好，那我们直接从项目切入。你做的 GoZero 订单服务，先讲一下订单创建接口的请求链路和幂等处理。",
			wantState:  types.StateQuestion,
			wantReason: "opening_question_signal",
		},
		{
			name:       "question mark transitions to question",
			reply:      "我们直接开始：你怎么理解 Go 的并发模型？",
			wantState:  types.StateQuestion,
			wantReason: "opening_question_signal",
		},
		{
			name:       "plain text without signals stays in start",
			reply:      "收到，我会结合你的上下文来继续。",
			wantState:  types.StateStart,
			wantReason: "no_transition",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotState, gotReason := sm.TransitionStateDetailed(types.StateStart, tt.reply)
			if gotState != tt.wantState || gotReason != tt.wantReason {
				t.Fatalf("TransitionStateDetailed(start, %q) = (%q, %q), want (%q, %q)", tt.reply, gotState, gotReason, tt.wantState, tt.wantReason)
			}
		})
	}
}

func TestTransitionStateDetailedFromQuestion(t *testing.T) {
	sm := &StateManager{}

	tests := []struct {
		name       string
		reply      string
		wantState  string
		wantReason string
	}{
		{
			name:       "follow up signal",
			reply:      "为什么这么设计？请详细说明具体实现。",
			wantState:  types.StateFollowUp,
			wantReason: "follow_up_signal",
		},
		{
			name:       "natural follow up phrase",
			reply:      "那你接着讲讲，P 为什么不能简单省掉？",
			wantState:  types.StateFollowUp,
			wantReason: "follow_up_signal",
		},
		{
			name:       "scenario follow up phrase",
			reply:      "那如果下游阻塞且不支持 ctx，你会怎么避免 goroutine 泄漏？",
			wantState:  types.StateFollowUp,
			wantReason: "follow_up_signal",
		},
		{
			name:       "natural implementation follow up",
			reply:      "这个方向可以，你再往下讲机制：本地消息表长期 pending 时，你怎么定位是库锁还是 MQ 故障？",
			wantState:  types.StateFollowUp,
			wantReason: "follow_up_signal",
		},
		{
			name:       "evaluation signal",
			reply:      "我们做个阶段性评估，总结一下你的优缺点。",
			wantState:  types.StateEvaluate,
			wantReason: "evaluation_signal",
		},
		{
			name:       "summary and explicit finish completes from question",
			reply:      "我做一个阶段性总结并结束这场面试。",
			wantState:  types.StateEnd,
			wantReason: "completion_signal",
		},
		{
			name:       "technical end word stays follow up",
			reply:      "请求结束后 goroutine 怎么释放？",
			wantState:  types.StateFollowUp,
			wantReason: "follow_up_signal",
		},
		{
			name:       "no transition",
			reply:      "好的，我继续听你展开。",
			wantState:  types.StateQuestion,
			wantReason: "no_transition",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotState, gotReason := sm.TransitionStateDetailed(types.StateQuestion, tt.reply)
			if gotState != tt.wantState || gotReason != tt.wantReason {
				t.Fatalf("TransitionStateDetailed(question, %q) = (%q, %q), want (%q, %q)", tt.reply, gotState, gotReason, tt.wantState, tt.wantReason)
			}
		})
	}
}

func TestTransitionStateDetailedFromFollowUp(t *testing.T) {
	sm := &StateManager{}

	tests := []struct {
		name       string
		reply      string
		wantState  string
		wantReason string
	}{
		{
			name:       "evaluation after follow up",
			reply:      "我先做个总结和评估，再看你的表现。",
			wantState:  types.StateEvaluate,
			wantReason: "evaluation_signal",
		},
		{
			name:       "new question after follow up",
			reply:      "我们进入下一个问题，聊聊 channel 和 mutex 的选择。",
			wantState:  types.StateQuestion,
			wantReason: "next_question_signal",
		},
		{
			name:       "summary and explicit finish completes from follow up",
			reply:      "这场面试我就先结束在这里，下面给你阶段性总结。",
			wantState:  types.StateEnd,
			wantReason: "completion_signal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotState, gotReason := sm.TransitionStateDetailed(types.StateFollowUp, tt.reply)
			if gotState != tt.wantState || gotReason != tt.wantReason {
				t.Fatalf("TransitionStateDetailed(follow_up, %q) = (%q, %q), want (%q, %q)", tt.reply, gotState, gotReason, tt.wantState, tt.wantReason)
			}
		})
	}
}

func TestTransitionStateDetailedFromEvaluate(t *testing.T) {
	sm := &StateManager{}

	tests := []struct {
		name       string
		reply      string
		wantState  string
		wantReason string
	}{
		{
			name:       "completion signal",
			reply:      "今天的面试就到这里，感谢参加。",
			wantState:  types.StateEnd,
			wantReason: "completion_signal",
		},
		{
			name:       "continue signal",
			reply:      "我们继续，下一个问题聊聊 GC 调优。",
			wantState:  types.StateQuestion,
			wantReason: "continue_signal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotState, gotReason := sm.TransitionStateDetailed(types.StateEvaluate, tt.reply)
			if gotState != tt.wantState || gotReason != tt.wantReason {
				t.Fatalf("TransitionStateDetailed(evaluate, %q) = (%q, %q), want (%q, %q)", tt.reply, gotState, gotReason, tt.wantState, tt.wantReason)
			}
		})
	}
}

func TestDetectCandidateIntentEnd(t *testing.T) {
	tests := []struct {
		name    string
		message string
		wantEnd bool
	}{
		{name: "explicit quit interview", message: "我不想面试了", wantEnd: true},
		{name: "end interview", message: "结束面试吧", wantEnd: true},
		{name: "stop here", message: "今天先到这里", wantEnd: true},
		{name: "do not continue", message: "不继续了", wantEnd: true},
		{name: "single word end", message: "结束", wantEnd: true},
		{name: "negated end", message: "先不结束，我们继续", wantEnd: false},
		{name: "technical end word", message: "请求结束后 goroutine 怎么释放？", wantEnd: false},
		{name: "normal answer", message: "我会先看 context 是否传到数据库调用。", wantEnd: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectCandidateIntent(tt.message)
			if got.End != tt.wantEnd {
				t.Fatalf("detectCandidateIntent(%q).End = %v, want %v", tt.message, got.End, tt.wantEnd)
			}
			if got.End && got.Reason != candidateEndIntentReason {
				t.Fatalf("detectCandidateIntent(%q).Reason = %q, want %q", tt.message, got.Reason, candidateEndIntentReason)
			}
		})
	}
}

func TestApplyCandidateEndIntentUpdatesFlowState(t *testing.T) {
	redisServer, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis.Run() error = %v", err)
	}
	defer redisServer.Close()

	client := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	defer client.Close()

	sm := &StateManager{
		ctx: context.Background(),
		svcCtx: &svc.ServiceContext{
			RedisClient: client,
		},
	}
	scope := ConversationScope{ChatID: "candidate-end-session", Mode: "interview"}
	snapshot, err := sm.ApplyCandidateEndIntent(scope)
	if err != nil {
		t.Fatalf("ApplyCandidateEndIntent() error = %v", err)
	}
	if snapshot.InterviewState != types.StateEnd {
		t.Fatalf("InterviewState = %q, want %q", snapshot.InterviewState, types.StateEnd)
	}
	if snapshot.LifecycleState != chatflow.LifecycleCompleted {
		t.Fatalf("LifecycleState = %q, want %q", snapshot.LifecycleState, chatflow.LifecycleCompleted)
	}
	if snapshot.ExecutionState != chatflow.ExecutionIdle {
		t.Fatalf("ExecutionState = %q, want %q", snapshot.ExecutionState, chatflow.ExecutionIdle)
	}
	if snapshot.LastReason != candidateEndIntentReason {
		t.Fatalf("LastReason = %q, want %q", snapshot.LastReason, candidateEndIntentReason)
	}
}

func TestTryAcquireExecutionLockRejectsConcurrentSession(t *testing.T) {
	client := newStateManagerRedisClient(t)
	defer client.Close()

	sm := &StateManager{
		ctx: context.Background(),
		svcCtx: &svc.ServiceContext{
			RedisClient: client,
		},
	}
	scope := ConversationScope{ChatID: "lock-session", Mode: "interview"}
	release, err := sm.TryAcquireExecutionLock(scope)
	if err != nil {
		t.Fatalf("TryAcquireExecutionLock() first error = %v", err)
	}
	defer release()

	_, err = sm.TryAcquireExecutionLock(scope)
	if err == nil {
		t.Fatal("TryAcquireExecutionLock() second error = nil, want conflict")
	}
	code, ok := statuserr.StatusCode(err)
	if !ok || code != http.StatusConflict {
		t.Fatalf("status = %d ok=%v, want 409/true; err=%v", code, ok, err)
	}
}

func newStateManagerRedisClient(t *testing.T) *redis.Client {
	t.Helper()

	server, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis.Run() error = %v", err)
	}
	t.Cleanup(server.Close)

	return redis.NewClient(&redis.Options{Addr: server.Addr()})
}
