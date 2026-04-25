# Planifest API Design Standards

> APIs are contracts. They are consumed by other components, external systems, and future agents. A well-designed API is self-documenting, consistent, and hard to misuse.

---

## 1. REST Conventions

| Concern | Standard |
|---------|----------|
| **Resource naming** | Plural nouns, kebab-case: `/api/v1/user-profiles` |
| **HTTP methods** | GET (read), POST (create), PUT (full replace), PATCH (partial update), DELETE (remove) |
| **Status codes** | Use standard codes. 200 OK, 201 Created, 204 No Content, 400 Bad Request, 401 Unauthorized, 403 Forbidden, 404 Not Found, 409 Conflict, 422 Unprocessable Entity, 500 Internal Server Error |
| **Versioning** | URL path prefix: `/api/v1/`. Never break a published version. |
| **Pagination** | Cursor-based preferred. Offset-based acceptable. Always include `total`, `next`, `previous`. |
| **Filtering** | Query parameters: `?status=active&created_after=2024-01-01` |
| **Sorting** | Query parameter: `?sort=created_at:desc` |

---

## 2. Request/Response Format

- All request and response bodies use JSON
- All dates use ISO-8601 format with timezone: `2024-01-15T10:30:00Z`
- All IDs use the format defined in the component's data contract
- Null fields are omitted from responses (do not send `"field": null`)
- Envelope responses for collections: `{ "data": [...], "pagination": {...} }`

---

## 3. Error Responses

All errors use a consistent structure:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Human-readable description",
    "details": [
      { "field": "email", "message": "must be a valid email address" }
    ]
  }
}
```

- `code` is a machine-readable constant (UPPER_SNAKE_CASE)
- `message` is human-readable, never exposes internal details
- `details` is optional, used for field-level validation errors
- Never expose stack traces, SQL queries, or internal paths in error responses

---

## 4. Authentication and Authorization

- All endpoints require authentication unless explicitly marked as public in the OpenAPI spec
- Use the auth strategy declared in the confirmed design (JWT, session, OAuth2, API key)
- Authorization checks happen at the handler level, not in middleware alone
- Rate limiting is required for all public-facing endpoints

---

## 5. OpenAPI Specification

- Every API must have a corresponding OpenAPI 3.1 spec at `plan/{feature-id}/openapi-spec.yaml`
- The spec is the source of truth - implementation must match it exactly
- Every endpoint, parameter, request body, and response must be documented
- Use `$ref` for shared schemas - do not duplicate type definitions
- Include example values for every schema

---

## 6. Breaking Changes

A breaking change is any change that could cause existing consumers to fail:
- Removing an endpoint or field
- Changing a field's type
- Adding a required field to a request body
- Changing the meaning of an existing field

Breaking changes require:
1. An ADR documenting the change and its impact
2. Human approval
3. A new API version (`v1` → `v2`)
4. A migration period where both versions are available

---

*Referenced by spec-agent and codegen-agent. Source of truth: `planifest-framework/standards/api-design-standards.md`*

