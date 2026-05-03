---
title: Background Jobs
document_type: production-notes
domain: backend
author: Patrick Delaney
updated_at: 2026-05-01T11:50:49Z
updated_at_display: "Friday, May 01, 2026 at 11:50 AM UTC"
update_reason: "Refresh the deferred jobs draft against landed readiness and startup-writer seams on upstream main at 051aac6."
status: draft
---

# Background Jobs

## Starter Draft Status

This is a lower-priority starter draft.

It exists to capture the likely future direction for asynchronous processing without confusing that work with the higher-priority runtime, DB, monitoring, and validation concerns described in the active production notes.

The current posture is explicit:

- background jobs are not part of the current production-hardening slice
- correctness comes before async offloading
- operational visibility comes before retry-heavy worker systems
- money-moving flows should not be pushed into background jobs to compensate for weak transaction boundaries

This draft was refreshed on Friday, May 01, 2026 against upstream `main` at `051aac6b2fefa5634b8c98cc38caf52acf0043a9`. The backend now has liveness and readiness probes, explicit startup-writer gating, and better request-boundary failure handling, but it still has no worker topology, no idempotency model, and no retry or outbox ownership.

## Executive Direction

If SocialPredict later introduces background-job infrastructure, it should do so as a narrow support system for explicitly idempotent, non-request-critical work.

That means:

1. Keep the active backend request path synchronous for accounting-sensitive flows such as betting, selling, and market resolution.
2. Introduce background processing only for tasks that are safe to decouple from the request path.
3. Treat retries, dead-letter handling, job ownership, and observability as prerequisites, not afterthoughts.
4. Avoid adding Redis, worker pools, cron frameworks, or queue tables just because they sound operationally mature.
5. Keep any future async system small and explicit rather than turning it into a second application architecture.

## Why This Is Deferred

The live backend still has more urgent concerns:

- the serving runtime baseline is stronger than the older draft assumed, but deployment still runs one HTTP-serving shape with no worker role
- atomic accounting-sensitive workflow boundaries still need more explicit hardening outside the place-bet path
- the runtime still lacks worker-specific signals such as lag, replay, retry, or dead-letter visibility
- no queue, outbox, scheduler, or worker ownership model exists yet in the codebase

Adding queues or workers before those concerns are stronger would add a new failure domain before the current system is fully ready to support it.

## Current Code Snapshot

### Runtime prerequisites are stronger, but still not worker-ready

The backend now has:

- `STARTUP_WRITER` mode in [runtime/startup_mutation.go](/workspace/socialpredict/backend/internal/app/runtime/startup_mutation.go)
- startup mutation and verification behavior in [startup_mutation.go](/workspace/socialpredict/backend/startup_mutation.go)
- liveness and readiness probes in [server.go](/workspace/socialpredict/backend/server/server.go)

That is useful baseline infrastructure, but it is still serving-process infrastructure. It does not define a queue contract, worker lifecycle, retry semantics, or async ownership model.

### There is no live background-job subsystem in the backend

The active backend does not currently have:

- a `jobs/` or `workers/` package
- a Redis queue runtime
- a cron or scheduler subsystem
- a job table or outbox pattern in the application code
- a second worker deployment topology in the repo

The current backend is still one primary server process.

### Current important flows are synchronous on purpose

Important flows currently execute inline through request and domain paths, including:

- `POST /v0/bet`
- `POST /v0/sell`
- market resolution flows

That is important because it means correctness and failure semantics are still concentrated in the request path rather than hidden behind async infrastructure.

The place-bet path now has stronger transaction behavior, but that still does not justify moving money-sensitive workflows behind background retries or workers.

### Deploy and workflow topology assumes one backend runtime shape

The current Docker and workflow topology publishes and deploys the backend as one primary HTTP-serving binary:

- [docker/backend/Dockerfile](/workspace/socialpredict/docker/backend/Dockerfile)
- [docker-compose-prod.yaml](/workspace/socialpredict/scripts/docker-compose-prod.yaml)
- [docker.yml](/workspace/socialpredict/.github/workflows/docker.yml)

There is no parallel worker deployment contract yet, which is another reason to keep this note deferred.

## Candidate Future Uses

If background jobs are introduced later, plausible early candidates are:

- email delivery
- periodic or triggered cache refreshes
- scheduled derived snapshots that are safe to recompute
- export generation
- non-critical notification fan-out

## Candidate Early Non-Targets

The following should not be early background-job targets:

- account-balance correctness
- bet placement correctness
- sell settlement correctness
- market-resolution correctness
- any flow where partial success would break system economics

## Preconditions Before Background Jobs

Before the backend should prioritize background-job infrastructure, it should first have:

- an operationally enforced startup and shutdown contract for more than one runtime role
- stronger readiness and monitoring posture for worker-specific failures
- explicit idempotency rules for candidate async tasks
- a clear retry and dead-letter policy
- visibility into failure, replay, and lag behavior
- at least one candidate task that is valuable, decouplable, and clearly non-financial

## Relationship To Other Notes

This note is intentionally downstream of:

- [04-database-layer.md](/workspace/socialpredict/README/PRODUCTION-NOTES/BACKEND/04-database-layer.md)
- [08-deployment-infrastructure.md](/workspace/socialpredict/README/PRODUCTION-NOTES/BACKEND/08-deployment-infrastructure.md)
- [09-monitoring-alerting.md](/workspace/socialpredict/README/PRODUCTION-NOTES/BACKEND/09-monitoring-alerting.md)
- [12-database-caching.md](/workspace/socialpredict/README/PRODUCTION-NOTES/BACKEND/12-database-caching.md)

Longer-term queue and worker ideas now belong in [FUTURE/07-long-term-background-jobs.md](/workspace/socialpredict/README/PRODUCTION-NOTES/BACKEND/FUTURE/07-long-term-background-jobs.md).
