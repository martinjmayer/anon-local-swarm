// Package router implements the semantic router and task dispatcher.
//
// REQ-003: Routes each task type to its configured Ollama model + skill.
// REQ-005: Runs variational execution for eligible tasks.
// REQ-006: Handles VIABILITY_REVIEW scores and triggers cascading prune.
// REQ-007: Validates GO_CODE output via go build; emits self-correction tasks.
// REQ-008: Checks context threshold; emits SUMMARY tasks when exceeded.
// REQ-009: Injects available MCP tool context into RESEARCH and GO_CODE prompts.
// REQ-010: Deposits artifacts on terminal state.
// REQ-011: Emits obs events on every state transition.
package router

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"als/obs"
	"als/orchestrator/internal/config"
	"als/orchestrator/internal/db"
	"als/orchestrator/internal/deposit"
	"als/orchestrator/internal/mcp"
	"als/orchestrator/internal/ollama"
	"als/orchestrator/internal/sidecar"
	"als/orchestrator/internal/skills"
	"als/orchestrator/internal/validator"
	"als/orchestrator/internal/variational"

	"github.com/google/uuid"
)

// Router dispatches tasks to Ollama and handles post-dispatch logic.
type Router struct {
	cfg    *config.Config
	db     *db.DB
	obs    *obs.Logger
	ollama *ollama.Client
	mcp    *mcp.Client
	skills *skills.Loader
}

// New creates a Router wired to all subsystems.
func New(cfg *config.Config, database *db.DB, obsLog *obs.Logger, ollamaClient *ollama.Client, mcpClient *mcp.Client, skillLoader *skills.Loader) *Router {
	return &Router{
		cfg:    cfg,
		db:     database,
		obs:    obsLog,
		ollama: ollamaClient,
		mcp:    mcpClient,
		skills: skillLoader,
	}
}

// Dispatch executes a single task through the full pipeline and updates its
// status in swarm.db. This is called by the scheduler for each ready task.
func (r *Router) Dispatch(ctx context.Context, task *db.Task) error {
	// Emit task_transition: pending → active is already done by scheduler.
	r.emitTransition(task, "active")

	mc, ok := r.cfg.ModelRouting[task.Type]
	if !ok {
		r.markFailed(ctx, task, "no model routing for task type: "+task.Type)
		return fmt.Errorf("router: no model config for %s", task.Type)
	}

	// REQ-008: Check context threshold before dispatching.
	if err := r.maybeEmitSummary(ctx, task, mc); err != nil {
		slog.Warn("router: context summary error", "task_id", task.ID, "error", err)
	}

	// Build the system prompt from the Agent Skill.
	systemPrompt := r.systemPrompt(task.Type)

	// Build the user prompt, injecting MCP tool context for RESEARCH / GO_CODE.
	userPrompt := r.buildPrompt(ctx, task)

	// Dispatch: variational for RESEARCH and GO_CODE, direct for others.
	output, err := r.generate(ctx, task, mc, systemPrompt, userPrompt)
	if err != nil {
		r.markFailed(ctx, task, err.Error())
		return err
	}

	// Handle output markers.
	parsed := ollama.ParseOutput(output)

	switch task.Type {
	case "GO_CODE":
		if err := r.handleGOCode(ctx, task, parsed, mc); err != nil {
			return err
		}
	case "VIABILITY_REVIEW":
		if err := r.handleViabilityReview(ctx, task, parsed); err != nil {
			return err
		}
	case "DECOMPOSE", "RESEARCH":
		if err := r.handleSubtaskEmission(ctx, task, parsed); err != nil {
			return err
		}
		r.appendSidecar(ctx, task, output)
		r.markCompleted(ctx, task, output)
	case "SUMMARY":
		if err := r.handleSummary(ctx, task, output); err != nil {
			return err
		}
	default:
		r.appendSidecar(ctx, task, output)
		r.markCompleted(ctx, task, output)
	}

	return nil
}

