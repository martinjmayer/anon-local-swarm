package validator_test

import (
	"os/exec"
	"testing"

	"als/orchestrator/internal/validator"
)

// req-007: SelfCorrectionPayload includes original code and compiler stderr.
func TestSelfCorrectionPayload_req007_IncludesBothParts(t *testing.T) {
	payload := validator.SelfCorrectionPayload("func main() {}", "undefined: foo")
	if !contains(payload, "func main() {}") {
		t.Error("req-007: payload must include original code")
	}
	if !contains(payload, "undefined: foo") {
		t.Error("req-007: payload must include compiler stderr")
	}
}

// req-007: RetryCount returns 0 for empty meta_data.
func TestRetryCount_req007_EmptyMeta(t *testing.T) {
	if n := validator.RetryCount(""); n != 0 {
		t.Errorf("want 0, got %d", n)
	}
}

// req-007: RetryCount parses count from meta_data JSON.
func TestRetryCount_req007_ParsesCount(t *testing.T) {
	if n := validator.RetryCount(`{"self_correction_count": 3}`); n != 3 {
		t.Errorf("want 3, got %d", n)
	}
}

// req-007: IncrementRetryMeta increments from zero.
func TestIncrementRetryMeta_req007_FromZero(t *testing.T) {
	meta, err := validator.IncrementRetryMeta("")
	if err != nil {
		t.Fatalf("IncrementRetryMeta: %v", err)
	}
	if n := validator.RetryCount(meta); n != 1 {
		t.Errorf("want 1, got %d", n)
	}
}

// req-007: IncrementRetryMeta increments from an existing count.
func TestIncrementRetryMeta_req007_FromExisting(t *testing.T) {
	updated, err := validator.IncrementRetryMeta(`{"self_correction_count": 4}`)
	if err != nil {
		t.Fatalf("IncrementRetryMeta: %v", err)
	}
	if n := validator.RetryCount(updated); n != 5 {
		t.Errorf("want 5, got %d", n)
	}
}

// req-007: Validate succeeds on valid Go code (requires go on PATH).
func TestValidate_req007_ValidCode(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go not on PATH — skipping build validation test")
	}
	result, err := validator.Validate(t.Context(), `package main

import "fmt"

func main() {
	fmt.Println("hello")
}
`)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if !result.Success {
		t.Errorf("req-007: valid code should pass build, stderr: %s", result.Stderr)
	}
}

// req-007: Validate returns failure and stderr for invalid code.
func TestValidate_req007_InvalidCode(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go not on PATH")
	}
	result, err := validator.Validate(t.Context(), `package main

func main() {
	undefined_function()
}
`)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if result.Success {
		t.Error("req-007: invalid code should fail build")
	}
	if result.Stderr == "" {
		t.Error("req-007: want non-empty stderr on build failure")
	}
}

func contains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
