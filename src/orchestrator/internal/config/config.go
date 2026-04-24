// Package config loads and validates the swarm configuration from config.yaml.
// All tuneable constants (model routing, thresholds, MCP servers) live here so
// that no swarm behaviour is hardcoded — REQ-003: model routing must be
// config-driven to allow future remapping without code changes.
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config is the top-level swarm configuration loaded from config.yaml.
type Config struct {
	// DBPath is the path to swarm.db (SQLite — app state).
	DBPath string `yaml:"db_path"`

	// ObsDBPath is the path to obs.db (DuckDB — event log).
	ObsDBPath string `yaml:"obs_db_path"`

	// OllamaBaseURL is the Ollama HTTP API base (default: http://localhost:11434).
	OllamaBaseURL string `yaml:"ollama_base_url"`

	// ContextThreshold is the fraction of a model's context window that triggers
	// a SUMMARY task (default: 0.70). REQ-008.
	ContextThreshold float64 `yaml:"context_threshold"`

	// ViabilityThreshold is the minimum score (0–10) a VIABILITY_REVIEW must
	// return to avoid a cascading prune (default: 5). REQ-006.
	ViabilityThreshold float64 `yaml:"viability_threshold"`

	// VariationCount is the number of candidates generated for variational
	// execution (default: 3). REQ-005.
	VariationCount int `yaml:"variation_count"`

	// MaxDepth is the maximum allowed task depth; tasks beyond this are pruned
	// immediately (default: 5). REQ-002.
	MaxDepth int `yaml:"max_depth"`

	// MaxSelfCorrections is the maximum number of self-correction attempts per
	// GO_CODE branch before the chain is marked failed (default: 5). REQ-007.
	MaxSelfCorrections int `yaml:"max_self_corrections"`

	// SchedulerPollInterval is the DAG scheduler poll interval in milliseconds
	// (default: 2000). REQ-002.
	SchedulerPollIntervalMS int `yaml:"scheduler_poll_interval_ms"`

	// ModelRouting maps task types to Ollama model tags. REQ-003.
	ModelRouting map[string]ModelConfig `yaml:"model_routing"`

	// MCPServers is the ordered list of MCP servers to register at startup. REQ-009.
	MCPServers []MCPServerConfig `yaml:"mcp_servers"`

	// SkillsDir is the path to the skills/ directory containing SKILL.md files.
	SkillsDir string `yaml:"skills_dir"`

	// OutputDir is the default project output directory (overridden per-project).
	OutputDir string `yaml:"output_dir"`
}

// ModelConfig describes the Ollama model and approximate context window for a
// task type.
type ModelConfig struct {
	// Tag is the Ollama model tag (e.g. "phi-4:mini").
	Tag string `yaml:"tag"`

	// ContextWindow is the approximate token limit for this model. Used by the
	// context compression check (REQ-008).
	ContextWindow int `yaml:"context_window"`
}

// MCPServerConfig describes a single MCP server that the orchestrator registers
// at startup. REQ-009.
type MCPServerConfig struct {
	// Name is the human-readable server identifier (e.g. "brave-search").
	Name string `yaml:"name"`

	// Command is the executable to spawn (e.g. "npx", "node").
	Command string `yaml:"command"`

	// Args are the command-line arguments passed to Command.
	Args []string `yaml:"args"`

	// Env is a map of environment variable names to their values.
	// Values may reference OS env vars using ${VAR} syntax; Load resolves them.
	Env map[string]string `yaml:"env"`

	// TaskTypes lists the task types that may invoke this server's tools.
	// An empty list means the server is available to all task types.
	TaskTypes []string `yaml:"task_types"`
}

// Defaults returns a Config with safe default values applied.
func Defaults() Config {
	return Config{
		DBPath:                  "swarm.db",
		ObsDBPath:               "obs.db",
		OllamaBaseURL:           "http://localhost:11434",
		ContextThreshold:        0.70,
		ViabilityThreshold:      5.0,
		VariationCount:          3,
		MaxDepth:                5,
		MaxSelfCorrections:      5,
		SchedulerPollIntervalMS: 2000,
		SkillsDir:               "skills",
		OutputDir:               "output",
		ModelRouting: map[string]ModelConfig{
			"DECOMPOSE":        {Tag: "phi-4:mini", ContextWindow: 4096},
			"RESEARCH":         {Tag: "llama3.1:8b", ContextWindow: 8192},
			"GO_CODE":          {Tag: "qwen2.5-coder:7b", ContextWindow: 8192},
			"VIABILITY_REVIEW": {Tag: "phi-4:mini", ContextWindow: 4096},
			"SUMMARY":          {Tag: "smollm3:135m", ContextWindow: 2048},
		},
	}
}

// Load reads config.yaml from path, overlays it onto Defaults(), resolves
// environment variable references in MCPServer.Env values, and validates the
// result.
func Load(path string) (*Config, error) {
	cfg := Defaults()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read %q: %w", path, err)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config: parse %q: %w", path, err)
	}

	resolveEnvRefs(&cfg)

	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("config: invalid: %w", err)
	}

	return &cfg, nil
}

// resolveEnvRefs expands ${VAR} patterns in MCPServer.Env values using the
// current OS environment. Unknown variables are left as empty strings.
func resolveEnvRefs(cfg *Config) {
	for i := range cfg.MCPServers {
		for k, v := range cfg.MCPServers[i].Env {
			cfg.MCPServers[i].Env[k] = os.ExpandEnv(v)
		}
	}
}

// validate checks that required config fields are in range.
func validate(cfg *Config) error {
	if cfg.ContextThreshold <= 0 || cfg.ContextThreshold >= 1 {
		return fmt.Errorf("context_threshold must be between 0 and 1 exclusive, got %f", cfg.ContextThreshold)
	}
	if cfg.ViabilityThreshold < 0 || cfg.ViabilityThreshold > 10 {
		return fmt.Errorf("viability_threshold must be 0–10, got %f", cfg.ViabilityThreshold)
	}
	if cfg.VariationCount < 1 {
		return fmt.Errorf("variation_count must be >= 1, got %d", cfg.VariationCount)
	}
	if cfg.MaxDepth < 1 || cfg.MaxDepth > 10 {
		return fmt.Errorf("max_depth must be 1–10, got %d", cfg.MaxDepth)
	}
	if cfg.MaxSelfCorrections < 1 {
		return fmt.Errorf("max_self_corrections must be >= 1, got %d", cfg.MaxSelfCorrections)
	}
	for taskType, mc := range cfg.ModelRouting {
		if mc.Tag == "" {
			return fmt.Errorf("model_routing[%s].tag must not be empty", taskType)
		}
	}
	return nil
}
