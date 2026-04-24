package db

// ddl is the initial schema — M-001 in the migration history.
// Applied idempotently via CREATE TABLE IF NOT EXISTS.
const ddl = `
CREATE TABLE IF NOT EXISTS projects (
    id         TEXT     NOT NULL PRIMARY KEY,
    name       TEXT     NOT NULL,
    status     TEXT     NOT NULL DEFAULT 'active'
                        CHECK (status IN ('active', 'pruned', 'completed')),
    output_dir TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_projects_status ON projects (status);

CREATE TABLE IF NOT EXISTS tasks (
    id           TEXT     NOT NULL PRIMARY KEY,
    project_id   TEXT     NOT NULL REFERENCES projects(id),
    parent_id    TEXT     REFERENCES tasks(id),
    blocked_by   TEXT     REFERENCES tasks(id),
    type         TEXT     NOT NULL
                          CHECK (type IN ('DECOMPOSE','RESEARCH','GO_CODE','VIABILITY_REVIEW','SUMMARY')),
    model_tag    TEXT     NOT NULL,
    payload      TEXT     NOT NULL,
    status       TEXT     NOT NULL DEFAULT 'blocked'
                          CHECK (status IN ('blocked','pending','active','completed','pruned','failed')),
    depth        INTEGER  NOT NULL DEFAULT 0
                          CHECK (depth >= 0 AND depth <= 5),
    context_path TEXT,
    meta_data    TEXT,
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_tasks_status_model   ON tasks (status, model_tag);
CREATE INDEX IF NOT EXISTS idx_tasks_blocked_by     ON tasks (blocked_by);
CREATE INDEX IF NOT EXISTS idx_tasks_project_id     ON tasks (project_id);
CREATE INDEX IF NOT EXISTS idx_tasks_parent_id      ON tasks (parent_id);

CREATE TABLE IF NOT EXISTS task_candidates (
    id               TEXT    NOT NULL PRIMARY KEY,
    task_id          TEXT    NOT NULL REFERENCES tasks(id),
    output_text      TEXT    NOT NULL,
    evaluation_score REAL,
    is_selected      INTEGER NOT NULL DEFAULT 0
                             CHECK (is_selected IN (0, 1)),
    created_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_candidates_task_id          ON task_candidates (task_id);
CREATE INDEX IF NOT EXISTS idx_candidates_task_is_selected ON task_candidates (task_id, is_selected);
`

// migrate applies the DDL idempotently.
func (d *DB) migrate() error {
	_, err := d.sql.Exec(ddl)
	return err
}
