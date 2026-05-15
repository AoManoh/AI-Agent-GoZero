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
	"GoZero-AI/internal/statuserr"

	"github.com/redis/go-redis/v9"
)

const maxStateEvents = int64(50)
const executionLockTTL = 10 * time.Minute

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
		if snapshot.LifecycleState == chatflow.LifecycleCompleted || snapshot.InterviewState == chatflow.InterviewStateEnd {
			return snapshot, nil, nil
		}
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

func (sm *StateManager) TryAcquireExecutionLock(scope ConversationScope) (func(), error) {
	if sm.svcCtx == nil || sm.svcCtx.RedisClient == nil {
		return func() {}, nil
	}

	key := chatflow.BuildContextKey(scope.ChatID, scope.UserID, scope.Mode)
	lockKey := chatflow.ExecutionLockRedisKey(key)
	token := fmt.Sprintf("%s:%d", key.SessionID, time.Now().UnixNano())
	acquired, err := sm.svcCtx.RedisClient.SetNX(sm.context(), lockKey, token, executionLockTTL).Result()
	if err != nil {
		return nil, fmt.Errorf("获取会话执行锁失败: %w", err)
	}
	if !acquired {
		return nil, statuserr.Conflict("当前会话正在生成回复，请稍后再试")
	}

	return func() {
		releaseCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = sm.svcCtx.RedisClient.Watch(releaseCtx, func(tx *redis.Tx) error {
			value, err := tx.Get(releaseCtx, lockKey).Result()
			if err == redis.Nil {
				return nil
			}
			if err != nil {
				return err
			}
			if value != token {
				return nil
			}
			_, err = tx.TxPipelined(releaseCtx, func(pipe redis.Pipeliner) error {
				pipe.Del(releaseCtx, lockKey)
				return nil
			})
			return err
		}, lockKey)
	}, nil
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
	if strings.ContainsAny(s, "？?") {
		return true
	}
	return looksLikeCandidateFacingPrompt(s, true)
}

func looksLikeCandidateFacingPrompt(s string, allowOpeningTopicTask bool) bool {
	for _, clause := range splitPromptClauses(s) {
		if isCandidateDirectedPromptClause(clause) || isExplicitAnswerRequestClause(clause) {
			return true
		}
		if allowOpeningTopicTask && isOpeningTopicTaskClause(clause) {
			return true
		}
	}
	return false
}

func splitPromptClauses(s string) []string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return strings.ContainsRune("。！？!?；;，,\n\r：:", r)
	})
	clauses := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			clauses = append(clauses, part)
		}
	}
	return clauses
}

func isCandidateDirectedPromptClause(clause string) bool {
	if isAssistantSelfExplanationClause(clause) {
		return false
	}
	if hasCandidateQuestionSubjectCue(clause) && hasQuestionCue(clause) {
		return true
	}
	return hasCandidateDirectiveSubjectCue(clause) && hasAnswerActionCue(clause)
}

func isExplicitAnswerRequestClause(clause string) bool {
	if isAssistantSelfExplanationClause(clause) {
		return false
	}
	return hasAnswerRequestCue(clause) && (hasQuestionCue(clause) || hasAnswerActionCue(clause))
}

func isOpeningTopicTaskClause(clause string) bool {
	if isAssistantSelfExplanationClause(clause) {
		return false
	}
	if containsAny(clause, []string{
		"我们来看看这个问题",
		"我们来看这个问题",
		"我们先看",
		"我们先聊",
		"我们从",
		"直接从",
		"先来",
		"先聊聊",
		"先说说",
		"先讲一下",
		"先讲讲",
		"先谈谈",
		"第一题",
		"第一个问题",
		"进入第一个问题",
	}) {
		return true
	}
	if hasAnswerRequestCue(clause) && containsAny(clause, []string{"先", "我们", "请", "问题", "场景", "项目"}) {
		return true
	}
	return false
}

func isAssistantSelfExplanationClause(clause string) bool {
	return containsAny(clause, []string{
		"我会",
		"我先",
		"我来",
		"我可以",
		"下面我",
		"接下来我",
	})
}

func hasCandidateQuestionSubjectCue(clause string) bool {
	return containsAny(clause, []string{
		"你会",
		"你具体",
		"你怎么",
		"你如何",
		"你先",
		"你第一",
		"你认为",
		"你觉得",
		"你理解",
		"你打算",
		"你准备",
		"你要",
		"你可以",
		"你在",
		"如果你",
		"假设你",
		"你能否",
		"你能不能",
	})
}

func hasCandidateDirectiveSubjectCue(clause string) bool {
	return containsAny(clause, []string{
		"你先",
		"请你",
		"让你",
		"给你",
	})
}

