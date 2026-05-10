// Package interviewer 负责构建 AI 面试官的系统提示词。
//
// 该包只接受纯数据结构，不依赖 logic、svc、数据库或 OpenAI SDK，
// 让面试官画像、会话配置和提示词安全策略可以独立演进。
package interviewer

type FocusArea struct {
	Key   string
	Label string
}

type KnowledgeChunk struct {
	Title   string
	Content string
}

type SessionConfig struct {
	DirectionKey          string
	DirectionLabel        string
	DifficultyLevel       int64
	DifficultyLabel       string
	InterviewerStyle      string
	InterviewerStyleLabel string
	FocusAreas            []FocusArea
	FollowUpDepth         string
	EstimatedMinutes      int64
	ProgressPercent       int64
}

const (
	ScenarioFormalInterview  = "formal_interview"
	ScenarioQuestionPractice = "question_practice"

	CandidateSignalNone              = "none"
	CandidateSignalStuck             = "candidate_stuck"
	CandidateSignalTeachingRequested = "teaching_requested"
	CandidateSignalSubstantiveAnswer = "substantive_answer"
)

type ScenarioConfig struct {
	Type             string
	QuestionKey      string
	QuestionSnapshot string
	StuckCount       int
	HelpOffered      bool
	TeachingMode     bool
	CandidateSignal  string
}

type BuildInput struct {
	ChatID            string
	State             string
	Session           *SessionConfig
	Scenario          *ScenarioConfig
	Knowledge         []KnowledgeChunk
	MaxKnowledgeRunes int
}

type Prompt struct {
	SystemMessage string
	Domain        DomainProfile
	Style         Style
	FocusLabels   []string
}
