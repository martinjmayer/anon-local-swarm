package obs

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	_ "github.com/marcboeker/go-duckdb" // DuckDB driver registration
)

// ddl is applied idempotently on Open. CREATE IF NOT EXISTS ensures re-runs
// against an existing obs.db are safe.
const ddl = `
CREATE SEQUENCE IF NOT EXISTS events_id_seq START 1;

CREATE TABLE IF NOT EXISTS events (
    id          UHUGEINT    NOT NULL DEFAULT nextval('events_id_seq') PRIMARY KEY,
    timestamp   TIMESTAMPTZ NOT NULL DEFAULT now(),
    event_type  VARCHAR     NOT NULL CHECK (event_type IN (
                    'task_transition',
                    'vram_swap',
                    'cascading_prune',
                    'self_correction',
                    'context_compression',
                    'mcp_tool_call',
                    'output_deposit',
                    'orchestrator_start',
                    'orchestrator_stop'
                )),
    task_id     VARCHAR,
    task_type   VARCHAR,
    from_status VARCHAR,
    to_status   VARCHAR,
    model_tag   VARCHAR,
    depth       INTEGER,
    detail      JSON
);

CREATE INDEX IF NOT EXISTS idx_events_timestamp  ON events (timestamp);
CREATE INDEX IF NOT EXISTS idx_events_task_id    ON events (task_id);
CREATE INDEX IF NOT EXISTS idx_events_event_type ON events (event_type);
`

// insertSQL is the single write path. All columns are positional parameters.
const insertSQL = `
INSERT INTO events
    (event_type, task_id, task_type, from_status, to_status, model_tag, depth, detail)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

// credentialSubstrings are lower-case substrings used to detect JSON keys that
// likely carry credential material. Matched keys are redacted before persistence.
var credentialSubstrings = []string{
	"api_key", "apikey", "token", "secret", "password",
	"passwd", "credential", "bearer", "auth",
}

// Logger owns the DuckDB connection to obs.db and exposes WriteEvent.
// Construct one with Open; release with Close.
type Logger struct {
	db  *sql.DB
	log *slog.Logger
}

// Open opens (or creates) a DuckDB database at path, applies the events schema
// idempotently, and returns a Logger ready to accept WriteEvent calls.
//
// Returns an error if the database cannot be opened or the schema migration fails.
// The caller must call Close when done.
func Open(path string) (*Logger, error) {
	db, err := sql.Open("duckdb", path)
	if err != nil {
		return nil, fmt.Errorf("obs: open duckdb %q: %w", path, err)
	}

	if err := db.PingContext(context.Background()); err != nil {
		db.Close()
		return nil, fmt.Errorf("obs: ping duckdb %q: %w", path, err)
	}

	if _, err := db.ExecContext(context.Background(), ddl); err != nil {
		db.Close()
		return nil, fmt.Errorf("obs: apply schema to %q: %w", path, err)
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	return &Logger{
		db:  db,
		log: slog.New(handler),
	}, nil
}

// Close releases the underlying DuckDB connection. Safe to call multiple times.
func (l *Logger) Close() error {
	return l.db.Close()
}

// WriteEvent persists e to obs.db and mirrors it to stdout via slog.
//
// Failures are absorbed — a DuckDB write failure is logged to stderr but never
// returned, because a logging failure must never crash the orchestrator (REQ-011).
// The stdout mirror is always attempted, even if the DuckDB write fails.
func (l *Logger) WriteEvent(e Event) {
	detailJSON := scrubAndMarshal(e.Detail)

	_, dbErr := l.db.ExecContext(
		context.Background(),
		insertSQL,
		string(e.Type),
		e.TaskID,
		e.TaskType,
		e.FromStatus,
		e.ToStatus,
		e.ModelTag,
		e.Depth,
		detailJSON,
	)
	if dbErr != nil {
		// Graceful failure: surface to stderr, do not propagate (REQ-011).
		slog.Error("obs: duckdb write failed", "event_type", string(e.Type), "error", dbErr)
	}

	l.mirrorToStdout(e, detailJSON)
}

// mirrorToStdout emits a structured slog line for the event. All nil fields are
// omitted so the log line stays compact.
func (l *Logger) mirrorToStdout(e Event, detailJSON []byte) {
	attrs := []any{
		"event_type", string(e.Type),
		"ts", time.Now().UTC().Format(time.RFC3339),
	}
	if e.TaskID != nil {
		attrs = append(attrs, "task_id", *e.TaskID)
	}
	if e.TaskType != nil {
		attrs = append(attrs, "task_type", *e.TaskType)
	}
	if e.FromStatus != nil {
		attrs = append(attrs, "from_status", *e.FromStatus)
	}
	if e.ToStatus != nil {
		attrs = append(attrs, "to_status", *e.ToStatus)
	}
	if e.ModelTag != nil {
		attrs = append(attrs, "model_tag", *e.ModelTag)
	}
	if e.Depth != nil {
		attrs = append(attrs, "depth", *e.Depth)
	}
	if detailJSON != nil {
		attrs = append(attrs, "detail", string(detailJSON))
	}
	l.log.Info("event", attrs...)
}

// scrubAndMarshal marshals v to JSON after redacting any top-level or nested
// object key whose name (case-insensitive) matches a known credential substring.
// Returns nil if v is nil or marshalling fails (failure is logged to stderr).
func scrubAndMarshal(v any) []byte {
	if v == nil {
		return nil
	}
	raw, err := json.Marshal(v)
	if err != nil {
		slog.Error("obs: marshal detail failed", "error", err)
		return nil
	}

	// Attempt to treat as a JSON object for key-level scrubbing.
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		// Not a JSON object — return as-is; no key scrubbing possible.
		return raw
	}

	scrubMap(m)

	out, err := json.Marshal(m)
	if err != nil {
		slog.Error("obs: re-marshal after scrub failed", "error", err)
		return nil
	}
	return out
}

// scrubMap redacts values for credential-like keys in-place, recursively
// traversing nested maps.
func scrubMap(m map[string]any) {
	for k, v := range m {
		if isCredentialKey(k) {
			m[k] = "[REDACTED]"
			continue
		}
		if nested, ok := v.(map[string]any); ok {
			scrubMap(nested)
		}
	}
}

// isCredentialKey returns true when the lower-cased key contains any known
// credential-related substring.
func isCredentialKey(key string) bool {
	lower := strings.ToLower(key)
	for _, sub := range credentialSubstrings {
		if strings.Contains(lower, sub) {
			return true
		}
	}
	return false
}
