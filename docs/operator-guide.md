# Operator Guide — Autonomous Local Swarm

Local-first autonomous R&D engine. Seeds a problem brief; runs until terminal state; deposits outputs.

---

## Prerequisites

| Requirement | Notes |
|---|---|
| Docker Desktop | [docker.com](https://www.docker.com/products/docker-desktop/) — used to build and run the swarm |
| Ollama | [ollama.ai](https://ollama.ai) — must be running on the host |
| Node.js | Already bundled in the Docker image; only needed on host for `mcp-reddit` dev |
| DuckDB CLI | Already installed — used to query `obs.db` |

### Pull required models (once)

```powershell
ollama pull phi-4:mini
ollama pull llama3.1:8b
ollama pull qwen2.5-coder:7b
ollama pull smollm3:135m
```

---

## 1. Configure credentials

```powershell
copy .env.example .env
# Edit .env — fill in API keys (never commit this file)
notepad .env
```

Required keys:

| Key | Where to get it |
|---|---|
| `BRAVE_API_KEY` | [brave.com/search/api](https://brave.com/search/api/) — free tier |
| `GITHUB_TOKEN` | GitHub → Settings → Developer settings → Personal access tokens |
| `REDDIT_CLIENT_ID` | [reddit.com/prefs/apps](https://www.reddit.com/prefs/apps) → create script app |
| `REDDIT_CLIENT_SECRET` | Same Reddit app page |
| `OBSIDIAN_API_KEY` | Obsidian → Settings → Local REST API plugin (optional) |

---

## 2. Build the image

```powershell
docker compose build
```

This runs in four stages:
1. **deps** — downloads Go modules
2. **tester** — runs `go test ./...` across all packages; build fails if any test fails
3. **builder** — compiles the Linux/amd64 `swarm` binary (CGO enabled for go-duckdb)
4. **runtime** — packages binary + Node.js + compiled mcp-reddit into the final image

First build takes ~5 minutes (dependency download + CGO compile). Subsequent builds use the layer cache.

> **To build and extract the binary without Docker runtime** (e.g. for WSL2):
> ```powershell
> docker build -f Dockerfile.build --output bin .
> # Binary is now at bin/swarm (Linux ELF)
> ```

---

## 3. Seed a project

The swarm reads work from `swarm.db`. The operator inserts a seed task directly.

```powershell
# Create swarm.db if it doesn't exist yet
duckdb  # (swarm.db is SQLite, not DuckDB — use sqlite3 CLI or the helper below)
```

Use the provided seed helper (or insert manually):

```powershell
# Seed via SQLite (DuckDB CLI cannot open SQLite files)
# Install sqlite3: winget install SQLite.SQLite
sqlite3 swarm.db "
  INSERT INTO projects (id, name, output_dir)
  VALUES ('proj-001', 'My Project', '/data/output/proj-001');

  INSERT INTO tasks (id, project_id, type, model_tag, payload, status, depth)
  VALUES (
    lower(hex(randomblob(16))),
    'proj-001',
    'DECOMPOSE',
    'phi-4:mini',
    'Research and build a Go CLI tool that converts XBRL financial reports to CSV.',
    'pending',
    0
  );
"
```

---

## 4. Run the swarm

```powershell
docker compose up
```

The container:
- Reads `config.yaml`, `swarm.db`, `skills/` from the current directory (volume-mounted)
- Talks to Ollama at `http://host.docker.internal:11434` (your Windows host)
- Writes artifacts to `./output/`
- Appends events to `obs.db`

Stop with `Ctrl+C`. The swarm is restartable — it picks up where it left off on the next `docker compose up`.

---

## 5. Monitor activity

### Live logs

```powershell
docker compose logs -f
```

Every task state transition is logged in structured format:

```
time=2026-04-24T10:00:00Z level=INFO msg=event event_type=task_transition task_id=abc123 task_type=RESEARCH from_status=pending to_status=active model_tag=llama3.1:8b depth=1
```

### Query obs.db with DuckDB CLI

`obs.db` is on your local filesystem (volume-mounted). Query it directly with the DuckDB CLI (already installed):

```powershell
# All events for a project's tasks
duckdb obs.db "SELECT timestamp, event_type, task_type, from_status, to_status FROM events ORDER BY timestamp"

# VRAM swap history
duckdb obs.db "SELECT timestamp, detail FROM events WHERE event_type = 'vram_swap'"

# Cascading prune events with score and count
duckdb obs.db "SELECT timestamp, detail FROM events WHERE event_type = 'cascading_prune'"

# Task timeline for a specific task
duckdb obs.db "SELECT timestamp, event_type, from_status, to_status FROM events WHERE task_id = '<uuid>'"

# Self-correction attempts
duckdb obs.db "SELECT timestamp, task_id, detail FROM events WHERE event_type = 'self_correction' ORDER BY timestamp"

# Count events by type
duckdb obs.db "SELECT event_type, count(*) AS n FROM events GROUP BY event_type ORDER BY n DESC"
```

> **Note:** `swarm.db` is SQLite (app state). `obs.db` is DuckDB (event log). Use `sqlite3 swarm.db` for the former, `duckdb obs.db` for the latter.

### Query swarm.db with sqlite3

```powershell
# Task status summary
sqlite3 swarm.db "SELECT type, status, count(*) FROM tasks GROUP BY type, status"

# All tasks for a project
sqlite3 swarm.db "SELECT id, type, status, depth FROM tasks WHERE project_id = 'proj-001'"
```

---

## 6. Retrieve outputs

When a project reaches terminal state (`completed` or `pruned`), artifacts are written to:

```
output/
└── proj-001/
    ├── research-summary.md
    ├── viability-report.md
    ├── main.go              (if GO_CODE tasks completed)
    └── prune-report.md      (if all branches were pruned)
```

---

## 7. Human on the Loop — Obsidian annotations

If Obsidian is running with the Local REST API plugin:

- **RESEARCH** tasks write findings to `ALS/runs/<date>/` in your vault
- **VIABILITY_REVIEW** tasks write score + rationale notes
- You can annotate any note; on the next `DECOMPOSE` the swarm reads those annotations

No restart required — annotations are picked up on the next task dispatch.

---

## 8. Troubleshooting

| Symptom | Likely cause | Fix |
|---|---|---|
| `ollama: health check failed` | Ollama not running | Start Ollama on the host |
| `ollama: HTTP 404` | Model not pulled | `ollama pull <model-tag>` |
| `mcp: server registration failed` | npm package not found or node error | Check stderr in `docker compose logs` |
| Tasks stuck in `active` forever | Ollama inference hung | Restart with `docker compose restart`; task will be re-picked up |
| `obs.db` write failed in logs | DuckDB issue | Check available disk space |
| All branches pruned immediately | Viability threshold too high | Lower `viability_threshold` in `config.yaml` |
