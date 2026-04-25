# Project: Autonomous Local Swarm (ALS)
## Session Transfer & Initialization Packet v1.1

### 1. The Initial Seed (Database Entry)
To align with the v1.7 design, the first task in `swarm.db` should be a `DECOMPOSE` type. This ensures the **Decomposition Engine** breaks the high-level goal into atomic tasks before research begins.

+++json
{
  "id": "seed-task-001",
  "project_id": "p-001",
  "name": "Niche Go-CLI Market Discovery",
  "type": "DECOMPOSE",
  "model_tag": "phi-4:mini",
  "depth": 0,
  "status": "pending",
  "payload": "Goal: Identify and develop high-margin Go-based CLI tools for the legal and accounting niches. Specific focus: Automated PDF metadata extraction or CSV-to-XBRL conversion. Task: Shatter this goal into 5-7 atomic tasks including RESEARCH, VIABILITY_REVIEW, and initial GO_CODE specifications.",
  "context_path": "/dropbox/context/seed-task-001.json"
}
+++

### 2. Orchestrator System Instructions (The "Ghost")
The Go Orchestrator must inject these constraints into every Ollama call to maintain autonomous behavior:

* **Instruction Parsing**: "Identify sub-tasks using the prefix `NEW_TASK:`. If a project specification is ready for vetting, use the prefix `REVIEW_REQUIRED:`."
* **Code Standard**: "All generated code must be valid Golang. Include a `main.go` and a `go.mod` declaration."
* **Decision Logic**: "Focus on high-margin, low-overhead software or data arbitrage logic."

### 3. Operational Constants (8GB VRAM Guardrails)
Hardcode these into your Go configuration to manage the hardware bottleneck:

* **MAX_VRAM_MODELS**: 1 (Strict `keep_alive: 0` during swaps to prevent OOM errors)
* **CONTEXT_THRESHOLD**: 0.70 (Trigger `SUMMARY` task at 70% capacity)
* **MAX_DEPTH**: 5 (Hard-stop for recursive project generation)
* **VIABILITY_THRESHOLD**: 5 (Threshold for **Cascading Pruning**)
* **VARIATION_COUNT**: 3 (Number of candidates to generate during **Variational Execution**)

### 4. Implementation Priority (Next Session Roadmap)
1. **Infrastructure**: Initialize SQLite with the schema from Design v1.7, including the `task_candidates` table.
2. **Connectivity**: Build the Ollama API wrapper in Go, ensuring the `keep_alive` parameter is correctly managed to clear VRAM between swaps.
3. **The DAG Loop**: Develop `RefreshQueue()` to handle `blocked_by` unblocking, `PRUNED` cascading logic, and `DECOMPOSE` output parsing.
4. **Variational Logic**: Implement the loop that triggers multiple runs for a task and the "Judge" logic to select the best `task_candidate`.
5. **Validation**: Implement the `os/exec` wrapper for `go build` and the self-healing feedback loop for compiler errors.

---
**End of Session Context Transfer**
