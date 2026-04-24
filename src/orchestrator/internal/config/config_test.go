package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"als/orchestrator/internal/config"
)

// req-003: Model routing must be config-driven without recompilation.
func TestLoad_req003_ModelRoutingFromFile(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "config.yaml")
	os.WriteFile(f, []byte(`
model_routing:
  DECOMPOSE:
    tag: "phi-4:mini"
    context_window: 4096
  RESEARCH:
    tag: "llama3.1:8b"
    context_window: 8192
  GO_CODE:
    tag: "qwen2.5-coder:7b"
    context_window: 8192
  VIABILITY_REVIEW:
    tag: "phi-4:mini"
    context_window: 4096
  SUMMARY:
    tag: "smollm3:135m"
    context_window: 2048
`), 0644)

	cfg, err := config.Load(f)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	mc, ok := cfg.ModelRouting["GO_CODE"]
	if !ok {
		t.Fatal("missing GO_CODE routing entry")
	}
	if mc.Tag != "qwen2.5-coder:7b" {
		t.Errorf("want qwen2.5-coder:7b, got %s", mc.Tag)
	}
}

// req-003: Missing file returns a wrapped error.
func TestLoad_MissingFile(t *testing.T) {
	_, err := config.Load("/nonexistent/config.yaml")
	if err == nil {
		t.Fatal("want error for missing file, got nil")
	}
}

// Defaults provides safe values for all numeric thresholds.
func TestDefaults_NumericBounds(t *testing.T) {
	d := config.Defaults()
	if d.ContextThreshold != 0.70 {
		t.Errorf("want ContextThreshold 0.70, got %f", d.ContextThreshold)
	}
	if d.ViabilityThreshold != 5.0 {
		t.Errorf("want ViabilityThreshold 5.0, got %f", d.ViabilityThreshold)
	}
	if d.VariationCount != 3 {
		t.Errorf("want VariationCount 3, got %d", d.VariationCount)
	}
	if d.MaxDepth != 5 {
		t.Errorf("want MaxDepth 5, got %d", d.MaxDepth)
	}
}

// req-009: MCP server env vars are resolved from OS environment.
func TestLoad_req009_EnvVarResolution(t *testing.T) {
	t.Setenv("BRAVE_API_KEY", "test-brave-key")

	dir := t.TempDir()
	f := filepath.Join(dir, "config.yaml")
	os.WriteFile(f, []byte(`
mcp_servers:
  - name: "brave-search"
    command: "npx"
    args: ["-y", "@brave/brave-search-mcp"]
    env:
      BRAVE_API_KEY: "${BRAVE_API_KEY}"
`), 0644)

	cfg, err := config.Load(f)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(cfg.MCPServers) == 0 {
		t.Fatal("want at least 1 MCP server")
	}
	got := cfg.MCPServers[0].Env["BRAVE_API_KEY"]
	if got != "test-brave-key" {
		t.Errorf("want resolved env value 'test-brave-key', got %q", got)
	}
}

// validate rejects out-of-range context_threshold.
func TestLoad_InvalidContextThreshold(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "config.yaml")
	os.WriteFile(f, []byte(`context_threshold: 1.5`), 0644)
	_, err := config.Load(f)
	if err == nil {
		t.Fatal("want validation error for context_threshold > 1")
	}
}
