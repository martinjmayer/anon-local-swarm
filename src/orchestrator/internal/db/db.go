// Package db owns swarm.db (SQLite) — the application state store for the
// Autonomous Local Swarm. It implements the schema described in
// src/orchestrator/docs/data-contract.md and exposes typed query functions.
//
// Only the orchestrator package may write to swarm.db (data ownership rule).
package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite" // pure-Go SQLite driver; no CGO required
)

// DB wraps *sql.DB with swarm-specific query methods.
type DB struct {
	sql *sql.DB
}

// Open opens (or creates) swarm.db at path, enables WAL mode, and applies the
// schema migration idempotently.
func Open(path string) (*DB, error) {
	raw, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("db: open %q: %w", path, err)
	}

	// WAL mode enables concurrent readers while the scheduler writes.
	// REQ-002: scheduler poll must not block task completion writes.
	if _, err := raw.Exec(`PRAGMA journal_mode=WAL`); err != nil {
		raw.Close()
		return nil, fmt.Errorf("db: set WAL mode: %w", err)
	}

	if _, err := raw.Exec(`PRAGMA foreign_keys=ON`); err != nil {
		raw.Close()
		return nil, fmt.Errorf("db: enable FK: %w", err)
	}

	db := &DB{sql: raw}
	if err := db.migrate(); err != nil {
		raw.Close()
		return nil, fmt.Errorf("db: migrate: %w", err)
	}
	return db, nil
}

// Close releases the database connection.
func (d *DB) Close() error {
	return d.sql.Close()
}

// SQL returns the underlying *sql.DB for use in tests that need raw access.
func (d *DB) SQL() *sql.DB { return d.sql }

// WithTx executes fn inside a serialisable transaction. If fn returns an error
// the transaction is rolled back; otherwise it is committed.
// REQ-002: task dispatch must be atomic (pending → active in a single write).
func (d *DB) WithTx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := d.sql.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("db: begin tx: %w", err)
	}
	if err := fn(tx); err != nil {
		tx.Rollback() //nolint:errcheck
		return err
	}
	return tx.Commit()
}

// ── Projects ────────────────────────────────────────────────────────────────

