package config

import "testing"

func TestEvaluationTemperatureUsesUnitTemperatureForGPT5(t *testing.T) {
	var cfg Config
	cfg.Evaluation.Model = "gpt-5.4"
	cfg.OpenAI.EvaluationTemp = 0.2

	if got := cfg.EvaluationTemperature(); got != 1 {
		t.Fatalf("EvaluationTemperature() = %v, want 1 for gpt-5 model", got)
	}
}

func TestEvaluationTemperatureKeepsDefaultForRegularModel(t *testing.T) {
	var cfg Config
	cfg.OpenAI.EvaluationModel = "qwen-plus"

	if got := cfg.EvaluationTemperature(); got != 0.2 {
		t.Fatalf("EvaluationTemperature() = %v, want 0.2", got)
	}
}