// generate runs the Ollama inference — variational for RESEARCH/GO_CODE,
// direct for all others. REQ-005.
func (r *Router) generate(ctx context.Context, task *db.Task, mc config.ModelConfig, system, prompt string) (string, error) {
	useVariational := (task.Type == "RESEARCH" || task.Type == "GO_CODE") && r.cfg.VariationCount > 1

	if !useVariational {
		resp, err := r.ollama.Generate(ctx, mc.Tag, system, prompt, nil)
		if err != nil {
			return "", fmt.Errorf("router: ollama generate: %w", err)
		}
		return resp.Response, nil
	}

	// Variational execution.
	varCfg := variational.Config{
		VariationCount: r.cfg.VariationCount,
		JudgeModel:     r.cfg.ModelRouting["VIABILITY_REVIEW"].Tag,
	}
	gen := func(ctx context.Context, model, sys, pmt string, opts map[string]any) (*ollama.GenerateResponse, error) {
		return r.ollama.Generate(ctx, model, sys, pmt, opts)
	}

	candidates, err := variational.Run(ctx, varCfg, gen, mc.Tag, system, prompt)
	if err != nil {
		return "", fmt.Errorf("router: variational run: %w", err)
	}

	// Persist candidates.
	var dbCandidates []*db.Candidate
	for _, c := range candidates {
		id := uuid.NewString()
		dbC := &db.Candidate{ID: id, TaskID: task.ID, OutputText: c.Output}
		if err := r.db.InsertCandidate(ctx, dbC); err != nil {
			slog.Warn("router: insert candidate failed", "error", err)
		}
		dbCandidates = append(dbCandidates, dbC)
	}

	// Judge selects best candidate.
	selectedIdx, score, err := variational.Judge(ctx, varCfg, gen, prompt, candidates)
	if err != nil {
		slog.Warn("router: judge failed, using candidate 0", "error", err)
		selectedIdx = 0
	}
	if selectedIdx < len(dbCandidates) {
		r.db.SelectCandidate(ctx, dbCandidates[selectedIdx].ID, task.ID, score) //nolint:errcheck
	}

	return candidates[selectedIdx].Output, nil
}

// handleGOCode runs go build validation and self-correction. REQ-007.
func (r *Router) handleGOCode(ctx context.Context, task *db.Task, parsed ollama.ParsedOutput, mc config.ModelConfig) error {
	meta := ""
	if task.MetaData != nil {
		meta = *task.MetaData
	}
	retryCount := validator.RetryCount(meta)

	result, err := validator.Validate(ctx, parsed.Raw)
	if err != nil {
		// Validation system error (not a compile error) — mark failed.
		r.markFailed(ctx, task, "validator error: "+err.Error())
		return err
	}

	r.obs.WriteEvent(obs.Event{
		Type:     obs.EventSelfCorrection,
		TaskID:   &task.ID,
		TaskType: &task.Type,
		ModelTag: &mc.Tag,
		Depth:    &task.Depth,
		Detail: map[string]any{
			"attempt":  retryCount + 1,
			"success":  result.Success,
			"stderr":   result.Stderr,
		},
	})

	if result.Success {
		r.appendSidecar(ctx, task, parsed.Raw)
		r.markCompleted(ctx, task, parsed.Raw)
		// Deposit generated code.
		if task.ProjectID != "" {
			proj, err := r.db.GetProject(ctx, task.ProjectID)
			if err == nil && proj.OutputDir != nil {
				deposit.Write(*proj.OutputDir, []deposit.Artifact{
					{Filename: "main.go", Content: parsed.Raw},
				}) //nolint:errcheck
			}
		}
		return nil
	}

	// Build failed — self-correction loop.
	if retryCount >= r.cfg.MaxSelfCorrections {
		r.obs.WriteEvent(obs.Event{
			Type:   obs.EventSelfCorrection,
			TaskID: &task.ID,
			Detail: map[string]any{"halt": true, "attempts": retryCount, "stderr": result.Stderr},
		})
		r.markFailed(ctx, task, fmt.Sprintf("max self-corrections (%d) reached", r.cfg.MaxSelfCorrections))
		return nil
	}

	// Emit a new GO_CODE task with the error payload.
	newMeta, _ := validator.IncrementRetryMeta(meta)
	newDepth := task.Depth + 1
	correctionPayload := validator.SelfCorrectionPayload(parsed.Raw, result.Stderr)
	newTask := &db.Task{
		ID:        uuid.NewString(),
		ProjectID: task.ProjectID,
		ParentID:  &task.ID,
		Type:      "GO_CODE",
		ModelTag:  mc.Tag,
		Payload:   correctionPayload,
		Status:    "pending",
		Depth:     newDepth,
		MetaData:  &newMeta,
	}
	if err := r.db.InsertTask(ctx, newTask); err != nil {
		return fmt.Errorf("router: insert self-correction task: %w", err)
	}
	r.markCompleted(ctx, task, parsed.Raw) // original task completes; correction continues
	return nil
}

