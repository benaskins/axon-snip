package snippets

import (
	"fmt"
	"sort"
	"strings"
)

// Compose produces a complete main.go source from the given snippets.
// Snippets are topologically sorted by Deps before composing.
//
// If axon-hand is among the selected snippets, main() is wrapped in
// hand.RunCLI and the axon-talk selectLLMClient helper is suppressed
// (the chassis provides the client).
func Compose(name string, snippets []Snippet) (string, error) {
	sorted, err := topoSort(snippets)
	if err != nil {
		return "", err
	}

	useChassis := hasModule(sorted, "axon-hand")

	// When using the chassis, drop axon-talk's setup and helpers
	// (selectLLMClient, envOrDefault) since hand.RunCLI provides the client.
	if useChassis {
		sorted = filterModule(sorted, "axon-talk")
	}

	var b strings.Builder

	b.WriteString("package main\n\n")

	// Imports
	imports := collectImports(sorted)
	if len(imports) > 0 {
		b.WriteString("import (\n")
		for _, imp := range imports {
			if imp.Path == "" {
				b.WriteString("\n") // blank line separator between stdlib and external
				continue
			}
			if imp.Alias != "" {
				fmt.Fprintf(&b, "\t%s %q\n", imp.Alias, imp.Path)
			} else {
				fmt.Fprintf(&b, "\t%q\n", imp.Path)
			}
		}
		b.WriteString(")\n\n")
	}

	if useChassis {
		composeWithChassis(&b, name, sorted)
	} else {
		composeFlat(&b, name, sorted)
	}

	return b.String(), nil
}

// composeWithChassis generates main() using hand.RunCLI.
func composeWithChassis(b *strings.Builder, name string, sorted []Snippet) {
	// CLI struct
	fmt.Fprintf(b, "type cli struct {\n")
	fmt.Fprintf(b, "\thand.CLI\n")
	fmt.Fprintf(b, "\tProjectDir string `kong:\"arg,required,help='Project directory'\"`\n")
	fmt.Fprintf(b, "}\n\n")

	// main()
	fmt.Fprintf(b, "func main() {\n")
	fmt.Fprintf(b, "\tc := &cli{}\n")
	fmt.Fprintf(b, "\thand.RunCLI(%q, \"0.1.0\", c, func(ctx context.Context, id hand.Identity, client talk.LLMClient) error {\n", name)

	// Get model from config
	prefix := strings.ToUpper(strings.ReplaceAll(name, "-", "_"))
	fmt.Fprintf(b, "\t\tcfg, err := hand.LoadConfig(%q)\n", prefix)
	fmt.Fprintf(b, "\t\tif err != nil {\n")
	fmt.Fprintf(b, "\t\t\treturn err\n")
	fmt.Fprintf(b, "\t\t}\n")
	fmt.Fprintf(b, "\t\t_ = cfg.Model\n\n")

	// Inner setup (axon-tool, axon-loop, etc.) indented one extra level
	for _, s := range sorted {
		if s.Setup == "" || s.Module == "axon-hand" {
			continue
		}
		// Re-indent setup lines for the closure
		lines := strings.Split(s.Setup, "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) == "" {
				b.WriteString("\n")
			} else {
				fmt.Fprintf(b, "\t%s\n", line)
			}
		}
		b.WriteString("\n")
	}

	fmt.Fprintf(b, "\t\t// TODO: wire %s agent logic here\n", name)
	fmt.Fprintf(b, "\t\treturn nil\n")
	fmt.Fprintf(b, "\t})\n")
	fmt.Fprintf(b, "}\n")

	// Helpers (skip axon-talk helpers since chassis replaces them)
	for _, s := range sorted {
		if s.Helpers == "" || s.Module == "axon-hand" {
			continue
		}
		b.WriteString("\n")
		b.WriteString(s.Helpers)
		b.WriteString("\n")
	}
}

// composeFlat generates a standard main() without the chassis.
func composeFlat(b *strings.Builder, name string, sorted []Snippet) {
	b.WriteString("func main() {\n")
	for _, s := range sorted {
		if s.Setup == "" {
			continue
		}
		fmt.Fprintf(b, "%s\n\n", s.Setup)
	}
	fmt.Fprintf(b, "\t// TODO: wire %s business logic here\n", name)
	fmt.Fprintf(b, "\tfmt.Fprintln(os.Stderr, %q, \"ready\")\n", name+":")
	b.WriteString("}\n")

	for _, s := range sorted {
		if s.Helpers == "" {
			continue
		}
		b.WriteString("\n")
		b.WriteString(s.Helpers)
		b.WriteString("\n")
	}
}

func hasModule(snippets []Snippet, module string) bool {
	for _, s := range snippets {
		if s.Module == module {
			return true
		}
	}
	return false
}

func filterModule(snippets []Snippet, module string) []Snippet {
	var out []Snippet
	for _, s := range snippets {
		if s.Module != module {
			out = append(out, s)
		}
	}
	return out
}

// collectImports deduplicates and sorts imports from all snippets.
// Standard library imports come first, then third-party.
func collectImports(snippets []Snippet) []Import {
	seen := map[string]Import{}
	for _, s := range snippets {
		for _, imp := range s.Imports {
			if _, exists := seen[imp.Path]; !exists {
				seen[imp.Path] = imp
			}
		}
	}

	var stdlib, external []Import
	for _, imp := range seen {
		if isStdlib(imp.Path) {
			stdlib = append(stdlib, imp)
		} else {
			external = append(external, imp)
		}
	}
	sort.Slice(stdlib, func(i, j int) bool { return stdlib[i].Path < stdlib[j].Path })
	sort.Slice(external, func(i, j int) bool { return external[i].Path < external[j].Path })

	if len(stdlib) > 0 && len(external) > 0 {
		// Add a blank separator import between stdlib and external
		return append(append(stdlib, Import{}), external...)
	}
	return append(stdlib, external...)
}

func isStdlib(path string) bool {
	return !strings.Contains(path, ".")
}

// topoSort orders snippets so that each snippet's Deps appear before it.
// Returns an error on cycles.
func topoSort(snippets []Snippet) ([]Snippet, error) {
	byModule := map[string]Snippet{}
	for _, s := range snippets {
		byModule[s.Module] = s
	}

	var order []Snippet
	state := map[string]int{} // 0=unvisited, 1=visiting, 2=visited

	var visit func(string) error
	visit = func(module string) error {
		switch state[module] {
		case 2:
			return nil
		case 1:
			return fmt.Errorf("snippets: dependency cycle involving %q", module)
		}
		state[module] = 1

		s, ok := byModule[module]
		if !ok {
			// Dependency not in the selected set — skip (it's not needed)
			state[module] = 2
			return nil
		}

		for _, dep := range s.Deps {
			if err := visit(dep); err != nil {
				return err
			}
		}

		state[module] = 2
		order = append(order, s)
		return nil
	}

	for _, s := range snippets {
		if err := visit(s.Module); err != nil {
			return nil, err
		}
	}
	return order, nil
}
