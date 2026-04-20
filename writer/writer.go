// Package writer deterministically writes a project scaffold from a ScaffoldSpec.
package writer

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/benaskins/axon-snip/analysis"
	"github.com/benaskins/axon-snip/catalogue"
)

//go:embed templates
var templateFS embed.FS

// templateData is the common data passed to every template.
type templateData struct {
	Name        string
	ModulePath  string // e.g. github.com/benaskins/my-service
	Type        analysis.ProjectType
	Description string
	Modules     []analysis.ModuleSelection
	Boundaries  []analysis.Boundary
	PlanSteps   []analysis.PlanStep
	Constraints []string
	Requires    []string // go.mod require paths from snippets
	Date        string
	Catalogue   *catalogue.Catalogue // nil when no catalogue was used
}

// ComposedOutput holds the pre-composed source files from snippet composition.
type ComposedOutput struct {
	MainGo   string   // composed main.go source
	Requires []string // deduplicated go.mod require paths
}

// Options holds optional configuration for Write.
type Options struct {
	Composed      *ComposedOutput
	Catalogue     *catalogue.Catalogue
	CataloguePath string // original YAML path; copied into scaffold as catalogue.yaml
}

// Write creates outDir and writes the scaffold files derived from spec.
// If opts.Composed is non-nil, main.go uses the composed source and go.mod
// includes the require lines. Otherwise falls back to the stub template.
func Write(spec *analysis.ScaffoldSpec, outDir string, opts *Options) error {
	if opts == nil {
		opts = &Options{}
	}
	composed := opts.Composed
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("writer: create output dir: %w", err)
	}

	funcMap := template.FuncMap{
		"add": func(a, b int) int { return a + b },
	}

	tmpl, err := template.New("").Funcs(funcMap).ParseFS(templateFS, "templates/*.tmpl")
	if err != nil {
		return fmt.Errorf("writer: parse templates: %w", err)
	}

	var requires []string
	if composed != nil {
		requires = composed.Requires
	}
	// Derive requires from selected modules when snippets don't provide them.
	if len(requires) == 0 {
		for _, m := range spec.Modules {
			requires = append(requires, analysis.ModulePrefix+"/"+m.Name)
		}
	}

	data := templateData{
		Name:        spec.Name,
		ModulePath:  spec.ModulePath,
		Type:        spec.Type,
		Description: descriptionFromSpec(spec),
		Modules:     spec.Modules,
		Boundaries:  spec.Boundaries,
		PlanSteps:   spec.PlanSteps,
		Constraints: spec.Constraints,
		Requires:    requires,
		Date:        time.Now().Format("2006-01-02"),
		Catalogue:   opts.Catalogue,
	}

	files := []struct {
		tmplName string
		path     string
	}{
		{"AGENTS.md.tmpl", "AGENTS.md"},
		{"CLAUDE.md.tmpl", "CLAUDE.md"},
		{"README.md.tmpl", "README.md"},
		{"go.mod.tmpl", "go.mod"},
		{"justfile.tmpl", "justfile"},
		{"plan.md.tmpl", filepath.Join("plans", data.Date+"-initial-build.md")},
	}

	// main.go: skip for libraries, use composed source or template for services/CLIs.
	if spec.Type != analysis.ProjectLibrary {
		mainPath := filepath.Join(outDir, "cmd", spec.Name, "main.go")
		if composed != nil && composed.MainGo != "" {
			if err := os.MkdirAll(filepath.Dir(mainPath), 0o755); err != nil {
				return fmt.Errorf("writer: mkdir %s: %w", filepath.Dir(mainPath), err)
			}
			if err := os.WriteFile(mainPath, []byte(composed.MainGo), 0o644); err != nil {
				return fmt.Errorf("writer: write main.go: %w", err)
			}
		} else {
			files = append(files, struct {
				tmplName string
				path     string
			}{"main.go.tmpl", filepath.Join("cmd", spec.Name, "main.go")})
		}
	}

	for _, f := range files {
		if err := writeFile(tmpl, f.tmplName, filepath.Join(outDir, f.path), data); err != nil {
			return err
		}
	}

	// Copy catalogue YAML into the scaffold so downstream tools (maestro, inspector) can read it.
	if opts.CataloguePath != "" {
		src, err := os.ReadFile(opts.CataloguePath)
		if err != nil {
			return fmt.Errorf("writer: read catalogue: %w", err)
		}
		if err := os.WriteFile(filepath.Join(outDir, "catalogue.yaml"), src, 0o644); err != nil {
			return fmt.Errorf("writer: write catalogue.yaml: %w", err)
		}
	}

	return nil
}

func writeFile(tmpl *template.Template, tmplName, destPath string, data templateData) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return fmt.Errorf("writer: mkdir %s: %w", filepath.Dir(destPath), err)
	}
	f, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("writer: create %s: %w", destPath, err)
	}
	defer f.Close()
	if err := tmpl.ExecuteTemplate(f, tmplName, data); err != nil {
		return fmt.Errorf("writer: execute template %s: %w", tmplName, err)
	}
	return nil
}

// descriptionFromSpec derives a one-line description from the module selections.
func descriptionFromSpec(spec *analysis.ScaffoldSpec) string {
	if len(spec.Modules) == 0 {
		return spec.Name + " -- axon-based " + string(spec.Type)
	}
	// Build description from primary module purposes
	var parts []string
	for _, m := range spec.Modules {
		if m.Name == "axon-hand" || m.Name == "axon-tool" || m.Name == "axon-talk" {
			continue // skip infrastructure modules
		}
		parts = append(parts, m.Reason)
	}
	if len(parts) == 0 {
		return spec.Name + " -- axon-based " + string(spec.Type)
	}
	return parts[0]
}
