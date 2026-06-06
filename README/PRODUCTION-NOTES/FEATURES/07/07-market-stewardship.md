---
title: Market Stewardship
document_type: production-notes
domain: features
author: Patrick Delaney
updated_at: 2026-06-03T00:00:00Z
updated_at_display: "Wednesday, June 3, 2026"
update_reason: "Define creator attribution versus operational market stewardship for moderator governance."
status: draft
---

# Market Stewardship

## Purpose

Market creators should remain immutable for attribution, proposal-cost history, and audit context. Operational authority over a market should be assignable by admins when the original moderator is suspended, inactive, conflicted, or otherwise unavailable to resolve or maintain the market.

## Ubiquitous Language

- Creator: immutable original market author.
- Steward: current moderator responsible for market-governance actions such as resolution and future edit/yank workflows.
- Stewardship reassignment: admin action that moves operational responsibility from one steward to another without changing the creator.
- Stewardship audit: append-only record of who reassigned stewardship, from whom, to whom, why, and when.

## Rules

- New markets default `steward_username` to `creator_username`.
- Existing markets backfill steward to creator through an additive migration.
- Creator attribution is not rewritten during stewardship changes.
- Admins can reassign stewardship to an active moderator only.
- Suspended moderators cannot receive stewardship and cannot use steward-only actions.
- The current steward, not necessarily the original creator, is the moderator allowed to resolve a market.
- Admins remain platform override authority.
- Reassignment must persist an audit row in the same unit of work as the market update.

## Design Alignment

This feature follows the canonical design plan at `/Users/patrick/Projects/spec-socialpredict-tasks/lib/design/design-plan.json`:

- Prediction Market Core Context owns market lifecycle and governance rules.
- Participant Account Context owns moderator status semantics.
- Backend Persistence and Migration Boundary owns additive timestamped Go migrations.
- Auditability requires reconstructable governance history rather than destructive creator rewrites.

## Implementation Notes

- Add `markets.steward_username` with default/backfill behavior.
- Add `market_stewardship_audits` for append-only reassignment facts.
- Add admin API: `PATCH /v0/admin/markets/{id}/steward`.
- Include `stewardUsername` in admin market review payloads and public market metadata.
- Keep reassignment UI controls as a later frontend slice unless immediate admin workflow requires it.
