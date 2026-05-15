package interviewer

import (
	"strings"
	"testing"
)

func TestSelectStyleIsStablePerChatID(t *testing.T) {
	first := SelectStyle("demo-session-1")
	second := SelectStyle("demo-session-1")
	if first.Key != second.Key {
		t.Fatalf("style should be stable for same chatId: %q != %q", first.Key, second.Key)
	}
}

func TestSelectStyleFallback(t *testing.T) {
	style := SelectStyle("")
	if style.Key != "structured" {
		t.Fatalf("empty chatId style = %q, want structured", style.Key)
	}
}

func TestSelectStyleByKeySupportsPublicMentorKeyAndLegacyAlias(t *testing.T) {
	mentor := SelectStyleByKey("mentor", "demo-session")
	if mentor.Key != "mentor" {
		t.Fatalf("mentor style key = %q, want mentor", mentor.Key)
	}

	coaching := SelectStyleByKey("coaching", "demo-session")
	if coaching.Key != "mentor" {
		t.Fatalf("coaching alias key = %q, want mentor", coaching.Key)
	}
}

func TestSelectStyleByKeySupportsPresentationPersonaAliases(t *testing.T) {
	tests := []struct {
		key  string
		want string
	}{
		{key: "engineer", want: "senior"},
		{key: "工程师型", want: "senior"},
		{key: "pressure", want: "pressure"},
		{key: "guiding", want: "mentor"},
		{key: "引导型", want: "mentor"},
		{key: "deep_dive", want: "system_design"},
		{key: "深挖型", want: "system_design"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			style := SelectStyleByKey(tt.key, "demo-session")
			if style.Key != tt.want {
				t.Fatalf("SelectStyleByKey(%q) = %q, want %q", tt.key, style.Key, tt.want)
			}
		})
	}
}

func TestBuildStylePrompt(t *testing.T) {
	style := SelectStyle("demo-session-2")
	prompt := BuildStylePrompt(style)
	if prompt == "" || !strings.Contains(prompt, style.Label) || !strings.Contains(prompt, "一次只问一个问题") {
		t.Fatalf("style prompt missing expected constraints: %q", prompt)
	}
}
