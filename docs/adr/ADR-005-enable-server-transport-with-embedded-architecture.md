# ADR-005: Enable Server Transport with Embedded Architecture

## Status
- Accepted

## Context
- After the CLI-only MVP phase, the project needed a real transport boundary between the CLI and business logic.
- The team wanted that separation without giving up single-binary local deployment.

## Decision
- Run an embedded HTTP server inside the CLI process.
- Make the CLI a thin client that talks to the server over localhost HTTP.
- Keep the API contract in `docs/api/openapi.yaml`.
- Use stdlib `net/http` for server transport.

## Consequences
- CLI and server transport are separated more cleanly.
- The project gains local HTTP overhead and server lifecycle management.
- The API contract becomes a stronger integration point for future clients.

## Related Documents
- `docs/prd.md`
- `docs/trd.md`
- `docs/api/openapi.yaml`
- `docs/adr/ADR-001-adopt-hexagonal-clean-architecture-go.md`
- `docs/adr/ADR-002-defer-server-separation-to-roadmap.md`
