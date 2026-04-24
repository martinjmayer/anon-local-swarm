package db_test

import (
	"context"
	"path/filepath"
	"testing"

	"als/orchestrator/internal/db"
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

func ptr[T any](v T) *T { return &v }

// req-001: Seed task with missing project_id is rejected by FK constraint.
func TestInsertTask_req001_FKConstraint(t *testing.T) {
	d := openTestDB(t)
	err := d.InsertTask(context.Background(), &db.Task{
		ID:        "task-001",
		ProjectID: "nonexistent-project",
		Type:      "DECOMPOSE",
		ModelTag:  "phi-4:mini",
		Payload:   "test",
		Status:    "pending",
		Depth:     0,
	})
	if err == nil {
		t.Fatal("req-001: want FK error for missing project_id, got nil")
	}
}

// req-001: Seed task with valid project_id is persisted.
func TestInsertTask_req001_ValidSeed(t *testing.T) {
	d := openTestDB(t)
	ctx := context.Background()

	_, err := d.InsertProject(ctx, "proj-001", "Test Project", nil)
	if err != nil {
		t.Fatalf("InsertProject: %v", err)
	}

	err = d.InsertTask(ctx, &db.Task{
		ID:        "task-seed-001",
		ProjectID: "proj-001",
		Type:      "DECOMPOSE",
		ModelTag:  "phi-4:mini",
		Payload:   "Build a Go CLI tool for X",
		Status:    "pending",
		Depth:     0,
	})
	if err != nil {
		t.Fatalf("req-001: InsertTask: %v", err)
	}

	got, err := d.GetTask(ctx, "task-seed-001")
	if err != nil {
		t.Fatalf("GetTask: %v", err)
	}
	if got.Depth != 0 {
		t.Errorf("req-001: want depth 0, got %d", got.Depth)
	}
	if got.Status != "pending" {
		t.Errorf("req-001: want status pending, got %s", got.Status)
	}
}

// req-002: PendingTasks returns tasks whose blocked_by is completed.
func TestPendingTasks_req002_UnblocksOnDependencyCompletion(t *testing.T) {
	d := openTestDB(t)
	ctx := context.Background()

	d.InsertProject(ctx, "proj-002", "P2", nil)

	// Insert a blocker task (completed).
	d.InsertTask(ctx, &db.Task{
		ID: "blocker", ProjectID: "proj-002",
		Type: "DECOMPOSE", ModelTag: "phi-4:mini",
		Payload: "decompose", Status: "completed", Depth: 0,
	})

	// Insert a blocked task pointing to the completed blocker.
	d.InsertTask(ctx, &db.Task{
		ID: "dependent", ProjectID: "proj-002",
		ParentID: ptr("blocker"), BlockedBy: ptr("blocker"),
		Type: "RESEARCH", ModelTag: "llama3.1:8b",
		Payload: "research X", Status: "pending", Depth: 1,
	})

	tasks, err := d.PendingTasks(ctx)
	if err != nil {
		t.Fatalf("PendingTasks: %v", err)
	}
	if len(tasks) != 1 || tasks[0].ID != "dependent" {
		t.Errorf("req-002: want [dependent], got %v", taskIDs(tasks))
	}
}

// req-002: PendingTasks does not return tasks blocked by a non-completed task.
func TestPendingTasks_req002_BlockedByActiveNotReturned(t *testing.T) {
	d := openTestDB(t)
	ctx := context.Background()

	d.InsertProject(ctx, "proj-003", "P3", nil)

	// Blocker still active.
	d.InsertTask(ctx, &db.Task{
		ID: "active-blocker", ProjectID: "proj-003",
		Type: "DECOMPOSE", ModelTag: "phi-4:mini",
		Payload: "x", Status: "active", Depth: 0,
	})
	d.InsertTask(ctx, &db.Task{
		ID: "still-blocked", ProjectID: "proj-003",
		BlockedBy: ptr("active-blocker"),
		Type: "RESEARCH", ModelTag: "llama3.1:8b",
		Payload: "y", Status: "pending", Depth: 1,
	})

	tasks, err := d.PendingTasks(ctx)
	if err != nil {
		t.Fatalf("PendingTasks: %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("req-002: want 0 pending tasks, got %v", taskIDs(tasks))
	}
}

// req-006: PruneTasksBranch recursively marks all descendants as pruned.
func TestPruneTasksBranch_req006_CascadingPrune(t *testing.T) {
	d := openTestDB(t)
	ctx := context.Background()

	d.InsertProject(ctx, "proj-prune", "Prune", nil)

	// root → child → grandchild
	d.InsertTask(ctx, &db.Task{
		ID: "root", ProjectID: "proj-prune",
		Type: "DECOMPOSE", ModelTag: "phi-4:mini",
		Payload: "r", Status: "pending", Depth: 0,
	})
	d.InsertTask(ctx, &db.Task{
		ID: "child", ProjectID: "proj-prune", ParentID: ptr("root"),
		Type: "RESEARCH", ModelTag: "llama3.1:8b",
		Payload: "c", Status: "pending", Depth: 1,
	})
	d.InsertTask(ctx, &db.Task{
		ID: "grandchild", ProjectID: "proj-prune", ParentID: ptr("child"),
		Type: "GO_CODE", ModelTag: "qwen2.5-coder:7b",
		Payload: "g", Status: "blocked", Depth: 2,
	})

	pruned, err := d.PruneTasksBranch(ctx, "root")
	if err != nil {
		t.Fatalf("PruneTasksBranch: %v", err)
	}
	if pruned != 3 {
		t.Errorf("req-006: want 3 pruned tasks, got %d", pruned)
	}

	for _, id := range []string{"root", "child", "grandchild"} {
		task, err := d.GetTask(ctx, id)
		if err != nil {
			t.Fatalf("GetTask %s: %v", id, err)
		}
		if task.Status != "pruned" {
			t.Errorf("req-006: task %s want pruned, got %s", id, task.Status)
		}
	}
}

// req-005: InsertCandidate and SelectCandidate enforce single selection per task.
func TestSelectCandidate_req005_SingleSelection(t *testing.T) {
	d := openTestDB(t)
	ctx := context.Background()

	d.InsertProject(ctx, "proj-var", "Var", nil)
	d.InsertTask(ctx, &db.Task{
		ID: "task-var", ProjectID: "proj-var",
		Type: "GO_CODE", ModelTag: "qwen2.5-coder:7b",
		Payload: "write main.go", Status: "active", Depth: 0,
	})

	for i, id := range []string{"c1", "c2", "c3"} {
		_ = i
		d.InsertCandidate(ctx, &db.Candidate{
			ID: id, TaskID: "task-var",
			OutputText: "output " + id,
		})
	}

	score := 8.5
	if err := d.SelectCandidate(ctx, "c2", "task-var", score); err != nil {
		t.Fatalf("SelectCandidate: %v", err)
	}

	candidates, err := d.CandidatesForTask(ctx, "task-var")
	if err != nil {
		t.Fatalf("CandidatesForTask: %v", err)
	}
	if len(candidates) != 3 {
		t.Fatalf("req-005: want 3 candidates, got %d", len(candidates))
	}

	selected := 0
	for _, c := range candidates {
		if c.IsSelected {
			selected++
			if c.ID != "c2" {
				t.Errorf("req-005: wrong candidate selected: %s", c.ID)
			}
		}
	}
	if selected != 1 {
		t.Errorf("req-005: want exactly 1 selected, got %d", selected)
	}
}

func taskIDs(tasks []*db.Task) []string {
	ids := make([]string, len(tasks))
	for i, t := range tasks {
		ids[i] = t.ID
	}
	return ids
}
