---
title: "Domain Glossary - Autonomous Local Swarm"
summary: "Ubiquitous language for the ALS. All agents and humans use these terms."
status: "active"
version: "0.1.0"
---
# Domain Glossary - Autonomous Local Swarm

**Skill:** spec-agent (updated by any agent that introduces a new domain term)
**Tool:** claude-code
**Model:** claude-sonnet-4-5
**Feature:** 0000001-autonomous-local-swarm
**Version:** 0.1.0

---

## Terms

| Term | Definition | Aliases | Used In |
|------|-----------|---------|---------|
| Swarm | The complete ALS system: orchestrator, models, MCP tools, databases, and output pipeline running as a single local process group | ALS, system | all components |
| Seed | The initial task inserted by the operator that starts a project run; always type DECOMPOSE, depth 0 | seed task, job brief | orchestrator, swarm.db |
| Project | A named run of the swarm against a single seed goal; owns a set of tasks and an output directory | run, job | orchestrator, swarm.db |
| Task | The atomic unit of work in the swarm; has a type, model, payload, status, depth, and optional blocked_by dependency | work item | orchestrator, swarm.db |
| Task Type | One of five fixed categories that determine which model handles the task: DECOMPOSE, RESEARCH, GO_CODE, VIABILITY_REVIEW, SUMMARY | type | router, orchestrator |
| DAG | Directed Acyclic Graph — the dependency structure of tasks within a project; no circular dependencies permitted | task graph | orchestrator, scheduler |
| Blocked | Task status indicating the task cannot be dispatched until its `blocked_by` prerequisite completes | waiting | scheduler, swarm.db |
| Pending | Task status indicating the task is ready for dispatch (unblocked, not yet active) | ready | scheduler, swarm.db |
| Active | Task status indicating the task is currently being processed by Ollama | in-progress, running | scheduler, swarm.db |
| Completed | Task status indicating the task finished successfully; dependents are unblocked | done, passed | scheduler, swarm.db |
| Pruned | Task status indicating the task was cancelled due to a failed viability review upstream | cancelled, killed | scheduler, prune engine |
| Failed | Task status indicating the self-correction loop exhausted its retry limit | error | scheduler, self-correction |
| Decomposition Engine | The DECOMPOSE task type handler; uses phi-4:mini to shatter a high-level goal into atomic tasks | pre-processor | orchestrator |
| Semantic Router | The component that maps task type to the correct Ollama model and MCP tool set | router | orchestrator |
| Sticky Model Strategy | Scheduler optimisation that batches tasks by model_tag to reduce VRAM swap frequency | model batching | scheduler |
| VRAM Swap Tax | The latency cost of unloading one model and loading another into GPU memory | loading tax, swap cost | orchestrator, obs |
| Variational Execution | Running the same task prompt N times (default 3) with varied temperature to generate a candidate set | ensemble generation | variational-engine |
| Judge | The phi-4:mini model invoked after variational execution to select the best candidate from the set | judge model, selector | variational-engine |
| Candidate | A single output from one variational execution run; stored in task_candidates | variant | variational-engine, swarm.db |
| Viability Review | The VIABILITY_REVIEW task type; a gatekeeper check that scores a branch 0-10 before implementation is unblocked | gatekeeper, review | orchestrator, swarm.db |
| Viability Threshold | The minimum score (default 5) a VIABILITY_REVIEW must return to unblock its dependents | threshold, pass score | orchestrator, config |
| Cascading Prune | The recursive process of marking all descendant tasks PRUNED when a VIABILITY_REVIEW fails | branch pruning | prune engine |
| Self-Correction | The loop that emits a new GO_CODE task when go build fails, including the compiler error in the payload | self-healing, auto-fix | validator |
| Context Sidecar | A per-branch JSON file storing accumulated context (entities, constraints, summaries) for a task lineage | sidecar, context file | orchestrator, swarm.db |
| Context Threshold | The proportion of a model's context window at which a SUMMARY task is emitted (default 70%) | compression threshold | orchestrator |
| Summary Task | The SUMMARY task type; uses smollm3:135m to compress branch context when the threshold is reached | compression task | orchestrator |
| Depth | The recursion level of a task within a project; DECOMPOSE, REVIEW, and SUMMARY do not increment depth | recursion depth, level | scheduler, swarm.db |
| Max Depth | The hard limit on task depth (default 5); tasks beyond this are pruned automatically | MAX_DEPTH | scheduler |
| MCP Client | The embedded client in the orchestrator that registers and calls MCP tool servers | tool client | orchestrator |
| MCP Server | An external process providing tools via the Model Context Protocol; registered at orchestrator startup | tool server | mcp-client |
| Terminal State | A project state where all tasks are in completed, pruned, or failed status; triggers output deposit | done, finished | orchestrator |
| Output Deposit | The act of writing all project artifacts to the configured output directory on reaching terminal state | artifact deposit | orchestrator |
| Prune Report | A markdown file written on terminal state when all branches were pruned; explains scores and reasons | pruning summary | orchestrator |
| obs.db | The DuckDB database storing the append-only event log for all swarm activity | observability db, event log | obs |
| swarm.db | The SQLite database storing application state: projects, tasks, task_candidates | app db, state db | orchestrator |
| Human on the Loop | The operator role; seeds the problem and can intervene but does not need to actively monitor | operator, HOTL | all |

---

*Generated by spec-agent. Updated by any agent that introduces a new domain term.*
