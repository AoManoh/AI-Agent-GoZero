package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"GoZero-AI/api/chat/internal/types"

	"github.com/sashabaranov/go-openai"
	"github.com/zeromicro/go-zero/core/logx"
)

const stateTransitionSystemPrompt = `你是 AI 面试系统的状态机判定器，只负责根据面试官刚生成的回复判断下一步 interview_state。

可选状态：
- start：尚未正式进入提问，只是初始化或寒暄。
- question：面试官提出了新的核心技术问题。
- follow_up：面试官基于候选人上一轮回答继续追问细节、原因、场景、取舍或线上处理。
- evaluate：面试官进入行为面试、综合素质考察、阶段性评价或复盘引导。
- end：面试官明确结束面试、告别或要求生成最终总结。

判定规则：
- 只看“面试官刚生成的回复”表达出的下一步动作。
- 对“那你接着讲讲”“如果线上...你会怎么处理”“具体怎么保证”这类自然追问，应判为 follow_up。
- 对“讲一次你处理冲突/推动方案/复盘失败的经历”这类行为题，应判为 evaluate。
- 不要因为回复中出现“总结一下你的回答”就轻易 end，只有明确结束才 end。
- 如果不确定，保持当前状态。

只输出 JSON，不要输出 markdown，不要解释。JSON 格式：
{"state":"question","reason":"简短中文原因","confidence":0.82}`

type stateTransitionDecision struct {
	FromState  string
	NextState  string
	Reason     string
	Source     string
	Confidence float64
}

type stateTransitionLLMResponse struct {
	State      string  `json:"state"`
	Reason     string  `json:"reason"`
	Confidence float64 `json:"confidence"`
}

func (sm *StateManager) decideTransition(currentState, aiResponse string) stateTransitionDecision {
	ruleNextState, ruleReason := sm.TransitionStateDetailed(currentState, aiResponse)
	decision := stateTransitionDecision{
		FromState: currentState,
		NextState: ruleNextState,
		Reason:    ruleReason,
		Source:    "rule",
	}

	if !sm.llmTransitionEnabled() {
		return decision
	}

	llmState, llmReason, confidence, err := sm.transitionStateByLLM(currentState, aiResponse)
	if err != nil {
		logx.WithContext(sm.context()).Errorf("状态转移模型失败，使用规则兜底: %v", err)
		decision.Reason = "llm_fallback_" + ruleReason
		return decision
	}

	return stateTransitionDecision{
		FromState:  currentState,
		NextState:  llmState,
		Reason:     formatLLMTransitionReason(llmReason, confidence),
		Source:     "llm",
		Confidence: confidence,
	}
}

func (sm *StateManager) llmTransitionEnabled() bool {
	return sm != nil &&
		sm.svcCtx != nil &&
		sm.svcCtx.Config.StateTransition.Enabled &&
		sm.svcCtx.StateTransitionClient != nil
}

func (sm *StateManager) transitionStateByLLM(currentState, aiResponse string) (string, string, float64, error) {
	timeout := time.Duration(sm.svcCtx.Config.StateTransitionTimeoutMillis()) * time.Millisecond
	ctx, cancel := context.WithTimeout(sm.context(), timeout)
	defer cancel()

	request := openai.ChatCompletionRequest{
		Model: sm.svcCtx.Config.StateTransitionModel(),
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: stateTransitionSystemPrompt,
			},
			{
				Role: openai.ChatMessageRoleUser,
				Content: fmt.Sprintf(
					"当前状态：%s\n允许下一状态：%s\n面试官刚生成的回复：\n%s",
					currentState,
					strings.Join(allowedNextStates(currentState), ", "),
					truncateForStateTransition(aiResponse, 1200),
				),
			},
		},
		MaxCompletionTokens: sm.svcCtx.Config.StateTransitionMaxCompletionTokens(),
		Temperature:         sm.svcCtx.Config.StateTransitionTemperature(),
		ReasoningEffort:     sm.svcCtx.Config.StateTransitionReasoningEffort(),
	}

	resp, err := sm.svcCtx.StateTransitionClient.CreateChatCompletion(ctx, request)
	if err != nil {
		return "", "", 0, err
	}
	if len(resp.Choices) == 0 {
		return "", "", 0, fmt.Errorf("状态转移模型未返回结果")
	}

	parsed, err := parseStateTransitionPayload(resp.Choices[0].Message.Content)
	if err != nil {
		return "", "", 0, err
	}
	nextState := normalizeInterviewState(parsed.State)
	if !isAllowedStateTransition(currentState, nextState) {
		return "", "", 0, fmt.Errorf("状态转移模型返回非法状态: current=%s next=%s", currentState, parsed.State)
	}
	if parsed.Confidence > 0 && parsed.Confidence < 0.35 {
		return "", "", parsed.Confidence, fmt.Errorf("状态转移模型置信度过低: %.2f", parsed.Confidence)
	}

	reason := strings.TrimSpace(parsed.Reason)
	if reason == "" {
		reason = "semantic_state_transition"
	}
	return nextState, reason, parsed.Confidence, nil
}

