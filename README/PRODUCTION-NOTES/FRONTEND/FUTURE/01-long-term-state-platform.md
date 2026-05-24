---
title: Long-Term Frontend State Platform
document_type: production-notes
domain: frontend
future: true
author: Patrick Delaney
updated_at: 2026-05-24T00:00:00Z
updated_at_display: "Sunday, May 24, 2026"
update_reason: "Keep global state platforms deferred after the first API/auth adapter seam until remaining transport and state orchestration evidence accumulates."
status: future
---

# Long-Term Frontend State Platform

## Purpose

This note holds broad state-management platform ideas that should not lead the next frontend work.

The active state note is [../01-state-management.md](../01-state-management.md).

## Deferred Topics

- Redux Toolkit or Zustand as an application-wide store.
- RTK Query or equivalent data-fetching/cache platform.
- Persisted auth or app state beyond the current compatibility model.
- Optimistic updates for market and account workflows.
- Offline synchronization.
- Time-travel debugging and advanced store tooling.

## Why Deferred

The current design problem is still dependency direction, not store selection. The first `authStorage` and `httpClient` seams exist, but remaining direct token reads, direct `fetch(API_URL...)` calls, and duplicated envelope parsing should be reduced before introducing a global state platform.

## Entry Criteria

Reconsider this when:

- The API/auth adapter seam has covered more than representative flows.
- Direct transport and token-storage dependencies are reduced or documented.
- Multiple workflows show repeated state orchestration that local hooks/adapters cannot handle cleanly.
- Server freshness, auth/session behavior, and failure mapping are stable enough for caching decisions.

## Guardrail

Do not introduce Redux, RTK Query, or another state platform as a substitute for clarifying auth/API/session boundaries.
