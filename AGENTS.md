# AGENTS.md — Infinita Personal Finance App

Guidance for coding agents working in this repository.

## Read First

Always treat these as canonical before making significant changes:

- PRD: `docs/prd.md`
- TRD: `docs/trd.md`
- ADRs: `docs/adr/`
- Architecture diagram: `docs/diagrams/trd/001-trd-system-architecture.mmd`
- Data model: `docs/diagrams/trd/001-trd-data-model-erd.mmd`

If a change affects scope, CLI behavior, validation, architecture, storage, or error handling, update the matching docs in the same change.

## Commands

```bash
go build -o bin/infinita ./cmd/cli
go run ./cmd/cli <command> [flags]
go test ./...
go test -cover ./...
golangci-lint run
go mod tidy
```

## Architecture Rules

Hexagonal architecture. Dependencies point inward:

```text
transport -> application -> domain
infrastructure -> application/domain
```

- `internal/domain`: business rules only
- `internal/application`: use cases and ports
- `internal/transport/cli`: CLI adapter only
- `internal/infrastructure`: concrete adapters

## SQL Generation

- Never edit generated SQL files manually. Always use `sqlc generate` to regenerate them.
- Edit only the source `.sql` query files and `sqlc.yaml` config, then run `sqlc generate`.

## Critical Constraints

- Money uses `int64` minor units only. No `float64`.
- Reject money inputs with more than 2 fractional digits. No rounding.
- Validate at the application/transport boundary.
- Preserve structured errors and CLI exit codes defined in the TRD.
- User-facing messages must stay in English.
- MVP stays local-only. Do not add remote storage, telemetry-by-default, or transaction edit/delete unless docs explicitly change.

## Documentation Rules

- New architecture or engineering decision: add an ADR using `docs/adr/README.md`.
- Product behavior or scope change: update the PRD.
- Technical contract change: update the TRD.
- Diagram-affecting change: update the related `.mmd` file.
