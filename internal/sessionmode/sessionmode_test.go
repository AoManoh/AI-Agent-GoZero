package sessionmode

import "testing"

func TestNormalizeKey(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty defaults to interview", input: "", want: KeyInterview},
		{name: "interview studio alias", input: "Interview Studio", want: KeyInterview},
		{name: "research desk alias", input: "Research Desk", want: KeyResearch},
		{name: "memory atlas alias", input: "Memory Atlas", want: KeyMemory},
		{name: "coach passthrough", input: "Coach", want: KeyCoach},
		{name: "unknown falls back", input: "Companion", want: KeyInterview},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeKey(tt.input); got != tt.want {
				t.Fatalf("NormalizeKey(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestDisplayName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "interview key", input: KeyInterview, want: LabelInterview},
		{name: "research key", input: KeyResearch, want: LabelResearch},
		{name: "memory alias", input: "Memory Atlas", want: LabelMemory},
		{name: "coach key", input: KeyCoach, want: LabelCoach},
		{name: "unknown falls back", input: "Companion", want: LabelInterview},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DisplayName(tt.input); got != tt.want {
				t.Fatalf("DisplayName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestAllKeys(t *testing.T) {
	got := AllKeys()
	want := []string{KeyInterview, KeyResearch, KeyMemory, KeyCoach}

	if len(got) != len(want) {
		t.Fatalf("len(AllKeys()) = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("AllKeys()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
