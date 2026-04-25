# Getting Started with Planifest

> Step-by-step instructions for humans setting up a Planifest project.

---

## Prerequisites

- An agentic coding tool: Claude Code, Cursor, Codex, Antigravity, GitHub Copilot, Windsurf, Cline, Roo Code, or OpenCode
- A terminal with Bash (macOS/Linux) or PowerShell (Windows)

---

## New Project

### 1. Add the framework

Copy the `planifest-framework/` folder into your repository root. This is the only thing you need - it contains the skills, templates, standards, and setup scripts.

### 2. Create the project structure

```
mkdir plan plan/changelog src docs
```

These are the core working directories:
- `plan/` - The current change being planned.
  - `plan/current/design.md` - Confirmed design and build plan.
  - `plan/current/feature-brief.md` - The initiating human-authored brief.
  - `plan/current/iteration-log.md` - Audit trail of the pipeline run.
  - `plan/archive/` - Historical plans filed here after merge.
  - `plan/changelog/` - A record of all changes ({feature-id}-{YYYY-MM-DD}.md).
- `src/` - Component source code, tests, and component manifests (`component.yml`).
- `docs/` - Living repository documentation (always current). Includes component registry and dependency graph.

See [feature-structure.md](../plan/feature-structure.md) for the full layout.

### 3. Run the setup script

This copies skills into the directory your agentic tool expects.

#### Basic setup

```bash
# macOS / Linux
chmod +x planifest-framework/setup.sh
./planifest-framework/setup.sh claude-code      # or cursor, codex, antigravity, copilot, windsurf, cline, roo-code, opencode, all
```

```powershell
# Windows (PowerShell)
.\planifest-framework\setup.ps1 claude-code     # or cursor, codex, antigravity, copilot, windsurf, cline, roo-code, opencode, all
```

Installs:
- Skill folders with YAML frontmatter (auto-discovered by your tool)
- Supporting files (templates, standards, schemas)
- A boot file for your tool (e.g. `CLAUDE.md`, `AGENTS.md`)
- Git guardrails (see below)

The agent uses native tools (`Grep`, `Bash`, `WebFetch`) directly. No context window protection.

#### Option: Context-Mode (recommended)

Follow the guidance above around the tool, then consider the option to use the Context Mode MCP service.

