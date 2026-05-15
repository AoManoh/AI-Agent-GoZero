package logic

import (
	"strings"
	"unicode"

	"GoZero-AI/api/chat/internal/interviewer"
	"GoZero-AI/internal/chatflow"
)

const (
	formalInterviewResponseBudgetRunes = 260
	teachingResponseBudgetRunes        = 420
	sessionEndingResponseBudgetRunes   = 800
	responseBudgetClosureRunes         = 10
)

type responseBudgetGuard struct {
	maxRunes int
	stopped  bool
}

func newResponseBudgetGuard(state string, scenario *interviewer.ScenarioConfig) *responseBudgetGuard {
	limit := formalInterviewResponseBudgetRunes
	if strings.TrimSpace(state) == chatflow.InterviewStateEnd {
		limit = sessionEndingResponseBudgetRunes
	}
	if scenario != nil && scenario.TeachingMode {
		limit = teachingResponseBudgetRunes
	}
	return &responseBudgetGuard{maxRunes: limit}
}

func (g *responseBudgetGuard) Filter(existing, delta string) (string, bool) {
	if g == nil || g.maxRunes <= 0 {
		return delta, false
	}
	if g.stopped {
		return "", true
	}
	if delta == "" {
		return "", false
	}

	candidate := existing + delta
	if stopAt, ok := multiQuestionStopRuneIndex(candidate); ok {
		g.stopped = true
		allowed := stopAt - runeLen(existing)
		if budgetAllowed := g.maxRunes - runeLen(existing) - responseBudgetClosureRunes; allowed > budgetAllowed {
			allowed = budgetAllowed
		}
		if allowed < 0 {
			allowed = 0
		}
		visible := takeRunes(delta, allowed)
		return g.withClosure(existing, visible), true
	}

	if runeLen(candidate)+responseBudgetClosureRunes <= g.maxRunes {
		return delta, false
	}

	g.stopped = true
	allowed := g.maxRunes - runeLen(existing) - responseBudgetClosureRunes
	if allowed < 0 {
		allowed = 0
	}
	visible := takeRunes(delta, allowed)
	return g.withClosure(existing, visible), true
}

func (g *responseBudgetGuard) withClosure(existing, visible string) string {
	prefix := strings.TrimRightFunc(existing+visible, unicode.IsSpace)
	closure := responseBudgetClosure(prefix)
	if strings.Contains(prefix, closure) {
		return visible
	}
	return takeSuffixAfterPrefix(existing, prefix+closure)
}

func responseBudgetClosure(prefix string) string {
	if strings.TrimSpace(prefix) == "" || endsWithSentencePunctuation(prefix) {
		return "我们先聚焦这一点。"
	}
	return "。我们先聚焦这一点。"
}

func multiQuestionStopRuneIndex(value string) (int, bool) {
	firstQuestion := -1
	questionCount := 0
	for idx, r := range []rune(value) {
		if r == '？' || r == '?' {
			questionCount++
			if firstQuestion < 0 {
				firstQuestion = idx + 1
			}
			if questionCount > 1 {
				return firstQuestion, true
			}
		}
	}

	if idx, ok := numberedListSecondItemRuneIndex(value); ok {
		return idx, true
	}
	return 0, false
}

func numberedListSecondItemRuneIndex(value string) (int, bool) {
	patterns := []string{"\n2.", "\n2、", "\n2)", "\n二、"}
	for _, pattern := range patterns {
		if byteIdx := strings.Index(value, pattern); byteIdx >= 0 {
			return runeLen(value[:byteIdx]), true
		}
	}
	return 0, false
}

func endsWithSentencePunctuation(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return false
	}
	runes := []rune(trimmed)
	switch runes[len(runes)-1] {
	case '。', '！', '？', '.', '!', '?', '；', ';':
		return true
	default:
		return false
	}
}

func takeRunes(value string, limit int) string {
	if limit <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
}

func runeLen(value string) int {
	return len([]rune(value))
}

func takeSuffixAfterPrefix(existing, full string) string {
	existingRunes := []rune(existing)
	fullRunes := []rune(full)
	if len(existingRunes) >= len(fullRunes) {
		return ""
	}
	return string(fullRunes[len(existingRunes):])
}
