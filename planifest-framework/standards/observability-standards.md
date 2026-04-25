# Planifest Observability Standards

> If you can't observe it, you can't operate it. Every component must emit structured logs, expose metrics, and propagate traces. These standards ensure agent-generated code is observable from day one.

---

## 1. Three Pillars

Every component must implement all three:

| Pillar | What | Tool Examples |
|--------|------|---------------|
| **Logs** | Structured event records | Pino, Winston, zerolog, slog |
| **Metrics** | Numerical measurements over time | Prometheus, CloudWatch, OpenTelemetry |
| **Traces** | Request flow across components | OpenTelemetry, Jaeger, Cloud Trace |

---

## 2. Logging

- **Structured logging only** - JSON format, never unstructured strings
- **Required fields:** `timestamp`, `level`, `message`, `service`, `requestId`
- **Log levels:** `error` (action required), `warn` (degraded but functional), `info` (significant events), `debug` (development only, never in production)
- **Never log:** credentials, PII, full request/response bodies, stack traces at info level
- **Always log:** request start/end, errors with context, authentication events, data mutations

---

## 3. Metrics

Required metrics for every service:

| Metric | Type | Description |
|--------|------|-------------|
| `http_requests_total` | Counter | Total requests by method, path, status |
| `http_request_duration_seconds` | Histogram | Request latency by method, path |
| `http_requests_in_flight` | Gauge | Current concurrent requests |
| `errors_total` | Counter | Errors by type and severity |

Additional metrics per the SLO definitions (e.g., business-specific counters).

---

## 4. Tracing

- Use OpenTelemetry SDK for trace instrumentation
- Propagate trace context across all component boundaries (HTTP headers, message metadata)
- Create spans for: HTTP requests, database queries, external API calls, queue operations
- Include relevant attributes: `user.id`, `order.id`, `component.id`

---

## 5. Health Checks

Every service exposes:

| Endpoint | Purpose | Response |
|----------|---------|----------|
| `/health` | Liveness - is the process running? | 200 OK |
| `/ready` | Readiness - can it serve traffic? | 200 OK or 503 |

Readiness checks verify database connectivity, cache availability, and downstream service health.

---

## 6. Alerting

Alerts are defined in the operational model. The codegen-agent implements the metric emission; the human configures the alerting thresholds and channels.

---

*Referenced by codegen-agent and spec-agent. Source of truth: `planifest-framework/standards/observability-standards.md`*
