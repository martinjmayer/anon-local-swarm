# Component Manifest - Template Guide

> This is the canonical template for `component.yml` - the manifest file that describes a built component. Agents write it after building; agents and humans read it before changing anything.

*Related: [Shared Schema Definitions](../schemas/shared.schema.json)*

---

## When is it created?

The `component.yml` is created at **two points** in the lifecycle:

1. **Pre-seeded by the human** (or orchestrator) before the pipeline runs - with stack decisions, feature mode, and domain knowledge path. Purpose, contract, and data sections are left empty or minimal. This gives the codegen-agent its constraints.

2. **Completed by the codegen-agent** after the implementation is built - the agent fills in purpose, contract, data ownership, risk, scope, and quality sections based on what was actually implemented.

From that point forward, every change-agent reads it before modifying the component.

---

## File location

```
src/{{component-id}}/component.yml
```

---

## Field Reference

### Identity

| Field | Required | Written by | Description |
|---|---|---|---|
| `id` | Yes | Human / Orchestrator | Kebab-case identifier. Must be unique across the feature. |
| `name` | Yes | Human / Orchestrator | Human-readable name. |
| `feature` | Yes | Human / Orchestrator | The feature this component belongs to. |
| `version` | Yes | Agent | Semver. Incremented by the agent on each meaningful change. |
| `status` | Yes | Agent | `planned` -> `active` -> `deprecated`. Only the orchestrator or a human sets `deprecated`. |

### Purpose

The most important section. This is what prevents agents from building overlapping components or modifying the wrong thing.

| Field | Required | Written by | Description |
|---|---|---|---|
| `purpose.summary` | Yes | Agent | One sentence. What this component exists to do. If you need "and", it's doing too much. |
| `purpose.systemContext` | Yes | Agent | Where it sits in the wider system. What calls it, what it calls. |
| `purpose.type` | Yes | Human / Agent | `microservice`, `microfrontend`, or `component-pack`. |
| `purpose.domain` | Yes | Human / Agent | The business domain this component belongs to (from the domain glossary). |
| `purpose.responsibilities` | Yes | Agent | Array of specific things this component does. |
| `purpose.notResponsibleFor` | Yes | Agent | Array of explicit exclusions. **Equally important as responsibilities.** An agent reading this knows what NOT to add to this component. |

### Stack

Pre-seeded by the human or orchestrator. The codegen-agent reads this - it does not choose the stack.

| Field | Required | Description |
|---|---|---|
| `stack.language` | Yes | Primary language. |
| `stack.runtime` | Yes | Runtime environment. |
| `stack.framework` | Yes | Web framework (or `none`). |
| `stack.frontend` | Conditional | Frontend framework (or `none` for backend-only). |
| `stack.database` | Conditional | Database engine (or `none` for stateless components). |
| `stack.orm` | Conditional | ORM or query builder. |
| `stack.testing` | Yes | Test framework. |
| `stack.iac` | Yes | Infrastructure as code tool. |
| `stack.cloud` | Yes | Target cloud provider. |
| `stack.compute` | Yes | Compute platform. |
| `stack.styling` | Conditional | CSS approach for frontend components (or `none`). |
| `stack.componentLibrary` | Conditional | UI component library (or `none`). |
| `stack.ci` | Yes | CI/CD platform. |

### Contract

Describes the component's interface - what it accepts, what it produces, who depends on it.

| Field | Required | Written by | Description |
|---|---|---|---|
| `contract.apiSpec` | Yes | Agent | Path to the OpenAPI spec (from project root). The spec-agent writes this to `plan/{feature-id}/openapi-spec.yaml` (if the component serves an API). |
| `contract.inputs` | Yes | Agent | Array of inputs (HTTP endpoints, events, queues) the component accepts. |
| `contract.outputs` | Yes | Agent | Array of outputs the component produces. |
| `contract.consumedBy` | Yes | Agent | Array of component IDs that depend on this component. Updated as the system grows. |
| `contract.consumes` | Yes | Agent | Array of component IDs this component depends on. |
| `contract.breakingChangePolicy` | Yes | Human / Agent | What happens when the contract changes: `requires-adr`, `requires-human-approval`, or `semver-major`. |

