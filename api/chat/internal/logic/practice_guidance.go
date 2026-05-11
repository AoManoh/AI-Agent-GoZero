package logic

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"GoZero-AI/api/chat/internal/interviewer"
	"GoZero-AI/api/chat/internal/svc"
	"GoZero-AI/internal/chatflow"

	"github.com/redis/go-redis/v9"
)

const (
	practiceGuidanceKeyPrefix        = "chat_practice_guidance:v1:"
	formalInterviewGuidanceKeyPrefix = "chat_formal_interview_guidance:v1:"
)

type practiceGuidanceSnapshot struct {
	Scenario      string `json:"scenario"`
	StuckCount    int    `json:"stuckCount"`
	HelpOffered   bool   `json:"helpOffered"`
	TeachingMode  bool   `json:"teachingMode"`
	LastSignal    string `json:"lastSignal"`
	LastMessageAt string `json:"lastMessageAt"`
}

func (sm *StateManager) UpdatePracticeGuidance(scope ConversationScope, message string) (practiceGuidanceSnapshot, error) {
	return sm.updateCandidateGuidance(scope, message, interviewer.ScenarioQuestionPractice, practiceGuidanceRedisKey)
}

func (sm *StateManager) UpdateFormalInterviewGuidance(scope ConversationScope, message string) (practiceGuidanceSnapshot, error) {
	return sm.updateCandidateGuidance(scope, message, interviewer.ScenarioFormalInterview, formalInterviewGuidanceRedisKey)
}

func (sm *StateManager) updateCandidateGuidance(scope ConversationScope, message, scenario string, keyFunc func(ConversationScope) string) (practiceGuidanceSnapshot, error) {
	snapshot := defaultCandidateGuidanceSnapshot(scenario)
	if sm == nil || sm.svcCtx == nil || sm.svcCtx.RedisClient == nil {
		return snapshot, nil
	}

	loaded, err := sm.loadCandidateGuidance(scope, keyFunc, scenario)
	if err != nil {
		return snapshot, err
	}
	snapshot = loaded

	signal := classifyCandidateSignal(message, snapshot.HelpOffered)
	switch signal {
	case interviewer.CandidateSignalTeachingRequested:
		snapshot.TeachingMode = true
		snapshot.HelpOffered = false
	case interviewer.CandidateSignalStuck:
		if !snapshot.TeachingMode {
			snapshot.StuckCount++
			if snapshot.StuckCount >= 3 {
				snapshot.HelpOffered = true
			}
		}
	case interviewer.CandidateSignalSubstantiveAnswer:
		if !snapshot.TeachingMode {
			snapshot.StuckCount = 0
			snapshot.HelpOffered = false
		}
	}
	snapshot.LastSignal = signal
	snapshot.LastMessageAt = time.Now().Format(time.RFC3339)

	payload, err := json.Marshal(snapshot)
	if err != nil {
		return snapshot, err
	}
	if err := sm.svcCtx.RedisClient.Set(sm.context(), keyFunc(scope), payload, chatflow.StateTTL).Err(); err != nil {
		return snapshot, err
	}
	return snapshot, nil
}

func (sm *StateManager) loadPracticeGuidance(scope ConversationScope) (practiceGuidanceSnapshot, error) {
	return sm.loadCandidateGuidance(scope, practiceGuidanceRedisKey, interviewer.ScenarioQuestionPractice)
}

func (sm *StateManager) loadCandidateGuidance(scope ConversationScope, keyFunc func(ConversationScope) string, scenario string) (practiceGuidanceSnapshot, error) {
	snapshot := defaultCandidateGuidanceSnapshot(scenario)
	if sm == nil || sm.svcCtx == nil || sm.svcCtx.RedisClient == nil {
		return snapshot, nil
	}
	raw, err := sm.svcCtx.RedisClient.Get(sm.context(), keyFunc(scope)).Result()
	if err != nil {
		if err == redis.Nil {
			return snapshot, nil
		}
		return snapshot, err
	}
	if err := json.Unmarshal([]byte(raw), &snapshot); err != nil {
		return defaultCandidateGuidanceSnapshot(scenario), err
	}
	if snapshot.Scenario == "" {
		snapshot.Scenario = scenario
	}
	if snapshot.LastSignal == "" {
		snapshot.LastSignal = interviewer.CandidateSignalNone
	}
	if snapshot.StuckCount < 0 {
		snapshot.StuckCount = 0
	}
	return snapshot, nil
}

