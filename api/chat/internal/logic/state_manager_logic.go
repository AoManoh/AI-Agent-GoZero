// Package logic 提供AI面试系统的业务逻辑处理层实现
// state_manager_logic.go 实现面试流程的状态管理和转换逻辑。
package logic

import (
	"context"
	"fmt"
	"strings"
	"time"

	"GoZero-AI/api/chat/internal/svc"
	"GoZero-AI/api/chat/internal/types"
	"GoZero-AI/internal/chatflow"
)

const maxStateEvents = int64(50)

const candidateEndIntentReason = "candidate_end_intent"

type StateManager struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

type ConversationScope struct {
	ChatID string
	UserID *int64
	Mode   string
}

type candidateIntent struct {
	End    bool
	Reason string
}

func NewStateManager(ctx context.Context, svcCtx *svc.ServiceContext) *StateManager {
	return &StateManager{ctx: ctx, svcCtx: svcCtx}
}

func (sm *StateManager) GetCurrentState(scope ConversationScope) (string, error) {
	snapshot, err := sm.GetFlowState(scope)
	if err != nil {
		return chatflow.InterviewStateStart, err
	}
	return snapshot.InterviewState, nil
}

func (sm *StateManager) GetFlowState(scope ConversationScope) (*chatflow.Snapshot, error) {
	key := chatflow.BuildContextKey(scope.ChatID, scope.UserID, scope.Mode)
	snapshot, err := chatflow.LoadSnapshot(sm.context(), sm.svcCtx.RedisClient, key, types.StateStart)
	if err != nil {
		return nil, fmt.Errorf("获取 flow state 失败: %w", err)
	}
	return &snapshot, nil
}

func (sm *StateManager) UpdateExecutionState(scope ConversationScope, executionState, reason string) (*chatflow.Snapshot, error) {
	key := chatflow.BuildContextKey(scope.ChatID, scope.UserID, scope.Mode)
	snapshot, err := chatflow.MutateSnapshot(sm.context(), sm.svcCtx.RedisClient, key, types.StateStart, maxStateEvents, func(snapshot chatflow.Snapshot) (chatflow.Snapshot, *chatflow.Event, error) {
		from := snapshot.ExecutionState
		snapshot.ExecutionState = executionState
		if executionState == chatflow.ExecutionFailed {
			snapshot.LifecycleState = chatflow.LifecycleErrored
		} else if snapshot.LifecycleState == "" || snapshot.LifecycleState == chatflow.LifecycleErrored {
			snapshot.LifecycleState = chatflow.LifecycleActive
		}
		snapshot.LastEvent = "execution." + executionState
		snapshot.LastReason = reason
		snapshot.UpdatedAt = time.Now().Format(time.RFC3339)

		return snapshot, &chatflow.Event{
			Type:   "execution",
			From:   from,
			To:     executionState,
			Reason: reason,
			At:     snapshot.UpdatedAt,
		}, nil
	})
	if err != nil {
		return nil, fmt.Errorf("加载 flow state 失败: %w", err)
	}
	return &snapshot, nil
}

func (sm *StateManager) RecordTurn(scope ConversationScope, role, reason string) error {
	key := chatflow.BuildContextKey(scope.ChatID, scope.UserID, scope.Mode)
	_, err := chatflow.MutateSnapshot(sm.context(), sm.svcCtx.RedisClient, key, types.StateStart, maxStateEvents, func(snapshot chatflow.Snapshot) (chatflow.Snapshot, *chatflow.Event, error) {
		if role == "user" || role == "assistant" {
			snapshot.TurnCount++
		}
		snapshot.LastEvent = "turn." + role
		snapshot.LastReason = reason
		snapshot.UpdatedAt = time.Now().Format(time.RFC3339)
		return snapshot, &chatflow.Event{
			Type:   "turn",
			Role:   role,
			Reason: reason,
			At:     snapshot.UpdatedAt,
		}, nil
	})
	if err != nil {
		return fmt.Errorf("加载 flow state 失败: %w", err)
	}
	return nil
}

