// Package examples extracts Example* test functions from Go source files
// and formats them as markdown documentation for LLM system prompts.
package examples

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// Example is a single extracted example function.
type Example struct {
	Name string // e.g. "ExampleNew", "ExamplePool_WithTransaction"
	Code string // function body (without func declaration and braces)
}

// ExtractFromDir parses all *_test.go files in dir and returns
// Example functions found. Returns nil (not error) if dir has no
// Go test files or no examples.
func ExtractFromDir(dir string) ([]Example, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, func(fi os.FileInfo) bool {
		return strings.HasSuffix(fi.Name(), "_test.go")
	}, parser.ParseComments)
	if err != nil {
		// No Go files or parse errors — not fatal.
		return nil, nil
	}

	var out []Example
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Recv != nil {
					continue
				}
				if !strings.HasPrefix(fn.Name.Name, "Example") {
					continue
				}
				code, err := extractBody(fset, fn)
				if err != nil {
					continue
				}
				out = append(out, Example{
					Name: fn.Name.Name,
					Code: code,
				})
			}
		}
	}
	return out, nil
}

// ExtractFromModule walks a module root directory and extracts examples
// from all sub-packages. Returns a map of relative package path to examples.
// Packages with no examples are omitted.
func ExtractFromModule(root string) (map[string][]Example, error) {
	result := make(map[string][]Example)

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip inaccessible dirs
		}
		if !info.IsDir() {
			return nil
		}
		// Skip hidden dirs and common non-source dirs.
		name := info.Name()
		if strings.HasPrefix(name, ".") || name == "testdata" || name == "vendor" {
			return filepath.SkipDir
		}

		exs, err := ExtractFromDir(path)
		if err != nil || len(exs) == 0 {
			return nil
		}

		rel, _ := filepath.Rel(root, path)
		if rel == "." {
			rel = filepath.Base(root)
		}
		result[rel] = exs
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Format renders examples as a markdown section for a module.
// Returns empty string if examples is empty.
func Format(module string, examples []Example) string {
	if len(examples) == 0 {
		return ""
	}

	var b strings.Builder
	fmt.Fprintf(&b, "### %s\n\n", module)
	for _, ex := range examples {
		fmt.Fprintf(&b, "**%s**\n\n```go\n%s\n```\n\n", ex.Name, ex.Code)
	}
	return b.String()
}

// extractBody reads the function body source, strips the outer braces,
// and removes one level of indentation.
func extractBody(fset *token.FileSet, fn *ast.FuncDecl) (string, error) {
	if fn.Body == nil {
		return "", fmt.Errorf("no body")
	}

	// Read the source file to get the raw body text.
	pos := fset.Position(fn.Body.Lbrace)
	end := fset.Position(fn.Body.Rbrace)

	data, err := os.ReadFile(pos.Filename)
	if err != nil {
		return "", err
	}

	// Extract between braces (exclusive).
	body := string(data[pos.Offset+1 : end.Offset])

	// Dedent: remove one tab from the start of each line.
	lines := strings.Split(body, "\n")
	var out []string
	for _, line := range lines {
		line = strings.TrimPrefix(line, "\t")
		out = append(out, line)
	}

	return strings.TrimSpace(strings.Join(out, "\n")), nil
}
