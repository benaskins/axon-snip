# axon-snip

Snippet composition library. Assembles axon module code fragments into a
working `main.go` with dependency-ordered initialisation.

Import: `github.com/benaskins/axon-snip`

## What it does

axon-snip provides composable code snippets, a YAML-driven catalogue of axon
modules, and runnable examples. Callers pick snippets by module name, compose
them, and emit Go source. axon-snip knows nothing about PRDs, ADRs, or LLM
orchestration.

## Packages

| Package | Purpose |
|---------|---------|
| `snippets` | Composable code fragments with topological sorting |
| `catalogue` | YAML framework catalogue loading and prompt rendering |
| `examples` | Runnable axon module examples |

## Usage

```go
import "github.com/benaskins/axon-snip/snippets"

all := snippets.GeneratedSnippets()
composed, err := snippets.Compose("my-service", pick(all, []string{"axon", "axon-talk"}))
```

## Dependencies

- `gopkg.in/yaml.v3` (catalogue parsing)

## Build & Test

```bash
go test ./...
go vet ./...
```
