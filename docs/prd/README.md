# PRD Writing Standard

Use Product Requirement Documents (PRDs) to define the business problem, product goals, feature scope, and success criteria.

## Purpose

A PRD should answer:
- What problem are we solving?
- Who is this for?
- What outcome do we want?
- What is in scope and out of scope?
- How will we measure success?

## Writing Principles

- Write in clear, simple English.
- Focus on **why** and **what**, not implementation details.
- Separate facts, assumptions, and open questions.
- Use bullets and short sections instead of long paragraphs.
- Make every requirement testable.
- Define explicit non-goals.
- Prefer measurable outcomes over vague statements.

## Required Metadata

Every PRD should include:
- Document ID
- Title
- Owner
- Reviewers
- Status: Draft / In Review / Approved / Deprecated
- Version
- Created date
- Last updated date
- Related documents

## Required Sections

1. **Title & Metadata**
2. **Problem Statement**
   - What is the problem?
   - Who is affected?
   - Why does it matter now?
3. **Goals**
   - Desired product or business outcomes
4. **Non-Goals**
   - What this document does not cover
5. **Target Users / Personas**
6. **User Stories / Use Cases**
7. **Scope**
   - In scope
   - Out of scope
8. **Functional Requirements**
9. **Non-Functional Requirements**
10. **Success Metrics**
11. **Acceptance Criteria**
12. **Risks & Dependencies**
13. **Open Questions**
14. **References**

## Requirement Rules

- Use unique IDs such as `PRD-FR-001` and `PRD-NFR-001`.
- Each requirement must describe observable behavior or constraint.
- Avoid technical design decisions unless they are product constraints.
- If implementation decisions are needed, capture them in TRD or ADR.

## Example Requirement Format

- `PRD-FR-001`: Users can sign in with email and password.
- `PRD-FR-002`: Users can reset their password through email verification.
- `PRD-NFR-001`: The sign-in flow should complete within 3 seconds for 95% of requests.

## Example Success Metrics

- Activation rate increases by 15% within one quarter.
- Checkout completion rate reaches at least 70%.
- Core task success rate remains above 98%.

## Review Checklist

- Is the problem clearly defined?
- Are goals outcome-focused?
- Is scope explicit?
- Are requirements testable?
- Are success metrics measurable?
- Are risks and dependencies listed?
- Are open questions visible?

## Recommended Template

```md
# PRD-XXX: [Feature Name]

## Metadata
- Owner:
- Reviewers:
- Status:
- Version:
- Created:
- Updated:
- Related Documents:

## Problem Statement

## Goals

## Non-Goals

## Target Users / Personas

## User Stories / Use Cases

## Scope
### In Scope
### Out of Scope

## Functional Requirements

## Non-Functional Requirements

## Success Metrics

## Acceptance Criteria

## Risks & Dependencies

## Open Questions

## References
```
