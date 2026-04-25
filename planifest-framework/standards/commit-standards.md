# Commit Message Standards

> Planifest commit standard — applies to all commits on Planifest-managed projects.

---

## Format

```
type(scope): short description
```

- **type** — one of: `feat` `fix` `docs` `chore` `refactor` `test` `perf`
- **scope** — feature ID (e.g. `0000003`) or component ID (e.g. `auth-service`)
- **short description** — imperative mood, present tense, ≤72 characters total subject line

### Examples

```
feat(0000003): add gate-write enforcement hook
fix(auth-service): correct token expiry calculation
docs(0000003): update ADR-003 with flag-file rationale
chore: update .gitattributes for LF enforcement
```

---

## Rules

### 1. Subject line ≤72 characters

The entire subject line (`type(scope): description`) must not exceed 72 characters.

### 2. Imperative mood, present tense

Write as a command: "add X", "fix Y", "remove Z" — not "added X", "fixed Y", "removes Z".

### 3. No AI attribution

Commit messages MUST NOT attribute authorship or co-development to any AI tool or LLM. Prohibited:

- `Co-Authored-By: Claude <...>`
- `co-developed with Claude / Copilot / Cursor / Windsurf / Cline`
- `AI-assisted`, `LLM-generated`, `with the help of [tool name]`
- Any model name (Claude, GPT, Gemini, etc.) in an authorship context

The commit is owned by the **human practitioner**. The AI tool is an instrument, not a contributor of record.

### 4. No affirmatory or confirmatory language

Messages are objective and scope-focused. Prohibited:

- `Done!`, `Fixed!`, `Working now`, `All done`, `Finally works`
- `Claude helped with this`, `Looks good now`
- Phrases that confirm rather than describe

### 5. No contradictory messaging

Do not describe a change and then reverse it in the same message. One atomic commit, one clear description.

### 6. Body (optional)

A blank line after the subject may be followed by a body explaining *why*, not *what*. The subject line already states what.

---

## Advisory hook

`planifest-framework/hooks/commit-msg` checks messages against these rules and prints a warning on violation. It never blocks the commit (ADR-008).

To install: run `./planifest-framework/setup.sh <tool>` — the hook is wired automatically via `git config core.hooksPath`.
