package ollama_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"als/orchestrator/internal/ollama"
)

// req-004: Every request must include keep_alive: 0.
func TestGenerate_req004_KeepAliveZero(t *testing.T) {
	var captured map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&captured)
		json.NewEncoder(w).Encode(map[string]any{
			"model": "phi-4:mini", "response": "ok", "done": true,
		})
	}))
	defer srv.Close()

	c := ollama.New(srv.URL)
	_, err := c.Generate(t.Context(), "phi-4:mini", "system", "prompt", nil)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if captured["keep_alive"] != "0" {
		t.Errorf("req-004: want keep_alive '0', got %v", captured["keep_alive"])
	}
}

// req-004: stream must be false (non-streaming).
func TestGenerate_StreamFalse(t *testing.T) {
	var captured map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&captured)
		json.NewEncoder(w).Encode(map[string]any{"response": "ok", "done": true})
	}))
	defer srv.Close()

	ollama.New(srv.URL).Generate(t.Context(), "phi-4:mini", "", "hello", nil)

	if v, _ := captured["stream"].(bool); v {
		t.Errorf("want stream: false, got true")
	}
}

// req-003: System prompt is included in the request.
func TestGenerate_req003_SystemPromptIncluded(t *testing.T) {
	var captured map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&captured)
		json.NewEncoder(w).Encode(map[string]any{"response": "done", "done": true})
	}))
	defer srv.Close()

	ollama.New(srv.URL).Generate(t.Context(), "phi-4:mini", "You are a decomposer", "decompose this", nil)

	if captured["system"] != "You are a decomposer" {
		t.Errorf("req-003: system prompt not in request, got %v", captured["system"])
	}
}

// req-003: ParseOutput extracts NEW_TASK lines.
func TestParseOutput_req003_NewTasks(t *testing.T) {
	raw := `
I'll break this down.
NEW_TASK: Research competitor pricing
NEW_TASK: Identify target market segments
This completes the decomposition.`

	out := ollama.ParseOutput(raw)
	if len(out.NewTasks) != 2 {
		t.Fatalf("req-003: want 2 new tasks, got %d", len(out.NewTasks))
	}
	if out.NewTasks[0] != "Research competitor pricing" {
		t.Errorf("req-003: want 'Research competitor pricing', got %q", out.NewTasks[0])
	}
}

// req-003: ParseOutput extracts REVIEW_REQUIRED marker.
func TestParseOutput_req003_ReviewRequired(t *testing.T) {
	raw := `Analysis complete.
REVIEW_REQUIRED: Market is too crowded
Continue after review.`

	out := ollama.ParseOutput(raw)
	if !out.ReviewRequired {
		t.Error("req-003: want ReviewRequired true")
	}
	if out.ReviewReason != "Market is too crowded" {
		t.Errorf("req-003: wrong reason: %q", out.ReviewReason)
	}
}

// req-006: ParseOutput extracts VIABILITY_SCORE.
func TestParseOutput_req006_ViabilityScore(t *testing.T) {
	raw := `Review complete.
VIABILITY_SCORE: 7
This opportunity scores well.`

	out := ollama.ParseOutput(raw)
	if out.ViabilityScore != 7.0 {
		t.Errorf("req-006: want score 7.0, got %f", out.ViabilityScore)
	}
}

// req-005: ParseOutput extracts SELECTED_CANDIDATE.
func TestParseOutput_req005_SelectedCandidate(t *testing.T) {
	raw := `After reviewing all candidates:
SELECTED_CANDIDATE: candidate-uuid-002
This candidate has the best structure.`

	out := ollama.ParseOutput(raw)
	if out.SelectedCandidateID != "candidate-uuid-002" {
		t.Errorf("req-005: want candidate-uuid-002, got %q", out.SelectedCandidateID)
	}
}

// HTTP error is propagated as an error.
func TestGenerate_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "model not found", http.StatusNotFound)
	}))
	defer srv.Close()

	_, err := ollama.New(srv.URL).Generate(t.Context(), "bad-model", "", "hi", nil)
	if err == nil {
		t.Fatal("want error for HTTP 404, got nil")
	}
}
