// Package sidecar manages the per-branch context sidecar — a JSON file on disk
// that accumulates research findings, constraints, and summaries for a task
// branch.
//
// ADR-012: context sidecars use atomic file writes (write to .tmp, rename).
// REQ-008: SUMMARY tasks replace (not append to) sidecar content.
package sidecar

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Context is the structured content of a branch context sidecar.
type Context struct {
	// BranchID is the root task ID of this branch.
	BranchID string `json:"branch_id"`

	// Entities is a free-form map of named entities discovered during research
	// (competitors, markets, technical constraints).
	Entities map[string]any `json:"entities,omitempty"`

	// Summaries holds compressed summaries from prior SUMMARY tasks.
	Summaries []string `json:"summaries,omitempty"`

	// Notes holds raw research notes appended during RESEARCH tasks.
	Notes []string `json:"notes,omitempty"`

	// TokenEstimate is the last computed token estimate. Updated on each Append.
	TokenEstimate int `json:"token_estimate"`
}

// Load reads and parses a context sidecar from path.
// Returns an empty Context (with branchID set) if the file does not exist.
func Load(path, branchID string) (*Context, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Context{BranchID: branchID, Entities: make(map[string]any)}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("sidecar: read %q: %w", path, err)
	}

	var ctx Context
	if err := json.Unmarshal(data, &ctx); err != nil {
		return nil, fmt.Errorf("sidecar: parse %q: %w", path, err)
	}
	return &ctx, nil
}

// Save writes ctx to path atomically — write to path+".tmp", then rename.
// ADR-012: atomic writes prevent partial reads by the scheduler.
func Save(path string, ctx *Context) error {
	ctx.TokenEstimate = estimateTokens(ctx)

	data, err := json.MarshalIndent(ctx, "", "  ")
	if err != nil {
		return fmt.Errorf("sidecar: marshal: %w", err)
	}

	// Ensure parent directory exists.
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("sidecar: mkdir: %w", err)
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return fmt.Errorf("sidecar: write tmp: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("sidecar: rename tmp→final: %w", err)
	}
	return nil
}

// Append adds a research note to ctx.Notes and saves.
func Append(path string, ctx *Context, note string) error {
	ctx.Notes = append(ctx.Notes, note)
	return Save(path, ctx)
}

// Replace discards all Notes and Summaries and replaces them with a single
// compressed summary. REQ-008: SUMMARY task output replaces sidecar content.
func Replace(path string, ctx *Context, compressedSummary string) error {
	ctx.Notes = nil
	ctx.Summaries = []string{compressedSummary}
	return Save(path, ctx)
}

// ExceedsThreshold returns true when the sidecar's estimated token count exceeds
// threshold fraction of the model's context window. REQ-008.
func ExceedsThreshold(ctx *Context, contextWindow int, threshold float64) bool {
	if contextWindow <= 0 {
		return false
	}
	return float64(ctx.TokenEstimate)/float64(contextWindow) > threshold
}

// estimateTokens produces a rough token count for the sidecar content.
// Uses the approximation: 1 token ≈ 4 chars (common for English prose).
func estimateTokens(ctx *Context) int {
	raw, _ := json.Marshal(ctx)
	return len(raw) / 4
}
