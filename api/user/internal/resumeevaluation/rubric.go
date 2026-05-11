package resumeevaluation

import (
	"fmt"
	"strings"

	"GoZero-AI/api/user/internal/types"
)

const RubricVersion = "resume-rubric-v1"

type Dimension struct {
	Key    string
	Label  string
	Weight float64
}

var rubricDimensions = []Dimension{
	{Key: "target_alignment", Label: "方向匹配度", Weight: 0.18},
	{Key: "technical_relevance", Label: "技术相关性", Weight: 0.18},
	{Key: "project_depth", Label: "项目深度", Weight: 0.18},
	{Key: "impact_evidence", Label: "结果证据", Weight: 0.16},
	{Key: "engineering_practice", Label: "工程实践", Weight: 0.14},
	{Key: "clarity_structure", Label: "表达结构", Weight: 0.08},
	{Key: "interview_readiness", Label: "可追问度", Weight: 0.08},
}

func Dimensions() []Dimension {
	return append([]Dimension(nil), rubricDimensions...)
}

func PromptSection() string {
	var builder strings.Builder
	builder.WriteString("请严格基于以下 resume rubric 维度输出 dimensions，禁止自造 key：\n")
	for _, dimension := range rubricDimensions {
		builder.WriteString(fmt.Sprintf("- key=%s, label=%s, maxScore=100\n", dimension.Key, dimension.Label))
	}
	return builder.String()
}

func ComputeOverallScore(dimensions []types.EvaluationDimension) float64 {
	if len(dimensions) == 0 {
		return 0
	}

	weightMap := make(map[string]float64, len(rubricDimensions))
	for _, dimension := range rubricDimensions {
		weightMap[dimension.Key] = dimension.Weight
	}

	var total float64
	var usedWeight float64
	for _, dimension := range dimensions {
		weight, ok := weightMap[dimension.Key]
		if !ok {
			continue
		}
		maxScore := dimension.MaxScore
		if maxScore <= 0 {
			maxScore = 100
		}
		total += (float64(dimension.Score) / float64(maxScore)) * weight
		usedWeight += weight
	}
	if usedWeight == 0 {
		return 0
	}
	return round2(total / usedWeight * 100)
}

func NormalizeDimensions(primary, fallback []types.EvaluationDimension) []types.EvaluationDimension {
	fallbackMap := make(map[string]types.EvaluationDimension, len(fallback))
	for _, dimension := range fallback {
		fallbackMap[dimension.Key] = dimension
	}

	primaryMap := make(map[string]types.EvaluationDimension, len(primary))
	for _, dimension := range primary {
		primaryMap[dimension.Key] = dimension
	}

	result := make([]types.EvaluationDimension, 0, len(rubricDimensions))
	for _, definition := range rubricDimensions {
		current, ok := primaryMap[definition.Key]
		if !ok {
			current = fallbackMap[definition.Key]
		}
		current.Key = definition.Key
		current.Label = definition.Label
		current.MaxScore = 100
		if current.Score < 0 {
			current.Score = 0
		}
		if current.Score > current.MaxScore {
			current.Score = current.MaxScore
		}
		if strings.TrimSpace(current.Summary) == "" {
			current.Summary = fallbackMap[definition.Key].Summary
		}
		result = append(result, current)
	}
	return result
}

func HasCompleteDimensions(dimensions []types.EvaluationDimension) bool {
	if len(dimensions) != len(rubricDimensions) {
		return false
	}
	seen := make(map[string]struct{}, len(dimensions))
	for _, dimension := range dimensions {
		seen[dimension.Key] = struct{}{}
	}
	for _, definition := range rubricDimensions {
		if _, ok := seen[definition.Key]; !ok {
			return false
		}
	}
	return true
}

func Level(score float64) string {
	switch {
	case score >= 85:
		return "strong"
	case score >= 70:
		return "interview_ready"
	case score >= 55:
		return "needs_polish"
	default:
		return "high_risk"
	}
}

func round2(value float64) float64 {
	return float64(int(value*100+0.5)) / 100
}
