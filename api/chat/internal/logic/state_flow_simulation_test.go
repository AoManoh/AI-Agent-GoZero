package logic

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"GoZero-AI/api/chat/internal/config"
	"GoZero-AI/api/chat/internal/svc"
	"GoZero-AI/api/chat/internal/types"

	"github.com/sashabaranov/go-openai"
)

func TestInterviewStateFlowSimulationRecords(t *testing.T) {
	sm := &StateManager{}
	rounds := []struct {
		name       string
		input      string
		fromState  string
		reply      string
		wantState  string
		wantReason string
	}{
		{
			name:       "start -> question",
			input:      "可以开始面试。",
			fromState:  types.StateStart,
			reply:      "你好，欢迎来到今天的 Go 后端面试。我们先来聊一个问题：你怎么理解 context 在服务超时控制里的作用？",
			wantState:  types.StateQuestion,
			wantReason: "welcome_signal",
		},
		{
			name:       "question -> follow_up",
			input:      "我会用 context 控制 goroutine，主流程取消后子任务也退出。",
			fromState:  types.StateQuestion,
			reply:      "你提到子任务退出，为什么只传 context 还不够，具体怎么保证数据库调用和 goroutine 都能及时释放？",
			wantState:  types.StateFollowUp,
			wantReason: "follow_up_signal",
		},
		{
			name:       "follow_up -> question",
			input:      "我会加超时和错误返回，并观察 goroutine 数量。",
			fromState:  types.StateFollowUp,
			reply:      "我们进入下一个问题，聊聊数据库连接池在高并发下的等待队列和超时设置。",
			wantState:  types.StateQuestion,
			wantReason: "next_question_signal",
		},
		{
			name:       "question -> evaluate",
			input:      "这块我没有线上经验，只看过一些资料。",
			fromState:  types.StateQuestion,
			reply:      "我们做个阶段性评估，总结一下你目前的表现和优缺点，再看后续怎么补强。",
			wantState:  types.StateEvaluate,
			wantReason: "evaluation_signal",
		},
		{
			name:       "follow_up -> evaluate",
			input:      "这个追问我暂时回答不上来。",
			fromState:  types.StateFollowUp,
			reply:      "我先做个总结和评估，再看你的表现：你能说出方向，但缺少可验证的项目证据。",
			wantState:  types.StateEvaluate,
			wantReason: "evaluation_signal",
		},
		{
			name:       "evaluate -> end",
			input:      "好的，不需要继续了。",
			fromState:  types.StateEvaluate,
			reply:      "今天的面试就到这里，感谢参加。",
			wantState:  types.StateEnd,
			wantReason: "completion_signal",
		},
	}

	for _, round := range rounds {
		t.Run(round.name, func(t *testing.T) {
			gotState, gotReason := sm.TransitionStateDetailed(round.fromState, round.reply)
			if gotState != round.wantState || gotReason != round.wantReason {
				t.Fatalf("input=%q reply=%q got=(%s,%s), want=(%s,%s)", round.input, round.reply, gotState, gotReason, round.wantState, round.wantReason)
			}
		})
	}
}

func TestInterviewStateFlowAmbiguousPhrases(t *testing.T) {
	sm := &StateManager{}
	rounds := []struct {
		name       string
		fromState  string
		reply      string
		wantState  string
		wantReason string
	}{
		{
			name:       "summary phrase with why remains follow up",
			fromState:  types.StateQuestion,
			reply:      "你先总结一下刚才的答案，然后继续说为什么 context 取消不一定能中断数据库调用？",
			wantState:  types.StateFollowUp,
			wantReason: "follow_up_signal",
		},
		{
			name:       "negated end in evaluate should continue instead of end",
			fromState:  types.StateEvaluate,
			reply:      "这个问题不结束，我们继续下一个点，聊聊你怎么补齐线上验证证据。",
			wantState:  types.StateQuestion,
			wantReason: "continue_signal",
		},
		{
			name:       "follow up and next question both present prefers explicit next question",
			fromState:  types.StateFollowUp,
			reply:      "这个追问到这里，我们进入下一个问题：为什么连接池等待队列会影响接口 P95？",
			wantState:  types.StateQuestion,
			wantReason: "next_question_signal",
		},
		{
			name:       "candidate asks to end but interviewer keeps probing",
			fromState:  types.StateQuestion,
			reply:      "你说想结束，但我再追问一个点：为什么线上只看平均延迟会掩盖尾延迟问题？",
			wantState:  types.StateFollowUp,
			wantReason: "follow_up_signal",
		},
		{
			name:       "negated end without continue stays evaluate",
			fromState:  types.StateEvaluate,
			reply:      "这不是结束，我只是先记录一下你的阶段表现。",
			wantState:  types.StateEvaluate,
			wantReason: "no_transition",
		},
	}

	for _, round := range rounds {
		t.Run(round.name, func(t *testing.T) {
			gotState, gotReason := sm.TransitionStateDetailed(round.fromState, round.reply)
			if gotState != round.wantState || gotReason != round.wantReason {
				t.Fatalf("reply=%q got=(%s,%s), want=(%s,%s)", round.reply, gotState, gotReason, round.wantState, round.wantReason)
			}
		})
	}
}

