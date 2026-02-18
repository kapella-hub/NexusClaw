# NexusClaw

Unified gateway for MCP server management, credential vaulting, and AI firewall.

## Modules

- **Nexus Pass** — Credential vault and authentication bridge (registration, login, encrypted vault CRUD, credential relay)
- **Nexus Nodes** — Managed MCP server registry and lifecycle (Docker containers, WebSocket proxy, OAuth2, discovery)
- **Nexus Sentry** — AI firewall, audit logging, budget enforcement, and rule engine

## Prerequisites

- Go 1.25+
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

# Start dependencies (postgres, redis, app)
docker compose -f deployments/docker-compose.yaml up -d

# Run migrations
bash scripts/migrate.sh

# Or run the server directly (requires Go)
go run ./cmd/nexusclaw serve
```

## Development Commands

If you have `make` installed:

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

Without `make`, run the equivalent commands directly:

```bash
# Build
go build -o bin/nexusclaw ./cmd/nexusclaw

# Test
go test ./... -race -count=1

# Docker
docker compose -f deployments/docker-compose.yaml up -d
docker compose -f deployments/docker-compose.yaml down
```

## API Endpoints

### Health
- `GET /healthz` — Liveness check
- `GET /readyz` — Readiness check (verifies DB connectivity)

### Pass (AuthBridge) — `/api/v1/pass`
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/register` | No | Create a new user |
| POST | `/sessions` | No | Login (returns token) |
| DELETE | `/sessions/{id}` | Yes | Logout |
| GET | `/vault` | Yes | List vault entries |
| POST | `/vault` | Yes | Store encrypted credential |
| DELETE | `/vault/{id}` | Yes | Remove vault entry |
| POST | `/relay/{provider}` | Yes | Relay credentials to provider |

### Nodes (Managed MCP) — `/api/v1/nodes`
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/` | Yes | List servers |
| POST | `/` | Yes | Register server |
| GET | `/discover?q=` | Yes | Search servers by name/image |
| GET | `/{id}` | Yes | Get server details |
| DELETE | `/{id}` | Yes | Remove server |
| POST | `/{id}/start` | Yes | Start server container |
| POST | `/{id}/stop` | Yes | Stop server container |
| GET | `/{id}/ws` | Yes | WebSocket proxy to MCP server |

### Nodes OAuth — `/api/v1/nodes/oauth`
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/initiate/{provider}` | Yes | Start OAuth2 flow |
| GET | `/callback/{provider}` | Yes | OAuth2 callback |

### Sentry (Firewall) — `/api/v1/sentry`
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/audit` | Yes | List audit entries |
| GET | `/rules` | Yes | List firewall rules |
| POST | `/rules` | Yes | Create rule |
| PUT | `/rules/{id}` | Yes | Update rule |
| DELETE | `/rules/{id}` | Yes | Delete rule |
| GET | `/budget` | Yes | Get token budget |
| PUT | `/budget` | Yes | Update token budget |

## Public SDKs

### MCP SDK (`pkg/mcpsdk`)
```go
client := mcpsdk.NewClient("http://localhost:8080", "your-token")
servers, _ := client.ListServers(ctx)
client.Connect(ctx, serverID)
result, _ := client.Call(ctx, serverID, "tools/list", nil)
```

### Sentry SDK (`pkg/sentryapi`)
```go
client := sentryapi.NewClient("http://localhost:8080", "your-token")
rules, _ := client.ListRules(ctx)
budget, _ := client.GetBudget(ctx)
```

## CLI

```bash
# Auth
nexusclaw pass register --email user@example.com --password secret
nexusclaw pass login --email user@example.com --password secret

# Servers
nexusclaw node list
nexusclaw node register --name my-mcp --image mcp-server:latest
nexusclaw node start <server-id>

# Sentry
nexusclaw sentry audit
nexusclaw sentry rules
nexusclaw sentry budget
```

## Architecture

```
cmd/nexusclaw/          CLI entrypoint (Cobra)
internal/
  app/                  HTTP handler wiring and DI
  cli/                  CLI commands + HTTP API client
  nodes/                MCP server management (Docker, WebSocket, OAuth, Registry)
  pass/                 Auth, sessions, encrypted vault, credential relay
  sentry/               Audit logging, rule engine, budget tracking
  platform/
    config/             Viper-based configuration
    crypto/             AES-256-GCM encryption, Argon2id hashing, token issuing
    database/           PostgreSQL pool + Redis client + migrations
    middleware/         Auth, rate limiting, logging, request ID
    respond/            JSON response helpers
    server/             Graceful HTTP server
pkg/
  mcpsdk/               Public MCP client SDK
  sentryapi/            Public Sentry client SDK
```
