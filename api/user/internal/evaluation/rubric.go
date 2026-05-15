package evaluation

import (
	"fmt"
	"strings"

	"GoZero-AI/api/user/internal/types"
)

const RubricVersion = "rubric-v1"

type Dimension struct {
	Key    string
	Label  string
	Weight float64
}

var rubricDimensions = []Dimension{
	{Key: "technical_depth", Label: "技术深度", Weight: 0.35},
	{Key: "engineering_practice", Label: "工程实践", Weight: 0.30},
	{Key: "architecture_sense", Label: "架构意识", Weight: 0.20},
	{Key: "communication", Label: "表达与沟通", Weight: 0.15},
}

func Dimensions() []Dimension {
	return append([]Dimension(nil), rubricDimensions...)
}

func PromptSection() string {
	var builder strings.Builder
	builder.WriteString("请严格基于以下 rubric 维度输出 dimensions，禁止自造 key：\n")
	for _, dimension := range rubricDimensions {
		builder.WriteString(fmt.Sprintf("- key=%s, label=%s, maxScore=5\n", dimension.Key, dimension.Label))
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
		total += (float64(dimension.Score) / float64(maxInt64(dimension.MaxScore, 1))) * weight
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
		current.MaxScore = 5

		if current.Score <= 0 {
			if fallbackDimension, ok := fallbackMap[definition.Key]; ok && fallbackDimension.Score > 0 {
				current.Score = fallbackDimension.Score
			} else {
				current.Score = 1
			}
		}
		if current.Score > current.MaxScore {
			current.Score = current.MaxScore
		}

		if strings.TrimSpace(current.Summary) == "" {
			if fallbackDimension, ok := fallbackMap[definition.Key]; ok {
				current.Summary = fallbackDimension.Summary
			}
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

func round2(value float64) float64 {
	return float64(int(value*100+0.5)) / 100
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
