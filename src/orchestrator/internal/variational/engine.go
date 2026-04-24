// Package variational implements REQ-005: multi-candidate generation and
// judge-model selection for high-priority tasks.
//
// For eligible tasks the engine runs the same prompt VARIATION_COUNT times
// (with varying temperature), stores each output as a task_candidate row, then
// invokes the judge model (phi-4:mini) to select the best candidate.
package variational

import (
	"context"
	"fmt"
	"strings"

	"als/orchestrator/internal/ollama"
)

// Candidate holds a single generated output before it is persisted.
type Candidate struct {
	ID     string
	Output string
}

// Config parameterises the variational engine.
type Config struct {
	// VariationCount is the number of candidates to generate (e.g. 3).
	VariationCount int

	// JudgeModel is the Ollama model tag used for evaluation (phi-4:mini).
	JudgeModel string

	// Temperatures is the list of temperature values used per candidate run.
	// Length must equal VariationCount. If nil, defaults are applied.
	Temperatures []float64
}

// DefaultTemperatures returns the default temperature ladder for n candidates.
// Spreads from focused (0.3) through standard (0.7) to exploratory (1.0).
func DefaultTemperatures(n int) []float64 {
	if n <= 0 {
		return nil
	}
	temps := make([]float64, n)
	for i := range temps {
		temps[i] = 0.3 + (0.7/float64(n-1))*float64(i)
		if n == 1 {
			temps[i] = 0.7
		}
	}
	return temps
}

// Generator is the function signature used by the engine to call Ollama.
// It matches ollama.Client.Generate so it can be swapped in tests.
type Generator func(ctx context.Context, model, system, prompt string, opts map[string]any) (*ollama.GenerateResponse, error)

// Run generates VARIATION_COUNT candidates for the given task and returns them
// unordered. REQ-005: each candidate run uses a different temperature.
func Run(ctx context.Context, cfg Config, gen Generator, model, system, prompt string) ([]Candidate, error) {
	if cfg.VariationCount <= 0 {
		return nil, fmt.Errorf("variational: VariationCount must be >= 1")
	}

	temps := cfg.Temperatures
	if len(temps) == 0 {
		temps = DefaultTemperatures(cfg.VariationCount)
	}

	candidates := make([]Candidate, 0, cfg.VariationCount)
	for i := 0; i < cfg.VariationCount; i++ {
		opts := map[string]any{"temperature": temps[i]}
		resp, err := gen(ctx, model, system, prompt, opts)
		if err != nil {
			return nil, fmt.Errorf("variational: candidate %d: %w", i, err)
		}
		candidates = append(candidates, Candidate{
			ID:     fmt.Sprintf("candidate-%d", i+1),
			Output: resp.Response,
		})
	}
	return candidates, nil
}

// Judge invokes the judge model to select the best candidate. Returns the index
// of the selected candidate and the evaluation score. REQ-005.
//
// The judge is given all candidates formatted as a numbered list and asked to
// output SELECTED_CANDIDATE: <index> and EVALUATION_SCORE: <0-10>.
func Judge(ctx context.Context, cfg Config, gen Generator, taskPrompt string, candidates []Candidate) (int, float64, error) {
	if len(candidates) == 0 {
		return 0, 0, fmt.Errorf("variational: no candidates to judge")
	}
	if len(candidates) == 1 {
		return 0, 10.0, nil // only one option; auto-select
	}

	judgePrompt := buildJudgePrompt(taskPrompt, candidates)
	const judgeSystem = `You are a code and content judge. You evaluate candidate outputs and select the best one.
Respond with:
SELECTED_CANDIDATE: <1-based index>
EVALUATION_SCORE: <0-10>
RATIONALE: <one sentence>`

	resp, err := gen(ctx, cfg.JudgeModel, judgeSystem, judgePrompt, map[string]any{"temperature": 0.1})
	if err != nil {
		return 0, 0, fmt.Errorf("variational: judge call: %w", err)
	}

	parsed := ollama.ParseOutput(resp.Response)

	// Prefer SELECTED_CANDIDATE ID format; fall back to scanning for a number.
	selectedIdx := 0
	if parsed.SelectedCandidateID != "" {
		// Parse "candidate-N" format or plain integer.
		var n int
		if _, err := fmt.Sscanf(strings.TrimPrefix(parsed.SelectedCandidateID, "candidate-"), "%d", &n); err == nil {
			selectedIdx = n - 1
		}
	}

	// Clamp to valid range.
	if selectedIdx < 0 || selectedIdx >= len(candidates) {
		selectedIdx = 0
	}

	score := parsed.ViabilityScore // reuse the same parser field for evaluation score
	if score == 0 {
		// Try to read from the raw response.
		for _, line := range strings.Split(resp.Response, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "EVALUATION_SCORE:") {
				fmt.Sscanf(strings.TrimPrefix(line, "EVALUATION_SCORE:"), "%f", &score)
			}
		}
	}

	return selectedIdx, score, nil
}

// buildJudgePrompt formats all candidates into a numbered list for the judge.
func buildJudgePrompt(taskPrompt string, candidates []Candidate) string {
	var sb strings.Builder
	sb.WriteString("## Task\n\n")
	sb.WriteString(taskPrompt)
	sb.WriteString("\n\n## Candidates\n\n")
	for i, c := range candidates {
		fmt.Fprintf(&sb, "### Candidate %d\n\n%s\n\n", i+1, c.Output)
	}
	sb.WriteString("Select the best candidate. Respond with SELECTED_CANDIDATE: <number>, EVALUATION_SCORE: <0-10>, RATIONALE: <one sentence>.")
	return sb.String()
}
