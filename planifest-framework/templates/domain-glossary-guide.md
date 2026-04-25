# Domain Glossary - Guide

> How to maintain the ubiquitous language that keeps agents and humans aligned.

*Related: [Spec Agent Skill](../skills/spec-agent-SKILL.md) | [Code Quality Standards](../standards/code-quality-standards.md)*

---

## Purpose

The Domain Glossary defines the ubiquitous language for a feature. Every term has one meaning. Agents and humans use these terms - in code, comments, file names, variable names, and documentation. If the glossary says "Order", the code says `Order` - not "Purchase", "Transaction", or "Booking".

This is not optional. Domain confusion is one of the most common sources of bugs in agent-generated code. An agent that calls the same concept by three different names will produce inconsistent APIs, mismatched types, and confusing docs.

---

## Who Writes It

- **spec-agent** creates it during Phase 1, extracting terms from the Feature Brief and Design Requirements
- **Any agent** that introduces a new domain term must add it to the glossary before using it in code

---

## What Goes In

### Include

- Business concepts: Order, Customer, Subscription, Invoice
- Technical concepts specific to this domain: Tenant, Workspace, Pipeline Run
- Status values: draft, active, suspended, cancelled (with precise definitions of transitions)
- Roles: admin, member, viewer (with precise definitions of permissions)

### Don't Include

- Generic programming terms: function, class, variable
- Framework-specific terms: component, hook, middleware (unless they have domain-specific meaning)
- Obvious English words that aren't used as domain concepts

---

## The Aliases Column

If a term has known alternatives that people might use, list them as aliases. This helps agents recognise when code uses the wrong term.

| Term | Aliases |
|------|---------|
| Order | Purchase, Transaction |
| Customer | User, Client, Account Holder |

The agent should use the **Term**, never the alias. If existing code uses an alias, flag it for renaming.

---

## The Used In Column

Track which components use each term. This helps the change-agent understand ripple effects - if "Order" changes meaning, every component that uses it needs updating.

---

## Common Mistakes

1. **Synonyms in code.** The glossary says "Customer" but the code has `User` in one service and `Client` in another. Each is a bug.
2. **Missing status definitions.** The glossary lists "active" as a status but doesn't define when something becomes active or what it means.
3. **No glossary at all.** The agent invents terms. Each service uses different words for the same concept. Integration becomes a translation exercise.

---

## File Location

`plan/current/domain-glossary.md`

---

*Template: [domain-glossary.template.md](domain-glossary.template.md)*

