// Package ollama wraps the Ollama HTTP API for inference calls.
//
// REQ-003: Every call routes to the model declared in config.ModelRouting.
// REQ-004: Every request includes keep_alive: 0 to unload the model immediately
// after the response, enforcing MAX_VRAM_MODELS = 1.
package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client calls the Ollama HTTP API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// New creates a Client targeting baseURL (e.g. "http://localhost:11434").
func New(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 10 * time.Minute, // SLMs can be slow on first token
		},
	}
}

// GenerateRequest is the payload for POST /api/generate.
// REQ-004: KeepAlive is always "0" to unload the model after each call.
type GenerateRequest struct {
	Model     string         `json:"model"`
	Prompt    string         `json:"prompt"`
	System    string         `json:"system,omitempty"`
	Stream    bool           `json:"stream"`
	KeepAlive string         `json:"keep_alive"`
	Options   map[string]any `json:"options,omitempty"`
}

// GenerateResponse is the non-streaming response from /api/generate.
type GenerateResponse struct {
	Model     string `json:"model"`
	Response  string `json:"response"`
	Done      bool   `json:"done"`
	CreatedAt string `json:"created_at"`
}

// Generate calls POST /api/generate with stream:false and keep_alive:0.
// REQ-003: system prompt is injected per task type.
// REQ-004: keep_alive is always "0".
func (c *Client) Generate(ctx context.Context, model, system, prompt string, opts map[string]any) (*GenerateResponse, error) {
	req := GenerateRequest{
		Model:     model,
		Prompt:    prompt,
		System:    system,
		Stream:    false,
		KeepAlive: "0", // REQ-004: unload immediately after response
		Options:   opts,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("ollama: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("ollama: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ollama: POST /api/generate: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("ollama: HTTP %d: %s", resp.StatusCode, string(errBody))
	}

	var result GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("ollama: decode response: %w", err)
	}
	return &result, nil
}

// HealthCheck calls GET /api/tags to verify Ollama is reachable.
func (c *Client) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/tags", nil)
	if err != nil {
		return fmt.Errorf("ollama: build health request: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("ollama: health check: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama: health check HTTP %d", resp.StatusCode)
	}
	return nil
}

// ParseOutput scans a model response for structured output markers.
// REQ-003: NEW_TASK: lines trigger subtask creation; REVIEW_REQUIRED: triggers viability gate.
func ParseOutput(response string) ParsedOutput {
	var out ParsedOutput
	out.Raw = response

	for _, line := range strings.Split(response, "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "NEW_TASK:"):
			payload := strings.TrimPrefix(line, "NEW_TASK:")
			out.NewTasks = append(out.NewTasks, strings.TrimSpace(payload))
		case strings.HasPrefix(line, "REVIEW_REQUIRED:"):
			out.ReviewRequired = true
			out.ReviewReason = strings.TrimSpace(strings.TrimPrefix(line, "REVIEW_REQUIRED:"))
		case strings.HasPrefix(line, "VIABILITY_SCORE:"):
			fmt.Sscanf(strings.TrimPrefix(line, "VIABILITY_SCORE:"), "%f", &out.ViabilityScore)
		case strings.HasPrefix(line, "SELECTED_CANDIDATE:"):
			out.SelectedCandidateID = strings.TrimSpace(strings.TrimPrefix(line, "SELECTED_CANDIDATE:"))
		}
	}
	return out
}

// ParsedOutput holds structured markers extracted from a model response.
type ParsedOutput struct {
	// Raw is the unmodified model response.
	Raw string

	// NewTasks holds payloads from NEW_TASK: lines. REQ-003.
	NewTasks []string

	// ReviewRequired is true when the output contains REVIEW_REQUIRED:. REQ-003.
	ReviewRequired bool

	// ReviewReason is the text after REVIEW_REQUIRED:.
	ReviewReason string

	// ViabilityScore is extracted from VIABILITY_SCORE: (0–10). REQ-006.
	ViabilityScore float64

	// SelectedCandidateID is the judge's selection from SELECTED_CANDIDATE:. REQ-005.
	SelectedCandidateID string
}
