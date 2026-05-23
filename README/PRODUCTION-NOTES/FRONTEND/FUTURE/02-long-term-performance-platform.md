---
title: Long-Term Frontend Performance Platform
document_type: production-notes
domain: frontend
future: true
author: Patrick Delaney
updated_at: 2026-05-23T00:00:00Z
updated_at_display: "Saturday, May 23, 2026"
update_reason: "Move broad performance optimization programs behind the active measurement baseline."
status: future
---

# Long-Term Frontend Performance Platform

## Purpose

This note holds performance work that should follow current build and bundle evidence.

The active performance note is [../02-performance-optimization.md](../02-performance-optimization.md).

## Deferred Topics

- Route-based code splitting campaign.
- Strict bundle budgets.
- Lighthouse or Core Web Vitals gates.
- Charting-library replacement.
- Asset pipeline redesign.
- CDN, service-worker, or edge-caching platform.
- Runtime performance monitoring dashboards.

## Why Deferred

The current frontend needs a reproducible build signal and build-size baseline first. Without current measurements, code splitting or dependency replacement is speculative and risks adding complexity without a clear target.

## Entry Criteria

Reconsider this when:

- Frontend PR build feedback exists.
- Current production build size is recorded.
- A repeated performance issue appears in user or deploy evidence.
- Heavy dependencies are identified with actual bundle data.
- A permissive budget can be introduced without blocking unrelated work.

## Guardrail

Optimize from evidence. Do not add performance tooling or rewrite routing/dependencies before the baseline proves what is expensive.
