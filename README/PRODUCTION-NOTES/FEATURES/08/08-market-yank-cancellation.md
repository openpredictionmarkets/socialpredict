---
title: Market Yank And Cancellation
document_type: production-notes
domain: features
author: Patrick Delaney
updated_at: 2026-06-04T00:00:00Z
updated_at_display: "Thursday, June 4, 2026"
update_reason: "Start the feature spec for admin market yanking, cancellation refunds, visibility, and audit posture."
status: draft
---

# Market Yank And Cancellation

## Purpose

Admins need a platform-level way to yank a market when it should no longer operate as an ordinary market. Examples include abusive content, invalid market construction, moderator misconduct, duplicated markets, broken resolution criteria, legal/safety concerns, or an operational emergency.

A yank is not a hard delete and is not an ordinary creator resolution. It is an admin governance action that cancels the market, records why it happened, stops further trading, and performs a cancellation accounting process that makes participants economically whole according to a clearly specified rule.

## Feature Artifact Map

This directory keeps the yanked-market feature work together:

- [08-market-yank-cancellation.md](./08-market-yank-cancellation.md): feature overview, open questions, and acceptance criteria.
- [DESIGN.md](./DESIGN.md): domain, accounting, visibility, and architecture design artifact.
- [PLAN.md](./PLAN.md): implementation checklist and PR sequencing plan.

## Ubiquitous Language

- Yank: admin action that removes a market from normal operation and transitions it to cancelled.
- Cancelled market: market that is no longer tradable or resolvable through ordinary moderator flow because an admin yanked it.
- Cancellation refund: accounting process that reverses participant exposure caused by the market.
- Obfuscated cancelled market: cancelled market whose public title/description/labels are hidden or replaced because the original content is unsafe or abusive.
- Public cancellation notice: public explanation that the market was yanked and why, safe to show to participants.
- Private cancellation reason: operator/admin-only rationale that may contain details not suitable for public display.
- Cancellation audit: append-only record of actor, market, reason, visibility choice, accounting outcome, and timestamp.

## Product Problem

Moderator mode and market stewardship give admins control over who can propose and steward markets, but they do not yet provide a safe emergency path for already-published markets.

Without a yank feature, operators must choose between inadequate workarounds:

- resolve the market as `N/A`, even when the market should be explicitly cancelled
- manually edit database state, which risks accounting and audit errors
- leave harmful or invalid content visible
- hard-delete a market, which destroys auditability and may strand participant positions

## Target Behavior

Admins can yank a market from the admin market governance UI.

Yanking should:

- require admin authentication
- require a reason
- optionally require public-safe explanation text
- optionally choose whether to obfuscate public market details
- transition the market to `cancelled`
- immediately block buying, selling, resolution, and stewardship actions that assume a live market
- preserve immutable creator attribution and stewardship history
- persist a cancellation audit row
- run cancellation accounting in one transaction
- surface cancelled markets distinctly from resolved markets
- keep participant-facing financial history coherent

## Visibility Policy

There are two likely visibility modes.

### Public Cancellation Notice

Default for ordinary invalid or duplicate markets.

The market page remains accessible, but the page clearly says the market was cancelled by admin action. Existing title, description, labels, creator, steward, and audit-safe cancellation reason can remain visible.

### Obfuscated Cancellation Notice

Used when the market title, description, or labels contain abusive, illegal, unsafe, or troll content.

The market page remains accessible for audit and user history, but public content fields are replaced with neutral copy such as:

```text
This market was cancelled by platform administrators.
```

The private/original market text should remain available only to admins through an audit/governance view if legally and operationally appropriate.

## Status And Display

Cancelled should be a first-class lifecycle state, not a disguised resolved market.

Public display can show:

- status: `cancelled`
- outcome: `N/A` or `not applicable`
- trading: disabled
- resolution: disabled
- explanation: public cancellation notice

Using `resolved as N/A` as the only representation is ambiguous because it hides the difference between ordinary no-result resolution and platform intervention. If backward compatibility requires an `N/A` outcome for old components, the domain state should still remain `cancelled`.

## Accounting Problem

