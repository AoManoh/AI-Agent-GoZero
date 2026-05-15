package sessionruntime

import "strings"

const (
	ScenarioFormalInterview  = "formal_interview"
	ScenarioQuestionPractice = "question_practice"

	StarterNone       = "none"
	StarterBank       = "bank"
	StarterResumePlan = "resume_plan"
	StarterManual     = "manual"
)

func NormalizeScenario(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case ScenarioQuestionPractice:
		return ScenarioQuestionPractice
	default:
		return ScenarioFormalInterview
	}
}

func NormalizeStarterSource(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case StarterBank:
		return StarterBank
	case StarterResumePlan:
		return StarterResumePlan
	case StarterManual:
		return StarterManual
	default:
		return StarterNone
	}
}
