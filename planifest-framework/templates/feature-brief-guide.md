# Feature Brief - Guide

> How to write a Feature Brief that gives the orchestrator everything it needs to coach you and build your system.

*Related: [Orchestrator Skill](../skills/planifest-orchestrator/SKILL.md) | [Feature Structure](../../plan/feature-structure.md)*

---

## Purpose

The Feature Brief is the **only human-authored input** to the confirmed design Agentic Iteration Loop. Everything the agent builds - execution plans, ADRs, code, tests, docs - derives from this document. If it's vague, everything downstream is vague. If it's specific, the agent has a fighting chance.

You are the Product Owner and Technical Architect. This brief is where you exercise both roles.

---

## When to Write It

- **Before** engaging the orchestrator
- **One per feature** - if you have multiple features, write multiple briefs
- **Updated** when the scope or architecture changes materially (the orchestrator re-reads it)

---

## Section-by-Section Guidance

### Business Goal

Write 2-3 sentences. Be specific about the problem, who has it, and how you'll know it's solved.

| Ã¢ÂÅ’ Bad | Ã¢Å“â€¦ Good |
|--------|---------|
| "Improve performance" | "Reduce checkout latency from 3s to under 500ms - 40% of users abandon at payment" |
| "Add user management" | "Enable self-service account creation so customer support stops onboarding 200 users/week manually" |

### Features

Each feature should be **small enough for one agent session**. Think:
- One API resource with its data model, validation, tests, and docs
- One UI screen with its state management, data fetching, and tests
- One integration with its adapter, contract, and error handling

**Rule of thumb:** If a feature has more than 3 user stories, split it.

The Priority column uses MoSCoW: must-have, should-have, could-have. The orchestrator builds must-haves first.

### Phases

Only needed if you have more than 5-6 features. Phases are sequential iterations of the Agentic Iteration Loop:
- Phase 1 ships before Phase 2 begins
- Phase 2's agent reads Phase 1's component manifests for context
- This is how Planifest manages context at scale

If you don't need phases, delete this section.

### Target Architecture

This is where you make the architectural decisions. The agent implements within these constraints - it does not choose the architecture.

**Components:** Name them, type them, say what each one does. If modifying existing components, say so.

**Data Ownership:** Each data store is owned by exactly one component. This is a hard limit. If two components need the same data, one owns it and the other reads it via an API or event.

**Integration Points:** How components talk to each other. REST, gRPC, events, queues - be explicit.

### Stack

Every technology choice must be stated. The agent builds with this - it does not choose differently. If you haven't decided, leave it blank. The orchestrator will ask.

The [Backend Stack Evaluation](../standards/backend-stack-evaluation.md) and [Frontend Stack Evaluation](../standards/frontend-stack-evaluation.md) can help you decide.

### Scope Boundaries

**In Scope** - what this feature delivers. Be specific.

**Out of Scope** - what it does NOT deliver. This is equally important. Without it, the agent may build features you didn't ask for. Explicit exclusions save cycles.

**Deferred** - what might be delivered later but not now. Note what's blocked until deferred items are resolved.

### Non-Functional Requirements

Every target must be **measurable**. "Fast" is not a requirement. "p95 < 200ms" is.

If you don't have a target, leave it blank. The orchestrator will ask. Better to have no target than a fake one.

### Acceptance Criteria

How you know the feature is done. These become the orchestrator's exit criteria. Be specific - the agent will test against them.

---

## Common Mistakes

1. **Too big.** The most common mistake. If the brief describes more than one system, it's too big. Split it.
2. **Vague stories.** "As a user, I want to manage my account" - manage how? What operations? What data?
3. **Missing stack.** The codegen-agent cannot begin without a stack declaration. Decide before you brief.
4. **No scope boundaries.** Without "Out of Scope", the agent has no guardrails. It may build things you didn't want.
5. **Assumed NFRs.** If you need p95 < 200ms, say so. The agent won't guess your latency requirements.

---

## What Happens Next

1. You write the brief and save it to `plan/current/feature-brief.md`
2. You tell the orchestrator to load it
3. The orchestrator assesses it against the three layers:
   - **Product**: Functional Requirements. What the system must do and why.
   - **Architecture**: Standards. The cross-cutting rules and non-functional requirements.
   - **Engineering**: Implementation. How the system was actually built.
   The orchestrator coaches you through any gaps - one question at a time.
4. Once complete, the orchestrator produces the **confirmed design** and begins the Agentic Iteration Loop

---

*Template: [feature-brief.template.md](feature-brief.template.md)*

