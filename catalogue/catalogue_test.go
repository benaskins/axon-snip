package catalogue

import (
	"strings"
	"testing"
)

func minCatalogue() *Catalogue {
	return &Catalogue{
		Name:     "Axon/Lamina",
		Language: "go",
	}
}

func TestSystemPrompt_ClassifiesPersistenceBeforeSelecting(t *testing.T) {
	prompt, err := minCatalogue().SystemPrompt()
	if err != nil {
		t.Fatalf("SystemPrompt: %v", err)
	}

	checks := []string{
		"Selection Discipline",
		"axon-base",
		"durable",
	}
	for _, want := range checks {
		if !strings.Contains(prompt, want) {
			t.Errorf("system prompt missing %q; got:\n%s", want, prompt)
		}
	}
}
