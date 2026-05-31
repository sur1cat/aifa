package openai

import (
	"strings"
	"testing"
)

func TestUniversalAgentPromptRestrictsScope(t *testing.T) {
	prompt := SystemPrompt(AgentUniversal, "")

	required := []string{
		"You are NOT a general-purpose assistant.",
		"You may help ONLY with these domains",
		"habits and goals related to habits",
		"tasks and planning",
		"personal finance, budgeting, spending, debt, savings",
		"Requests outside these supported domains",
	}

	for _, part := range required {
		if !strings.Contains(prompt, part) {
			t.Fatalf("universal prompt is missing required restriction text: %q", part)
		}
	}
}

func TestCommandPromptRestrictsUnsupportedRequests(t *testing.T) {
	prompt := CommandPrompt()

	required := []string{
		"SCOPE RESTRICTION:",
		"Supported domains ONLY:",
		"If the request is outside these domains, you MUST return intent=\"unsupported\".",
		"supported-domain conversation only",
		"Any request outside supported domains MUST become \"unsupported\", not \"chat\"",
	}

	for _, part := range required {
		if !strings.Contains(prompt, part) {
			t.Fatalf("command prompt is missing required restriction text: %q", part)
		}
	}
}