// Project mirrors the projects table row.
type Project struct {
	ID        string
	Name      string
	Status    string // active | pruned | completed
	OutputDir *string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// InsertProject inserts a new project row and returns the created Project.
func (d *DB) InsertProject(ctx context.Context, id, name string, outputDir *string) (*Project, error) {
	const q = `INSERT INTO projects (id, name, output_dir) VALUES (?, ?, ?)`
	if _, err := d.sql.ExecContext(ctx, q, id, name, outputDir); err != nil {
		return nil, fmt.Errorf("db: insert project: %w", err)
	}
	return d.GetProject(ctx, id)
}

// GetProject returns a project by ID.
func (d *DB) GetProject(ctx context.Context, id string) (*Project, error) {
	const q = `SELECT id, name, status, output_dir, created_at, updated_at FROM projects WHERE id = ?`
	row := d.sql.QueryRowContext(ctx, q, id)
	return scanProject(row)
}

// UpdateProjectStatus sets the project status and bumps updated_at.
func (d *DB) UpdateProjectStatus(ctx context.Context, id, status string) error {
	const q = `UPDATE projects SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := d.sql.ExecContext(ctx, q, status, id)
	return err
}

func scanProject(row *sql.Row) (*Project, error) {
	var p Project
	var outputDir sql.NullString
	err := row.Scan(&p.ID, &p.Name, &p.Status, &outputDir, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("db: scan project: %w", err)
	}
	if outputDir.Valid {
		p.OutputDir = &outputDir.String
	}
	return &p, nil
}

// ── Tasks ────────────────────────────────────────────────────────────────────

// Task mirrors the tasks table row.
type Task struct {
	ID          string
	ProjectID   string
	ParentID    *string
	BlockedBy   *string
	Type        string // DECOMPOSE | RESEARCH | GO_CODE | VIABILITY_REVIEW | SUMMARY
	ModelTag    string
	Payload     string
	Status      string // blocked | pending | active | completed | pruned | failed
	Depth       int
	ContextPath *string
	MetaData    *string // JSON
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// InsertTask inserts a task row. status defaults to 'blocked' if blocked_by is
// set; otherwise 'pending'. The caller must ensure project_id exists.
func (d *DB) InsertTask(ctx context.Context, t *Task) error {
	const q = `
		INSERT INTO tasks
			(id, project_id, parent_id, blocked_by, type, model_tag, payload,
			 status, depth, context_path, meta_data)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := d.sql.ExecContext(ctx, q,
		t.ID, t.ProjectID, t.ParentID, t.BlockedBy,
		t.Type, t.ModelTag, t.Payload,
		t.Status, t.Depth, t.ContextPath, t.MetaData,
	)
	if err != nil {
		return fmt.Errorf("db: insert task %s: %w", t.ID, err)
	}
	return nil
}

// GetTask returns a task by ID.
func (d *DB) GetTask(ctx context.Context, id string) (*Task, error) {
	const q = `SELECT id, project_id, parent_id, blocked_by, type, model_tag, payload,
		status, depth, context_path, meta_data, created_at, updated_at
		FROM tasks WHERE id = ?`
	return scanTask(d.sql.QueryRowContext(ctx, q, id))
}

// UpdateTaskStatus sets task status and bumps updated_at. Use within WithTx
// for atomic dispatch (pending → active).
func (d *DB) UpdateTaskStatus(ctx context.Context, id, status string) error {
	const q = `UPDATE tasks SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := d.sql.ExecContext(ctx, q, status, id)
	return err
}

// UpdateTaskStatusTx is UpdateTaskStatus scoped to an existing transaction.
func UpdateTaskStatusTx(ctx context.Context, tx *sql.Tx, id, status string) error {
	const q = `UPDATE tasks SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := tx.ExecContext(ctx, q, status, id)
	return err
}

// UpdateTaskMetaData updates the meta_data JSON column.
func (d *DB) UpdateTaskMetaData(ctx context.Context, id, metaJSON string) error {
	const q = `UPDATE tasks SET meta_data = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := d.sql.ExecContext(ctx, q, metaJSON, id)
	return err
}

// UpdateTaskContextPath sets the context sidecar path.
func (d *DB) UpdateTaskContextPath(ctx context.Context, id, path string) error {
	const q = `UPDATE tasks SET context_path = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := d.sql.ExecContext(ctx, q, path, id)
	return err
}

// PendingTasks returns all tasks with status=pending whose blocked_by is either
// NULL or points to a completed task. REQ-002: scheduler query.
func (d *DB) PendingTasks(ctx context.Context) ([]*Task, error) {
	const q = `
		SELECT t.id, t.project_id, t.parent_id, t.blocked_by, t.type, t.model_tag,
		       t.payload, t.status, t.depth, t.context_path, t.meta_data,
		       t.created_at, t.updated_at
		FROM tasks t
		WHERE t.status = 'pending'
		  AND (
		        t.blocked_by IS NULL
		        OR EXISTS (
		            SELECT 1 FROM tasks dep
		            WHERE dep.id = t.blocked_by
		              AND dep.status = 'completed'
		        )
		      )`
	rows, err := d.sql.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("db: pending tasks: %w", err)
	}
	defer rows.Close()
	return scanTasks(rows)
}

// TasksByProject returns all tasks for a project (used in cascading prune).
func (d *DB) TasksByProject(ctx context.Context, projectID string) ([]*Task, error) {
	const q = `SELECT id, project_id, parent_id, blocked_by, type, model_tag, payload,
		status, depth, context_path, meta_data, created_at, updated_at
		FROM tasks WHERE project_id = ?`
	rows, err := d.sql.QueryContext(ctx, q, projectID)
	if err != nil {
		return nil, fmt.Errorf("db: tasks by project %s: %w", projectID, err)
	}
	defer rows.Close()
	return scanTasks(rows)
}

// ChildTasks returns all tasks whose parent_id equals parentID.
func (d *DB) ChildTasks(ctx context.Context, parentID string) ([]*Task, error) {
	const q = `SELECT id, project_id, parent_id, blocked_by, type, model_tag, payload,
		status, depth, context_path, meta_data, created_at, updated_at
		FROM tasks WHERE parent_id = ?`
	rows, err := d.sql.QueryContext(ctx, q, parentID)
	if err != nil {
		return nil, fmt.Errorf("db: child tasks of %s: %w", parentID, err)
	}
	defer rows.Close()
	return scanTasks(rows)
}

// PruneTasksBranch recursively marks all descendants of rootID as pruned.
// It returns the count of rows updated. REQ-006.
func (d *DB) PruneTasksBranch(ctx context.Context, rootID string) (int, error) {
	// Collect all descendant IDs via BFS using the parent_id graph.
	seen := map[string]bool{rootID: true}
	queue := []string{rootID}
	var all []string

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		all = append(all, cur)

		children, err := d.ChildTasks(ctx, cur)
		if err != nil {
			return 0, err
		}
		for _, c := range children {
			if !seen[c.ID] {
				seen[c.ID] = true
				queue = append(queue, c.ID)
			}
		}
	}

	pruned := 0
	for _, id := range all {
		if err := d.UpdateTaskStatus(ctx, id, "pruned"); err != nil {
			return pruned, fmt.Errorf("db: prune task %s: %w", id, err)
		}
		pruned++
	}
	return pruned, nil
}