// handleViabilityReview processes the VIABILITY_SCORE and triggers cascading
// prune if score < viability_threshold. REQ-006.
func (r *Router) handleViabilityReview(ctx context.Context, task *db.Task, parsed ollama.ParsedOutput) error {
	score := parsed.ViabilityScore
	r.obs.WriteEvent(obs.Event{
		Type:   obs.EventTaskTransition,
		TaskID: &task.ID,
		Detail: map[string]any{"viability_score": score},
	})

	if score < r.cfg.ViabilityThreshold {
		// Cascading prune of the entire branch.
		var rootID string
		if task.ParentID != nil {
			rootID = *task.ParentID
		} else {
			rootID = task.ID
		}
		pruned, err := r.db.PruneTasksBranch(ctx, rootID)
		if err != nil {
			return fmt.Errorf("router: prune branch: %w", err)
		}
		r.obs.WriteEvent(obs.Event{
			Type:   obs.EventCascadingPrune,
			TaskID: &task.ID,
			Depth:  &task.Depth,
			Detail: map[string]any{"score": score, "pruned_count": pruned, "threshold": r.cfg.ViabilityThreshold},
		})
		slog.Info("router: cascading prune triggered",
			"task_id", task.ID, "score", score, "pruned_count", pruned)

		// Check if all project tasks are pruned — deposit prune report.
		r.maybeDepositPruneReport(ctx, task.ProjectID, score, pruned)
		return nil
	}

	// Score passes — mark completed, unblock dependents.
	r.markCompleted(ctx, task, parsed.Raw)
	return nil
}

// handleSubtaskEmission creates NEW_TASK subtasks from parsed output. REQ-003.
func (r *Router) handleSubtaskEmission(ctx context.Context, task *db.Task, parsed ollama.ParsedOutput) error {
	for _, payload := range parsed.NewTasks {
		taskType, modelTag := inferTaskType(payload, r.cfg)
		newDepth := task.Depth
		if task.Type != "DECOMPOSE" && task.Type != "SUMMARY" {
			newDepth = task.Depth + 1
		}
		if newDepth > r.cfg.MaxDepth {
			slog.Warn("router: subtask would exceed max depth, skipping",
				"parent_id", task.ID, "depth", newDepth)
			continue
		}
		newTask := &db.Task{
			ID:        uuid.NewString(),
			ProjectID: task.ProjectID,
			ParentID:  &task.ID,
			BlockedBy: &task.ID,
			Type:      taskType,
			ModelTag:  modelTag,
			Payload:   payload,
			Status:    "blocked",
			Depth:     newDepth,
		}
		if err := r.db.InsertTask(ctx, newTask); err != nil {
			slog.Warn("router: insert subtask failed", "error", err)
		}
	}
	return nil
}

