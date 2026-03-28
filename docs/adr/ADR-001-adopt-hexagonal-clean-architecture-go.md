# ADR-001: Adopt Hexagonal Clean Architecture for Go CLI/Server Boundaries

## Status
Superseded

## Date
2026-03-28

Superseded by: `docs/adr/ADR-002-defer-server-separation-to-roadmap.md`

> **Superseded Notice**
> This ADR is retained as a historical record of the original decision.
> For current MVP scope and implementation requirements, use `ADR-002` as the authoritative decision.

## Owners / Decision Makers
- Engineering

## Context
The MVP is a CLI-first personal finance application with local SQLite persistence. The project also needs clear extensibility toward server transport without mixing concerns between command handling, business rules, and persistence implementation.

Without explicit architecture boundaries, CLI command code can directly depend on database adapters, making testing harder and future server transport adoption expensive.

## Decision
Adopt Hexagonal (Ports & Adapters) architecture in Go with these layers:
- `domain` for entities/value objects/rules
- `application` for use cases and input/output ports
- `transport` for adapters (`cli`, `server`)
- `infrastructure` for adapter implementations (SQLite/config/observability)

For MVP:
- Release artifact is CLI-first.
- Server runtime exposure is optional.
- Server boundaries remain mandatory as scaffolding (`cmd/server`, `internal/transport/server`) to preserve separation and forward compatibility.

## Options Considered
1. Minimal monolith by feature folders without strict ports/adapters.
2. Layered architecture without explicit ports.
3. Hexagonal architecture with explicit ports/adapters (chosen).

## Rationale
- Keeps dependency direction explicit and testable.
- Allows CLI and server transports to reuse identical use cases.
- Reduces coupling to SQLite and transport/framework details.
- Aligns with the chosen team reference (`go-clean-arch-poc`).

## Consequences
### Positive
- Better maintainability and testability for core business logic.
- Easier future server/API rollout without domain refactor.
- Clear package boundaries for contributors.

### Negative
- More initial boilerplate (ports, adapters, DTO mapping).
- Requires discipline to enforce dependency rules.

### Operational / Maintenance
- CI/lint/review should guard against forbidden imports from core to adapters.
- Scaffolding directories may exist before full server implementation.

## Implementation Notes
- Entrypoints: `cmd/cli`, `cmd/server`.
- Transport adapters: `internal/transport/cli`, `internal/transport/server`.
- Core layers must not import transport/infrastructure concrete packages.

## Related Documents
- `docs/trd/001-infinita-personal-finance-app-mvp-cli.md`
- `docs/prd/001-infinita-personal-finance-app.md`
- `https://raw.githubusercontent.com/handiism/go-clean-arch-poc/8a94e96666ed9715cabaa46ede768163a86ebe6a/README.md`
