// Package deposit writes terminal-state output artifacts to the project output
// directory. REQ-010.
//
// On terminal state (completed or pruned), the orchestrator calls deposit.Write
// which collects all relevant task outputs and writes them as discrete files.
package deposit

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Artifact is a named output file to be written to the project directory.
type Artifact struct {
	// Filename is the deterministic, human-readable output filename.
	Filename string
	// Content is the file content to write.
	Content string
}

// Write creates outputDir if it does not exist, then writes each artifact as a
// discrete file. REQ-010: filenames are deterministic and human-readable.
func Write(outputDir string, artifacts []Artifact) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("deposit: mkdir %q: %w", outputDir, err)
	}

	for _, a := range artifacts {
		path := filepath.Join(outputDir, a.Filename)
		if err := os.WriteFile(path, []byte(a.Content), 0644); err != nil {
			return fmt.Errorf("deposit: write %q: %w", path, err)
		}
	}
	return nil
}

// PruneReport builds the prune-report.md content for a fully-pruned project.
// REQ-010: if the project is fully pruned, a prune-report.md is written.
func PruneReport(projectName string, pruneEvents []PruneEvent) string {
	var sb strings.Builder
	sb.WriteString("# Prune Report\n\n")
	fmt.Fprintf(&sb, "**Project:** %s\n", projectName)
	fmt.Fprintf(&sb, "**Generated:** %s\n\n", time.Now().UTC().Format(time.RFC3339))
	sb.WriteString("## Pruned Branches\n\n")
	if len(pruneEvents) == 0 {
		sb.WriteString("No branches were pruned.\n")
		return sb.String()
	}
	sb.WriteString("| Branch Task ID | Viability Score | Pruned Task Count | Reason |\n")
	sb.WriteString("|---|---|---|---|\n")
	for _, e := range pruneEvents {
		fmt.Fprintf(&sb, "| %s | %.1f | %d | %s |\n",
			e.BranchTaskID, e.ViabilityScore, e.PrunedCount, e.Reason)
	}
	return sb.String()
}

// PruneEvent describes a single cascading prune event for the report.
type PruneEvent struct {
	BranchTaskID   string
	ViabilityScore float64
	PrunedCount    int
	Reason         string
}

// RunSummary builds the run-summary.md content for a completed project.
func RunSummary(projectName string, taskOutputs []TaskOutput) string {
	var sb strings.Builder
	sb.WriteString("# Run Summary\n\n")
	fmt.Fprintf(&sb, "**Project:** %s\n", projectName)
	fmt.Fprintf(&sb, "**Generated:** %s\n\n", time.Now().UTC().Format(time.RFC3339))
	sb.WriteString("## Outputs\n\n")
	for _, t := range taskOutputs {
		fmt.Fprintf(&sb, "### %s (%s)\n\n%s\n\n", t.TaskID, t.TaskType, t.Output)
	}
	return sb.String()
}

// TaskOutput is a completed task's final output for inclusion in the run summary.
type TaskOutput struct {
	TaskID   string
	TaskType string
	Output   string
}
