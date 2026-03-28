# AGENTS.md — Infinita Personal Finance App

Guidance for agentic coding agents working in this repository.

## Project Overview

Infinita is a CLI-first personal finance application written in Go. MVP scope covers daily transaction logging, category-based budgeting, simple reporting, and local-only SQLite storage. No server transport, no GUI, no bank integrations in MVP.

**Authoritative docs — always read these before implementing:**
- PRD: `docs/prd/001-infinita-personal-finance-app.md`
- TRD: `docs/trd/001-infinita-personal-finance-app-mvp-cli.md`
- ADR-002 (current): `docs/adr/ADR-002-defer-server-separation-to-roadmap.md`
- Architecture diagram: `docs/diagrams/trd/001-trd-system-architecture.mmd`
- Data model: `docs/diagrams/trd/001-trd-data-model-erd.mmd`

## Build / Run / Test Commands

```bash
# Build CLI binary
go build -o bin/infinita ./cmd/cli

# Run CLI
go run ./cmd/cli <command> [flags]

# Run all tests
go test ./...

# Run tests for a single package
go test ./internal/domain/entity/

# Run a single test by name
go test -run TestFunctionName ./internal/domain/entity/

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...

# Run linter (if golangci-lint is installed)
golangci-lint run

# Tidy modules
go mod tidy
```

## Architecture Reference

Hexagonal (Ports & Adapters). Dependencies point strictly inward:

```
transport → application → domain
infrastructure → application/domain
```

For the full directory layout, layer responsibilities, and dependency rules, see TRD section "Directory Structure" and "Architecture Overview".

### Quick Layer Rules

- **`domain/`**: Pure Go types and business rules. Zero imports from `transport/`, `infrastructure/`, or external frameworks.
- **`application/`**: Orchestrates domain objects through input/output port interfaces. Imports only `domain/`.
- **`transport/cli/`**: Adapts CLI input to application input ports. Never accesses repositories directly.
- **`infrastructure/`**: Implements output port interfaces. Wires dependencies in `main.go` or a dedicated composition root.

## Code Style Guidelines

### Go Conventions

- Follow [Effective Go](https://go.dev/doc/effective_go) and standard `gofmt` formatting.
- Run `go vet` and `golangci-lint` before committing.
- Use `goimports` for import ordering: stdlib, then blank line, then external packages, then blank line, then internal packages.

### Naming

- **Packages**: lowercase, single word, no underscores (e.g., `entity`, `usecase`, `sqlite`).
- **Files**: `snake_case.go` (e.g., `transaction_repository.go`, `budget_service.go`).
- **Interfaces**: noun or verb+er in `CamelCase` (e.g., `TransactionRepository`, `BudgetCalculator`).
- **Structs**: `CamelCase` exported; `camelCase` unexported.
- **Functions/Methods**: `CamelCase` exported; `camelCase` unexported.
- **Constants**: `CamelCase` exported; `camelCase` unexported. Use `iota` for enum-like sequences.
- **Test files**: `*_test.go` colocated with the code being tested. Functions: `TestXxx`, `TestXxx_Yyy`, or `TestXxx/TableDriven`.

### Testing

- Unit tests for validation, budget math, and summary aggregation.
- Integration tests for CLI command flows and error handling.
- Table-driven tests preferred for multi-case scenarios:
  ```go
  func TestNormalizeAmount(t *testing.T) {
      tests := []struct{
          name    string
          input   string
          want    int64
          wantErr bool
      }{ ... }
      for _, tt := range tests { ... }
  }
  ```

## Critical Rules for Agents

These are the most frequently violated constraints. For full details, see TRD.

### Monetary Values

- **Always `int64` minor units**, never `float64`. Field names end in `Minor` (e.g., `amountMinor`).
- Exact decimal-to-integer conversion only (`amount * 100`). Reject >2 fractional digits. No rounding.
- All monetary CLI inputs share one normalization function. See TRD "Monetary Input Normalization".

### Error Handling

- Structured format: `{code, message, field?, hint?}`. See TRD `TRD-VAL-002` for stable error codes.
- Wrap errors with `fmt.Errorf("verb: %w", err)`. Never lose error context.
- CLI exit codes: `0` success, `2` validation/domain errors, `3` runtime/storage errors.
- All user-facing messages in English.

### Types and Validation

- Value objects for domain primitives: `Money` (minor units), `CategoryKey` (case-insensitive), `EntryType` (`income`/`expense`).
- Validate at application/transport boundary. Centralized in `internal/application/validation/`.

### MVP Scope Boundaries

- **Local-only SQLite**. No remote storage modes.
- **No telemetry by default**. Analytics requires explicit user opt-in.
- **No transaction edit/delete**. Only add, list, filter.
- **No server scaffolding**. `cmd/server` and `internal/transport/server` are roadmap-only.
- **No floating-point for money**. Always integer minor units.

## CLI Command Surface

See TRD section "CLI Command Surface (MVP Canonical)" for the full command reference and contracts.
