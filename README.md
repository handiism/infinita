# Infinita

CLI-first personal finance app. Track income, expenses, budgets, and summaries — all stored locally on your device.

## Status

Pre-implementation. Documentation is ready; code is in progress.

## Features (MVP)

- **Transaction logging** — add income/expense entries with category, date, and optional description
- **Category management** — default categories + custom categories (case-insensitive)
- **Monthly budgets** — set spending limits per category with over-limit warnings
- **Reports** — daily and monthly summaries with income, expense, net balance, and top spending categories
- **Initial balance** — optional starting balance for accurate closing-balance calculations
- **Local-only storage** — SQLite on-device, no cloud sync, no telemetry by default

## Architecture

Hexagonal (Ports & Adapters) in Go. Dependencies point inward:

```
transport (CLI) → application (use cases + ports) → domain (entities + rules)
infrastructure (SQLite) → application / domain
```

See the full architecture diagram: [`docs/diagrams/trd/001-trd-system-architecture.mmd`](docs/diagrams/trd/001-trd-system-architecture.mmd)

## Documentation

| Document | Path |
|---|---|
| Product Requirements (PRD) | [`docs/prd/001-infinita-personal-finance-app.md`](docs/prd/001-infinita-personal-finance-app.md) |
| Technical Requirements (TRD) | [`docs/trd/001-infinita-personal-finance-app-mvp-cli.md`](docs/trd/001-infinita-personal-finance-app-mvp-cli.md) |
| ADR-002: CLI-only MVP | [`docs/adr/ADR-002-defer-server-separation-to-roadmap.md`](docs/adr/ADR-002-defer-server-separation-to-roadmap.md) |
| Data Model (ERD) | [`docs/diagrams/trd/001-trd-data-model-erd.mmd`](docs/diagrams/trd/001-trd-data-model-erd.mmd) |
| Main User Flow | [`docs/diagrams/prd/001-prd-main-flow.mmd`](docs/diagrams/prd/001-prd-main-flow.mmd) |
| PRD Writing Standard | [`docs/prd/README.md`](docs/prd/README.md) |
| TRD Writing Standard | [`docs/trd/README.md`](docs/trd/README.md) |
| ADR Writing Standard | [`docs/adr/README.md`](docs/adr/README.md) |

## License

All rights reserved.
