# Planifest Deployment Standards

> Deployment is the last mile between code and value. A good deployment is boring - predictable, automated, and reversible. These standards ensure agent-generated deployment configuration meets production requirements.

---

## 1. Deployment Strategy

Use the strategy declared in the operational model:

| Strategy | When to Use | Rollback Speed |
|----------|-------------|---------------|
| **Rolling** | Default for most services | Medium (replace instances) |
| **Blue-Green** | Zero-downtime required, stateless services | Fast (switch traffic) |
| **Canary** | High-risk changes, large user base | Fast (route traffic back) |

---

## 2. Container Standards

If the stack uses containers:

- **Multi-stage builds** - separate build and runtime stages
- **Minimal base images** - use `alpine`, `distroless`, or `slim` variants
- **Non-root user** - never run as root in production
- **Health check** - `HEALTHCHECK` instruction in Dockerfile
- **No secrets in images** - use runtime injection via environment variables or secrets manager
- **Pin versions** - use exact image tags, never `latest`

---

## 3. CI/CD Pipeline

The CI pipeline must:

1. Lint and type-check
2. Run unit tests
3. Run integration tests
4. Build the artifact (container, binary, bundle)
5. Run security scan (dependency audit)
6. Deploy to staging
7. Run smoke tests against staging
8. Deploy to production (with human approval gate)

---

## 4. Environment Configuration

- All configuration is injected via environment variables
- Secrets are stored in a secrets manager (never in code, never in env files committed to git)
- Each environment has its own configuration - never share config between staging and production
- Feature flags use a structured format: `FEATURE_{NAME}_ENABLED=true|false`

---

## 5. Rollback

- Every deployment must be reversible within 5 minutes
- Keep the previous version's artifacts available for at least 7 days
- Database migrations must be backward-compatible - the previous code version must work with the new schema
- Rollback procedures are documented in the operational model

---

*Referenced by codegen-agent. Source of truth: `planifest-framework/standards/deployment-standards.md`*
