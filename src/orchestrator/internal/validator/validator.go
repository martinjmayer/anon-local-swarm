// Package validator runs `go build` on generated Go source and implements the
// self-correction loop described in REQ-007.
//
// REQ-007: After every GO_CODE task, go build is executed. On failure a new
// GO_CODE task is emitted with the original code + stderr in the payload.
// The loop halts after MaxSelfCorrections consecutive failures on the same branch.
package validator

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Result is the outcome of a build validation.
type Result struct {
	// Success is true when go build exited 0.
	Success bool

	// Stderr contains the compiler output on failure. Empty on success.
	Stderr string
}

// Validate writes code to a temporary directory and runs `go build ./...` on it.
// REQ-007: uses os/exec; returns the compiler stderr on failure.
func Validate(ctx context.Context, code string) (*Result, error) {
	dir, err := os.MkdirTemp("", "als-validate-*")
	if err != nil {
		return nil, fmt.Errorf("validator: mktemp: %w", err)
	}
	defer os.RemoveAll(dir)

	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(code), 0644); err != nil {
		return nil, fmt.Errorf("validator: write main.go: %w", err)
	}
	gomod := "module als/generated\n\ngo 1.23\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(gomod), 0644); err != nil {
		return nil, fmt.Errorf("validator: write go.mod: %w", err)
	}

	cmd := exec.CommandContext(ctx, "go", "build", "./...")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return &Result{Success: false, Stderr: string(out)}, nil
	}
	return &Result{Success: true}, nil
}

// SelfCorrectionPayload constructs the payload for a new GO_CODE self-correction
// task from the original code and the compiler stderr. REQ-007.
func SelfCorrectionPayload(originalCode, compilerStderr string) string {
	return fmt.Sprintf(
		"## Original Code\n\n%s\n\n## Compiler Error\n\n%s\n\n## Instruction\n\nFix the compiler error above. Return only the corrected Go code.",
		originalCode, compilerStderr,
	)
}

// RetryCount extracts the self_correction_count from a task's meta_data JSON.
// Returns 0 if not set or unparseable.
func RetryCount(metaJSON string) int {
	if metaJSON == "" {
		return 0
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(metaJSON), &m); err != nil {
		return 0
	}
	v, ok := m["self_correction_count"]
	if !ok {
		return 0
	}
	if n, ok := v.(float64); ok {
		return int(n)
	}
	return 0
}

// IncrementRetryMeta returns a new meta_data JSON string with self_correction_count
// incremented by 1. REQ-007.
func IncrementRetryMeta(metaJSON string) (string, error) {
	m := make(map[string]any)
	if metaJSON != "" {
		_ = json.Unmarshal([]byte(metaJSON), &m)
	}
	count := 0
	if v, ok := m["self_correction_count"]; ok {
		if n, ok := v.(float64); ok {
			count = int(n)
		}
	}
	m["self_correction_count"] = count + 1
	b, err := json.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("validator: marshal meta: %w", err)
	}
	return string(b), nil
}
