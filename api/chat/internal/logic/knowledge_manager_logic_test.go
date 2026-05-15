package logic

import "testing"

func TestKnowledgeDocumentChunksAllowsReaderSizedLimit(t *testing.T) {
	if got := boundedKnowledgeLimit(500, 50, 500); got != 500 {
		t.Fatalf("boundedKnowledgeLimit(500, 50, 500) = %d, want 500", got)
	}
	if got := boundedKnowledgeLimit(800, 50, 500); got != 500 {
		t.Fatalf("boundedKnowledgeLimit(800, 50, 500) = %d, want 500", got)
	}
}
