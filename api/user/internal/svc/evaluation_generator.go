package svc

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"GoZero-AI/api/user/internal/config"
	"GoZero-AI/api/user/internal/evaluation"
	"GoZero-AI/api/user/internal/types"

	"github.com/sashabaranov/go-openai"
)

type EvaluationGenerator struct {
	client *openai.Client
	cfg    config.Config
}

type EvaluationMessage struct {
	Role    string
	Content string
}

type GeneratedEvaluation struct {
	Summary     string                      `json:"summary"`
	Dimensions  []types.EvaluationDimension `json:"dimensions"`
	Strengths   []string                    `json:"strengths"`
	Risks       []string                    `json:"risks"`
	Suggestions []string                    `json:"suggestions"`
}

func NewEvaluationGenerator(client *openai.Client, cfg config.Config) *EvaluationGenerator {
	return &EvaluationGenerator{
		client: client,
		cfg:    cfg,
	}
}

func (g *EvaluationGenerator) Generate(ctx context.Context, messages []EvaluationMessage) (*GeneratedEvaluation, error) {
	if len(messages) == 0 {
		return nil, fmt.Errorf("no messages available for evaluation")
	}

	var transcript strings.Builder
	for _, message := range messages {
		transcript.WriteString(message.Role)
		transcript.WriteString(": ")
		transcript.WriteString(message.Content)
		transcript.WriteString("\n")
	}

	req := openai.ChatCompletionRequest{
		Model: g.cfg.EvaluationModel(),
		Messages: []openai.ChatCompletionMessage{
			{
				Role: openai.ChatMessageRoleSystem,
				Content: "你是一名 Go 技术面试复盘助手。请只输出 JSON，不要输出 markdown 代码块。" +
					"返回结构必须包含 summary、dimensions、strengths、risks、suggestions。" +
					"dimensions 必须是数组，每项包含 key、label、score、maxScore、summary。" +
					"score 和 maxScore 都必须是整数，maxScore 固定为 5。" +
					"请基于真实对话内容，给出克制、结构化、可执行的评估，不要虚构不存在的经历。\n" +
					evaluation.PromptSection(),
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "请基于下面的会话记录输出结构化评估 JSON：\n\n" + transcript.String(),
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
		return nil, fmt.Errorf("evaluation model returned no choices")
	}

	content := sanitizeJSONPayload(resp.Choices[0].Message.Content)
	var generated GeneratedEvaluation
	if err := json.Unmarshal([]byte(content), &generated); err != nil {
		return nil, fmt.Errorf("parse evaluation json failed: %w", err)
	}

	return &generated, nil
}

func sanitizeJSONPayload(content string) string {
	trimmed := strings.TrimSpace(content)
	trimmed = strings.TrimPrefix(trimmed, "```json")
	trimmed = strings.TrimPrefix(trimmed, "```")
	trimmed = strings.TrimSuffix(trimmed, "```")
	return strings.TrimSpace(trimmed)
}