The refund rule needs explicit design before implementation.

The initial intuition is: take the total participant exposure caused by this market, claw back market-specific payouts/recoveries if needed, refund users so net market effect becomes zero, and mark market-specific values as zero.

This is accounting-sensitive and must not be guessed in a handler.

Questions to answer before implementation:

- What exactly counts as user spend after buys, fees, subsidies, and partial sells?
- Are fees refunded?
- Are creator proposal costs refunded on yank? If yes, always or only for admin-error cases?
- Are trader bonuses or subsidies reversed?
- How are users handled if their current account balance cannot absorb a clawback?
- Is cancellation refund based on net unrecovered exposure, gross buys, or some ledger-derived transaction replay?
- Should cancellation create explicit compensating transactions rather than mutating historical bets?
- Should cancelled market positions remain visible as cancelled/zeroed rather than deleted?

## Recommended Accounting Direction

Prefer compensating ledger entries over destructive mutation.

The safer model is:

- retain original bets and sales as historical facts
- calculate each participant's net market-specific financial effect
- write explicit cancellation refund/clawback transactions that make the market-specific net effect zero according to policy
- mark the market lifecycle as cancelled
- make market positions display as cancelled and no longer economically active

Do not hard-delete bets or rewrite old bet rows unless a later accounting design proves that is both correct and auditable.

## Rules

- Only admins can yank markets.
- Admins cannot yank a market without a reason.
- A market can be yanked from proposed, published, closed, or possibly resolved states only if the domain rules explicitly allow it.
- Rejected proposal handling should remain separate unless later design chooses to let admins cancel rejected proposals.
- Cancelled markets cannot be traded, sold, resolved, or reassigned for operational stewardship unless the action is cancellation-audit-specific.
- Cancelled markets should remain queryable for admins.
- Public visibility depends on the cancellation visibility mode.
- Cancellation accounting must be transaction-scoped.
- Cancellation audit must be transaction-scoped with market update and accounting writes.

## API Shape Candidate

Candidate admin route:

```http
PATCH /v0/admin/markets/{id}/yank
```

Candidate request body:

```json
{
  "reason": "Resolution criteria are invalid and cannot be fairly adjudicated.",
  "publicReason": "This market was cancelled because the resolution criteria were invalid.",
  "obfuscatePublicContent": false,
  "confirm": true
}
```

Candidate response:

```json
{
  "ok": true,
  "result": {
    "market": {
      "id": 123,
      "lifecycleStatus": "cancelled",
      "status": "cancelled",
      "resolutionResult": "N/A"
    },
    "refundSummary": {
      "usersAffected": 42,
      "totalRefunded": 1200,
      "totalClawedBack": 0
    }
  }
}
```

The exact accounting response should wait for the accounting design.

## Frontend Shape Candidate

Admin market governance should expose a yank action from the stewardship/governance view for eligible markets.

The modal should show:

- market title, creator, steward, status, and volume summary
- warning that yanking is irreversible
- required private reason
- optional public reason
- public content visibility choice
- refund/accounting preview if the backend can provide one
- confirmation checkbox or typed confirmation

Public market pages should show cancelled-state copy and hide trade/sell/resolve controls.

## Acceptance Criteria

Baseline acceptance criteria:

- Admin can yank an eligible market through an API.
- Non-admin users cannot yank markets.
- Yank requires a reason.
- Yank records an audit row.
- Yank transitions lifecycle to `cancelled`.
- Cancelled markets cannot accept buy/sell requests.
- Cancelled markets cannot be resolved through ordinary steward resolution.
- Cancelled markets appear in admin governance search.
- Cancelled markets do not appear as active tradable markets.
- Public market page communicates cancellation state.
- If obfuscation is enabled, public unsafe fields are replaced with neutral copy.
- Cancellation accounting is covered by unit tests and transaction tests before production use.

## Out Of Scope For First Planning Pass

- Implementing the accounting math.
- Adding AI content moderation.
- Hard deletion.
- General-purpose legal hold tooling.
- Bulk yanking multiple markets.
- Automatic yank decisions.
- Reopening a cancelled market.
