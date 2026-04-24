# Dockerfile — runtime image for the ALS orchestrator.
#
# Contains the swarm binary + Node.js (required for MCP server subprocesses).
# The binary is built by Dockerfile.build; this image packages it for running.
#
# Full workflow:
#   1. docker build -f Dockerfile.build --output bin .   # build + test
#   2. docker compose up                                  # run
#
# Ollama must be running on the host. The compose file sets
# OLLAMA_BASE_URL=http://host.docker.internal:11434 automatically.

# ── Stage 1: build Go binary + run tests ────────────────────────────────────
FROM golang:1.23-bullseye AS builder

RUN apt-get update \
 && apt-get install -y --no-install-recommends gcc libc-dev \
 && rm -rf /var/lib/apt/lists/*

WORKDIR /src
COPY go.work ./
COPY src/obs/go.mod         ./src/obs/
COPY src/orchestrator/go.mod ./src/orchestrator/

RUN cd src/obs          && go mod download || true
RUN cd src/orchestrator && go mod download || true

COPY src/obs/         ./src/obs/
COPY src/orchestrator/ ./src/orchestrator/

# Tests must pass before the binary is produced.
RUN CGO_ENABLED=1 go test -count=1 ./src/obs/... \
 && CGO_ENABLED=1 go test -count=1 ./src/orchestrator/...

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
    go build -C src/orchestrator -o /out/swarm ./cmd/swarm

# ── Stage 2: build mcp-reddit ────────────────────────────────────────────────
FROM node:20-bullseye-slim AS mcp-builder

WORKDIR /mcp
COPY src/mcp-reddit/package.json src/mcp-reddit/tsconfig.json ./
RUN npm install

COPY src/mcp-reddit/src/ ./src/
RUN npm run build

# ── Stage 3: runtime ─────────────────────────────────────────────────────────
FROM node:20-bullseye-slim AS runtime

# Install DuckDB CLI so the operator can query obs.db from inside the container
# (optional — the host DuckDB CLI also works on volume-mounted obs.db).
RUN apt-get update \
 && apt-get install -y --no-install-recommends curl ca-certificates \
 && rm -rf /var/lib/apt/lists/*

# Copy swarm binary.
COPY --from=builder /out/swarm /usr/local/bin/swarm

# Copy compiled mcp-reddit.
COPY --from=mcp-builder /mcp/dist /app/src/mcp-reddit/dist
COPY --from=mcp-builder /mcp/node_modules /app/src/mcp-reddit/node_modules

# Working directory is /data — all runtime files (db, config, skills, output)
# are mounted here by docker-compose.
WORKDIR /data

ENTRYPOINT ["swarm"]
CMD ["-config", "config.yaml"]
