# Infinita

CLI-first personal finance app in Go.

## Quick Start

### Prerequisites

- Go 1.22+
- CGO enabled (requires a C compiler — `gcc` or `clang`)

### Build

```bash
go build -o bin/infinita ./cmd/cli
```

### Run

```bash
go run ./cmd/cli <command> [flags]
```

Available commands: `add`, `list`, `category`, `budget`, `report`, `settings`.

### Test

```bash
go test ./...
go test -cover ./...
```

### Lint

```bash
golangci-lint run
```

## Architecture

Hexagonal (Clean) Architecture with an embedded HTTP server. The CLI communicates with the embedded server via an in-process HTTP client.

```
transport (CLI/HTTP client) → application (use cases) → domain (entities, value objects)
infrastructure (SQLite, YAML settings) → application/domain
```

## Docs

- PRD: [`docs/prd.md`](docs/prd.md)
- TRD: [`docs/trd.md`](docs/trd.md)
- ADRs: [`docs/adr/`](docs/adr/)
- Agent guidelines: [`AGENTS.md`](AGENTS.md)