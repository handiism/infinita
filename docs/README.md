# Infinita Documentation

Simple docs structure:

```txt
 repo/
 ├── docs/
 │   ├── prd/
 │   ├── trd/
 │   ├── diagrams/
 │   │   ├── prd/
 │   │   └── trd/
 │   ├── adr/
 │   ├── security/
 │   └── README.md
```

## PRD
- `/docs/prd/001-infinita-personal-finance-app.md`
- `/docs/prd/README.md` — PRD writing standard

## TRD
- `/docs/trd/001-infinita-personal-finance-app-mvp-cli.md`
- `/docs/trd/README.md` — TRD writing standard

## API
- (No active API documentation yet)

## Diagrams
- `/docs/diagrams/prd/001-prd-main-flow.mmd` — PRD main flow Mermaid source
- `/docs/diagrams/trd/001-trd-system-architecture.mmd` — TRD system architecture Mermaid source (includes sqlc and golang-migrate)
- `/docs/diagrams/trd/001-trd-sequence-add-transaction.mmd` — TRD add transaction sequence Mermaid source
- `/docs/diagrams/trd/001-trd-sequence-monthly-report.mmd` — TRD monthly report sequence Mermaid source
- `/docs/diagrams/trd/001-trd-data-model-erd.mmd` — TRD data model ERD Mermaid source

## ADR
- `/docs/adr/README.md` — ADR writing standard
- `/docs/adr/ADR-001-adopt-hexagonal-clean-architecture-go.md` — Hexagonal architecture (Superseded by ADR-002)
- `/docs/adr/ADR-002-defer-server-separation-to-roadmap.md` — CLI-only MVP scope (Accepted)
- `/docs/adr/ADR-003-adopt-sqlc-golang-migrate-tiered-testing.md` — sqlc, golang-migrate, tiered testing strategy (Accepted)

## Security
- `/docs/security/local-data-handling.md` — Local data path, permission model, backup/restore, and deletion guidance
