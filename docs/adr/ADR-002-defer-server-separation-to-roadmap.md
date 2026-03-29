# ADR-002: Defer Server Separation to Roadmap (CLI-Only MVP)

## Status
- Superseded
- Superseded by: `docs/adr/ADR-005-enable-server-transport-with-embedded-architecture.md`

## Context
- Mandatory server scaffolding increased MVP scope while product direction prioritized CLI delivery first.
- The team needed a smaller implementation target without dropping the architectural direction entirely.

## Decision
- Keep MVP focused on CLI-only runtime.
- Retain Hexagonal boundaries for implemented layers.
- Defer dedicated server entrypoint and transport separation to a later phase.

## Consequences
- MVP scope became smaller and easier to deliver.
- Future server work required a later ADR and follow-up technical updates.

## Related Documents
- `docs/trd.md`
- `docs/prd.md`
- `docs/adr/ADR-001-adopt-hexagonal-clean-architecture-go.md`
- `docs/adr/ADR-005-enable-server-transport-with-embedded-architecture.md`