func TestDecideTransitionFallsBackToRulesWhenLLMRequestFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "state transition model unavailable", http.StatusInternalServerError)
	}))
	defer server.Close()

	clientConfig := openai.DefaultConfig("test-key")
	clientConfig.BaseURL = strings.TrimRight(server.URL, "/") + "/v1"

	sm := &StateManager{
		svcCtx: &svc.ServiceContext{
			Config: config.Config{
				OpenAI: config.OpenAIConfig{
					Model: "mock-chat-model",
				},
				StateTransition: config.StateTransitionConfig{
					Enabled:             true,
					Model:               "mock-state-model",
					MaxCompletionTokens: 32,
					TimeoutMillis:       500,
				},
			},
			StateTransitionClient: openai.NewClientWithConfig(clientConfig),
		},
	}

	decision := sm.decideTransition(types.StateQuestion, "那你接着讲讲，具体怎么保证取消信号能传到所有 goroutine？")
	if decision.NextState != types.StateFollowUp {
		t.Fatalf("NextState = %q, want %q", decision.NextState, types.StateFollowUp)
	}
	if decision.Source != "rule" {
		t.Fatalf("Source = %q, want rule", decision.Source)
	}
	if decision.Reason != "llm_fallback_follow_up_signal" {
		t.Fatalf("Reason = %q, want llm_fallback_follow_up_signal", decision.Reason)
	}
}

func TestDecideTransitionLLMPaths(t *testing.T) {
	tests := []struct {
		name       string
		response   string
		statusCode int
		current    string
		reply      string
		wantState  string
		wantSource string
		wantReason string
	}{
		{
			name:       "legal json wins even when it conflicts with rules",
			response:   stateTransitionMockResponse(`{"state":"evaluate","reason":"行为面试判断","confidence":0.91}`),
			statusCode: http.StatusOK,
			current:    types.StateQuestion,
			reply:      "为什么你会这么设计？",
			wantState:  types.StateEvaluate,
			wantSource: "llm",
			wantReason: "llm_行为面试判断_0.91",
		},
		{
			name:       "invalid json falls back to rules",
			response:   stateTransitionMockResponse(`not json`),
			statusCode: http.StatusOK,
			current:    types.StateQuestion,
			reply:      "为什么只传 context 还不够？",
			wantState:  types.StateFollowUp,
			wantSource: "rule",
			wantReason: "llm_fallback_follow_up_signal",
		},
		{
			name:       "illegal state falls back to rules",
			response:   stateTransitionMockResponse(`{"state":"archived","reason":"错误状态","confidence":0.88}`),
			statusCode: http.StatusOK,
			current:    types.StateStart,
			reply:      "你好，欢迎来到今天的面试，我们先聊聊数据库索引。",
			wantState:  types.StateQuestion,
			wantSource: "rule",
			wantReason: "llm_fallback_welcome_signal",
		},
		{
			name:       "low confidence falls back to rules",
			response:   stateTransitionMockResponse(`{"state":"evaluate","reason":"不确定","confidence":0.21}`),
			statusCode: http.StatusOK,
			current:    types.StateQuestion,
			reply:      "那你接着讲讲，具体怎么验证？",
			wantState:  types.StateFollowUp,
			wantSource: "rule",
			wantReason: "llm_fallback_follow_up_signal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.response))
			}))
			defer server.Close()

			sm := newStateTransitionTestManager(server.URL)
			decision := sm.decideTransition(tt.current, tt.reply)
			if decision.NextState != tt.wantState || decision.Source != tt.wantSource || decision.Reason != tt.wantReason {
				t.Fatalf("decision=(state:%s source:%s reason:%s), want=(state:%s source:%s reason:%s)", decision.NextState, decision.Source, decision.Reason, tt.wantState, tt.wantSource, tt.wantReason)
			}
		})
	}
}

func newStateTransitionTestManager(serverURL string) *StateManager {
	clientConfig := openai.DefaultConfig("test-key")
	clientConfig.BaseURL = strings.TrimRight(serverURL, "/") + "/v1"

	return &StateManager{
		svcCtx: &svc.ServiceContext{
			Config: config.Config{
				OpenAI: config.OpenAIConfig{
					Model: "mock-chat-model",
				},
				StateTransition: config.StateTransitionConfig{
					Enabled:             true,
					Model:               "mock-state-model",
					MaxCompletionTokens: 32,
					TimeoutMillis:       500,
				},
			},
			StateTransitionClient: openai.NewClientWithConfig(clientConfig),
		},
	}
}

func stateTransitionMockResponse(content string) string {
	payload := map[string]any{
		"id":      "chatcmpl-state-test",
		"object":  "chat.completion",
		"created": 1710000000,
		"model":   "mock-state-model",
		"choices": []map[string]any{
			{
				"index": 0,
				"message": map[string]any{
					"role":    "assistant",
					"content": content,
				},
				"finish_reason": "stop",
			},
		},
	}
	raw, _ := json.Marshal(payload)
	return string(raw)
}
