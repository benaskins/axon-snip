@AGENTS.md

## Conventions
- Snippet composition uses topological sorting by dependency
- Catalogue files are YAML describing framework components and patterns
- Snippets must be self-contained: each snippet declares its own imports, setup, and dependencies

## Constraints
- Plain snippet library: no LLM calls, no PRD analysis, no ADR writing
- Depends only on gopkg.in/yaml.v3
- PRD analysis, scaffold writing, and ADR recording belong in the caller (plan-lead)

## Testing
- `go test ./...` runs all tests
- `go vet ./...` for lint
