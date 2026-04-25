# Planifest Testing Standards

> Tests are the requirements' executable counterpart. Every acceptance criterion becomes a test. Every contract becomes a contract test. Every edge case the requirements identify become a test case. If it's not tested, it's not verified.

---

## Why This Matters for Agent-Generated Code

Agents generate tests alongside code. Without clear standards, agent-generated tests tend toward two failure modes: (1) trivial tests that assert implementation details rather than behavior, or (2) sprawling integration tests that are slow, flaky, and unmaintainable. These standards prevent both.

---

## 1. Test Pyramid

Every component must have tests at three levels:

| Level | What It Tests | Speed | Scope | Required Coverage |
|-------|--------------|-------|-------|-------------------|
| **Unit** | Pure functions, business logic, transformations | Fast (< 100ms each) | Single function/class | Every exported pure function |
| **Integration** | API endpoints, database queries, service interactions | Medium (< 5s each) | Single component boundary | Every endpoint, every query |
| **Contract** | Cross-component interface agreements | Medium | Component boundary pair | Every consumed interface |

E2E tests are recommended but not required at the component level. The orchestrator may request E2E tests for critical user flows during Phase 3.

---

## 2. Agentic TDD & Requirement Traceability

Agents must employ **Agentic TDD (Test-Driven Development)** to guarantee semantic correctness. The LLM must not evaluate its own success probabilistically; it must prove it deterministically.

**The Agentic TDD Loop:**
1. Read the functional requirement (`plan/current/requirements/req-*.md`).
2. Write the failing test case *first*.
3. Execute the test to verify it fails (establishing the boundary).
4. Write the implementation logic.
5. Execute the test to verify it passes.

**Requirement Traceability:**
Every functional requirement must be explicitly traceable to a test. You MUST include the requirement ID in the test description or suite name.

```javascript
// ✅ Correct: Traceable to req-001-auth
describe("req-001-auth: user login flow", () => {
  test("returns 401 when password is mathematically invalid", () => {
    // Arrange - set up preconditions
    // Act - perform the operation
    // Assert - verify the outcome
  })
})
```

**Rules:**
- Test names describe behavior, not implementation: "returns 404 when order does not exist" not "test getOrder"
- One assertion concept per test. Multiple `expect` calls are fine if they verify the same logical assertion.
- No shared mutable state between tests. Each test sets up its own data.
- No test interdependence. Tests must pass in any order.

---

## 3. What to Test

**Always test:**
- Happy path for every user story's acceptance criteria
- Error cases: invalid input, missing data, unauthorized access
- Boundary conditions: empty arrays, zero values, maximum lengths
- State transitions: before/after for any operation that changes state

**Never test:**
- Framework internals (don't test that Express calls your handler)
- Private implementation details (test the public interface)
- Third-party library behavior (trust it or replace it)
- Trivial getters/setters with no logic

---

## 4. Test Data

- Use factories or builders, not raw object literals repeated across tests
- Test data should be minimal - only include fields relevant to the test
- Never use production data in tests
- Never hardcode dates - use relative dates or freezable clocks
- Database tests must clean up after themselves or use transactions that roll back

---

## 5. Mocking

- **Mock at the boundary, not inside the component.** Mock external services, databases (for unit tests), and third-party APIs. Do not mock internal modules.
- **Integration tests use real dependencies.** Database integration tests hit a real database (test instance). API tests hit the real API server.
- **Contract tests verify the mock matches reality.** If you mock a service in a consumer test, the provider must have a corresponding contract test that verifies the mock's behavior.

---

## 6. Flakiness Policy

- A flaky test is worse than no test - it erodes trust in the suite
- If a test fails intermittently, fix it or delete it. Do not retry-and-ignore.
- Common flakiness sources to avoid: time-dependent assertions, shared state, port conflicts, network calls to external services
- All tests must be deterministic. Use dependency injection for time, randomness, and external state.

---

## 7. Coverage

Coverage is a proxy, not a target. The goal is confidence that the requirements are implemented correctly.

| Metric | Minimum | Target |
|--------|---------|--------|
| Line coverage | 80% | 90%+ |
| Branch coverage | 70% | 85%+ |
| Critical path coverage | 100% | 100% |

"Critical path" = the primary user-facing flows identified in the design requirements' acceptance criteria.

---

*Referenced by codegen-agent and validate-agent. Source of truth: `planifest-framework/standards/testing-standards.md`*
