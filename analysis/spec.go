// Package analysis implements the structured output call that converts a PRD
// into a ScaffoldSpec.
package analysis

import (
	"bufio"
	"fmt"
	"os"
)

// ProjectType indicates whether the scaffold is a library or a runnable service/CLI.
type ProjectType string

const (
	ProjectLibrary ProjectType = "library"
	ProjectService ProjectType = "service"
	ProjectCLI     ProjectType = "cli"
)

// ModulePrefix is the Go module path prefix used for all factory projects.
const ModulePrefix = "github.com/benaskins"

// ScaffoldSpec is the machine-readable output of the analysis call.
type ScaffoldSpec struct {
	Name        string            `json:"name"`
	ModulePath  string            `json:"module_path"` // e.g. github.com/benaskins/my-service
	Type        ProjectType       `json:"type"`
	Modules     []ModuleSelection `json:"modules"`
	Boundaries  []Boundary        `json:"boundaries"`
	PlanSteps   []PlanStep        `json:"plan_steps"`
	Constraints []string          `json:"constraints"`
	Gaps        []Gap             `json:"gaps"`
}

type Resolver interface {
	Resolve(gap *Gap) (string, error)
}

type StdinResolver struct{}

func (s *StdinResolver) Resolve(gap *Gap) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Fprintf(os.Stdout, "%s (Context: %s): ", gap.Question, gap.Context)
	return reader.ReadString('\n')
}

func WithResolver(resolver Resolver) func(*Gap) {
	return func(gap *Gap) {
		gap.Resolver = resolver
	}
}

// ModuleSelection records which axon module was selected and why.
type ModuleSelection struct {
	Name            string `json:"name"`
	Reason          string `json:"reason"`
	IsDeterministic bool   `json:"is_deterministic"`
}

// Boundary describes the interface between two components.
type Boundary struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"` // "det" or "non-det"
}

// PlanStep is one commit-sized implementation step for the generated plan.
type PlanStep struct {
	Title         string `json:"title"`
	Description   string `json:"description"`
	CommitMessage string `json:"commit_message"`
}

type Gap struct {
	Question string `json:"question"`
	Context  string `json:"context"`
	Resolver Resolver
}