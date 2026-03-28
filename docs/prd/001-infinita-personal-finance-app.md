# PRD-001: Infinita Personal Finance App (MVP CLI)

## Metadata
- Document ID: PRD-001
- Title: Infinita Personal Finance App (MVP CLI)
- Owner:
- Reviewers:
- Status: Draft
- Version: 0.8
- Created date: 2026-03-28
- Last updated date: 2026-03-28
- Related Documents:
  - `docs/prd/README.md`
  - `docs/trd/001-infinita-personal-finance-app-mvp-cli.md`
  - `docs/adr/ADR-002-defer-server-separation-to-roadmap.md`

## Problem Statement
Many individuals struggle to consistently track income, expenses, budgets, and financial summaries because existing tools are either too complex or not private enough.

The users most affected are general individual users, budget-conscious users, and privacy-focused users who want lightweight personal finance tracking without accounting-heavy workflows or third-party data sharing.

This matters now because the MVP needs a simple, private, English-first workflow that lowers the barrier to building a consistent financial tracking habit.

## Goals
- Enable fast daily transaction logging through a CLI-first MVP.
- Help users understand spending through categories, budgets, and simple summaries.
- Keep MVP storage local-only on user device with no telemetry by default.
- Provide a clear English-only CLI experience with consistent terminology.

## Non-Goals
- Automated bank or e-wallet integrations.
- Investment, tax, or advanced bookkeeping features.
- AI-based spending insights in MVP.
- GUI web or mobile dashboard in MVP.
- Server transport/API exposure in MVP.

## Target Users / Personas
- **General individual users**: want to start tracking personal finances without complex accounting workflows.
- **Budget-conscious users**: want category-level spending control versus monthly targets.
- **Privacy-focused users**: want financial data control with local-only storage in MVP and no third-party data sharing.

## User Stories / Use Cases
1. **As a user**, I want to add income and expense transactions quickly so I can consistently log daily finances.
2. **As a user**, I want to categorize transactions so I can understand where my money goes.
3. **As a user**, I want to set monthly budgets by category so I can control spending.
4. **As a user**, I want to view cashflow summaries and simple reports so I can understand my financial condition.
5. **As a user**, I want the app interface to be in clear English so I can use commands and messages consistently.
6. **As a user**, I want my data to remain private and secure so my financial information is protected.
7. **As a user**, I want to optionally set an initial balance so my summaries can start from my real current financial position.
8. **As a user**, I want to clearly see storage behavior in settings so I know my data is stored locally in MVP.

## Main Flow Diagram
- Mermaid source: `docs/diagrams/prd/001-prd-main-flow.mmd`

## Scope
### In Scope
- Add income and expense transactions via CLI.
- Store transactions locally on-device in MVP.
- Support required transaction fields: type, amount, category, and date; description is optional.
- Use fixed MVP currency `IDR` with decimal amount input allowed up to 2 fractional digits.
- Support optional initial balance input for first-time setup or manual initialization.
- Provide default categories and support custom categories.
- Set and view monthly budget limits per category.
- View daily summaries including income, expense, and net balance.
- View monthly summaries including income, expense, net balance, top spending categories, and closing balance.
- Provide English-only CLI help, output, and error messages with consistent terminology.
- Support example CLI commands such as `add`, `list`, `budget`, and `report`; these examples illustrate the intended CLI interaction model and are not final command syntax.

### Out of Scope
- Editing transactions in MVP.
- Deleting transactions in MVP.
- Full cross-device synchronization in MVP.
- Automated bank or e-wallet integrations.
- Investment, tax, or advanced bookkeeping features.
- AI-based spending insights in MVP.
- GUI web or mobile dashboard in MVP.
- Server transport/API exposure in MVP.

## Functional Requirements
- `PRD-FR-001`: Users can add income transactions through the CLI.
- `PRD-FR-002`: Users can add expense transactions through the CLI.
- `PRD-FR-003`: The system requires transaction type, amount, category, and date for each transaction; description is optional. In MVP, currency is fixed to `IDR` and amount input supports up to 2 fractional digits.
- `PRD-FR-004`: The system provides default transaction categories.
- `PRD-FR-005`: Users can create and use custom transaction categories.
- `PRD-FR-006`: Users can list stored transactions.
- `PRD-FR-007`: Users can filter listed transactions by category.
- `PRD-FR-008`: Users can set monthly budget limits by category.
- `PRD-FR-009`: Users can view remaining budget and over-limit warnings by category.
- `PRD-FR-010`: Users can view daily summaries with period totals: income, expense, and net balance for the requested day only.
- `PRD-FR-011`: Users can view monthly summaries with period totals: income, expense, net balance, closing balance, and top spending categories for the requested month only.
- `PRD-FR-012`: CLI help text, command output, and error messages are presented in English using consistent terminology.
- `PRD-FR-013`: User financial data is stored locally on the user device in MVP.
- `PRD-FR-014`: The MVP does not require bank integration.
- `PRD-FR-015`: Users can set an optional initial balance, and the value is used as the starting point for cumulative closing-balance calculations shown in summaries.
- `PRD-FR-016`: CLI settings must clearly show that active storage mode in MVP is local.

