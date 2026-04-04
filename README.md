# Infinita

CLI-first personal finance app in Go. Track income, expenses, budgets, and financial summaries — locally, privately, and fast.

## Features

- **Add transactions** — income and expense with category, date, and optional description
- **List & filter** — browse transactions with category filtering and pagination
- **Categories** — default categories + custom category creation
- **Budgeting** — set monthly budget limits per category, check remaining and over-limit status
- **Reports** — daily summaries (income, expense, net) and monthly summaries (with closing balance and top spending categories)
- **Initial balance** — set a starting balance for cumulative closing-balance calculations
- **Local-first** — all data stored on-device via SQLite, no telemetry by default
- **Embedded HTTP server** — CLI communicates with business logic through a localhost REST API (hexagonal architecture)

## Quick Start

### Prerequisites

- Go 1.24+
- CGO enabled (requires a C compiler — `gcc` or `clang`)

### Build

```bash
make build
# or: go build -o bin/infinita ./cmd/cli
```

### Run

```bash
make run ARGS="<command> [flags]"
# or: go run ./cmd/cli <command> [flags]
```

### Available Commands

| Command | Description |
|---------|-------------|
| `add` | Create a transaction (income/expense) |
| `list` | List transactions with optional category filter |
| `category list` | List all available categories |
| `category create` | Create a custom category |
| `budget set` | Set monthly budget for a category |
| `budget status` | Check budget status for a month |
| `report daily` | View daily summary |
| `report monthly` | View monthly summary with closing balance |

### Test

```bash
make test       # go test ./...
make cover      # go test -cover ./...
```

### Lint

```bash
make lint       # golangci-lint run
```

### CI

```bash
make ci         # lint + test + build
```

## Architecture

Hexagonal (Ports & Adapters) Architecture. Dependencies point inward:

```
transport (CLI/HTTP client) → application (use cases) → domain (entities, value objects)
infrastructure (SQLite, YAML settings) → application/domain
```

The CLI embeds an HTTP server in the same process and communicates with it via localhost HTTP. This keeps the CLI as a thin client while maintaining clean layer separation.

```
cmd/cli ──┬── internal/transport/cli (CLI adapter, spf13/cobra)
          ├── internal/transport/client (HTTP client)
          └── internal/transport/server (HTTP handlers, stdlib net/http)
                    │
              internal/application (use cases + ports)
                    │
              internal/domain (entities, value objects, domain errors)
                    ▲
              internal/infrastructure (SQLite repos, YAML settings)
```

### Key Design Decisions

- **Money as `int64` minor units** — no `float64` for monetary values
- **ULID identifiers** — lexicographically sortable, collision-resistant (ADR-004)
- **sqlc for SQL-to-Go** — type-safe queries from `.sql` files (ADR-003)
- **golang-migrate** — embedded schema migrations (ADR-003)
- **SQLite** — local-only persistence with foreign-key enforcement
- **JSend-style responses** — consistent API error format

## Project Structure

```
.
├── cmd/cli/                          # CLI entrypoint (embeds HTTP server)
├── cmd/server/                       # Standalone server (dev use)
├── internal/
│   ├── bootstrap/                    # Shared composition root
│   ├── domain/                       # Business rules (entities, value objects)
│   ├── application/                  # Use cases and ports
│   ├── transport/                    # CLI, HTTP client, HTTP server
│   ├── infrastructure/               # SQLite repos, YAML settings
│   └── testutil/                     # Test helpers
├── internal/infrastructure/database/sqlite/
│   ├── migrations/                   # golang-migrate files
│   ├── queries/                      # sqlc query definitions
│   └── sqlc/                         # sqlc generated code (committed)
├── docs/
│   ├── prd.md                        # Product Requirements Document
│   ├── trd.md                        # Technical Requirements Document
│   ├── adr/                          # Architecture Decision Records
│   ├── api/openapi.yaml              # REST API specification
│   ├── diagrams/                     # Mermaid architecture diagrams
│   └── security/local-data.md        # Local data handling guide
├── sqlc.yaml                         # sqlc configuration
└── Makefile                          # Build, test, migration helpers
```

## Database Migrations

```bash
make mg-create name=<description>   # Create a new migration
make mgu                            # Run all pending migrations (up)
make mgd                            # Rollback last migration (down 1)
make mgd-all                        # Rollback all migrations
make mgf version=<ver>              # Force migration version
make mgv                            # Check current migration version
```

## sqlc

After editing `.sql` query files or `sqlc.yaml`:

```bash
make sg                             # sqlc generate
```

## Documentation

- **PRD**: [`docs/prd.md`](docs/prd.md) — Product requirements and roadmap
- **TRD**: [`docs/trd.md`](docs/trd.md) — Technical requirements and contracts
- **ADRs**: [`docs/adr/`](docs/adr/) — Architecture decision records
  - [ADR-001: Adopt Hexagonal Clean Architecture](docs/adr/ADR-001-adopt-hexagonal-clean-architecture-go.md)
  - [ADR-002: Defer Server Separation to Roadmap](docs/adr/ADR-002-defer-server-separation-to-roadmap.md)
  - [ADR-003: Adopt sqlc, golang-migrate, Tiered Testing](docs/adr/ADR-003-adopt-sqlc-golang-migrate-tiered-testing.md)
  - [ADR-004: Adopt ULID for Transaction Identifiers](docs/adr/ADR-004-adopt-ulid-for-transaction-identifiers.md)
  - [ADR-005: Enable Server Transport with Embedded Architecture](docs/adr/ADR-005-enable-server-transport-with-embedded-architecture.md)
- **API Spec**: [`docs/api/openapi.yaml`](docs/api/openapi.yaml)
- **Security**: [`docs/security/local-data.md`](docs/security/local-data.md)
- **Diagrams**: [`docs/diagrams/`](docs/diagrams/)

## Roadmap

### MVP (Current)
- Add, list, and filter transactions
- Default and custom categories
- Monthly budgeting per category
- Daily and monthly cashflow summaries
- Optional initial balance
- Local SQLite storage
- Embedded HTTP server with REST API

### v1.1
- CSV and TXT report export
- Expanded transaction filtering and search
- Logging reminders
- Local backup and restore

### v2.0
- Cross-device sync
- Web or mobile GUI dashboard
- Automated insights and forecasting
- Shared household budgeting

## License

MIT
