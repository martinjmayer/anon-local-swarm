package deposit_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"als/orchestrator/internal/deposit"
)

// req-010: Write creates output directory and writes all artifacts.
func TestWrite_req010_CreatesFilesInOutputDir(t *testing.T) {
	dir := t.TempDir()
	outputDir := filepath.Join(dir, "project-output")

	artifacts := []deposit.Artifact{
		{Filename: "research-summary.md", Content: "# Research\n\nFindings here."},
		{Filename: "viability-report.md", Content: "# Viability\n\nScore: 7"},
		{Filename: "main.go", Content: "package main\n\nfunc main() {}"},
	}

	if err := deposit.Write(outputDir, artifacts); err != nil {
		t.Fatalf("Write: %v", err)
	}

	for _, a := range artifacts {
		path := filepath.Join(outputDir, a.Filename)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("req-010: artifact %q not written: %v", a.Filename, err)
			continue
		}
		if string(data) != a.Content {
			t.Errorf("req-010: artifact %q content mismatch", a.Filename)
		}
	}
}

// req-010: Fully-pruned project writes prune-report.md.
func TestPruneReport_req010_ContainsBranchInfo(t *testing.T) {
	events := []deposit.PruneEvent{
		{BranchTaskID: "task-001", ViabilityScore: 3.5, PrunedCount: 5, Reason: "Score below threshold"},
		{BranchTaskID: "task-002", ViabilityScore: 2.0, PrunedCount: 3, Reason: "Market too crowded"},
	}

	report := deposit.PruneReport("Test Project", events)

	if !strings.Contains(report, "task-001") {
		t.Error("req-010: prune report missing branch task-001")
	}
	if !strings.Contains(report, "task-002") {
		t.Error("req-010: prune report missing branch task-002")
	}
	if !strings.Contains(report, "3.5") {
		t.Error("req-010: prune report missing viability score 3.5")
	}
}

// req-010: Write does not write to external systems (no side effects beyond dir).
func TestWrite_req010_OnlyWritesToOutputDir(t *testing.T) {
	dir := t.TempDir()
	outputDir := filepath.Join(dir, "output")

	deposit.Write(outputDir, []deposit.Artifact{
		{Filename: "test.md", Content: "test"},
	})

	// Verify only one subdirectory was created under temp dir.
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Errorf("req-010: expected only output dir, found %d entries", len(entries))
	}
}

// req-010: Output directory is created if it does not exist.
func TestWrite_req010_CreatesOutputDirIfMissing(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "a", "b", "c")

	if err := deposit.Write(nested, []deposit.Artifact{{Filename: "f.md", Content: "x"}}); err != nil {
		t.Fatalf("Write with nested path: %v", err)
	}
	if _, err := os.Stat(filepath.Join(nested, "f.md")); err != nil {
		t.Errorf("req-010: artifact not found in nested output dir: %v", err)
	}
}
