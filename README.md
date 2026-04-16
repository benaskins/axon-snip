# axon-snip

Code assembly engine: PRD analysis, module selection, and scaffold generation.

Import: `github.com/benaskins/axon-snip`

## What it does

axon-snip transforms a product requirements document (PRD) into a complete project scaffold. It operates in stages:

1. **Analysis** (`analysis`): LLM reads the PRD and calls tools to select modules, define boundaries, plan implementation steps, and extract constraints. Output is a `ScaffoldSpec`.
2. **Gap resolution** (`gaps`): Resolves ambiguities in the spec through conversational iteration.
3. **Snippet composition** (`snippets`): Assembles axon module snippets into a working `main.go` with dependency-ordered initialisation.
4. **Writing** (`writer`): Generates the project scaffold from the spec using embedded templates.

## Packages

| Package | Purpose |
|---------|---------|
| `analysis` | LLM-driven PRD to ScaffoldSpec via tool-calling |
| `gaps` | Conversational gap resolution |
| `snippets` | Composable code fragments with topological sorting |
| `writer` | Deterministic file generation from spec + templates |
| `catalogue` | YAML framework catalogue loading and prompt rendering |
| `patterns` | Embedded system prompt for analysis |

## Usage

```go
import (
    "github.com/benaskins/axon-snip/analysis"
    "github.com/benaskins/axon-snip/snippets"
    "github.com/benaskins/axon-snip/writer"
)

// Analyse PRD
spec, err := analysis.Analyse(ctx, prd, client, model)

// Compose snippets
composed, err := snippets.Compose(spec.Name, selectedSnippets)

// Write scaffold
err = writer.Write(spec, outDir, &writer.Options{Composed: composed})
```

## Dependencies

- axon-loop, axon-talk, axon-tool

## Build & Test

```bash
go test ./...
go vet ./...
```
