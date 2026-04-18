package sessionmode

import "strings"

const (
	KeyInterview = "Interview"
	KeyResearch  = "Research"
	KeyMemory    = "Memory"
	KeyCoach     = "Coach"

	LabelInterview = "Interview Studio"
	LabelResearch  = "Research Desk"
	LabelMemory    = "Memory Atlas"
	LabelCoach     = "Coach"

	DefaultKey   = KeyInterview
	DefaultLabel = LabelInterview
)

func NormalizeKey(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "interview", "interview studio":
		return KeyInterview
	case "research", "research desk":
		return KeyResearch
	case "memory", "memory atlas":
		return KeyMemory
	case "coach":
		return KeyCoach
	default:
		return DefaultKey
	}
}

func DisplayName(value string) string {
	switch NormalizeKey(value) {
	case KeyResearch:
		return LabelResearch
	case KeyMemory:
		return LabelMemory
	case KeyCoach:
		return LabelCoach
	default:
		return LabelInterview
	}
}

func AllKeys() []string {
	return []string{KeyInterview, KeyResearch, KeyMemory, KeyCoach}
}
