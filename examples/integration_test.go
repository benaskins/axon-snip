package examples_test

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/benaskins/axon-snip/examples"
)

// TestExtractRealModule extracts examples from the real axon-base module
// in the Go module cache. Skips if not available.
func TestExtractRealModule(t *testing.T) {
	// Find axon-base in the module cache.
	cmd := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/benaskins/axon-base")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Skipf("axon-base not in module cache: %v", err)
	}
	modDir := strings.TrimSpace(string(out))

	results, err := examples.ExtractFromModule(modDir)
	if err != nil {
		t.Fatalf("ExtractFromModule: %v", err)
	}

	// axon-base should have examples in pool, scan, and errors packages.
	if len(results) < 3 {
		t.Errorf("expected at least 3 packages with examples, got %d", len(results))
		for k, v := range results {
			t.Logf("  %s: %d examples", k, len(v))
		}
	}

	// Check pool examples exist.
	var poolKey string
	for k := range results {
		if strings.HasSuffix(k, "pool") || k == "pool" {
			poolKey = k
		}
	}
	if poolKey == "" {
		t.Fatal("expected pool examples")
	}

	poolExamples := results[poolKey]
	var names []string
	for _, ex := range poolExamples {
		names = append(names, ex.Name)
	}
	t.Logf("pool examples: %v", names)

	// Should include WithTransaction example.
	found := false
	for _, ex := range poolExamples {
		if ex.Name == "ExamplePool_WithTransaction" {
			found = true
			if !strings.Contains(ex.Code, "WithTransaction") {
				t.Errorf("ExamplePool_WithTransaction code should contain WithTransaction call:\n%s", ex.Code)
			}
		}
	}
	if !found {
		t.Error("expected ExamplePool_WithTransaction in pool examples")
	}

	// Format should produce readable markdown.
	md := examples.Format("axon-base/pool", poolExamples)
	if !strings.Contains(md, "```go") {
		t.Error("expected code fences in formatted output")
	}
	t.Logf("formatted output:\n%s", md)
}