func parseStateTransitionPayload(content string) (stateTransitionLLMResponse, error) {
	var parsed stateTransitionLLMResponse
	payload := sanitizeJSONPayload(content)
	if err := json.Unmarshal([]byte(payload), &parsed); err != nil {
		return parsed, fmt.Errorf("解析状态转移 JSON 失败: %w", err)
	}
	return parsed, nil
}

func sanitizeJSONPayload(content string) string {
	trimmed := strings.TrimSpace(content)
	trimmed = strings.TrimPrefix(trimmed, "```json")
	trimmed = strings.TrimPrefix(trimmed, "```")
	trimmed = strings.TrimSuffix(trimmed, "```")
	return strings.TrimSpace(trimmed)
}

func normalizeInterviewState(state string) string {
	normalized := strings.ToLower(strings.TrimSpace(state))
	normalized = strings.ReplaceAll(normalized, "-", "_")
	normalized = strings.ReplaceAll(normalized, " ", "_")
	switch normalized {
	case types.StateStart, "开始":
		return types.StateStart
	case types.StateQuestion, "ask", "asking", "core_question", "问题", "核心问题", "提问":
		return types.StateQuestion
	case types.StateFollowUp, "followup", "follow_up_question", "probe", "deep_dive", "追问", "深挖":
		return types.StateFollowUp
	case types.StateEvaluate, "evaluation", "behavior", "behavioral", "评估", "行为面试":
		return types.StateEvaluate
	case types.StateEnd, "done", "finish", "finished", "complete", "completed", "结束":
		return types.StateEnd
	default:
		return normalized
	}
}

func allowedNextStates(currentState string) []string {
	switch currentState {
	case types.StateStart:
		return []string{types.StateStart, types.StateQuestion}
	case types.StateQuestion:
		return []string{types.StateQuestion, types.StateFollowUp, types.StateEvaluate, types.StateEnd}
	case types.StateFollowUp:
		return []string{types.StateFollowUp, types.StateQuestion, types.StateEvaluate, types.StateEnd}
	case types.StateEvaluate:
		return []string{types.StateEvaluate, types.StateQuestion, types.StateEnd}
	case types.StateEnd:
		return []string{types.StateEnd}
	default:
		return []string{types.StateStart, types.StateQuestion, types.StateFollowUp, types.StateEvaluate, types.StateEnd}
	}
}

func isAllowedStateTransition(currentState, nextState string) bool {
	if nextState == "" {
		return false
	}
	for _, state := range allowedNextStates(currentState) {
		if state == nextState {
			return true
		}
	}
	return false
}

func formatLLMTransitionReason(reason string, confidence float64) string {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = "semantic_state_transition"
	}
	if len([]rune(reason)) > 80 {
		reason = string([]rune(reason)[:80])
	}
	if confidence > 0 {
		return fmt.Sprintf("llm_%s_%.2f", reason, confidence)
	}
	return "llm_" + reason
}

func truncateForStateTransition(text string, maxRunes int) string {
	if maxRunes <= 0 {
		return ""
	}
	runes := []rune(strings.TrimSpace(text))
	if len(runes) <= maxRunes {
		return string(runes)
	}
	return string(runes[:maxRunes]) + "..."
}
