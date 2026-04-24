// Command swarm is the entry point for the Autonomous Local Swarm orchestrator.
//
// Usage:
//
//	swarm [-config config.yaml]
//
// The orchestrator:
//  1. Loads config.yaml (or the path given by -config flag)
//  2. Opens swarm.db (SQLite) and obs.db (DuckDB)
//  3. Registers all configured MCP servers
//  4. Loads Agent Skills from the skills/ directory
//  5. Starts the DAG scheduler loop
//  6. Blocks until SIGTERM / SIGINT
package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"als/obs"
	"als/orchestrator/internal/config"
	"als/orchestrator/internal/db"
	"als/orchestrator/internal/mcp"
	"als/orchestrator/internal/ollama"
	"als/orchestrator/internal/router"
	"als/orchestrator/internal/scheduler"
	"als/orchestrator/internal/skills"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to config.yaml")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		slog.Error("failed to load config", "path", *configPath, "error", err)
		os.Exit(1)
	}

	// ── Observability ──────────────────────────────────────────────────────
	obsLog, err := obs.Open(cfg.ObsDBPath)
	if err != nil {
		slog.Error("failed to open obs.db", "path", cfg.ObsDBPath, "error", err)
		os.Exit(1)
	}
	defer obsLog.Close()

	obsLog.WriteEvent(obs.Event{Type: obs.EventOrchestratorStart})

	// ── Application state DB ───────────────────────────────────────────────
	swarmDB, err := db.Open(cfg.DBPath)
	if err != nil {
		slog.Error("failed to open swarm.db", "path", cfg.DBPath, "error", err)
		os.Exit(1)
	}
	defer swarmDB.Close()

	// ── Ollama health check ────────────────────────────────────────────────
	ollamaClient := ollama.New(cfg.OllamaBaseURL)
	checkCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := ollamaClient.HealthCheck(checkCtx); err != nil {
		slog.Warn("ollama health check failed — inference will fail at dispatch", "error", err)
	}

	// ── MCP servers ────────────────────────────────────────────────────────
	mcpClient := mcp.New()
	regCtx, regCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer regCancel()

	for _, srv := range cfg.MCPServers {
		if err := mcpClient.Register(regCtx, mcp.ServerConfig{
			Name:      srv.Name,
			Command:   srv.Command,
			Args:      srv.Args,
			Env:       srv.Env,
			TaskTypes: srv.TaskTypes,
		}); err != nil {
			// REQ-009: non-fatal; log and continue with degraded tool availability.
			slog.Warn("mcp: server registration failed", "name", srv.Name, "error", err)
		}
	}
	defer mcpClient.Close()
	slog.Info("mcp: registered servers", "count", len(mcpClient.HealthyServers()))

	// ── Skills ─────────────────────────────────────────────────────────────
	skillLoader := skills.NewLoader(cfg.SkillsDir)
	loaded, errs := skillLoader.LoadAll()
	for _, e := range errs {
		slog.Warn("skills: load error", "error", e)
	}
	slog.Info("skills: loaded", "count", loaded)

	// ── Router + Scheduler ─────────────────────────────────────────────────
	r := router.New(cfg, swarmDB, obsLog, ollamaClient, mcpClient, skillLoader)

	pollInterval := time.Duration(cfg.SchedulerPollIntervalMS) * time.Millisecond
	s := scheduler.New(swarmDB, obsLog, r.Dispatch, pollInterval, cfg.MaxDepth)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	slog.Info("swarm: starting", "config", *configPath, "db", cfg.DBPath, "obs_db", cfg.ObsDBPath)
	s.Run(ctx) // blocks until signal

	obsLog.WriteEvent(obs.Event{Type: obs.EventOrchestratorStop})
	slog.Info("swarm: stopped")
}
