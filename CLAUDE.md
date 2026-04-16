@AGENTS.md

## Conventions
- Analysis phase uses axon-loop tool-calling; the LLM calls tools to build the ScaffoldSpec incrementally
- Snippet composition uses topological sorting by dependency
- Writer uses embedded Go templates for file generation
- Catalogue files are YAML describing framework components and patterns

## Constraints
- Depends on axon-loop, axon-talk, axon-tool only
- analysis.Analyse is the non-deterministic boundary; everything after it is deterministic
- Do not add file I/O to the analysis package; that belongs in writer
- Snippets must be self-contained: each snippet declares its own imports, setup, and dependencies

## Testing
- `go test ./...` runs all tests
- `go vet ./...` for lint
