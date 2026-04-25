# Test Coverage — obs

## Summary

| Type | Coverage | Notes |
|---|---|---|
| Unit | ~80% | Core write path, scrubbing, event marshalling covered |
| Integration | 0% | No cross-component integration tests |
| E2E | 0% | No end-to-end tests (no HTTP surface) |

## What is tested

- `WriteEvent` persists events to DuckDB and mirrors to stdout
- `scrubAndMarshal` redacts credential-keyed fields
- `Open` initialises schema idempotently
- `Logger.Close` releases the connection cleanly
- Obs must not create `swarm.db` (cross-component isolation guard)

## What is not tested

- Very high event volume / write throughput benchmarking
- DuckDB schema migration path
- Concurrent writes from multiple goroutines (WriteEvent is called from a single scheduler goroutine in practice)

## Test files

- `src/obs/obs_test.go`
