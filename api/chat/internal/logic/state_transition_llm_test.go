package logic

import (
	"strings"
	"testing"

	"GoZero-AI/api/chat/internal/types"
)

func TestParseStateTransitionPayload(t *testing.T) {
	parsed, err := parseStateTransitionPayload("```json\n{\"state\":\"follow_up\",\"reason\":\"自然追问\",\"confidence\":0.91}\n```")
	if err != nil {
		t.Fatalf("parseStateTransitionPayload() error = %v", err)
	}
	if parsed.State != types.StateFollowUp {
		t.Fatalf("State = %q, want %q", parsed.State, types.StateFollowUp)
	}
	if parsed.Reason != "自然追问" {
		t.Fatalf("Reason = %q", parsed.Reason)
	}
	if parsed.Confidence != 0.91 {
		t.Fatalf("Confidence = %v", parsed.Confidence)
	}
}

func TestNormalizeInterviewStateAliases(t *testing.T) {
	tests := map[string]string{
		"followup":      types.StateFollowUp,
		"deep-dive":     types.StateFollowUp,
		"evaluation":    types.StateEvaluate,
		"behavioral":    types.StateEvaluate,
		"completed":     types.StateEnd,
		"core_question": types.StateQuestion,
	}

	for input, want := range tests {
		if got := normalizeInterviewState(input); got != want {
			t.Fatalf("normalizeInterviewState(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestAllowedStateTransition(t *testing.T) {
	if !isAllowedStateTransition(types.StateQuestion, types.StateFollowUp) {
		t.Fatal("question -> follow_up should be allowed")
	}
	if isAllowedStateTransition(types.StateStart, types.StateEnd) {
		t.Fatal("start -> end should not be allowed")
	}
	if !isAllowedStateTransition(types.StateEnd, types.StateEnd) {
		t.Fatal("end -> end should be allowed")
	}
}

func TestDecideTransitionFallsBackToRulesWhenLLMDisabled(t *testing.T) {
	sm := &StateManager{}
	decision := sm.decideTransition(types.StateQuestion, "那你接着讲讲，具体怎么保证取消信号能传到所有 goroutine？")
	if decision.NextState != types.StateFollowUp {
		t.Fatalf("NextState = %q, want %q", decision.NextState, types.StateFollowUp)
	}
	if decision.Source != "rule" {
		t.Fatalf("Source = %q, want rule", decision.Source)
	}
}

func TestTruncateForStateTransition(t *testing.T) {
	got := truncateForStateTransition("abcdef", 3)
	if got != "abc..." {
		t.Fatalf("truncateForStateTransition() = %q", got)
	}
	if strings.TrimSpace(truncateForStateTransition("  abc  ", 10)) != "abc" {
		t.Fatal("truncateForStateTransition should trim short text")
	}
}