### Data

Components that own data declare it here. This is what the `change-agent` reads before proposing any schema migration.

| Field | Required | Written by | Description |
|---|---|---|---|
| `data.ownsData` | Yes | Agent | Whether this component owns a data store. |
| `data.dataContract` | Conditional | Agent | Path to the data contract document. Required if `ownsData` is true. |
| `data.tables` | Conditional | Agent | Array of table names owned by this component. |
| `data.schemaVersion` | Conditional | Agent | Current schema version. Incremented on every migration. |
| `data.migrationPath` | Conditional | Agent | Path to the migrations directory. |

### Risk

Populated by the security-agent and updated by any agent that identifies a risk.

| Field | Required | Written by | Description |
|---|---|---|---|
| `risk.level` | Yes | Agent | Overall risk level: `low`, `medium`, `high`, `critical`. |
| `risk.items` | Yes | Agent | Array of specific risk items (see p007 `RiskItem` schema). |

### Scope

Mirrors the feature brief's scope boundaries, narrowed to this component.

| Field | Required | Written by | Description |
|---|---|---|---|
| `scope.in` | Yes | Human / Agent | What is in scope for this component. |
| `scope.out` | Yes | Human / Agent | What is explicitly out of scope. |
| `scope.deferred` | No | Human / Agent | What is deferred to a future feature. |

### Quality

Living record of the component's quality posture. Updated by the validate-agent.

| Field | Required | Written by | Description |
|---|---|---|---|
| `quality.testCoverage` | Yes | Agent | Coverage percentages for unit, integration, and e2e tests. |
| `quality.techDebt` | No | Agent | Array of known tech debt items (strings or structured objects). |
| `quality.quirks` | No | Agent | Array of known quirks or workarounds (see p007 `Quirk` schema). |

### Pipeline

Operational metadata used by the CI/CD pipeline and template stamping system.

| Field | Required | Written by | Description |
|---|---|---|---|
| `pipeline.templateVersion` | Yes | Agent / Orchestrator | Which pipeline template version this component is on. |
| `pipeline.featureMode` | Yes | Human | `greenfield`, `retrofit`, or `agent-interface`. |
| `pipeline.domainKnowledgePath` | Yes | Agent | Path to the component's domain knowledge docs directory. |

### Metadata

| Field | Required | Written by | Description |
|---|---|---|---|
| `metadata.createdAt` | Yes | Agent | ISO 8601 timestamp. |
| `metadata.updatedAt` | Yes | Agent | ISO 8601 timestamp. Updated on every change. |
| `metadata.createdBy` | Yes | System | `agent` or `human`. |
| `metadata.lastModifiedBy` | Yes | System | `agent` or `human`. |
| `metadata.skill` | Yes | Agent | Which skill produced or last modified this manifest (e.g., `codegen-agent`, `change-agent`). |
| `metadata.tool` | Yes | Agent | Which agentic tool was used (e.g., `claude-code`, `cursor`, `antigravity`). |
| `metadata.model` | Yes | Agent | Which model produced the output (e.g., `claude-sonnet-4-20250514`). |

---

## Rules for Agents

1. **Never invent stack fields.** The stack section is pre-seeded by the human or orchestrator. If it's empty, ask - don't guess.
2. **Always populate `notResponsibleFor`.** This is the single most important field for preventing scope creep across components.
3. **Leave `consumedBy` empty on first write.** You don't know who will consume this component. It gets populated as the system grows.
4. **Update `version` on every meaningful change.** Patch for fixes, minor for new capabilities, major for contract changes.
5. **Update `metadata.updatedAt` on every write.** This is how humans and other agents know when the manifest was last touched.
6. **Match field values to the domain glossary.** If the glossary says "Order", the manifest says "Order" - not "Purchase" or "Transaction".

---

*Template: [component.template.yml](component.template.yml)*