// handleSummary replaces the branch context sidecar with the compressed output.
// REQ-008.
func (r *Router) handleSummary(ctx context.Context, task *db.Task, output string) error {
	if task.ContextPath != nil {
		ctxData, err := sidecar.Load(*task.ContextPath, task.ID)
		if err == nil {
			sidecar.Replace(*task.ContextPath, ctxData, output) //nolint:errcheck
		}
	}
	r.markCompleted(ctx, task, output)
	return nil
}

// maybeEmitSummary checks whether the branch context exceeds the threshold and
// emits a SUMMARY task if so. REQ-008.
func (r *Router) maybeEmitSummary(ctx context.Context, task *db.Task, mc config.ModelConfig) error {
	if task.ContextPath == nil {
		return nil
	}
	ctxData, err := sidecar.Load(*task.ContextPath, task.ID)
	if err != nil {
		return err
	}
	if !sidecar.ExceedsThreshold(ctxData, mc.ContextWindow, r.cfg.ContextThreshold) {
		return nil
	}

	summaryMC := r.cfg.ModelRouting["SUMMARY"]
	ctxJSON, _ := json.Marshal(ctxData)
	summaryTask := &db.Task{
		ID:          uuid.NewString(),
		ProjectID:   task.ProjectID,
		ParentID:    &task.ID,
		BlockedBy:   &task.ID,
		Type:        "SUMMARY",
		ModelTag:    summaryMC.Tag,
		Payload:     "Compress the following branch context into a concise summary:\n\n" + string(ctxJSON),
		Status:      "pending",
		Depth:       task.Depth, // SUMMARY does not increment depth
		ContextPath: task.ContextPath,
	}
	r.obs.WriteEvent(obs.Event{
		Type:   obs.EventContextCompression,
		TaskID: &task.ID,
		Depth:  &task.Depth,
		Detail: map[string]any{"token_estimate": ctxData.TokenEstimate, "context_window": mc.ContextWindow},
	})
	return r.db.InsertTask(ctx, summaryTask)
}

// buildPrompt constructs the user-facing prompt, injecting MCP tool descriptions
// for RESEARCH and GO_CODE tasks. REQ-009.
func (r *Router) buildPrompt(ctx context.Context, task *db.Task) string {
	if task.Type != "RESEARCH" && task.Type != "GO_CODE" {
		return task.Payload
	}
	servers := r.mcp.HealthyServers()
	if len(servers) == 0 {
		return task.Payload
	}
	var sb strings.Builder
	sb.WriteString(task.Payload)
	sb.WriteString("\n\n## Available MCP Tools\n\nYou have access to the following tool servers: ")
	sb.WriteString(strings.Join(servers, ", "))
	sb.WriteString(".\nCall tools using the MCP protocol to gather information.")
	return sb.String()
}

// appendSidecar writes output to the branch context sidecar.
func (r *Router) appendSidecar(ctx context.Context, task *db.Task, output string) {
	if task.ContextPath == nil {
		return
	}
	ctxData, err := sidecar.Load(*task.ContextPath, task.ID)
	if err != nil {
		slog.Warn("router: load sidecar failed", "path", *task.ContextPath, "error", err)
		return
	}
	sidecar.Append(*task.ContextPath, ctxData, output) //nolint:errcheck
}

// systemPrompt returns the SKILL.md body for a task type, or a default.
func (r *Router) systemPrompt(taskType string) string {
	skillName := taskTypeToSkillName(taskType)
	if s := r.skills.Get(skillName); s != nil {
		return s.Body
	}
	return "You are an autonomous agent. Complete the task described in the prompt."
}

// taskTypeToSkillName maps task type strings to skill names.
func taskTypeToSkillName(taskType string) string {
	switch taskType {
	case "DECOMPOSE":
		return "als-decompose"
	case "RESEARCH":
		return "als-research"
	case "GO_CODE":
		return "als-go-code"
	case "VIABILITY_REVIEW":
		return "als-viability-review"
	case "SUMMARY":
		return "als-summary"
	default:
		return ""
	}
}

