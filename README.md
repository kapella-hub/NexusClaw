# NexusClaw

Unified gateway for MCP server management, credential vaulting, and AI firewall.

## Modules

- **Nexus Pass** — Credential vault and authentication bridge
- **Nexus Nodes** — Managed MCP server registry and lifecycle
- **Nexus Sentry** — AI firewall, audit logging, and budget enforcement

## Prerequisites

- Go 1.22+
- PostgreSQL 16+
- Redis 7+
- Docker & Docker Compose (for local dev)

## Quick Start

```bash
# Clone and enter the project
git clone https://github.com/kapella-hub/NexusClaw.git
cd NexusClaw

# Copy environment config
cp .env.example .env

# Start dependencies
make docker-up

# Run migrations
make migrate

# Start the server
make run
```

## Make Targets

| Target | Description |
|---|---|
| `make build` | Build binary to `bin/nexusclaw` |
| `make run` | Run the server |
| `make test` | Run tests with race detector |
| `make lint` | Run golangci-lint |
| `make migrate` | Run database migrations |
| `make docker-up` | Start local dev stack |
| `make docker-down` | Stop local dev stack |
| `make ci` | Run full CI pipeline |
| `make clean` | Remove build artifacts |

## API

Health check: `GET /healthz`
Readiness: `GET /readyz`

All module endpoints under `/api/v1/` return 501 (Not Implemented) until modules are built out.
