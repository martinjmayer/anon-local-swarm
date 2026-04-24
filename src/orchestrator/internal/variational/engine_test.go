package variational_test

import (
	"context"
	"fmt"
	"testing"

	"als/orchestrator/internal/ollama"
	"als/orchestrator/internal/variational"
)

// Stub generator that returns a fixed response per call index.
type stubGenerator struct {
	responses []string
	calls     int
}

func (s *stubGenerator) generate(ctx context.Context, model, system, prompt string, opts map[string]any) (*ollama.GenerateResponse, error) {
	if s.calls >= len(s.responses) {
		return nil, fmt.Errorf("stub: unexpected call %d", s.calls)
	}
	resp := &ollama.GenerateResponse{Response: s.responses[s.calls]}
	s.calls++
	return resp, nil
}

// req-005: Run produces exactly VARIATION_COUNT candidates.
func TestRun_req005_ProducesVariationCountCandidates(t *testing.T) {
	stub := &stubGenerator{responses: []string{"output-1", "output-2", "output-3"}}
	cfg := variational.Config{
		VariationCount: 3,
		JudgeModel:     "phi-4:mini",
	}

	candidates, err := variational.Run(context.Background(), cfg, stub.generate, "phi-4:mini", "system", "task prompt")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(candidates) != 3 {
		t.Errorf("req-005: want 3 candidates, got %d", len(candidates))
	}
}

// req-005: Each candidate has a non-empty output.
func TestRun_req005_CandidatesHaveOutput(t *testing.T) {
	stub := &stubGenerator{responses: []string{"alpha", "beta", "gamma"}}
	cfg := variational.Config{VariationCount: 3, JudgeModel: "phi-4:mini"}

	candidates, _ := variational.Run(context.Background(), cfg, stub.generate, "phi-4:mini", "", "prompt")
	for i, c := range candidates {
		if c.Output == "" {
			t.Errorf("req-005: candidate %d has empty output", i)
		}
	}
}

// req-005: Different temperatures are passed for each candidate run.
func TestRun_req005_DifferentTemperatures(t *testing.T) {
	capturedTemps := make([]float64, 0)
	gen := func(ctx context.Context, model, system, prompt string, opts map[string]any) (*ollama.GenerateResponse, error) {
		if temp, ok := opts["temperature"].(float64); ok {
			capturedTemps = append(capturedTemps, temp)
		}
		return &ollama.GenerateResponse{Response: "ok"}, nil
	}
	cfg := variational.Config{VariationCount: 3, JudgeModel: "phi-4:mini"}
	variational.Run(context.Background(), cfg, gen, "phi-4:mini", "", "prompt")

	if len(capturedTemps) != 3 {
		t.Fatalf("want 3 temperature values, got %d", len(capturedTemps))
	}
	// Temperatures must not all be the same.
	allSame := capturedTemps[0] == capturedTemps[1] && capturedTemps[1] == capturedTemps[2]
	if allSame {
		t.Error("req-005: all temperatures are identical — variation not applied")
	}
}

// req-005: Judge returns valid index and score.
func TestJudge_req005_ReturnsValidSelection(t *testing.T) {
	judgeGen := func(ctx context.Context, model, system, prompt string, opts map[string]any) (*ollama.GenerateResponse, error) {
		return &ollama.GenerateResponse{
			Response: "SELECTED_CANDIDATE: candidate-2\nEVALUATION_SCORE: 8\nRATIONALE: Best structure.",
		}, nil
	}
	cfg := variational.Config{VariationCount: 3, JudgeModel: "phi-4:mini"}
	candidates := []variational.Candidate{
		{ID: "c1", Output: "output 1"},
		{ID: "c2", Output: "output 2"},
		{ID: "c3", Output: "output 3"},
	}

	idx, score, err := variational.Judge(context.Background(), cfg, judgeGen, "task prompt", candidates)
	if err != nil {
		t.Fatalf("Judge: %v", err)
	}
	if idx != 1 { // 0-based index for candidate-2
		t.Errorf("req-005: want index 1 (candidate-2), got %d", idx)
	}
	if score != 8.0 {
		t.Errorf("req-005: want score 8.0, got %f", score)
	}
}

// req-005: Single candidate auto-selects without calling judge.
func TestJudge_req005_SingleCandidateAutoSelect(t *testing.T) {
	callCount := 0
	gen := func(ctx context.Context, model, system, prompt string, opts map[string]any) (*ollama.GenerateResponse, error) {
		callCount++
		return &ollama.GenerateResponse{Response: "unused"}, nil
	}
	cfg := variational.Config{VariationCount: 1, JudgeModel: "phi-4:mini"}
	candidates := []variational.Candidate{{ID: "c1", Output: "only output"}}

	idx, score, err := variational.Judge(context.Background(), cfg, gen, "prompt", candidates)
	if err != nil {
		t.Fatalf("Judge: %v", err)
	}
	if idx != 0 {
		t.Errorf("want index 0, got %d", idx)
	}
	if score != 10.0 {
		t.Errorf("want score 10.0 for auto-select, got %f", score)
	}
	if callCount != 0 {
		t.Errorf("req-005: judge must not be called for single candidate, got %d calls", callCount)
	}
}

// DefaultTemperatures returns the right count and spread.
func TestDefaultTemperatures_Spread(t *testing.T) {
	temps := variational.DefaultTemperatures(3)
	if len(temps) != 3 {
		t.Fatalf("want 3, got %d", len(temps))
	}
	if temps[0] >= temps[2] {
		t.Errorf("temperatures must increase: %v", temps)
	}
}