func scanTask(row *sql.Row) (*Task, error) {
	var t Task
	var parentID, blockedBy, contextPath, metaData sql.NullString
	err := row.Scan(
		&t.ID, &t.ProjectID, &parentID, &blockedBy,
		&t.Type, &t.ModelTag, &t.Payload,
		&t.Status, &t.Depth, &contextPath, &metaData,
		&t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("db: scan task: %w", err)
	}
	if parentID.Valid {
		t.ParentID = &parentID.String
	}
	if blockedBy.Valid {
		t.BlockedBy = &blockedBy.String
	}
	if contextPath.Valid {
		t.ContextPath = &contextPath.String
	}
	if metaData.Valid {
		t.MetaData = &metaData.String
	}
	return &t, nil
}

func scanTasks(rows *sql.Rows) ([]*Task, error) {
	var tasks []*Task
	for rows.Next() {
		var t Task
		var parentID, blockedBy, contextPath, metaData sql.NullString
		if err := rows.Scan(
			&t.ID, &t.ProjectID, &parentID, &blockedBy,
			&t.Type, &t.ModelTag, &t.Payload,
			&t.Status, &t.Depth, &contextPath, &metaData,
			&t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("db: scan task row: %w", err)
		}
		if parentID.Valid {
			t.ParentID = &parentID.String
		}
		if blockedBy.Valid {
			t.BlockedBy = &blockedBy.String
		}
		if contextPath.Valid {
			t.ContextPath = &contextPath.String
		}
		if metaData.Valid {
			t.MetaData = &metaData.String
		}
		tasks = append(tasks, &t)
	}
	return tasks, rows.Err()
}

// ── task_candidates ──────────────────────────────────────────────────────────

// Candidate mirrors a task_candidates row.
type Candidate struct {
	ID              string
	TaskID          string
	OutputText      string
	EvaluationScore *float64
	IsSelected      bool
	CreatedAt       time.Time
}

// InsertCandidate adds a new candidate row. REQ-005.
func (d *DB) InsertCandidate(ctx context.Context, c *Candidate) error {
	const q = `INSERT INTO task_candidates (id, task_id, output_text, evaluation_score, is_selected)
		VALUES (?, ?, ?, ?, ?)`
	isSelected := 0
	if c.IsSelected {
		isSelected = 1
	}
	_, err := d.sql.ExecContext(ctx, q, c.ID, c.TaskID, c.OutputText, c.EvaluationScore, isSelected)
	if err != nil {
		return fmt.Errorf("db: insert candidate %s: %w", c.ID, err)
	}
	return nil
}

// SelectCandidate marks a single candidate as selected (is_selected=1) and
// ensures no other candidate for the same task is selected. REQ-005.
func (d *DB) SelectCandidate(ctx context.Context, candidateID, taskID string, score float64) error {
	return d.WithTx(ctx, func(tx *sql.Tx) error {
		// Clear any prior selection for this task.
		if _, err := tx.ExecContext(ctx,
			`UPDATE task_candidates SET is_selected = 0 WHERE task_id = ?`, taskID,
		); err != nil {
			return err
		}
		// Set score and select.
		if _, err := tx.ExecContext(ctx,
			`UPDATE task_candidates SET is_selected = 1, evaluation_score = ? WHERE id = ?`,
			score, candidateID,
		); err != nil {
			return err
		}
		return nil
	})
}

// CandidatesForTask returns all candidates for a task. REQ-005.
func (d *DB) CandidatesForTask(ctx context.Context, taskID string) ([]*Candidate, error) {
	const q = `SELECT id, task_id, output_text, evaluation_score, is_selected, created_at
		FROM task_candidates WHERE task_id = ? ORDER BY created_at`
	rows, err := d.sql.QueryContext(ctx, q, taskID)
	if err != nil {
		return nil, fmt.Errorf("db: candidates for task %s: %w", taskID, err)
	}
	defer rows.Close()

	var candidates []*Candidate
	for rows.Next() {
		var c Candidate
		var score sql.NullFloat64
		var isSelected int
		if err := rows.Scan(&c.ID, &c.TaskID, &c.OutputText, &score, &isSelected, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("db: scan candidate: %w", err)
		}
		if score.Valid {
			c.EvaluationScore = &score.Float64
		}
		c.IsSelected = isSelected == 1
		candidates = append(candidates, &c)
	}
	return candidates, rows.Err()
}
