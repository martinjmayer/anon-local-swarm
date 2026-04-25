# Project: Autonomous Local Swarm (ALS)
## Full Solution Design Document v1.7

### 1. Executive Summary
The ALS is a local-first, autonomous R&D engine optimized for 8GB VRAM environments. It bridges the intelligence gap of Small Language Models (SLMs) by utilizing a specialized **Decomposition Engine**, **Variational Execution**, and a **Directed Acyclic Graph (DAG)** to manage complex task dependencies. The system identifies and develops income-generating Go-based assets with rigorous automated validation and pruning.

---

### 2. Core Architecture

#### A. The Decomposition Engine (Pre-Processor)
To ensure high-quality output from 3B-7B models, all "Seed" ideas pass through a reasoning-tuned model (e.g., Phi-4 Mini).
* **Role:** Shatters high-level goals into atomic, actionable steps.
* **Mechanism:** Outputs a structured JSON array of tasks, ensuring at least one `VIABILITY_REVIEW` is tethered to every development branch.

#### B. The Orchestration Layer (Golang)
* **DAG Scheduler:** Manages the `blocked_by` state, ensuring implementation only occurs after a "Pass" from the Gatekeeper.
* **Sticky Model Strategy:** Groups tasks by `model_tag` to minimize the SSD-to-VRAM "loading tax".
* **Ensemble Voting:** For high-priority tasks, the orchestrator triggers **Variational Execution**—running a prompt multiple times with varied temperatures—before selecting the best output via a "Judge" model.

#### C. Intent-Based Semantic Routing
| Task Type | Purpose | Primary Model |
| :--- | :--- | :--- |
| `DECOMPOSE` | Pre-processing seeds into atomic tasks. | Phi-4 Mini |
| `RESEARCH` | Market exploration and data mining. | Llama-3.1 8B |
| `GO_CODE` | Scripting and CLI development. | Qwen-2.5-Coder 7B |
| `VIABILITY_REVIEW`| The "Gatekeeper" viability check. | Phi-4 Mini |
| `SUMMARY` | Context compression at 70% limit. | SmolLM3 135M |

---

### 3. Logic & State Management

#### I. Dependency & Cascading Pruning
* **Blocked Status:** Implementation tasks inherit a `blocked_by` status from a Review ID.
* **Purge Logic:** If a `VIABILITY_REVIEW` returns a score < 5/10, the Go Orchestrator recursively marks all dependent tasks as `PRUNED`, instantly clearing the queue of dead-end ideas.

#### II. Depth & Recursion Control
* **Evolutionary Tasks:** `RESEARCH` and `GO_CODE` increment project `depth` (Max 5).
* **Administrative Tasks:** `DECOMPOSE`, `REVIEW`, and `SUMMARY` do not increment depth; they inherit the parent's level.

#### III. Self-Correction & Validation
* **Go Build Check:** All code is passed to `exec.Command("go", "build")`. 
* **Self-Healing:** Failed builds generate a new `GO_CODE` task containing the original code and the specific compiler error for immediate debugging.

---

### 4. The Data Model

#### A. SQLite Schema (swarm.db)
+++sql
CREATE TABLE projects (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    status TEXT DEFAULT 'active' -- 'active', 'pruned', 'completed'
);

CREATE TABLE tasks (
    id UUID PRIMARY KEY,
    project_id UUID REFERENCES projects(id),
    parent_id UUID REFERENCES tasks(id),
    blocked_by UUID REFERENCES tasks(id),
    
    type TEXT NOT NULL, -- DECOMPOSE, RESEARCH, GO_CODE, REVIEW, SUMMARY
    model_tag TEXT NOT NULL,
    payload TEXT NOT NULL,
    
    status TEXT DEFAULT 'blocked', -- blocked, pending, active, completed, pruned
    depth INTEGER DEFAULT 0,
    context_path TEXT, -- JSON Sidecar path
    meta_data JSON -- Scores, candidate selections, compiler logs
);

-- For Variational Execution tracking
CREATE TABLE task_candidates (
    id UUID PRIMARY KEY,
    task_id UUID REFERENCES tasks(id),
    output_text TEXT,
    evaluation_score FLOAT,
    is_selected BOOLEAN DEFAULT FALSE
);
+++

---

### 5. Appendices

#### Appendix A: Context Sidecar Pattern
Long-term memory is stored in branch-specific `.json` files to keep the SQLite database light. This includes extracted entities, business constraints, and the most recent 70% summary.

#### Appendix B: Intelligence Boosting
* **Few-Shot Injection:** Orchestrator prepends 2-3 high-quality examples to prompts to "anchor" SLM output quality.
* **Chain-of-Thought (CoT):** Reasoning models are prompted to "Think" before outputting final JSON/Code.

#### Appendix C: Delivery Pipeline
* **Dropbox Sync:** Final reports and compiled binaries are moved to a local Dropbox-watched folder.
* **SMTP Notify:** A Go routine sends a formatted Markdown summary to the user's email upon project breakthroughs.

#### Appendix D: Testing Strategy & Reliability Framework
* **DAG Dependency Test:** Verify `GetNextPendingTask()` ignores blocked tasks until the prerequisite is `COMPLETED_PASS`.
* **Cascading Pruning Validation:** Assert that failing a `VIABILITY_REVIEW` correctly marks all recursive child tasks as `PRUNED`.
* **Consensus/Judge Test:** Benchmark the Judge model (e.g., Phi-4) to ensure it correctly selects high-quality code from candidate sets.
* **VRAM Guardrail:** Use mock HTTP calls to verify that `"keep_alive": 0` is sent to Ollama between model swaps to prevent OOM errors.
* **Compiler Feedback Test:** Ensure `stderr` from failed Go builds triggers a new "Self-Correction" task rather than terminating the branch.

#### Appendix E: Technical Glossary
* **Decomposition Engine**: A specialized "Pre-Processor" where a reasoning model shatters a high-level goal into atomic, actionable tasks.
* **Variational Execution**: Running the same task multiple times with slight prompt/temperature modifications to generate a candidate set.
* **Ensemble Voting / Selection**: A quality-control mechanism where a "Judge" model selects the most robust version from candidates.
* **Directed Acyclic Graph (DAG)**: The workflow structure where tasks are "blocked" by prerequisites, ensuring no circular logic.
* **Cascading Pruning**: Logic where a failed "Viability Review" cancels all downstream child tasks automatically.
* **Sticky Model Strategy**: An optimization that groups tasks by model to avoid the "Loading Tax" of swapping files in VRAM.
