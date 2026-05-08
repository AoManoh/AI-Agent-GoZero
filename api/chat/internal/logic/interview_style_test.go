package logic

import "testing"

func TestSelectInterviewStyleIsStablePerChatID(t *testing.T) {
	first := selectInterviewStyle("demo-session-1")
	second := selectInterviewStyle("demo-session-1")
	if first.Key != second.Key {
		t.Fatalf("style should be stable for same chatId: %q != %q", first.Key, second.Key)
	}
}

func TestSelectInterviewStyleFallback(t *testing.T) {
	style := selectInterviewStyle("")
	if style.Key != "structured" {
		t.Fatalf("empty chatId style = %q, want structured", style.Key)
	}
}

func TestBuildInterviewStylePrompt(t *testing.T) {
	style := selectInterviewStyle("demo-session-2")
	prompt := buildInterviewStylePrompt(style)
	if prompt == "" || !containsAny(prompt, []string{style.Label, "一次只问一个问题"}) {
		t.Fatalf("style prompt missing expected constraints: %q", prompt)
	}
}
