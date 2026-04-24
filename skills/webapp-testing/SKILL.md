---
name: webapp-testing
description: Test strategy and patterns for GO_CODE validation — used to guide self-correction and test writing
bundle: anthropic
source: https://github.com/anthropics/skills/tree/main/skills/webapp-testing
task_types: [GO_CODE]
---

# Webapp Testing (Go adaptation)

Reference skill for GO_CODE tasks that write or validate tests for the generated code.

## Go test strategy for ALS-generated code

### Unit tests — pure functions

Every pure function must have a test. Use `testing.T` directly:

```go
func TestMyFunction_DescribesWhatItDoes(t *testing.T) {
    got := MyFunction(input)
    if got != want {
        t.Errorf("want %v, got %v", want, got)
    }
}
```

Name tests as `Test<FunctionName>_<brief description>`. Reference requirement IDs where applicable:

```go
func TestValidate_req007_ValidCodePasses(t *testing.T) { ... }
```

### Table-driven tests — multiple inputs

```go
func TestSomething(t *testing.T) {
    cases := []struct {
        name  string
        input string
        want  string
    }{
        {"empty input", "", ""},
        {"normal case", "hello", "HELLO"},
    }
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            got := Something(tc.input)
            if got != tc.want {
                t.Errorf("want %q, got %q", tc.want, got)
            }
        })
    }
}
```

### Integration tests — I/O boundaries

Use `t.TempDir()` for temporary files — never hardcode paths.
Use `httptest.NewServer` for HTTP clients — never call real external services.

```go
func TestClientCallsExternalAPI(t *testing.T) {
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        json.NewEncoder(w).Encode(map[string]string{"result": "ok"})
    }))
    defer srv.Close()
    client := NewClient(srv.URL)
    // ... assert
}
```

### Self-correction test patterns

When fixing a compiler error, verify the fix does not break adjacent tests:
1. Fix the stated error.
2. Check that surrounding functions still compile.
3. Add a test for the fixed case if one does not exist.

## Coverage targets for ALS-generated code

| Type | Target |
|---|---|
| Pure functions | 90%+ |
| HTTP clients | 80%+ (httptest) |
| File I/O | 70%+ (t.TempDir) |
| Main binary | smoke test only |

## What not to test

- Third-party library internals (go-duckdb, go-sqlite3)
- Ollama API responses (test the parser, not the HTTP call)
- MCP server responses (test the tool handler logic)
