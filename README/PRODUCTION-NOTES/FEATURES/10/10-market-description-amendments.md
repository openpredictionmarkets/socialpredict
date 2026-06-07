---
title: Market Description Amendments
document_type: production-notes
domain: features
author: Patrick Delaney
updated_at: 2026-06-06T00:00:00Z
updated_at_display: "Saturday, June 6, 2026"
update_reason: "Start the feature spec for immutable market titles, additive-only market description amendments, markdown-lite rendering, and amendment traceability."
status: draft
---

# Market Description Amendments

## Purpose

Market descriptions can contain resolution criteria, eligibility rules, or other contract-like language. Once a market is proposed or published, changing that language can change the meaning of the market.

SocialPredict therefore needs a governed way for moderators and admins to add clarification text without silently rewriting the original description.

This feature defines a description amendment system:

- market titles are immutable
- the original description remains the original contract text
- later description changes are additive-only amendments
- every amendment receives an explicit version and timestamp
- users can clearly see what was added and when
- amendment authority is tied to moderator stewardship and admin review
- markdown-lite formatting can be supported safely without allowing arbitrary HTML

## Feature Artifact Map

This directory keeps the market description amendment feature work together:

- [10-market-description-amendments.md](./10-market-description-amendments.md): feature overview, product behavior, and acceptance criteria.
- [DESIGN.md](./DESIGN.md): domain, markdown-lite, persistence, API, audit, and frontend design.
- [PLAN.md](./PLAN.md): backend-first implementation checklist and PR sequencing.

## Current Behavior

Current market descriptions are plain market fields:

- `markets.question_title` stores the market title.
- `markets.description` stores the original description.
- create-market validation limits descriptions to 2000 characters.
- the market page renders the description as escaped React text with preserved line breaks.
- the backend sanitizer allows only very limited HTML, but the current public market page does not render descriptions as HTML.
- there is no general market title or description update endpoint.

This means titles are effectively immutable today because no update route exists. The product should codify that behavior before adding any description update workflow.

## Problem Framing

A prediction market is partly a user interface object and partly a contract. The title, description, custom labels, tags, resolution date, and resolution criteria shape how users understand their risk.

If a moderator can silently edit a description after users have traded, then users cannot know which version of the market they agreed to. The platform needs traceability.

The design should avoid two bad outcomes:

1. Silent edits that mutate the original contract text.
2. Overly rigid descriptions that cannot be clarified when real-world ambiguity appears.

The amendment model solves this by preserving the original description and adding visible append-only clarifications.

## Ubiquitous Language

- Original description: the description supplied at market creation/proposal time.
- Amendment: append-only clarification or addition to the original description.
- Description version: chronological version number where version 1 is the original description and later versions are amendments.
- Markdown-lite: constrained formatting syntax for readable text, without arbitrary HTML.
- Steward: the moderator currently responsible for operational governance of the market.
- Amendment review: admin approval gate for amendments that affect published markets.
- Contract text: original description plus approved amendments, shown in order.

Avoid confusing:

- Amendment is not an edit to the original description.
- Amendment version is not the same as market lifecycle status.
- Markdown-lite is not arbitrary HTML.
- Stewardship is operational authority, not historical authorship.

## Product Rules

### Title Immutability

Market titles are immutable after market creation.

Rules:

- no moderator route can change a title
- no admin route should change a title in the baseline feature
- if future title correction is required, it should be a separate admin-only correction feature with explicit audit and user-facing disclosure

### Original Description

The original description is version 1.

Rules:

- keep the original `markets.description` value intact
- do not overwrite the original description when adding clarifications
- show the original description before amendments on public market pages
- keep the original description visible in admin review and audit contexts

### Additive Amendments

Later changes are appended as amendments.

Rules:

- amendment text is additive only
- amendments receive server-assigned versions, e.g. v2, v3, v4
- amendments are shown in chronological order
- amendments must show author, timestamp, and approval status where applicable
- amendments cannot delete or obscure prior text

### Published Market Governance

Published markets require stronger controls because users may have already traded.

Recommended baseline:

- current steward can propose an amendment
- admin can approve or reject the amendment
- approved amendments become visible in the public contract text
- rejected amendments remain visible only in admin/moderator governance views
- admin can create an amendment directly if needed, but it still records actor and reason

### Proposed Market Governance

Before publication, amendments can be treated as proposal clarifications.

Recommended baseline:

- creator/steward can append amendments while the market is proposed
- admin review sees the original description and all proposed amendments
- approving the market publishes the original description plus approved/proposal amendments

## Markdown-Lite Position

Descriptions should support markdown-lite, not arbitrary HTML.

Recommended allowed formatting:

- paragraphs and line breaks
- bold
- italic
- bullet lists
- numbered lists
- blockquotes
- inline code
- safe links using `http` or `https`

Recommended disallowed formatting:

- raw HTML
- scripts
- iframes
- embedded media
- images
- tables, initially
- custom CSS/classes/styles

The backend should validate and sanitize amendment text. The frontend can render markdown-lite, but it should not be the only enforcement point.

## User-Facing Display

Public market pages should clearly separate original and amended text.

Suggested display:

```text
Description
[original description]

Amendments
Amendment v2
Added June 6, 2026 by @moderator
[amendment body]

Amendment v3
Approved June 7, 2026 by @admin
[amendment body]
```

The UI should make amendments prominent enough that users do not mistake the original description as the full contract once amendments exist.

## Admin And Moderator UX

Moderator side:

- steward can open an amendment form from the market page or profile/governance area
- amendment form explains that the title is immutable and description changes are append-only
- markdown-lite preview is available before submission
- submitted amendment shows pending/approved/rejected state

Admin side:

- Market Review should include a description amendments tab or panel
- admin can review pending amendments
- admin can approve or reject with a reason
- audit trail should show all amendment decisions

## Acceptance Criteria

- Market titles cannot be changed by the amendment flow.
- Original market description is not overwritten.
- Amendments are persisted as separate records.
- Amendment versions are assigned server-side.
- Public market detail shows original description plus approved amendments.
- Pending/rejected amendments are not accidentally shown as approved contract text.
- Moderator authority respects current steward/suspended moderator rules.
- Published-market amendments require admin review unless an explicit admin-direct path is used.
- Markdown-lite is rendered safely and raw HTML is rejected or neutralized.
- OpenAPI documents all new amendment endpoints and response fields.
- Tests cover title immutability, append-only behavior, authorization, sanitization, and version ordering.

## Deferred Questions

- Should any amendment require user notification after publication?
- Should an amendment affect only future trades or the entire market contract historically?
- Should published-market amendments be blocked after a market closes?
- Should amendment text have a stricter length limit than the original description?
- Should admin-direct amendments bypass approval or simply auto-approve with an explicit reason?
- Should amendments be included in Open Graph descriptions or only on market pages?
