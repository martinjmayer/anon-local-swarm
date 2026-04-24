# Quirks — obs

## Q-001: Go not installed at codegen time

**Status:** Open
**Discovered:** P3 codegen

Go is not present in the system PATH at the time of code generation. `go mod tidy` and `go test ./...` cannot be run in-session.

**Resolution:** Install Go ≥ 1.23, then from `src/obs/`:
```
go mod tidy
go test ./...
```

`go-duckdb` requires CGO and a C compiler (gcc via MinGW on Windows or MSYS2). Ensure `gcc` is on PATH before `go mod tidy`.

## Q-002: go-duckdb requires CGO

`github.com/marcboeker/go-duckdb` uses CGO. On Windows this requires a MinGW/GCC toolchain. Install via:
```
winget install -e --id MSYS2.MSYS2
```
Then from MSYS2: `pacman -S mingw-w64-x86_64-gcc`

The orchestrator also uses CGO for DuckDB (it imports `als/obs`). The `modernc.org/sqlite` driver for swarm.db is pure-Go and does not require CGO.
