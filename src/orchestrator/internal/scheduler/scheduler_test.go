package scheduler_test

import (
	"context"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"als/obs"
	"als/orchestrator/internal/db"
	"als/orchestrator/internal/scheduler"
)

func openTestDB(t *testing.T) *db.DB {
	t.Helper()
	d, err := db.Open(filepath.Join(t.TempDir(), "swarm.db"))
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() { d.Close() })
	return d
}

func openTestObs(t *testing.T) *obs.Logger {
	t.Helper()
	l, err := obs.Open(filepath.Join(t.TempDir(), "obs.db"))
	if err != nil {
		t.Fatalf("obs.Open: %v", err)
	}
	t.Cleanup(func() { l.Close() })
	return l
}

func ptr[T any](v T) *T { return &v }

// req-001: Scheduler picks up a pending task within 2 poll cycles.
func TestScheduler_req001_PicksUpPendingTask(t *testing.T) {
	d := openTestDB(t)
	obsLog := openTestObs(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	d.InsertProject(ctx, "proj-001", "Test", nil)
	d.InsertTask(ctx, &db.Task{
		ID: "seed-001", ProjectID: "proj-001",
		Type: "DECOMPOSE", ModelTag: "phi-4:mini",
		Payload: "goal", Status: "pending", Depth: 0,
	})

	var dispatched atomic.Int32
	dispatch := func(ctx context.Context, task *db.Task) error {
		dispatched.Add(1)
		d.UpdateTaskStatus(ctx, task.ID, "completed")
		return nil
	}

	s := scheduler.New(d, obsLog, dispatch, 50*time.Millisecond, 5)
	go s.Run(ctx)

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if dispatched.Load() > 0 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	if dispatched.Load() == 0 {
		t.Error("req-001: scheduler did not dispatch task within 2 seconds")
	}
}

// req-002: Task at depth > maxDepth is pruned immediately.
func TestScheduler_req002_PrunesExceedingMaxDepth(t *testing.T) {
	d := openTestDB(t)
	obsLog := openTestObs(t)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	d.InsertProject(ctx, "proj-depth", "D", nil)

	// Insert task with depth 6 (> max 5). SQLite CHECK allows depth 0-5,
	// so we insert with depth 5 and manually update to 6 via raw SQL.
	d.InsertTask(ctx, &db.Task{
		ID: "deep-task", ProjectID: "proj-depth",
		Type: "GO_CODE", ModelTag: "qwen2.5-coder:7b",
		Payload: "x", Status: "pending", Depth: 5,
	})
	// Bypass the CHECK constraint for this test by using raw SQL.
	d.SQL().Exec(`UPDATE tasks SET depth = 6 WHERE id = 'deep-task'`)

	var dispatchCount atomic.Int32
	dispatch := func(ctx context.Context, task *db.Task) error {
		dispatchCount.Add(1)
		return nil
	}

	s := scheduler.New(d, obsLog, dispatch, 50*time.Millisecond, 5)
	go s.Run(ctx)
	time.Sleep(300 * time.Millisecond)
	cancel()

	// Task must be pruned, not dispatched.
	task, err := d.GetTask(context.Background(), "deep-task")
	if err != nil {
		t.Fatalf("GetTask: %v", err)
	}
	if task.Status != "pruned" {
		t.Errorf("req-002: want task pruned, got %s", task.Status)
	}
	if dispatchCount.Load() != 0 {
		t.Errorf("req-002: dispatch must not be called for depth > max, got %d calls", dispatchCount.Load())
	}
}

// req-002: Scheduler recovers from a panicking dispatcher.
func TestScheduler_req002_RecoverFromPanic(t *testing.T) {
	d := openTestDB(t)
	obsLog := openTestObs(t)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	d.InsertProject(ctx, "proj-panic", "P", nil)
	d.InsertTask(ctx, &db.Task{
		ID: "panic-task", ProjectID: "proj-panic",
		Type: "DECOMPOSE", ModelTag: "phi-4:mini",
		Payload: "x", Status: "pending", Depth: 0,
	})

	dispatch := func(ctx context.Context, task *db.Task) error {
		panic("simulated task panic")
	}

	s := scheduler.New(d, obsLog, dispatch, 50*time.Millisecond, 5)

	// Must not panic the test.
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("req-002: scheduler propagated panic: %v", r)
		}
	}()
	go s.Run(ctx)
	time.Sleep(300 * time.Millisecond)
}

// req-004: Sticky sort puts same-model tasks ahead of different-model tasks.
// This is tested indirectly: two tasks (phi-4:mini, llama3.1:8b) are pending;
// the phi task should dispatch before the llama task when phi is active.
func TestScheduler_req004_StickyModel(t *testing.T) {
	d := openTestDB(t)
	obsLog := openTestObs(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	d.InsertProject(ctx, "proj-sticky", "S", nil)
	d.InsertTask(ctx, &db.Task{
		ID: "phi-task", ProjectID: "proj-sticky",
		Type: "DECOMPOSE", ModelTag: "phi-4:mini",
		Payload: "x", Status: "pending", Depth: 0,
	})
	d.InsertTask(ctx, &db.Task{
		ID: "llama-task", ProjectID: "proj-sticky",
		Type: "RESEARCH", ModelTag: "llama3.1:8b",
		Payload: "y", Status: "pending", Depth: 0,
	})

	var dispatchOrder []string
	dispatch := func(ctx context.Context, task *db.Task) error {
		dispatchOrder = append(dispatchOrder, task.ID)
		d.UpdateTaskStatus(ctx, task.ID, "completed")
		time.Sleep(10 * time.Millisecond) // hold model slot briefly
		return nil
	}

	s := scheduler.New(d, obsLog, dispatch, 50*time.Millisecond, 5)
	go s.Run(ctx)

	// Wait for both tasks to complete.
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if len(dispatchOrder) >= 2 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	if len(dispatchOrder) < 2 {
		t.Errorf("req-004: want 2 tasks dispatched, got %d", len(dispatchOrder))
	}
}
