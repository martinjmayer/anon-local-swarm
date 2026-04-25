---
name: retrofit
description: Onboard an existing codebase into the confirmed design framework - reads the codebase, infers architecture, and reconciles with a feature brief. Use this when adding Planifest to a project that already has code.
---

# Retrofit

Onboard an existing codebase into the confirmed design framework.

## Prerequisites

- Existing code in `src/`
- A feature brief at `plan/current/feature-brief.md` describing the intended changes
- Component manifests (`component.yml`) in each existing component - use the [component manifest template](../templates/component.template.yml) and [guide](../templates/component-guide.md)

## Steps

1. **Load the orchestrator skill**
2. **Read the codebase** - the orchestrator reads `src/` to:
   - Identify existing components
   - Infer the architecture (stack, patterns, decisions)
   - Surface tech debt, undocumented decisions, naming conventions
3. **Reconcile** - compare the feature brief against the discovered reality:
   - What already exists that the brief describes?
   - What conflicts between the brief and the codebase?
   - What gaps remain?
4. **Coach** - Phase 0 coaching, but informed by codebase reality:
   - The human may need fewer questions (codebase already answers them)
   - Or more questions (codebase reveals conflicts)
5. **Proceed** - once the Design is confirmed, execute the pipeline as normal (Phases 1-6)

## Notes

- The spec-agent also operates in retrofit mode - it reads the codebase before producing artifacts
- Existing architecture decisions should be captured as ADRs, not re-decided
- The adoption mode is recorded in the confirmed design: `adoption_mode: retrofit`

