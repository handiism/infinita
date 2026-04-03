# ADR-003: Adopt sqlc, golang-migrate, and Tiered Testing Strategy

## Status
- Accepted

## Context
- The project needed type-safe SQL access, repeatable schema migration, and a clear testing strategy for SQLite-based development.
- Hand-written queries, manual schema changes, and unclear test boundaries would increase runtime risk and maintenance cost.

## Decision
- Use `sqlc` for SQL-to-Go code generation.
- Use `golang-migrate` for schema migration.
- Use four testing tiers: repository integration, service unit, HTTP handler, and command integration.
- Use `mattn/go-sqlite3` as the SQLite driver.

## Consequences
- Query safety and schema changes became more deterministic.
- The project now depends on code generation, migration files, and CGO-capable builds.
- Testing boundaries are clearer across infrastructure, business logic, and CLI flows.

## Related Documents
- `docs/trd.md`
- `docs/prd.md`
- `docs/adr/ADR-001-adopt-hexagonal-clean-architecture-go.md`
- `docs/adr/ADR-002-defer-server-separation-to-roadmap.md`
