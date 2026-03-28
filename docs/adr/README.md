# ADR Writing Standard

Use Architecture Decision Records (ADRs) to document important architecture and engineering decisions, including context, alternatives, rationale, and consequences.

## Purpose

An ADR should answer:
- What decision was made?
- What context led to the decision?
- What alternatives were considered?
- Why was this option chosen?
- What trade-offs and consequences follow?

## Writing Principles

- One ADR should capture one significant decision.
- Write decisions in clear, specific language.
- Include rejected or unselected options.
- Be explicit about trade-offs.
- Keep decision records immutable; if the decision changes, create a new ADR and mark the old one as superseded.

## Required Metadata

Every ADR should include:
- ADR ID
- Title
- Status: Proposed / Accepted / Rejected / Superseded / Deprecated
- Date
- Owners or decision makers
- Related documents

## Required Sections

1. **Title**
   - Example: `Use PostgreSQL for transactional data`
2. **Status**
3. **Context**
   - Problem statement
   - Constraints
   - Relevant background
4. **Decision**
   - Clear statement of the chosen approach
5. **Options Considered**
   - Option A
   - Option B
   - Option C
6. **Rationale**
   - Why the chosen option won
7. **Consequences**
   - Positive outcomes
   - Negative outcomes
   - Operational or maintenance implications
8. **Implementation Notes**
   - Optional
9. **Related Documents**

## Naming Convention

Use sequential IDs and short descriptive slugs.

Examples:
- `ADR-001-use-postgresql.md`
- `ADR-002-adopt-feature-flags.md`
- `ADR-003-use-event-driven-notifications.md`

## Example Decision Format

- **Context**: The system needs a reliable transactional database with strong relational querying.
- **Decision**: Use PostgreSQL as the primary transactional datastore.
- **Options Considered**: PostgreSQL, MySQL, MongoDB.
- **Consequences**: Strong relational support and mature tooling, with additional operational tuning needed at higher scale.

## Review Checklist

- Is the decision clearly stated?
- Is the context sufficient?
- Are alternatives documented?
- Is the rationale explicit?
- Are trade-offs honest and specific?
- Is the status correct?
- Are related PRDs or TRDs linked?

## Recommended Template

```md
# ADR-XXX: [Decision Title]

## Status

## Date

## Owners / Decision Makers

## Context

## Decision

## Options Considered

## Rationale

## Consequences

## Implementation Notes

## Related Documents
```