func defaultPracticeGuidanceSnapshot() practiceGuidanceSnapshot {
	return defaultCandidateGuidanceSnapshot(interviewer.ScenarioQuestionPractice)
}

func defaultCandidateGuidanceSnapshot(scenario string) practiceGuidanceSnapshot {
	if strings.TrimSpace(scenario) == "" {
		scenario = interviewer.ScenarioFormalInterview
	}
	return practiceGuidanceSnapshot{
		Scenario:   scenario,
		LastSignal: interviewer.CandidateSignalNone,
	}
}

func practiceGuidanceRedisKey(scope ConversationScope) string {
	key := chatflow.BuildContextKey(scope.ChatID, scope.UserID, scope.Mode)
	return fmt.Sprintf("%s%s:%s:%s", practiceGuidanceKeyPrefix, key.OwnerScope, key.Lane, key.SessionID)
}

func formalInterviewGuidanceRedisKey(scope ConversationScope) string {
	key := chatflow.BuildContextKey(scope.ChatID, scope.UserID, scope.Mode)
	return fmt.Sprintf("%s%s:%s:%s", formalInterviewGuidanceKeyPrefix, key.OwnerScope, key.Lane, key.SessionID)
}

func classifyPracticeCandidateSignal(message string, helpOffered bool) string {
	return classifyCandidateSignal(message, helpOffered)
}

func classifyCandidateSignal(message string, helpOffered bool) string {
	compact := normalizeIntentText(message)
	if compact == "" {
		return interviewer.CandidateSignalNone
	}
	if looksLikeTeachingRequest(compact, helpOffered) {
		return interviewer.CandidateSignalTeachingRequested
	}
	if looksLikeCandidateStuck(compact) {
		return interviewer.CandidateSignalStuck
	}
	if len([]rune(compact)) >= 8 {
		return interviewer.CandidateSignalSubstantiveAnswer
	}
	return interviewer.CandidateSignalNone
}

func looksLikeTeachingRequest(compact string, helpOffered bool) bool {
	if containsAny(compact, []string{
		"不用讲",
		"不用详细",
		"不用解释",
		"不需要讲",
		"先不讲",
		"别讲",
	}) {
		return false
	}
	if helpOffered && (compact == "可以" || compact == "好" || compact == "需要" || compact == "讲") {
		return true
	}
	return containsAny(compact, []string{
		"讲一下",
		"讲讲",
		"详细讲",
		"详细解释",
		"解释一下",
		"帮我讲",
		"给我讲",
		"需要讲解",
		"进入讲解",
		"开始讲解",
		"想听",
	})
}

func looksLikeCandidateStuck(compact string) bool {
	if containsAny(compact, []string{
		"不是不知道",
		"并不是不知道",
		"并非不知道",
		"不是不会",
		"并不是不会",
		"并非不会",
		"不是没思路",
	}) {
		return false
	}
	if compact == "不知道" || compact == "不会" || compact == "不懂" || compact == "没思路" || compact == "没想法" {
		return true
	}
	return containsAny(compact, []string{
		"我不知道",
		"我不会",
		"还是不会",
		"也不会",
		"不会答",
		"答不上来",
		"没有思路",
		"没什么思路",
		"没头绪",
		"没接触过",
		"不了解",
		"不太了解",
		"不清楚",
		"不确定",
		"完全不会",
		"完全没思路",
		"想不出来",
	})
}

func formalInterviewScenarioConfig(guidance practiceGuidanceSnapshot) interviewer.ScenarioConfig {
	return interviewer.ScenarioConfig{
		Type:            interviewer.ScenarioFormalInterview,
		StuckCount:      guidance.StuckCount,
		HelpOffered:     guidance.HelpOffered,
		TeachingMode:    guidance.TeachingMode,
		CandidateSignal: guidance.LastSignal,
	}
}

func practiceScenarioConfig(context svc.SessionPracticeContext, guidance practiceGuidanceSnapshot) interviewer.ScenarioConfig {
	return interviewer.ScenarioConfig{
		Type:             interviewer.ScenarioQuestionPractice,
		QuestionKey:      context.QuestionKey,
		QuestionSnapshot: strings.TrimSpace(context.QuestionSnapshot),
		StuckCount:       guidance.StuckCount,
		HelpOffered:      guidance.HelpOffered,
		TeachingMode:     guidance.TeachingMode,
		CandidateSignal:  guidance.LastSignal,
	}
}