[context-mode](https://github.com/mksglu/context-mode) routes large output — search results, file analysis, web fetches — into a sandboxed knowledge base. Only summaries enter the context window, so the agent stays fast and focused on large codebases.

Install context-mode first, then pass `--context-mode-mcp` during setup, after the tool selection argument:

```bash
# macOS / Linux
chmod +x planifest-framework/setup.sh
./planifest-framework/setup.sh claude-code --context-mode-mcp
```

```powershell
# Windows (PowerShell)
.\planifest-framework\setup.ps1 claude-code --context-mode-mcp
```

Installs everything above, plus routing rules (`AGENTS.md`) and (for Claude Code) enforcement hooks that physically block native tool use (`Grep`, `Bash` web/grep patterns, `WebFetch`) to prevent the agent from bypassing context-mode.

See [docs/context-mode.md](../docs/context-mode.md) for how it works and prerequisites.

#### Option: Structured Telemetry

Emit structured events from skills and hooks into a local telemetry backend for observability — pipeline phase timings, context-pressure alerts, and skill execution traces.

Requires the [structured-telemetry-mcp](https://github.com/anthropics/structured-telemetry-mcp) server to be running, then pass `--structured-telemetry-mcp` during setup:

```bash
# macOS / Linux
./planifest-framework/setup.sh claude-code --structured-telemetry-mcp
```

```powershell
# Windows (PowerShell)
.\planifest-framework\setup.ps1 claude-code --structured-telemetry-mcp
```

Installs everything in Basic setup, plus:
- `.claude/telemetry-enabled` sentinel — opt-in gate that authorises skills to emit events

When combined with `--context-mode-mcp`, also installs the `context-pressure.mjs` PostToolUse hook, which emits a telemetry event whenever context window pressure exceeds the configured threshold.

The default backend URL is `http://localhost:3741`. Override it with `--backend-url`:

```bash
./planifest-framework/setup.sh claude-code --structured-telemetry-mcp --backend-url http://myhost:4000
```

See [tool-setup-reference.md](tool-setup-reference.md) for what each tool expects.

### 3a. Git Guardrails (activated automatically)

The setup script also activates Planifest's **Progressive Guardrail System** - a three-tier enforcement model that protects `main` without blocking atomic commits:

| Tier | When | What happens |
|------|------|--------------|
| **1 - Advisory pre-commit** | Every local commit | Prints a warning if code was staged without docs. Commit **succeeds**. |
| **2 - Branch pre-push** | Every `git push` | Checks the *cumulative branch diff*. Push **fails** if `src/` was changed with no updates to `plan/`, `docs/`, or `component.yml` - **unless** all commits use the `fix(fast-path):` prefix, in which case only `component.yml` or `plan/changelog/` is required. |
| **3 - CI/CD pipeline** | Every Pull Request | Same check in GitHub Actions. Recognises the `fix(fast-path):` prefix and applies the same relaxed rule. Blocks the merge button if the rule is violated. |

The hooks live in `planifest-framework/hooks/` and are wired via `git config core.hooksPath` - no `.git/` modifications required.

The CI workflow is copied to `.github/workflows/planifest.yml` on first setup.

### 4. Write your first feature brief

Use the template:
```
cp planifest-framework/templates/feature-brief.template.md plan/current/feature-brief.md
```

Fill it in. The [feature brief guide](templates/feature-brief-guide.md) walks you through each section.

### 4a. Understanding phase indicators (REQ-018)

Every agent response begins with a phase prefix so you always know where you are in the pipeline:

| Prefix | Phase | What the agent is doing |
|--------|-------|-------------------------|
| `P0:` | Assess & Coach | Reviewing the brief; asking gap questions |
| `P1:` | Spec | Writing requirements, scope, glossary, risk register |
| `P2:` | ADRs | Documenting architecture decisions |
| `P3:` | Codegen | Generating implementation |
| `P4:` | Validate | Running CI checks; self-correcting |
| `P5:` | Security | Security review |
| `P6:` | Docs | Documentation artifacts |
| `P7:` | Ship | PR, changelog, archive |
| `PC:` | Change | Change pipeline (existing feature modification) |

If you see `Px: Resuming…` at the start of a session, the orchestrator detected existing artefacts in `plan/current/` and is continuing where it left off — no P0 briefing.

### 5. Start the orchestrator

Open your agentic tool. The orchestrator skill is now auto-discovered. Tell it:

```
Execute the confirmed design Agentic Iteration Loop.
Feature brief: plan/current/feature-brief.md
```

The orchestrator will:
1. Assess your brief against the three layers:
   - **Product**: Functional Requirements. What the system must do and why.
   - **Architecture**: Standards. The cross-cutting rules and non-functional requirements.
   - **Engineering**: Implementation. How the system was actually built.
2. Coach you through any gaps - one question at a time
3. Produce the **confirmed design** at `plan/current/design.md` (**Design confirmed**)
4. Execute the **Agentic Iteration Loop** (Phases 1-6): Requirements → ADRs → Code → Validate → Security → Docs
   (Executing: planifest-spec-agent → planifest-adr-agent → planifest-codegen-agent → ...)

---

## Retrofit an Existing Project

1. Copy `planifest-framework/` into your repo root
2. Run the setup script for your tool
3. Add a `component.yml` manifest to each existing component in `src/` - use the [component manifest template](templates/component.template.yml) and [guide](templates/component-guide.md)
4. Tell the orchestrator to use **retrofit** adoption mode:

```
Execute the confirmed design Agentic Iteration Loop in retrofit mode.
Feature brief: plan/current/feature-brief.md
```

The orchestrator will read your codebase, infer the existing architecture, and reconcile the brief against reality.

---

## Trivial Fixes (Fast Path)

For trivial changes - styling tweaks, copy corrections, isolated pure-function bugs - the orchestrator can route to the Fast Path, bypassing the requirements and ADR overhead:

```
fix(fast-path): updated button colour to match brand guidelines
```

The orchestrator evaluates four criteria before allowing Fast Path:
1. No new external dependencies
2. No schema or data model changes
3. No changes to security, auth, or routing logic
4. Change is confined to UI styling, copy, or isolated pure-function bugs

If all criteria pass: implement → validate → update `component.yml` (patch bump) → log in `plan/changelog/`.
If any criterion fails: route to Change Pipeline instead.

Fast Path commits use the `fix(fast-path):` prefix - the pre-push hook and CI recognise this and only require `component.yml` or a changelog update, not full `plan/` or `docs/` changes.

---

## Making Changes

For modifications to an existing feature:

```
Execute the confirmed design Change Pipeline.
Feature ID: my-feature
Component ID: auth-service
Change request: Add refresh token rotation
```

The **planifest-change-agent** handles it - no need to re-run the full Feature Pipeline.

---

## Updating the Framework

After updating any files in `planifest-framework/` (skills, templates, standards):

```bash
# Re-run setup to sync changes to your tool's directory
./planifest-framework/setup.sh claude-code                                                        # macOS / Linux
./planifest-framework/setup.sh claude-code --context-mode-mcp                                    # include if context-mode is installed
./planifest-framework/setup.sh claude-code --context-mode-mcp --structured-telemetry-mcp         # include if both MCPs are installed
```


```powershell
.\planifest-framework\setup.ps1 claude-code                                                       # Windows (PowerShell)
.\planifest-framework\setup.ps1 claude-code --context-mode-mcp                                    # include if context-mode is installed
.\planifest-framework\setup.ps1 claude-code --context-mode-mcp --structured-telemetry-mcp         # include if both MCPs are installed#
```

The setup script overwrites the generated copies. The source of truth is always `planifest-framework/`.

---

## What to Commit

| Path | Commit? | Why |
|------|:-------:|-----|
| `planifest-framework/` | ✅ | Source of truth - shared with team |
| `planifest-framework/hooks/` | ✅ | Git hooks and CI workflow - applied by `setup.sh` / `setup.ps1` |
| `.github/workflows/planifest.yml` | ✅ | CI/CD strict gate - deployed by setup, must be committed to take effect |
| `plan/` | ✅ | Feature briefs, execution plans, ADRs, scope docs |
| `src/` | ✅ | Component code and manifests |
| `docs/` | ✅ | Repo-wide registry and dependency graph |
| `.claude/`, `.cursor/`, `.agents/`, `.gemini/`, `.github/skills/` | Optional | Generated copies - can be `.gitignore`d and regenerated |
| `CLAUDE.md`, `AGENTS.md` | Optional | Boot files - tool-specific |
| `.claude/telemetry-enabled` | Optional | Telemetry opt-in sentinel - commit if the whole team uses structured telemetry |

If your team all uses the same tool, commit the generated files. If different team members use different tools, `.gitignore` them and let each person run the setup script.

