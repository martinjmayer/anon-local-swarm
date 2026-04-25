# Risk Register - Guide

> How risks are identified, assessed, and tracked throughout a feature.

*Related: [Spec Agent Skill](../skills/spec-agent-SKILL.md) | [Scope Guide](scope-guide.md)*

---

## Purpose

The Risk Register is a living document. The spec-agent seeds it during Phase 1. Any agent that identifies a new risk during any phase adds to it. It tracks specific, actionable risks - not generic hand-wringing.

---

## Who Writes It

- **spec-agent** creates it during Phase 1
- **Any agent** can add risks during their phase (codegen-agent discovers a library limitation, security-agent finds a vulnerability pattern, etc.)
- The `lastModifiedBy` fields in the header track who last updated it

---

## What Makes a Good Risk

A good risk entry is:
- **Specific to this feature** - not generic ("data loss is bad")
- **Actionable** - has a mitigation plan, not just a description
- **Assessed** - likelihood and impact rated, not left blank

| Ã¢ÂÅ’ Bad | Ã¢Å“â€¦ Good |
|--------|---------|
| "Security risk" | "R-003: The auth-service stores refresh tokens in a PostgreSQL table without encryption at rest. Likelihood: medium. Impact: high. Mitigation: enable TDE on the token table and rotate tokens on every use." |
| "Performance might be an issue" | "R-005: The search endpoint joins 4 tables without a covering index. Likelihood: high at >1000 records. Impact: medium (latency > SLO). Mitigation: add composite index on (tenant_id, created_at, status)." |

---

## Assumptions as Risks

Assumptions from the Design Requirements are automatically logged as risk items with `likelihood: medium`. This ensures they're tracked - if an assumption turns out to be wrong, the risk register already has an entry.

---

## Risk Levels

| Level | Meaning |
|-------|---------|
| **low** | Unlikely and/or low impact. Monitor only. |
| **medium** | Possible and material. Mitigation plan required. |
| **high** | Likely or high-impact. Mitigation must be implemented before ship. |
| **critical** | Near-certain or catastrophic. Blocks the pipeline - escalate to human. |

The **Overall Risk Level** in the header is the highest individual risk level in the register.

---

## Status Values

| Status | Meaning |
|--------|---------|
| **open** | Risk identified, mitigation planned but not yet implemented |
| **mitigated** | Mitigation implemented and verified |
| **accepted** | Human explicitly accepted this risk without mitigation |

---

## File Location

`plan/current/risk-register.md`

---

*Template: [risk-register.template.md](risk-register.template.md)*

