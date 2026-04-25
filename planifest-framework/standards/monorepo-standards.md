# Planifest Monorepo Standards

> When multiple components live in the same repository, clear boundaries prevent coupling. These standards ensure agent-generated code respects component isolation within a monorepo.

---

## 1. Directory Structure

```
src/
├── {component-a}/
│   ├── component.yml
│   ├── package.json (or equivalent)
│   ├── src/
│   ├── tests/
│   └── docs/
├── {component-b}/
│   ├── component.yml
│   ├── ...
└── shared/
    ├── types/          - shared type definitions
    ├── utils/          - shared utilities (minimal)
    └── contracts/      - cross-component interface definitions
```

---

## 2. Dependency Rules

- **Components never import from each other's `src/` directory.** Cross-component communication goes through defined interfaces (APIs, events, shared types).
- **Shared code lives in `src/shared/`.** It must be genuinely shared - used by 2+ components. Do not preemptively create shared modules.
- **Each component has its own dependency manifest** (`package.json`, `go.mod`, etc.). Dependency versions may differ between components.
- **Workspace tools** (npm workspaces, pnpm workspaces, Go workspaces) manage the monorepo. Each component is a workspace member.

---

## 3. Build and Test Isolation

- Each component builds independently
- Each component's tests run independently
- CI runs only the affected component's checks when changes are scoped to one component
- A full build/test run is required when shared code changes

---

## 4. Component Boundaries

- Data ownership is per-component - one component, one database/schema
- API contracts are defined in OpenAPI specs, not ad-hoc imports
- Event contracts are defined in schemas (JSON Schema, Protobuf, Avro)
- If two components need the same data, one owns it and exposes an API - the other consumes it

---

## 5. Versioning

- Each component has its own version in `component.yml`
- Shared packages follow semver independently
- Breaking changes to shared packages require updating all consumers in the same PR

---

*Referenced by codegen-agent and orchestrator. Source of truth: `planifest-framework/standards/monorepo-standards.md`*
