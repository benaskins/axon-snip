package examples_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/benaskins/axon-snip/examples"
)

func TestExtractFromDir(t *testing.T) {
	// Create a temp directory with a fake module containing example tests.
	dir := t.TempDir()

	// Write a package with example functions.
	pkgDir := filepath.Join(dir, "widget")
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Main source file (not examples — should be ignored).
	if err := os.WriteFile(filepath.Join(pkgDir, "widget.go"), []byte(`package widget

// Widget does things.
type Widget struct{ Name string }

// New creates a Widget.
func New(name string) *Widget { return &Widget{Name: name} }
`), 0o644); err != nil {
		t.Fatal(err)
	}

	// Example test file.
	if err := os.WriteFile(filepath.Join(pkgDir, "example_test.go"), []byte(`package widget_test

import (
	"fmt"
	"example.com/widget"
)

func ExampleNew() {
	w := widget.New("bolt")
	fmt.Println(w.Name)
	// Output: bolt
}

func ExampleWidget_String() {
	w := widget.New("nut")
	fmt.Println(w)
}
`), 0o644); err != nil {
		t.Fatal(err)
	}

	results, err := examples.ExtractFromDir(pkgDir)
	if err != nil {
		t.Fatalf("ExtractFromDir: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 examples, got %d", len(results))
	}

	// Check first example.
	ex := results[0]
	if ex.Name != "ExampleNew" {
		t.Errorf("expected ExampleNew, got %s", ex.Name)
	}
	if !strings.Contains(ex.Code, `widget.New("bolt")`) {
		t.Errorf("expected code to contain widget.New call, got:\n%s", ex.Code)
	}
	// Code should not include the func declaration line itself.
	if strings.HasPrefix(ex.Code, "func ") {
		t.Errorf("code should not start with func declaration:\n%s", ex.Code)
	}

	// Check second example.
	if results[1].Name != "ExampleWidget_String" {
		t.Errorf("expected ExampleWidget_String, got %s", results[1].Name)
	}
}

func TestExtractFromDir_NoExamples(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	results, err := examples.ExtractFromDir(dir)
	if err != nil {
		t.Fatalf("ExtractFromDir: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 examples, got %d", len(results))
	}
}

func TestExtractFromDir_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	results, err := examples.ExtractFromDir(dir)
	if err != nil {
		t.Fatalf("ExtractFromDir: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 examples, got %d", len(results))
	}
}

func TestFormat(t *testing.T) {
	exs := []examples.Example{
		{Name: "ExampleNew", Code: "w := widget.New(\"bolt\")\nfmt.Println(w.Name)"},
		{Name: "ExampleWidget_String", Code: "w := widget.New(\"nut\")\nfmt.Println(w)"},
	}

	out := examples.Format("widget", exs)
	if !strings.Contains(out, "### widget") {
		t.Errorf("expected module header, got:\n%s", out)
	}
	if !strings.Contains(out, "ExampleNew") {
		t.Errorf("expected ExampleNew in output, got:\n%s", out)
	}
	if !strings.Contains(out, "```go") {
		t.Errorf("expected code fence, got:\n%s", out)
	}
}

func TestFormatEmpty(t *testing.T) {
	out := examples.Format("widget", nil)
	if out != "" {
		t.Errorf("expected empty string for no examples, got: %q", out)
	}
}

func TestExtractFromModule(t *testing.T) {
	// Create a module-like structure with multiple packages.
	dir := t.TempDir()

	// Package with examples.
	aDir := filepath.Join(dir, "alpha")
	os.MkdirAll(aDir, 0o755)
	os.WriteFile(filepath.Join(aDir, "alpha.go"), []byte("package alpha\n"), 0o644)
	os.WriteFile(filepath.Join(aDir, "example_test.go"), []byte(`package alpha_test

func ExampleDo() {
	// does stuff
}
`), 0o644)

	// Package without examples.
	bDir := filepath.Join(dir, "beta")
	os.MkdirAll(bDir, 0o755)
	os.WriteFile(filepath.Join(bDir, "beta.go"), []byte("package beta\n"), 0o644)

	// Nested package with examples.
	cDir := filepath.Join(dir, "gamma", "sub")
	os.MkdirAll(cDir, 0o755)
	os.WriteFile(filepath.Join(cDir, "sub.go"), []byte("package sub\n"), 0o644)
	os.WriteFile(filepath.Join(cDir, "example_test.go"), []byte(`package sub_test

func ExampleRun() {
	// runs stuff
}
`), 0o644)

	results, err := examples.ExtractFromModule(dir)
	if err != nil {
		t.Fatalf("ExtractFromModule: %v", err)
	}

	// Should find examples from alpha and gamma/sub, but not beta.
	if len(results) != 2 {
		t.Fatalf("expected 2 packages with examples, got %d: %v", len(results), keys(results))
	}

	if _, ok := results["alpha"]; !ok {
		t.Error("expected alpha in results")
	}
	if _, ok := results["gamma/sub"]; !ok {
		t.Error("expected gamma/sub in results")
	}
	if _, ok := results["beta"]; ok {
		t.Error("did not expect beta in results")
	}
}

func keys(m map[string][]examples.Example) []string {
	var ks []string
	for k := range m {
		ks = append(ks, k)
	}
	return ks
}
