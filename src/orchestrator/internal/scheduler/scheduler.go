// Package scheduler implements the DAG task scheduler. REQ-002.
//
// The scheduler polls swarm.db every 2 seconds (configurable), finds tasks
// that are unblocked, and dispatches them to the dispatcher. It enforces:
//   - Sticky model strategy (REQ-004): prefer tasks with the same model_tag
//     as the currently active task to minimise VRAM swaps.
//   - MAX_DEPTH = 5 (REQ-002): tasks at depth > max are pruned immediately.
//   - Atomic dispatch: pending → active in a single transaction (REQ-002).
//   - Graceful recovery from task panics (REQ-002).
package scheduler

import (
	"context"
	"database/sql"
	"log/slog"
	"sync"
	"time"

	"als/obs"
	"als/orchestrator/internal/db"
)

// Dispatcher is the function called by the scheduler to execute a task.
// Implementations run the task to completion and return an error on failure.
type Dispatcher func(ctx context.Context, task *db.Task) error

// Scheduler polls swarm.db and dispatches ready tasks.
type Scheduler struct {
	db           *db.DB
	obs          *obs.Logger
	dispatch     Dispatcher
	pollInterval time.Duration
	maxDepth     int

	mu           sync.Mutex
	activeModels map[string]bool // model_tag → currently dispatching
}

// New creates a Scheduler.
func New(database *db.DB, obsLog *obs.Logger, dispatch Dispatcher, pollInterval time.Duration, maxDepth int) *Scheduler {
	return &Scheduler{
		db:           database,
		obs:          obsLog,
		dispatch:     dispatch,
		pollInterval: pollInterval,
		maxDepth:     maxDepth,
		activeModels: make(map[string]bool),
	}
}

// Run starts the scheduler loop and blocks until ctx is cancelled.
// REQ-002: polls every pollInterval; recovers from panics in dispatch.
func (s *Scheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()

	slog.Info("scheduler: started", "poll_interval", s.pollInterval)

	for {
		select {
		case <-ctx.Done():
			slog.Info("scheduler: stopped")
			return
		case <-ticker.C:
			s.poll(ctx)
		}
	}
}

// poll fetches all unblocked pending tasks and dispatches eligible ones.
func (s *Scheduler) poll(ctx context.Context) {
	pending, err := s.db.PendingTasks(ctx)
	if err != nil {
		slog.Error("scheduler: poll error", "error", err)
		return
	}
	if len(pending) == 0 {
		return
	}

	// REQ-004: sticky model — sort tasks so those matching an active model come first.
	ordered := s.stickySort(pending)

	for _, task := range ordered {
		// REQ-002: enforce max depth — prune and skip.
		if task.Depth > s.maxDepth {
			slog.Warn("scheduler: task exceeds max depth, pruning",
				"task_id", task.ID, "depth", task.Depth, "max", s.maxDepth)
			s.db.UpdateTaskStatus(ctx, task.ID, "pruned") //nolint:errcheck
			s.obs.WriteEvent(obs.Event{
				Type:       obs.EventCascadingPrune,
				TaskID:     &task.ID,
				TaskType:   &task.Type,
				Depth:      &task.Depth,
				Detail:     map[string]any{"reason": "max_depth_exceeded", "depth": task.Depth},
			})
			continue
		}

		// REQ-004: only one task per model at a time.
		s.mu.Lock()
		if s.activeModels[task.ModelTag] {
			s.mu.Unlock()
			continue
		}
		s.activeModels[task.ModelTag] = true
		s.mu.Unlock()

		// REQ-002: atomic dispatch — transition pending → active before Ollama call.
		if err := s.db.WithTx(ctx, func(tx *sql.Tx) error {
			return db.UpdateTaskStatusTx(ctx, tx, task.ID, "active")
		}); err != nil {
			slog.Error("scheduler: failed to activate task", "task_id", task.ID, "error", err)
			s.mu.Lock()
			delete(s.activeModels, task.ModelTag)
			s.mu.Unlock()
			continue
		}

		go s.runTask(ctx, task)
	}
}

// runTask executes a task in a goroutine, recovers from panics, and releases
// the model slot on completion. REQ-002: scheduler survives panics.
func (s *Scheduler) runTask(ctx context.Context, task *db.Task) {
	defer func() {
		s.mu.Lock()
		delete(s.activeModels, task.ModelTag)
		s.mu.Unlock()

		if r := recover(); r != nil {
			slog.Error("scheduler: task panicked", "task_id", task.ID, "panic", r)
			s.db.UpdateTaskStatus(ctx, task.ID, "failed") //nolint:errcheck
		}
	}()

	if err := s.dispatch(ctx, task); err != nil {
		slog.Error("scheduler: task dispatch error", "task_id", task.ID, "error", err)
	}
}

// stickySort reorders tasks so those matching an already-active model_tag come
// first. REQ-004: minimise VRAM swaps by prioritising the current model.
func (s *Scheduler) stickySort(tasks []*db.Task) []*db.Task {
	s.mu.Lock()
	active := make(map[string]bool, len(s.activeModels))
	for k, v := range s.activeModels {
		active[k] = v
	}
	s.mu.Unlock()

	if len(active) == 0 {
		return tasks
	}

	sticky := make([]*db.Task, 0, len(tasks))
	other := make([]*db.Task, 0, len(tasks))
	for _, t := range tasks {
		if active[t.ModelTag] {
			sticky = append(sticky, t)
		} else {
			other = append(other, t)
		}
	}
	return append(sticky, other...)
}
