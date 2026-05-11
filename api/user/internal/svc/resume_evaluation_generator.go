package svc

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"GoZero-AI/api/user/internal/config"
	"GoZero-AI/api/user/internal/resumeevaluation"
	"GoZero-AI/api/user/internal/types"

	"github.com/sashabaranov/go-openai"
)

type ResumeEvaluationGenerator struct {
	client *openai.Client
	cfg    config.Config
}

type ResumeEvaluationInput struct {
	Title        string
	Filename     string
	DirectionKey string
	Chunks       []string
}

type GeneratedResumeEvaluation struct {
	Summary     string                      `json:"summary"`
	Dimensions  []types.EvaluationDimension `json:"dimensions"`
	Strengths   []string                    `json:"strengths"`
	Risks       []types.ResumeRiskSignal    `json:"risks"`
	Suggestions []string                    `json:"suggestions"`
}

func NewResumeEvaluationGenerator(client *openai.Client, cfg config.Config) *ResumeEvaluationGenerator {
	return &ResumeEvaluationGenerator{
		client: client,
		cfg:    cfg,
	}
}

func (g *ResumeEvaluationGenerator) Generate(ctx context.Context, input ResumeEvaluationInput) (*GeneratedResumeEvaluation, error) {
	if len(input.Chunks) == 0 {
		return nil, fmt.Errorf("no resume chunks available for evaluation")
	}

	req := openai.ChatCompletionRequest{
		Model: g.cfg.EvaluationModel(),
		Messages: []openai.ChatCompletionMessage{
			{
				Role: openai.ChatMessageRoleSystem,
				Content: "你是一名严谨的技术简历面试准备度评估助手。请只输出 JSON，不要输出 markdown 代码块。" +
					"返回结构必须包含 summary、dimensions、strengths、risks、suggestions。" +
					"dimensions 必须是数组，每项包含 key、label、score、maxScore、summary，score 为 0-100 的整数，maxScore 固定为 100。" +
					"risks 必须是数组，每项包含 key、label、severity、suggestion，severity 只能是 low、medium、high。" +
					"必须严格基于简历原文证据评估；缺失项写成风险，不得虚构不存在的经历、指标或项目。" +
					"禁止依据姓名、年龄、性别、照片、籍贯、学校光环等非能力因素评分。\n" +
					resumeevaluation.PromptSection(),
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: buildResumeEvaluationPrompt(input),
			},
		},
		MaxCompletionTokens: g.cfg.EvaluationMaxTokens(),
		Temperature:         g.cfg.EvaluationTemperature(),
	}

	resp, err := g.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, err
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("resume evaluation model returned no choices")
	}

	content := sanitizeJSONPayload(resp.Choices[0].Message.Content)
	var generated GeneratedResumeEvaluation
	if err := json.Unmarshal([]byte(content), &generated); err != nil {
		return nil, fmt.Errorf("parse resume evaluation json failed: %w", err)
	}
	return &generated, nil
}

func buildResumeEvaluationPrompt(input ResumeEvaluationInput) string {
	var builder strings.Builder
	builder.WriteString("请基于下面的简历资料输出结构化评估 JSON。\n")
	builder.WriteString("目标方向：")
	if strings.TrimSpace(input.DirectionKey) == "" {
		builder.WriteString("未指定，请根据简历内容判断面试准备度")
	} else {
		builder.WriteString(input.DirectionKey)
	}
	builder.WriteString("\n简历标题：")
	builder.WriteString(input.Title)
	builder.WriteString("\n文件名：")
	builder.WriteString(input.Filename)
	builder.WriteString("\n\n简历正文片段：\n")
	for idx, chunk := range input.Chunks {
		if idx >= 10 {
			break
		}
		builder.WriteString(fmt.Sprintf("\n[片段 %d]\n", idx+1))
		builder.WriteString(truncateResumeEvaluationChunk(chunk, 900))
		builder.WriteString("\n")
	}
	return builder.String()
}

func truncateResumeEvaluationChunk(content string, maxLen int) string {
	trimmed := strings.TrimSpace(content)
	if maxLen <= 0 {
		return ""
	}
	runes := []rune(trimmed)
	if len(runes) <= maxLen {
		return trimmed
	}
	return string(runes[:maxLen]) + "..."
}
