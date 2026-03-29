# ADR-002: Defer Server Separation to Roadmap (CLI-Only MVP)

## Status
Superseded

Superseded by: `docs/adr/ADR-005-enable-server-transport-with-embedded-architecture.md`

## Date
2026-03-28

## Owners / Decision Makers
- Engineering

## Context
TRD-001 originally required mandatory server scaffolding (`cmd/server`, `internal/transport/server`) even though MVP release is CLI-first. This increased implementation scope for MVP and reduced focus on core CLI value delivery.

Product direction for current phase is to prioritize CLI delivery and postpone server separation work.

## Decision
For current MVP:
- Implement CLI-only runtime and package scope.
- Keep Hexagonal principles for implemented layers (`domain`, `application`, `transport/cli`, `infrastructure`).
- Defer explicit server package separation (`cmd/server`, `internal/transport/server`) to roadmap.

When server work is prioritized in future phase:
- Add dedicated server entrypoint and transport adapter.
- Reuse the same application use cases and contracts used by CLI.

## Options Considered
1. Keep mandatory server scaffolding in MVP.
2. Remove all server considerations entirely.
3. Keep CLI-only MVP and move server separation into roadmap (chosen).

## Rationale
- Reduces MVP scope and delivery risk.
- Keeps architecture clean for current implementation without premature scaffolding.
- Preserves forward path for server via roadmap commitment.

## Consequences
### Positive
- Faster MVP implementation and simpler codebase for current phase.
- Clearer acceptance criteria for “done” in MVP.

### Negative
- Future server enablement requires additional structure work later.

### Operational / Maintenance
- Planning docs must explicitly track server separation as roadmap item.
- Future ADR/TRD update will be needed when server implementation starts.

## Implementation Notes
- Current required entrypoint: `cmd/cli`.
- Current required transport adapter: `internal/transport/cli`.
- Server entrypoint/adapter remain roadmap targets.

## Related Documents
- `docs/trd/001-infinita-personal-finance-app-mvp-cli.md`
- `docs/adr/ADR-001-adopt-hexagonal-clean-architecture-go.md`
- `docs/prd/001-infinita-personal-finance-app.md`
