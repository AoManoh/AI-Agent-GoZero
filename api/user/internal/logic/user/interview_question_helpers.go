package user

import (
	"encoding/json"

	"GoZero-AI/api/user/internal/types"
	"GoZero-AI/api/user/model"
)

const (
	interviewQuestionSchemaVersion = "interview-question-bank-v1"
	defaultQuestionListLimit       = int64(50)
	maxQuestionListLimit           = int64(2000)
)

func buildInterviewQuestionItem(question model.InterviewQuestion) types.InterviewQuestionItem {
	item := types.InterviewQuestionItem{
		Id:                   question.Id,
		Key:                  question.QuestionKey,
		DirectionKey:         question.DirectionKey,
		FocusKey:             question.FocusKey,
		FocusLabel:           question.FocusLabel,
		DifficultyLevel:      question.DifficultyLevel,
		DifficultyLabel:      question.DifficultyLabel,
		Title:                question.Title,
		Prompt:               question.Prompt,
		ExpectedSignals:      decodeStringJSON(question.ExpectedSignals),
		FollowUps:            decodeStringJSON(question.FollowUps),
		EvaluationDimensions: decodeStringJSON(question.EvaluationDimensions),
		Tags:                 decodeStringJSON(question.Tags),
		SourceRefs:           decodeStringJSON(question.SourceRefs),
		BatchKey:             question.BatchKey,
		BatchLabel:           question.BatchLabel,
		Sequence:             question.Sequence,
		BatchSequence:        question.BatchSequence,
		Status:               question.Status,
		QualityScore:         question.QualityScore,
		UsageCount:           question.UsageCount,
		UpdatedAt:            question.UpdatedAt.Format(timeLayout),
		SourceCount:          question.SourceCount,
	}
	if question.LastUsedAt.Valid {
		item.LastUsedAt = question.LastUsedAt.Time.Format(timeLayout)
	}
	return item
}

func buildInterviewQuestionSourceItem(source model.InterviewQuestionSource) types.InterviewQuestionSourceItem {
	return types.InterviewQuestionSourceItem{
		SourceKey:   source.SourceKey,
		SourceTitle: source.SourceTitle,
		SourceUrl:   source.SourceUrl,
		SourceType:  source.SourceType,
		LicenseNote: source.LicenseNote,
		BatchKey:    source.BatchKey,
	}
}

func buildInterviewPlanQuestionFromBank(question model.InterviewQuestion) types.InterviewPlanQuestion {
	return types.InterviewPlanQuestion{
		Key:                  question.QuestionKey,
		DirectionKey:         question.DirectionKey,
		FocusKey:             question.FocusKey,
		FocusLabel:           question.FocusLabel,
		DifficultyLevel:      question.DifficultyLevel,
		DifficultyLabel:      question.DifficultyLabel,
		Title:                question.Title,
		Prompt:               question.Prompt,
		ExpectedSignals:      decodeStringJSON(question.ExpectedSignals),
		FollowUps:            decodeStringJSON(question.FollowUps),
		EvaluationDimensions: decodeStringJSON(question.EvaluationDimensions),
	}
}

func decodeStringJSON(raw []byte) []string {
	if len(raw) == 0 {
		return []string{}
	}
	var values []string
	if err := json.Unmarshal(raw, &values); err != nil {
		return []string{}
	}
	return values
}

func normalizeQuestionListLimit(limit int64) int64 {
	if limit <= 0 {
		return defaultQuestionListLimit
	}
	if limit > maxQuestionListLimit {
		return maxQuestionListLimit
	}
	return limit
}

func normalizeQuestionListOffset(offset int64) int64 {
	if offset < 0 {
		return 0
	}
	return offset
}
