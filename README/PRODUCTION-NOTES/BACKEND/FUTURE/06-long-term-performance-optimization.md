---
title: Long-Term Performance Optimization
document_type: production-notes
domain: backend
author: Patrick Delaney
updated_at: 2026-04-27T02:03:51Z
updated_at_display: "Monday, April 27, 2026 at 2:03 AM UTC"
update_reason: "Capture broader performance-platform ideas separately from the active evidence-driven optimization note."
status: draft
---

# Long-Term Performance Optimization

## Purpose

This note holds broader performance ideas that should not drive the active production-hardening sequence.

The active performance work remains in [08-performance-optimization.md](/workspace/socialpredict/README/PRODUCTION-NOTES/BACKEND/08-performance-optimization.md), and caching remains separately deferred in [13-database-caching.md](/workspace/socialpredict/README/PRODUCTION-NOTES/BACKEND/13-database-caching.md).

## Deferred Topics

Deferred optimization ideas include:

- load and stress programs
- always-on profiling strategy
- `pprof` exposure policy
- broad cache hierarchies
- response caching
- advanced memory-pooling work
- CDN or larger edge-acceleration choices
- wide query-plan programs beyond targeted hotspot fixes

## Preconditions

These ideas should stay deferred until the backend has:

- clearer runtime DB ownership
- explicit pool and connection-lifecycle tuning
- stronger operational signals
- evidence of real hotspots
- safer correctness and transaction posture for accounting-sensitive flows

## Guardrail

This document is non-binding on the active design plan and task queue until the active runtime, monitoring, and DB notes have landed further.
