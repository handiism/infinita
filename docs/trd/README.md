# TRD Writing Standard

Use Technical Requirement Documents (TRDs) to translate product needs into concrete technical requirements for implementation.

## Purpose

A TRD should answer:
- What technical problem are we solving?
- What system behavior is required?
- What constraints must engineering satisfy?
- How will the solution be validated?

## Writing Principles

- Write in clear, precise English.
- Focus on technical requirements, not decision history.
- Map technical requirements back to PRD needs when applicable.
- Make all requirements specific and verifiable.
- Distinguish requirements from implementation notes.
- Keep major architecture decisions in ADRs, not embedded implicitly in the TRD.

## Required Metadata

Every TRD should include:
- Document ID
- Title
- Owner
- Reviewers
- Status: Draft / In Review / Approved / Deprecated
- Version
- Created date
- Last updated date
- Related PRD
- Related ADRs

## Required Sections

1. **Title & Metadata**
2. **Context**
   - Feature summary
   - Related PRD
3. **Technical Goals**
4. **Scope**
   - In scope
   - Out of scope
5. **Requirements Mapping**
   - Product requirement to technical requirement mapping
6. **Architecture Overview**
   - Key components, services, and boundaries
7. **Technical Requirements**
   - Frontend
   - Backend
   - Data
   - API
   - Infrastructure
   - Security
8. **Non-Functional Requirements**
   - Performance
   - Reliability
   - Scalability
   - Observability
   - Compliance
9. **API / Data Contracts**
10. **Dependencies**
11. **Technical Risks**
12. **Testing Strategy**
13. **Rollout Plan**
14. **Rollback Plan**
15. **Open Questions**
16. **References**

## Requirement Rules

- Use unique IDs such as `TRD-API-001`, `TRD-BE-001`, or `TRD-NFR-001`.
- Write requirements in a way that can be checked by testing, observability, or inspection.
- State thresholds explicitly when relevant, for example latency, throughput, or retention.
- Record assumptions clearly.
- If a requirement depends on a major design choice, link the related ADR.

## Example Requirement Format

- `TRD-API-001`: The system must expose `POST /v1/orders` to create a new order.
- `TRD-NFR-001`: `POST /v1/orders` must achieve p95 latency of 500 ms or less under normal production load.
- `TRD-SEC-001`: Sensitive data must be encrypted in transit and at rest.

## Review Checklist

- Is the technical scope explicit?
- Are product requirements mapped to technical requirements?
- Are performance and security constraints measurable?
- Are dependencies and risks identified?
- Is the testing strategy sufficient?
- Are rollout and rollback plans defined?
- Are major architecture decisions linked to ADRs?

## Recommended Template

```md
# TRD-XXX: [Feature Name]

## Metadata
- Owner:
- Reviewers:
- Status:
- Version:
- Created:
- Updated:
- Related PRD:
- Related ADRs:

## Context

## Technical Goals

## Scope
### In Scope
### Out of Scope

## Requirements Mapping

## Architecture Overview

## Technical Requirements

## Non-Functional Requirements

## API / Data Contracts

## Dependencies

## Technical Risks

## Testing Strategy

## Rollout Plan

## Rollback Plan

## Open Questions

## References
```
