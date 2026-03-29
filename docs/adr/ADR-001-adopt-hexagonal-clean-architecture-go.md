# ADR-001: Adopt Hexagonal Clean Architecture for Go CLI/Server Boundaries

## Status
- Superseded
- Superseded by: `docs/adr/ADR-002-defer-server-separation-to-roadmap.md`

## Context
- The project needed clear boundaries between CLI handling, business logic, and persistence.
- Without explicit boundaries, transport code could couple directly to database adapters and make future changes harder.

## Decision
- Use Hexagonal architecture with `domain`, `application`, `transport`, and `infrastructure` layers.
- Keep dependency direction inward.
- At the time of this decision, server scaffolding was included to preserve transport separation.

## Consequences
- Package boundaries became explicit and testable.
- This ADR was later replaced when MVP scope moved away from mandatory server scaffolding.

## Related Documents
- `docs/trd.md`
- `docs/prd.md`
- `docs/adr/ADR-002-defer-server-separation-to-roadmap.md`