func containsAny(s string, subStrings []string) bool {
	for _, sub := range subStrings {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

func detectCandidateIntent(message string) candidateIntent {
	compact := normalizeIntentText(message)
	if compact == "" {
		return candidateIntent{}
	}

	if containsAny(compact, []string{
		"不结束",
		"不要结束",
		"别结束",
		"先不结束",
		"还不结束",
		"不是结束",
		"不算结束",
		"不用结束",
		"不能结束",
		"没有结束",
	}) {
		return candidateIntent{}
	}

	if compact == "结束" || compact == "退出" || compact == "停止" {
		return candidateIntent{End: true, Reason: candidateEndIntentReason}
	}

	if containsAny(compact, []string{
		"我不想面试了",
		"不想面试了",
		"不想面试",
		"不面试了",
		"不面了",
		"结束面试",
		"结束本次面试",
		"结束吧",
		"到此为止",
		"先到这里",
		"先到这",
		"今天到这里",
		"今天先到",
		"退出面试",
		"停止面试",
		"中止面试",
		"不想继续了",
		"不继续了",
		"不想继续",
		"别问了",
		"不问了",
		"不聊了",
		"放弃面试",
		"面试到这里",
		"面试先到这里",
	}) {
		return candidateIntent{End: true, Reason: candidateEndIntentReason}
	}

	return candidateIntent{}
}

func normalizeIntentText(message string) string {
	normalized := strings.ToLower(strings.TrimSpace(message))
	replacer := strings.NewReplacer(
		" ", "",
		"\t", "",
		"\n", "",
		"\r", "",
		"。", "",
		"，", "",
		",", "",
		".", "",
		"！", "",
		"!", "",
	)
	return replacer.Replace(normalized)
}

func looksLikeOpeningQuestion(s string) bool {
	if containsAny(s, []string{
		"先来",
		"我们来",
		"请介绍",
		"介绍一下",
		"聊聊",
		"说说",
		"谈谈",
		"你提到",
		"问你",
		"问题",
	}) {
		return true
	}

	return strings.ContainsAny(s, "？?")
}

func looksLikeFollowUpSignal(s string) bool {
	return containsAny(s, []string{
		"追问",
		"详细说明",
		"为什么",
		"怎么实现",
		"接着讲讲",
		"继续讲讲",
		"展开一下",
		"具体说说",
		"具体讲讲",
		"你会怎么",
		"会怎么",
		"怎么避免",
		"怎么保证",
		"哪些操作",
		"哪些场景",
		"如果",
		"假设",
		"换个角度",
	})
}

func looksLikeEvaluationSignal(s string) bool {
	if containsAny(s, []string{"总结", "评估", "表现", "优缺点"}) {
		return !looksLikeFollowUpSignal(s) && !containsAny(s, []string{"继续说", "继续讲", "接着说", "接着讲"})
	}
	return false
}

func looksLikeCompletionSignal(s string) bool {
	if !containsAny(s, []string{"结束", "再见", "感谢参加"}) {
		return false
	}
	return !containsAny(s, []string{"不结束", "不算结束", "没有结束", "不是结束", "还不结束", "先不结束"})
}

func (sm *StateManager) TransitionState(currentState, aiRes string) string {
	nextState, _ := sm.TransitionStateDetailed(currentState, aiRes)
	return nextState
}

func (sm *StateManager) TransitionStateDetailed(currentState, aiRes string) (string, string) {
	lowerRes := strings.ToLower(aiRes)

	switch currentState {
	case types.StateStart:
		if containsAny(lowerRes, []string{"你好", "欢迎", "面试开始"}) {
			return types.StateQuestion, "welcome_signal"
		}
		if looksLikeOpeningQuestion(lowerRes) {
			return types.StateQuestion, "opening_question_signal"
		}
	case types.StateQuestion:
		if looksLikeFollowUpSignal(lowerRes) {
			return types.StateFollowUp, "follow_up_signal"
		}
		if looksLikeEvaluationSignal(lowerRes) {
			return types.StateEvaluate, "evaluation_signal"
		}
	case types.StateFollowUp:
		if looksLikeEvaluationSignal(lowerRes) {
			return types.StateEvaluate, "evaluation_signal"
		}
		if containsAny(lowerRes, []string{"下一个问题", "新问题", "换个主题", "另一个问题"}) {
			return types.StateQuestion, "next_question_signal"
		}
	case types.StateEvaluate:
		if looksLikeCompletionSignal(lowerRes) {
			return types.StateEnd, "completion_signal"
		}
		if containsAny(lowerRes, []string{"继续", "下一个问题"}) {
			return types.StateQuestion, "continue_signal"
		}
	case types.StateEnd:
	}

	return currentState, "no_transition"
}

func (sm *StateManager) EvaluateAndUpdateState(scope ConversationScope, aiResponse string) (*chatflow.Snapshot, error) {
	key := chatflow.BuildContextKey(scope.ChatID, scope.UserID, scope.Mode)
	transition := stateTransitionDecision{}
	if snapshot, err := sm.GetFlowState(scope); err == nil {
		transition = sm.decideTransition(snapshot.InterviewState, aiResponse)
	} else {
		transition = sm.decideTransition(types.StateStart, aiResponse)
		transition.FromState = ""
	}

	snapshot, err := chatflow.MutateSnapshot(sm.context(), sm.svcCtx.RedisClient, key, types.StateStart, maxStateEvents, func(snapshot chatflow.Snapshot) (chatflow.Snapshot, *chatflow.Event, error) {
		from := snapshot.InterviewState
		nextState := transition.NextState
		reason := transition.Reason
		if transition.FromState != from {
			nextState, reason = sm.TransitionStateDetailed(from, aiResponse)
			reason = "rule_after_state_changed_" + reason
		}
		snapshot.InterviewState = nextState
		if nextState == types.StateEnd {
			snapshot.LifecycleState = chatflow.LifecycleCompleted
		} else if snapshot.LifecycleState != chatflow.LifecycleErrored {
			snapshot.LifecycleState = chatflow.LifecycleActive
		}
		snapshot.ExecutionState = chatflow.ExecutionIdle
		if nextState != from {
			snapshot.LastEvent = "state.transition"
		} else {
			snapshot.LastEvent = "state.stable"
		}
		snapshot.LastReason = reason
		snapshot.UpdatedAt = time.Now().Format(time.RFC3339)

		return snapshot, &chatflow.Event{
			Type:   "state",
			From:   from,
			To:     nextState,
			Reason: reason,
			At:     snapshot.UpdatedAt,
		}, nil
	})
	if err != nil {
		return &snapshot, err
	}
	return &snapshot, nil
}

func (sm *StateManager) ApplyCandidateEndIntent(scope ConversationScope) (*chatflow.Snapshot, error) {
	key := chatflow.BuildContextKey(scope.ChatID, scope.UserID, scope.Mode)
	snapshot, err := chatflow.MutateSnapshot(sm.context(), sm.svcCtx.RedisClient, key, types.StateStart, maxStateEvents, func(snapshot chatflow.Snapshot) (chatflow.Snapshot, *chatflow.Event, error) {
		from := snapshot.InterviewState
		snapshot.InterviewState = types.StateEnd
		snapshot.LifecycleState = chatflow.LifecycleCompleted
		snapshot.ExecutionState = chatflow.ExecutionIdle
		if from == types.StateEnd {
			snapshot.LastEvent = "state.stable"
		} else {
			snapshot.LastEvent = "state.transition"
		}
		snapshot.LastReason = candidateEndIntentReason
		snapshot.UpdatedAt = time.Now().Format(time.RFC3339)

		return snapshot, &chatflow.Event{
			Type:   "state",
			From:   from,
			To:     types.StateEnd,
			Reason: candidateEndIntentReason,
			At:     snapshot.UpdatedAt,
		}, nil
	})
	if err != nil {
		return &snapshot, err
	}
	return &snapshot, nil
}

func (sm *StateManager) context() context.Context {
	if sm.ctx != nil {
		return sm.ctx
	}
	return context.Background()
}