func hasQuestionCue(clause string) bool {
	return containsAny(clause, []string{
		"怎么",
		"如何",
		"什么",
		"哪些",
		"哪个",
		"为什么",
		"是否",
		"能否",
		"能不能",
		"可不可以",
		"有没有",
		"哪里",
		"多少",
		"多大",
		"该不该",
	})
}

func hasAnswerActionCue(clause string) bool {
	return containsAny(clause, []string{
		"回答",
		"说明",
		"解释",
		"分析",
		"描述",
		"介绍",
		"讲",
		"说",
		"谈",
		"阐述",
		"拆解",
		"设计",
		"定位",
		"排查",
		"处理",
		"选择",
		"保证",
		"优化",
		"实现",
		"验证",
		"复盘",
		"止损",
		"发现",
		"判断",
	})
}

func hasAnswerRequestCue(clause string) bool {
	return containsAny(clause, []string{
		"请说说",
		"请讲讲",
		"请谈谈",
		"请介绍",
		"请解释",
		"请说明",
		"请分析",
		"请描述",
		"请回答",
		"请展开",
		"麻烦你",
		"先说说",
		"先讲讲",
		"先介绍",
		"先解释",
		"先说明",
		"先分析",
		"说说",
		"聊聊",
		"谈谈",
		"讲一下",
		"讲讲",
		"介绍一下",
		"解释一下",
		"说明一下",
		"分析一下",
		"展开一下",
	})
}

func looksLikeFollowUpSignal(s string) bool {
	if containsAny(s, []string{
		"追问",
		"详细说明",
		"为什么",
		"为什么先",
		"怎么实现",
		"接着讲讲",
		"继续讲讲",
		"继续说说",
		"继续说",
		"继续讲",
		"再往下讲",
		"往下讲",
		"展开一下",
		"展开讲讲",
		"具体说说",
		"具体讲讲",
		"刚才",
		"上一轮",
		"你刚才",
		"你提到",
		"只说",
		"收敛",
		"第一眼",
		"单一指标",
		"你会怎么",
		"会怎么",
		"怎么做",
		"怎么控制",
		"怎么定位",
		"怎么排查",
		"怎么处理",
		"怎么释放",
		"怎么避免",
		"怎么保证",
		"哪些操作",
		"哪些场景",
		"如果",
		"假设",
		"换个角度",
	}) {
		return true
	}
	return looksLikeCandidateFacingPrompt(s, false)
}

func looksLikeNextQuestionSignal(s string) bool {
	if containsAny(s, []string{
		"不换题",
		"不换主题",
		"不是下一个问题",
		"不是新问题",
	}) {
		return false
	}
	return containsAny(s, []string{
		"下一个问题",
		"下一题",
		"新问题",
		"换个主题",
		"换一个主题",
		"换到",
		"另一个问题",
		"再看一个问题",
		"接下来聊",
		"下一个点",
		"换一个点",
	})
}

func looksLikeEvaluationSignal(s string) bool {
	if containsAny(s, []string{"总结", "评估", "表现", "优缺点"}) {
		return !looksLikeFollowUpSignal(s) && !containsAny(s, []string{"继续说", "继续讲", "接着说", "接着讲"})
	}
	return false
}

func looksLikeCompletionSignal(s string) bool {
	if containsAny(s, []string{"不结束", "不算结束", "没有结束", "不是结束", "还不结束", "先不结束"}) {
		return false
	}
	return containsAny(s, []string{
		"结束这场面试",
		"结束本次面试",
		"结束今天的面试",
		"本次面试结束",
		"这场面试结束",
		"面试结束",
		"面试就到这里",
		"面试先到这里",
		"就先结束在这里",
		"先结束在这里",
		"总结并结束",
		"阶段性总结并结束",
		"今天就到这里",
		"今天先到这里",
		"今天的面试就到这里",
		"感谢参加",
		"感谢你参加",
		"再见",
	})
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
		if looksLikeCompletionSignal(lowerRes) {
			return types.StateEnd, "completion_signal"
		}
		if looksLikeFollowUpSignal(lowerRes) {
			return types.StateFollowUp, "follow_up_signal"
		}
		if looksLikeEvaluationSignal(lowerRes) {
			return types.StateEvaluate, "evaluation_signal"
		}
		if looksLikeNextQuestionSignal(lowerRes) {
			return types.StateQuestion, "next_question_signal"
		}
	case types.StateFollowUp:
		if looksLikeCompletionSignal(lowerRes) {
			return types.StateEnd, "completion_signal"
		}
		if looksLikeEvaluationSignal(lowerRes) {
			return types.StateEvaluate, "evaluation_signal"
		}
		if looksLikeNextQuestionSignal(lowerRes) {
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
