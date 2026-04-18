package logic

import (
	"testing"

	"GoZero-AI/api/chat/internal/types"
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
			name:       "evaluation signal",
			reply:      "我们做个阶段性评估，总结一下你的优缺点。",
			wantState:  types.StateEvaluate,
			wantReason: "evaluation_signal",
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
