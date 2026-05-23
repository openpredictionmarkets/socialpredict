---
title: Long-Term Frontend PWA and Offline Platform
document_type: production-notes
domain: frontend
future: true
author: Patrick Delaney
updated_at: 2026-05-23T00:00:00Z
updated_at_display: "Saturday, May 23, 2026"
update_reason: "Move service-worker, offline, push, and installability work behind freshness and stale/read-only guardrails."
status: future
---

# Long-Term Frontend PWA and Offline Platform

## Purpose

This note holds PWA and offline capabilities that should not be added until freshness semantics are explicit.

## Deferred Topics

- Service worker runtime.
- Offline mode.
- Push notifications.
- Background sync.
- App installation prompts.
- API response caching.
- Offline write queues.

## Why Deferred

Prediction markets and participant accounts are sensitive to freshness. Cached market status, balances, positions, payouts, and resolution state can mislead users if presented as actionable truth.

## Entry Criteria

Reconsider this when:

- The design plan answers which frontend states require server freshness before participant action.
- Cached market/account views have stale/read-only labels.
- API/auth/error seams are stable.
- There is a product requirement for offline or installable behavior.
- Server-side semantics exist for any action that might be attempted after offline or stale reads.

## Guardrail

No offline writes or participant actions from cached accounting-sensitive state unless the backend provides an explicit freshness and conflict-resolution contract.
