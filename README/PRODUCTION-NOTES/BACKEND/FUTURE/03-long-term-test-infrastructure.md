---
title: Long-Term Test Infrastructure
document_type: production-notes
domain: backend
author: Patrick Delaney
updated_at: 2026-04-27T00:13:14Z
updated_at_display: "Monday, April 27, 2026 at 12:13 AM UTC"
update_reason: "Create a non-binding holding note for deferred containerized and broader test-infrastructure ideas so the active testing note can stay focused on the current backend and the active HA-focused modernization slice."
status: future
---

# Long-Term Test Infrastructure

## Purpose

This note is a holding area for deferred testing-infrastructure ideas that are not part of the active SocialPredict backend design plan, not part of the current production-note wave sequence, and not part of the current runnable task queue.

Its purpose is to preserve larger testing-infrastructure ideas without letting them distort the active near-term architecture in [07-testing-strategy.md](/workspace/socialpredict/README/PRODUCTION-NOTES/BACKEND/07-testing-strategy.md).

## Current Status

As of 2026-04-27:

- the active backend testing note should stay focused on package-local tests, current helper packages, contract checks, runtime-risk verification, and targeted DB truthfulness
- the active design plan is still prioritizing runtime ownership, failure handling, database runtime hardening, security alignment, and API contract cleanup
- the ideas in this document are explicitly deferred until those nearer-term seams are more stable

This document is non-binding on the active design plan and on `TASKS.json`.

## Candidate Future Topics

The following ideas are reasonable future candidates, but they are not current architecture commitments.

### Postgres-backed integration posture

- a small dedicated Postgres-backed suite for transaction, locking, isolation, and startup semantics
- explicit separation between convenience DB tests and source-of-truth DB tests
- a clearer policy for which behaviors SQLite must never be trusted to verify

### `testcontainers-go` evaluation

- using `testcontainers-go` to start Postgres for targeted integration tests
- evaluating Docker/runtime cost versus confidence gained
- deciding whether containerized DB tests should run locally, only in CI, or both

This is a plausible future direction, but it should be introduced only where a concrete verification need justifies the runtime cost and flake risk.

### Centralized cross-boundary integration tests

- a small central location for tests that span multiple packages and do not fit one owning directory cleanly
- a clear boundary between package-local tests and true system-level tests
- rules for when centralization is justified

This topic is deferred because the active backend is still well served by package-local tests as the default.

### Fast versus slow test orchestration

- explicit short/long-running test splits
- separate local and CI execution tiers
- optional tagging or selection strategy for heavier DB/runtime cases

This is deferred because the current suite still benefits more from architectural focus than from taxonomy expansion.

### Performance and load verification

- benchmark programs
- load or stress testing
- steady-state or failure-mode performance evaluation

These ideas may matter later, but they are not current prerequisites for the active modernization slice.

### Shared test-helper distribution across repos

- whether helper packages ever need to move to a separate shared repository
- versioning and ownership rules if multiple repos truly need the same helpers

This is explicitly speculative for now. Repo-local helpers are the default unless a real multi-repo sharing problem appears.

## Entry Criteria

This note should only become active planning input after the current architecture is substantially more stable.

Reasonable entry criteria are:

- the active production notes through at least `07` are aligned to live code
- the runtime, DB, security, and API waves have reduced the main transitional seams
- there is a concrete Postgres/runtime verification gap that package-local tests and current helpers cannot cover well
- there is a real CI or developer-workflow reason to introduce heavier infrastructure
- there is a measurable need for cross-repo helper sharing rather than a speculative one

## What Is Explicitly Deferred

The following items are explicitly not part of the current queue unless later reactivated on purpose:

- testcontainers-first rollout
- broad containerized integration infrastructure
- a generalized `TestSuite` platform
- a top-level `testing/` subsystem
- performance/load/stress platform work
- benchmark gating
- broad test taxonomy or orchestration redesign
- a separate test-helper repository without a concrete multi-repo need

## Re-Entry Questions

Before pulling any of these topics into the active plan, SocialPredict should answer:

- What exact verification gap exists that package-local tests and current helpers cannot cover?
- Is the problem truly Postgres-specific, multi-process, or deployment-specific?
- Would `testcontainers-go` improve confidence enough to justify added runtime cost, Docker dependency, and possible flake risk?
- Does the candidate test really need a central location, or is package-local placement still clearer?
- Is a separate helper repo solving a real multi-repo ownership problem, or just moving complexity outward?

## Relationship To The Active Note

The active note at [07-testing-strategy.md](/workspace/socialpredict/README/PRODUCTION-NOTES/BACKEND/07-testing-strategy.md) is the binding current-state-first architecture note.

This `FUTURE` note exists so that:

- the active note can stay grounded in idiomatic Go testing and the live backend
- package-local tests remain the default
- larger infrastructure ideas are not lost
- deferred test-platform ideas do not get mistaken for near-term design commitments
