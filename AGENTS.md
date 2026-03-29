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
  - **Test package exception (Go convention)**: `package <name>_test` is allowed for external/black-box tests.
  - Use `package <name>` for same-package/white-box tests when access to unexported symbols is needed.
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

## Documentation-First Policy

Every significant change must be documentation-aligned. Code and docs must never drift apart.

### What Counts as a Significant Change

Any change that does one or more of the following:
- Introduces a new domain concept, entity, or value object
- Alters existing domain types or business rules
- Adds, removes, or modifies CLI commands or their flags/behavior
- Changes architectural boundaries (layers, ports, adapters)
- Modifies data model (schema, tables, columns, constraints)
- Introduces a new external dependency or library
- Changes error handling strategy or error codes
- Affects monetary handling, validation logic, or scope boundaries

### Before Making Any Significant Change

1. **Read existing ADRs first.** Check `docs/adr/` for any ADR whose decision might be affected by or relevant to the planned change. List all ADRs in the directory to find applicable ones.
2. **Read the PRD.** Verify the change aligns with product scope defined in `docs/prd/001-infinita-personal-finance-app.md`. If the change falls outside MVP scope, stop and ask the user.
3. **Read the TRD.** Verify the change is consistent with technical contracts in `docs/trd/001-infinita-personal-finance-app-mvp-cli.md`. If the TRD contradicts the change, either update the TRD or reconsider the change.
4. **Read architecture and data model diagrams.** Check `docs/diagrams/trd/` if the change affects system structure or data model.

### When Writing Documentation Updates

- **New architectural/engineering decision → write an ADR.** Follow the template in `docs/adr/README.md`. Use the next sequential ID. Link related PRDs/TRDs.
- **Product scope change → update PRD.** If the change adds, removes, or redefines a feature, update the relevant PRD section. Keep PRD as the single source of truth for "what" the system does.
- **Technical contract change → update TRD.** If the change alters CLI surface, data model, validation rules, error codes, or architectural boundaries, update the TRD. Keep TRD as the single source of truth for "how" the system is built.
- **Diagram-affected change → update diagrams.** If the change modifies system architecture or data model, update the corresponding `.mmd` files in `docs/diagrams/`.
- **ADR superseded → create new ADR.** Never modify an accepted ADR. Create a new ADR with a "Supersedes" reference and update the old ADR's status to "Superseded".

### Doc-Code Consistency Checklist

Before completing any significant change, verify:
1. All affected ADRs have been reviewed and are still consistent with the change
2. PRD accurately reflects the feature scope after the change
3. TRD accurately reflects CLI surface, data model, validation rules, and architecture after the change
4. Diagrams (if affected) are updated to match the new state
5. No documented behavior is contradicted by the implementation
6. No implemented behavior is missing from documentation

### Quick Reference — Doc Locations

| Document | Location | Purpose |
|----------|----------|---------|
| ADRs | `docs/adr/ADR-XXX-*.md` | Architecture & engineering decisions |
| PRD | `docs/prd/001-*.md` | Product requirements & scope |
| TRD | `docs/trd/001-*.md` | Technical requirements & contracts |
| Architecture diagram | `docs/diagrams/trd/001-trd-system-architecture.mmd` | System structure |
| Data model diagram | `docs/diagrams/trd/001-trd-data-model-erd.mmd` | Database schema |
| ADR template | `docs/adr/README.md` | How to write ADRs |

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
