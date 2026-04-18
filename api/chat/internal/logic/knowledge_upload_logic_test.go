package logic

import (
	"testing"

	"GoZero-AI/api/chat/internal/types"
)

func TestValidateKnowledgeUploadInput(t *testing.T) {
	tests := []struct {
		name    string
		input   *types.KnowledgeUploadInput
		wantErr bool
	}{
		{
			name:    "nil input",
			input:   nil,
			wantErr: true,
		},
		{
			name: "empty title",
			input: &types.KnowledgeUploadInput{
				Title:   "   ",
				Content: "有效内容",
			},
			wantErr: true,
		},
		{
			name: "empty content",
			input: &types.KnowledgeUploadInput{
				Title:   "文档标题",
				Content: "   ",
			},
			wantErr: true,
		},
		{
			name: "valid input",
			input: &types.KnowledgeUploadInput{
				Title:   "文档标题",
				Content: "第一段内容",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateKnowledgeUploadInput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateKnowledgeUploadInput(%+v) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}
