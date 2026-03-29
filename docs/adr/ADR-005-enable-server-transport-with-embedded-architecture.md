# ADR-005: Enable Server Transport with Embedded Architecture

## Status

Accepted

## Date

2026-03-29

## Owners / Decision Makers

- Engineering

## Context

ADR-002 deferred server separation to roadmap. The CLI-only MVP is functionally complete. The next step is to add server transport so the CLI communicates with a local HTTP server instead of accessing the database directly.

This separation enables:
- Clear transport boundary between CLI and business logic
- Future ability to expose the API to external clients (web GUI, mobile, etc.)
- Testability: the server can be tested independently of the CLI
- Single binary deployment remains possible via embedded server

## Decision

Embed an HTTP server (stdlib `net/http`) as a goroutine within the CLI process.

Key decisions:
- **Architecture**: Embedded HTTP server using Go stdlib `net/http`. Server starts as a goroutine when CLI launches and stops on CLI exit. Single binary deployment preserved.
- **Transport**: HTTP/REST with JSON. Flat paths (e.g., `/transactions`, `/categories`). No version prefix since the API is internal-only.
- **CLI role**: Thin client. CLI no longer accesses SQLite directly. All operations go through the embedded HTTP server via localhost.
- **Server hosting**: Auto-managed by CLI. Server listens on a random available port. CLI polls `/health` until ready before executing commands. Graceful shutdown on CLI exit.
- **API contract**: OpenAPI 3.1 spec at `docs/api/openapi.yaml` as single source of truth for request/response shapes.
- **Response format**: JSend-style generic wrapper with `status` field (`success`/`fail`/`error`), `data`, and `meta`. Validation errors return array of error objects with per-field per-rule codes.
- **Framework**: stdlib `net/http` ServeMux (Go 1.22+ method-based routing). No external HTTP framework dependency.

## Options Considered

1. Separate server process spawned by CLI.
   - More complex lifecycle management, two processes to coordinate.
   - Rejected: unnecessary complexity for internal-only API.
2. Remote server deployment.
   - Adds network dependency, authentication, TLS complexity.
   - Deferred to future phase if external access is needed.
3. Embedded HTTP server as goroutine in CLI process (chosen).
   - Simple lifecycle, single process, single binary.
   - localhost-only communication, no authentication needed.

## Rationale

- Preserves single binary deployment (no separate server process to manage).
- Clean separation of concerns: CLI handles terminal I/O, server handles HTTP and business logic orchestration.
- Domain and application layers remain unchanged — only transport layer is added.
- Go 1.22+ ServeMux provides method-based routing without external dependencies.
- Random port allocation avoids conflicts on user machines.

## Consequences

### Positive

- CLI and server transport are cleanly separated in the codebase.
- Future external API exposure requires only exposing the listener (or adding TLS/auth).
- Server can be tested independently via HTTP handler tests.
- OpenAPI spec provides machine-readable API contract.
- Single binary deployment maintained.

### Negative

- CLI startup has slight latency (server boot + health check poll).
- HTTP overhead for local-only communication (negligible for CLI use case).
- All CLI error handling must map HTTP status codes to exit codes.

### Operational / Maintenance

- CI must verify OpenAPI spec consistency with implementation.
- Server goroutine lifecycle must be tested (startup, shutdown, panic recovery).
- Random port allocation must be handled gracefully (health check with timeout).

## Implementation Notes

- New directory: `internal/transport/server/` for HTTP handlers, router, DTOs.
- New directory: `internal/transport/cli/client/` for HTTP client implementation.
- New entrypoint: `cmd/server/main.go` for standalone server binary (development use).
- CLI entrypoint (`cmd/cli/main.go`) refactored: starts server goroutine, creates HTTP client, runs CLI commands, shuts down server on exit.
- OpenAPI spec: `docs/api/openapi.yaml`.

## Related Documents

- `docs/adr/ADR-002-defer-server-separation-to-roadmap.md` (superseded)
- `docs/adr/ADR-001-adopt-hexagonal-clean-architecture-go.md`
- `docs/api/openapi.yaml`
- `docs/trd/001-infinita-personal-finance-app-mvp-cli.md`
- `docs/prd/001-infinita-personal-finance-app.md`