// inferTaskType guesses the task type and model tag from a NEW_TASK payload.
func inferTaskType(payload string, cfg *config.Config) (string, string) {
	lower := strings.ToLower(payload)
	taskType := "RESEARCH"
	switch {
	case strings.Contains(lower, "implement") ||
		strings.Contains(lower, "write code") ||
		strings.Contains(lower, "build") ||
		strings.Contains(lower, "go function"):
		taskType = "GO_CODE"
	case strings.Contains(lower, "review") || strings.Contains(lower, "viability"):
		taskType = "VIABILITY_REVIEW"
	case strings.Contains(lower, "decompose") || strings.Contains(lower, "break down"):
		taskType = "DECOMPOSE"
	}
	mc := cfg.ModelRouting[taskType]
	return taskType, mc.Tag
}

// markCompleted transitions the task to completed and emits obs event.
func (r *Router) markCompleted(ctx context.Context, task *db.Task, output string) {
	r.db.UpdateTaskStatus(ctx, task.ID, "completed") //nolint:errcheck
	// Unblock dependent tasks by setting their status to pending.
	r.unblockDependents(ctx, task.ID)
	r.emitTransition(task, "completed")
	// Update metadata with output summary.
	meta := map[string]any{"output_preview": truncate(output, 200)}
	if b, err := json.Marshal(meta); err == nil {
		r.db.UpdateTaskMetaData(ctx, task.ID, string(b)) //nolint:errcheck
	}
}

// markFailed transitions to failed and emits obs event.
func (r *Router) markFailed(ctx context.Context, task *db.Task, reason string) {
	r.db.UpdateTaskStatus(ctx, task.ID, "failed") //nolint:errcheck
	r.emitTransition(task, "failed")
	slog.Error("router: task failed", "task_id", task.ID, "reason", reason)
}

// unblockDependents finds tasks blocked by completedID and sets them to pending.
// It looks up the project ID from the completed task to scope the query.
func (r *Router) unblockDependents(ctx context.Context, completedID string) {
	completedTask, err := r.db.GetTask(ctx, completedID)
	if err != nil {
		return
	}
	tasks, err := r.db.TasksByProject(ctx, completedTask.ProjectID)
	if err != nil {
		return
	}
	for _, t := range tasks {
		if t.BlockedBy != nil && *t.BlockedBy == completedID && t.Status == "blocked" {
			r.db.UpdateTaskStatus(ctx, t.ID, "pending") //nolint:errcheck
		}
	}
}

// maybeDepositPruneReport checks if all project tasks are pruned and writes report.
func (r *Router) maybeDepositPruneReport(ctx context.Context, projectID string, score float64, pruned int) {
	proj, err := r.db.GetProject(ctx, projectID)
	if err != nil || proj.OutputDir == nil {
		return
	}
	events := []deposit.PruneEvent{{
		BranchTaskID:   projectID,
		ViabilityScore: score,
		PrunedCount:    pruned,
		Reason:         "Viability score below threshold",
	}}
	report := deposit.PruneReport(proj.Name, events)
	deposit.Write(*proj.OutputDir, []deposit.Artifact{{Filename: "prune-report.md", Content: report}}) //nolint:errcheck
	r.obs.WriteEvent(obs.Event{
		Type:   obs.EventOutputDeposit,
		Detail: map[string]any{"file": "prune-report.md", "dir": *proj.OutputDir},
	})
}

// emitTransition writes a task_transition obs event.
func (r *Router) emitTransition(task *db.Task, toStatus string) {
	r.obs.WriteEvent(obs.Event{
		Type:     obs.EventTaskTransition,
		TaskID:   &task.ID,
		TaskType: &task.Type,
		ToStatus: &toStatus,
		ModelTag: &task.ModelTag,
		Depth:    &task.Depth,
	})
}

// truncate shortens s to at most n runes.
func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "…"
}
