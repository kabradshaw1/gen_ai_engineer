package guardrails

import (
	"testing"

	"github.com/kabradshaw1/portfolio/go/ai-service/internal/llm"
)

func TestTruncateHistory_ShortPassThrough(t *testing.T) {
	msgs := []llm.Message{
		{Role: llm.RoleUser, Content: "a"},
		{Role: llm.RoleAssistant, Content: "b"},
	}
	out := TruncateHistory(msgs, 5)
	if len(out) != 2 {
		t.Errorf("len = %d", len(out))
	}
}

func TestTruncateHistory_LongNoSystem(t *testing.T) {
	msgs := make([]llm.Message, 30)
	for i := range msgs {
		msgs[i] = llm.Message{Role: llm.RoleUser, Content: string(rune('a' + i%26))}
	}
	out := TruncateHistory(msgs, 5)
	if len(out) != 5 {
		t.Fatalf("len = %d", len(out))
	}
	if out[0].Content != msgs[25].Content {
		t.Errorf("expected tail start = %q, got %q", msgs[25].Content, out[0].Content)
	}
}

func TestTruncateHistory_LongWithSystem(t *testing.T) {
	msgs := []llm.Message{{Role: llm.RoleSystem, Content: "sys"}}
	for i := 0; i < 29; i++ {
		msgs = append(msgs, llm.Message{Role: llm.RoleUser, Content: string(rune('a' + i%26))})
	}
	out := TruncateHistory(msgs, 5)
	if len(out) != 5 {
		t.Fatalf("len = %d", len(out))
	}
	if out[0].Role != llm.RoleSystem {
		t.Errorf("expected system first, got %s", out[0].Role)
	}
	// Last 4 should be the tail of the user messages
	if out[4].Content != msgs[len(msgs)-1].Content {
		t.Errorf("tail wrong")
	}
}

func TestIsRefusal(t *testing.T) {
	cases := map[string]bool{
		"I can't help with that.": true,
		"I cannot do that":        true,
		"I'm not able to":         true,
		"Sorry, I can't":          true,
		"Here's what I found:":    false,
		"":                        false,
	}
	for text, want := range cases {
		if got := IsRefusal(text); got != want {
			t.Errorf("IsRefusal(%q) = %v, want %v", text, got, want)
		}
	}
}
