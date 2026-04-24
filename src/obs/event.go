// Package obs owns obs.db (DuckDB) and writes all swarm events as an
// append-only log, mirrored to stdout via log/slog.
//
// REQ-011: Every task state transition and supplementary event (VRAM swap,
// cascading prune, self-correction, context compression, MCP tool call,
// output deposit, orchestrator lifecycle) is persisted here.
package obs

// EventType is the enumerated set of event categories the swarm can emit.
// These values are written verbatim to the event_type column in obs.db and
// are constrained by a CHECK constraint at the database level.
type EventType string

const (
	// EventTaskTransition is emitted on every task status change.
	EventTaskTransition EventType = "task_transition"

	// EventVRAMSwap is emitted when the orchestrator unloads one model and
	// loads another (keep_alive: 0 strategy).
	EventVRAMSwap EventType = "vram_swap"

	// EventCascadingPrune is emitted when a VIABILITY_REVIEW score triggers a
	// recursive prune of the branch.
	EventCascadingPrune EventType = "cascading_prune"

	// EventSelfCorrection is emitted on each self-correction attempt by the
	// validator (go build failure → new GO_CODE task).
	EventSelfCorrection EventType = "self_correction"

	// EventContextCompression is emitted when a SUMMARY task is spawned at the
	// 70% context threshold.
	EventContextCompression EventType = "context_compression"

	// EventMCPToolCall is emitted on each MCP tool invocation (success or failure).
	EventMCPToolCall EventType = "mcp_tool_call"

	// EventOutputDeposit is emitted when a terminal-state artifact is written to
	// the project output directory.
	EventOutputDeposit EventType = "output_deposit"

	// EventOrchestratorStart is emitted once when the orchestrator process starts.
	EventOrchestratorStart EventType = "orchestrator_start"

	// EventOrchestratorStop is emitted once when the orchestrator process shuts down.
	EventOrchestratorStop EventType = "orchestrator_stop"
)

// Event is the payload passed to WriteEvent. Type is required; all other fields
// are optional and are written as SQL NULL when nil.
//
// The Detail field MUST NOT contain API credentials, tokens, or secrets.
// WriteEvent applies credential scrubbing before persisting, but callers
// should not rely on this as the primary defence.
type Event struct {
	// Type is the event category. Required.
	Type EventType

	// TaskID is the UUID of the related task. Nil for system-level events
	// (orchestrator_start, orchestrator_stop, vram_swap).
	TaskID *string

	// TaskType is one of DECOMPOSE / RESEARCH / GO_CODE / VIABILITY_REVIEW / SUMMARY.
	TaskType *string

	// FromStatus is the task status before a transition (task_transition events only).
	FromStatus *string

	// ToStatus is the task status after a transition (task_transition events only).
	ToStatus *string

	// ModelTag is the Ollama model tag involved in the event (e.g. "qwen2.5-coder:7b").
	ModelTag *string

	// Depth is the task depth in the DAG at the time of the event.
	Depth *int

	// Detail is an arbitrary JSON-serialisable value carrying event-specific
	// context (scores, error messages, counts). MUST NOT contain credentials.
	Detail any
}
