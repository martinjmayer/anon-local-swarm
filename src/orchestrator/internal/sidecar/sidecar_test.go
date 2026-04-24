package sidecar_test

import (
	"os"
	"path/filepath"
	"testing"

	"als/orchestrator/internal/sidecar"
)

// ADR-012: Save uses atomic write (tmp + rename); no .tmp file lingers.
func TestSave_ADR012_AtomicWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ctx.json")

	ctx := &sidecar.Context{
		BranchID: "branch-001",
		Notes:    []string{"research finding A"},
	}

	if err := sidecar.Save(path, ctx); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// tmp file must not linger after rename.
	if _, err := os.Stat(path + ".tmp"); err == nil {
		t.Error("tmp file should not exist after Save")
	}

	// Final file must exist and parse correctly.
	loaded, err := sidecar.Load(path, "branch-001")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(loaded.Notes) != 1 {
		t.Errorf("want 1 note, got %d", len(loaded.Notes))
	}
}

// req-008: Replace discards notes and replaces with a single compressed summary.
func TestReplace_req008_ReplacesContent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ctx.json")

	ctx := &sidecar.Context{
		BranchID: "branch-002",
		Notes:    []string{"note1", "note2", "note3"},
	}
	sidecar.Save(path, ctx)

	if err := sidecar.Replace(path, ctx, "compressed summary"); err != nil {
		t.Fatalf("Replace: %v", err)
	}

	loaded, err := sidecar.Load(path, "branch-002")
	if err != nil {
		t.Fatalf("Load after Replace: %v", err)
	}
	if len(loaded.Notes) != 0 {
		t.Errorf("req-008: notes must be cleared after Replace, got %d", len(loaded.Notes))
	}
	if len(loaded.Summaries) != 1 || loaded.Summaries[0] != "compressed summary" {
		t.Errorf("req-008: wrong summaries: %v", loaded.Summaries)
	}
}

// req-008: ExceedsThreshold triggers when token estimate > threshold × window.
func TestExceedsThreshold_req008_TriggerAbove70Percent(t *testing.T) {
	// Build a context large enough to exceed 70% of a small window.
	// estimateTokens = len(json)/4. To exceed 70 tokens in a 100-token window
	// we need > 280 JSON chars.
	notes := make([]string, 20)
	for i := range notes {
		notes[i] = "This is a substantial research note with enough text to push the token estimate past the threshold."
	}
	ctx := &sidecar.Context{BranchID: "b", Notes: notes}
	path := filepath.Join(t.TempDir(), "ctx.json")
	sidecar.Save(path, ctx)

	loaded, err := sidecar.Load(path, "b")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if !sidecar.ExceedsThreshold(loaded, 100, 0.70) {
		t.Errorf("req-008: want threshold exceeded (estimate=%d, window=100)", loaded.TokenEstimate)
	}
}

// req-008: ExceedsThreshold does not trigger below threshold.
func TestExceedsThreshold_req008_NoTriggerBelow70Percent(t *testing.T) {
	ctx := &sidecar.Context{BranchID: "b", TokenEstimate: 69}
	if sidecar.ExceedsThreshold(ctx, 100, 0.70) {
		t.Error("req-008: must not trigger at 69/100 (69% < 70%)")
	}
}

// Load returns an empty Context (not an error) when the file does not exist.
func TestLoad_MissingFile_ReturnsEmpty(t *testing.T) {
	ctx, err := sidecar.Load("/nonexistent/ctx.json", "b-001")
	if err != nil {
		t.Fatalf("Load missing file should return empty context, got error: %v", err)
	}
	if ctx.BranchID != "b-001" {
		t.Errorf("want branchID b-001, got %s", ctx.BranchID)
	}
	if len(ctx.Notes) != 0 {
		t.Errorf("want empty notes, got %v", ctx.Notes)
	}
}

// Append adds a note and persists it.
func TestAppend_AddsNote(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ctx.json")
	ctx := &sidecar.Context{BranchID: "b"}
	sidecar.Save(path, ctx)

	if err := sidecar.Append(path, ctx, "new finding"); err != nil {
		t.Fatalf("Append: %v", err)
	}

	loaded, _ := sidecar.Load(path, "b")
	if len(loaded.Notes) != 1 || loaded.Notes[0] != "new finding" {
		t.Errorf("want ['new finding'], got %v", loaded.Notes)
	}
}
