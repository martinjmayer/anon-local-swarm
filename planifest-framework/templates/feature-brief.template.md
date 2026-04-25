---
title: "Feature Brief - {{feature-name}}"
summary: "The business case, scope, and product requirements for the feature."
status: "draft | approved"
version: "0.1.0"
---
# Feature Brief - {{feature-name}}

**Feature ID:** {{0000000}}-{{kebab-case-name}}

> Written by a human. This is the input document that kicks off the confirmed design Agentic Iteration Loop. The orchestrator reads this and coaches you through any gaps before passing it to the spec-agent.
>
> **Feature ID format:** A 7-digit zero-padded number followed by a kebab-case name (e.g., `0000001-user-auth`, `0000002-payment-gateway`). The numeric prefix keeps features in chronological order when sorted alphabetically. Choose the next available number.

---

## Business Goal

What problem does this feature solve? Who benefits and how?

> Write 2-3 sentences. Be specific. "Improve performance" is not a goal - "Reduce checkout latency from 3s to under 500ms because 40% of users abandon at payment" is.

{{business-goal}}

---

## Features

Break the feature into discrete features. Each feature should be small enough that an agent can implement it within a single session (roughly: one API endpoint with its data model, validation, tests, and docs - or one UI screen with its state management and tests).

> **Rule of thumb:** If a feature has more than 3 user stories, it's too big. Split it further. Big features should have many small features, not a few large ones.

| Feature | User Stories | Priority | Phase |
|---------|-------------|----------|-------|
| {{feature-name}} | {{story-1}}, {{story-2}} | must-have / should-have / could-have | 1 |
| {{feature-name}} | {{story-1}}, {{story-2}} | must-have / should-have / could-have | 1 |
| {{feature-name}} | {{story-1}} | must-have / should-have / could-have | 2 |

---

## Phases

If the feature has more than 5-6 features, split it into phases. Each phase becomes a separate iteration of the Agentic Iteration Loop with its own execution plan, ADRs, and codegen pass. Earlier phases ship before later phases begin.

> **Why phases matter:** An agent working on phase 2 only needs the context from phase 2's brief plus the component manifests from phase 1. It doesn't need to hold the entire feature in context. This is how Planifest manages context at scale.

| Phase | Features Included | Ships When |
|-------|-------------------|------------|
| 1 | {{feature-list}} | {{criteria}} |
| 2 | {{feature-list}} | {{criteria}} |

---

## Target Architecture

What architectural decisions have you already made? The agent implements within these constraints - it does not choose the architecture.

### Components

What components does this feature create or modify?

| Component | Type | New or Existing | Responsibility |
|-----------|------|-----------------|---------------|
| {{component-name}} | microservice / microfrontend / component-pack | new / existing | {{what it does}} |

### Data Ownership

Which components own which data? Each data store is owned by exactly one component.

| Data Store | Owner Component | Shared With |
|------------|----------------|-------------|
| {{store}} | {{component}} | {{consumers - read only}} |

### Integration Points

How do components communicate?

| From | To | Method | Contract |
|------|-----|--------|----------|
| {{component}} | {{component}} | REST / gRPC / event / queue | {{brief description}} |

---

## Stack

What technology stack has been decided? The agent builds with this - it does not choose a different stack.

> If you haven't decided yet, leave this section empty. The orchestrator will ask you to fill it before the pipeline proceeds.

| Concern | Decision |
|---------|----------|
| Language | {{e.g. TypeScript, Go, Python}} |
| Runtime | {{e.g. Node, Deno, Bun}} |
| Framework | {{e.g. Fastify, Hono, Gin}} |
| Frontend | {{e.g. React 19, Vue, none}} |
| Database | {{e.g. PostgreSQL, none}} |
| ORM | {{e.g. Drizzle, Prisma, none}} |
| Testing | {{e.g. Vitest, Jest}} |
| IaC | {{e.g. Pulumi, Terraform, none}} |
| Cloud | {{e.g. GCP, AWS, none}} |
| Compute | {{e.g. Cloud Run, Lambda, k8s}} |
| CI | {{e.g. GitHub Actions, GitLab CI}} |

---

## Scope Boundaries

What is explicitly in scope, out of scope, and deferred?

### In Scope
- {{what this feature will deliver}}

### Out of Scope
- {{what this feature will NOT deliver - be explicit}}

### Deferred
- {{what might be delivered later but not now}}
- {{note what is blocked until deferred items are resolved}}

---

## Non-Functional Requirements

Specific, measurable targets. If you don't have a target, leave it blank - the spec-agent will ask.

| NFR | Target | Measurement |
|-----|--------|-------------|
| Latency | {{e.g. p95 < 200ms}} | {{how measured}} |
| Availability | {{e.g. 99.9%}} | {{SLO definition}} |
| Throughput | {{e.g. 1000 req/s}} | {{peak or sustained}} |
| Security | {{e.g. SOC2 compliance}} | {{certification or audit}} |

---

## Constraints and Assumptions

Anything the agents need to know that doesn't fit elsewhere.

### Constraints
- {{hard constraints - regulatory, contractual, technical}}

### Assumptions
- {{assumptions you've made - agents will flag if these conflict with the spec}}

---

## Acceptance Criteria

How do you know this feature is done?

- [ ] {{criterion 1}}
- [ ] {{criterion 2}}
- [ ] {{criterion 3}}

---

*This brief will be read by the orchestrator skill. See [planifest/skills/orchestrator/SKILL.md](../skills/orchestrator/SKILL.md)*