## Non-Functional Requirements
- `PRD-NFR-001`: The system validates that transaction amount is greater than 0, supports up to 2 fractional digits, and rejects invalid input.
- `PRD-NFR-002`: Sensitive config and user data use secure local file permissions and documented local data handling rules.
- `PRD-NFR-003`: Core CLI flows represented by example commands such as `add`, `list`, `budget`, and `report` achieve at least a 99% success rate in defined MVP test scenarios.
- `PRD-NFR-004`: New transaction entry completes within 15 seconds for familiar users.
- `PRD-NFR-005`: No telemetry is enabled by default in MVP.

## Measurement & Privacy Strategy
- Behavioral success metrics are measured only for users who explicitly opt in to anonymous analytics.
- Opt-in analytics are off by default and require clear consent during setup or via a dedicated settings command.
- Analytics events must exclude financial payloads (no amounts, descriptions, categories, or raw transaction content).
- Minimum analytics payload for KPI measurement: anonymous installation ID, event type (`transaction_added`, `budget_used`), and timestamp bucket.
- Users can review and revoke analytics consent at any time.
- Product KPIs that depend on user behavior are evaluated against the opt-in cohort; if cohort size is insufficient, fallback to structured usability/beta test cohorts.

## Success Metrics
- At least 60% of first-week active users in the analytics opt-in cohort log at least 5 transactions.
- At least 40% of users in the analytics opt-in cohort who create budgets use the budgeting feature at least once per week.
- At least 99% of MVP command execution scenarios covering core CLI flows complete without runtime errors across at least 1,000 executions.
- New transaction entry time is 15 seconds or less for familiar users.
- 100% of users use local-only storage in MVP, and telemetry remains off by default.

## Acceptance Criteria
- [ ] Given valid input, when the user runs an example transaction-add command, then the transaction is saved and confirmation is shown.
- [ ] Given existing categories, when the user runs an example transaction-listing or filtering command, then category filtering works correctly.
- [ ] Given monthly budgets are set, when the user runs an example budgeting command after expenses increase, then remaining budget and over-limit warnings are shown.
- [ ] Given a reporting flow, when the user runs an example daily summary command, then period totals (income, expense, net balance) are accurate for the requested day bucket.
- [ ] Given a reporting flow, when the user runs an example monthly summary command, then period totals and top spending categories are accurate for the requested month bucket and closing balance is accurate as of month end.
- [ ] Given CLI core flows, when help, output, and errors are rendered, then wording is in English and terminology is consistent.
- [ ] Given default app behavior, when data is persisted, then data stays local with secure file permissions and no telemetry by default.
- [ ] Given invalid transaction input, when the user submits an amount less than or equal to 0 or with more than 2 fractional digits, then the system rejects the input and shows a validation error.
- [ ] Given first-time setup or manual initialization, when the user provides an initial balance value, then the system stores it and uses it as the starting point for closing-balance calculations in summaries; when omitted, the system defaults initial balance to 0.
- [ ] Given storage settings in MVP, when the user checks storage configuration, then storage mode is shown as local and data is stored only on-device.

## Risks & Dependencies
### Risks
- **CLI adoption risk**: general users may prefer visual or mobile UI.
- **UX copy consistency risk**: inconsistent English terminology may confuse users.
- **Privacy implementation risk**: secure local storage depends on the chosen stack and file handling approach.
- **Reporting consistency risk**: budget and summary calculations require strong test coverage.
- **Scope risk**: MVP complexity can grow if non-MVP features are added prematurely.

### Dependencies
- A CLI-capable application framework or runtime must be selected for implementation.
- Local storage and file-permission handling must support private on-device persistence.
- Budgeting and reporting flows depend on consistent category and transaction data models.

## Open Questions
- Should the MVP support filtering transactions by date in addition to category?
- Should custom categories be created implicitly during transaction entry or through a separate flow?

## References
- `docs/prd/README.md`
- `docs/diagrams/prd/001-prd-main-flow.mmd`

## Roadmap
**MVP**
- Add transactions
- List and filter transactions
- Default and custom categories
- Monthly budgeting per category
- Basic daily and monthly cashflow summaries
- English-only command and output experience
- Local-only storage in MVP
- Optional initial balance setup

**v1.1**
- CSV and TXT report export
- Expanded transaction filtering and search
- Logging reminders
- Local backup and restore
- Server/API capability as a roadmap architecture milestone

**v2.0**
- Cross-device sync
- Web or mobile GUI dashboard
- Automated insights and forecasting
- Shared household budgeting
