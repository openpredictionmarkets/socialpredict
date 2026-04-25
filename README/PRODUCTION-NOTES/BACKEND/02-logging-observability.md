---
title: Logging and Observability
document_type: production-notes
domain: backend
author: Patrick Delaney
updated_at: 2026-04-25T00:00:00-05:00
updated_at_display: "Saturday, April 25, 2026 (CDT)"
update_reason: "Replace the greenfield observability plan with guidance aligned to the current codebase and the decision to standardize on backend/logger."
status: active
---

# Logging and Observability

## Update Summary

This note was updated on Saturday, April 25, 2026 to replace an older greenfield observability-platform plan with guidance that matches the current SocialPredict backend, the active design-plan posture, and the package-ownership decision to standardize on `backend/logger`.

| Topic | Prior to April 25, 2026 | After April 25, 2026 |
| --- | --- | --- |
| Core framing | Treated logging and observability as a new platform to build from scratch | Treats logging and observability as operational runtime boundaries to harden incrementally |
| Package direction | Proposed new `logging/` and `observability/` package structure | Standardize on `backend/logger`, deprecate and delete `backend/logging` |
| Current-state accuracy | Claimed health and metrics primitives were missing | Acknowledges existing `/health` and `/v0/system/metrics` routes in the live backend |
| Sequencing | Tried to introduce structured logging, metrics, health, tracing, and aggregation all at once | Unify one logging package first, then improve health/readiness, then add metrics and tracing in controlled slices |
| HA posture | Optimized for feature breadth first | Optimizes for deterministic runtime behavior, low-risk rollout, safe redaction, and operational clarity |

## Executive Direction

SocialPredict should treat logging and observability as distinct operational concerns at the runtime boundary, not as one vague platform abstraction.

The backend direction is:

1. Standardize on one logging package: `backend/logger`
2. Deprecate and delete `backend/logging`
3. Keep logging as a runtime/infrastructure adapter, not a domain concern
4. Harden existing health and metrics surfaces before inventing new subsystems
5. Add richer metrics and tracing only after package ownership and log vocabulary are stable

For a high-availability, fault-tolerant, enterprise-ready system, the backend should prefer:

- deterministic startup and shutdown logs
- stable field and message vocabulary
- stdout/stderr-first emission suitable for container and process aggregation
- request/correlation information at middleware and runtime boundaries
- explicit health/readiness semantics
- safe redaction and no secret leakage

This note explicitly rejects creating a second backend logging dialect or a new top-level `observability/` mega-tree before ownership is unified.

## Why This Matters

Two logging packages mean two competing models of backend runtime truth. That increases migration cost, confuses conventions, and makes later observability work harder instead of easier.

For high availability and fault tolerance, observability has to support:

- fast diagnosis during deploy and restart events
- consistent runtime behavior across replicas
- clear health and readiness semantics for orchestration
- low-risk operational rollout
- safe production logging that never emits secrets or uncontrolled dumps

That requires one owned logging package, one migration direction, and one current-state document.

## Current Code Snapshot

As of 2026-04-25, the backend already has partial logging and observability primitives, but they are split and inconsistent.

### Logging packages are duplicated

The backend currently contains two overlapping packages:

```text
backend/
├── logger/
│   ├── simplelogging.go
│   ├── simplelogging_test.go
│   └── README_SIMPLELOGGING.md
└── logging/
    ├── loggingutils.go
    └── mocklogging.go
```

`backend/logger` is the stronger current package:

- it has a concrete implementation in [simplelogging.go](/workspace/socialpredict/backend/logger/simplelogging.go)
- it has tests in [simplelogging_test.go](/workspace/socialpredict/backend/logger/simplelogging_test.go)
- it has repo-local usage guidance in [README_SIMPLELOGGING.md](/workspace/socialpredict/backend/logger/README_SIMPLELOGGING.md)

`backend/logging` is legacy overlap:

- it provides a thin wrapper and ad hoc debug helpers in [loggingutils.go](/workspace/socialpredict/backend/logging/loggingutils.go)
- it includes a mock in [mocklogging.go](/workspace/socialpredict/backend/logging/mocklogging.go)
- it appears to have only one remaining active non-test import in [resolvemarket.go](/workspace/socialpredict/backend/handlers/markets/resolvemarket.go)

### `backend/logger` is the live standardization target

Current `backend/logger` usage already exists in live backend code, for example:

- [changepassword.go](/workspace/socialpredict/backend/handlers/users/changepassword.go)
- [20251013_080000_core_models.go](/workspace/socialpredict/backend/migration/migrations/20251013_080000_core_models.go)

That makes `backend/logger` the right package to standardize around.

### Health and metrics are not starting from zero

The current note used to claim that health and metrics primitives did not exist. That is no longer true.

The backend already serves:

- `/health` via [server.go](/workspace/socialpredict/backend/server/server.go)
- `/v0/system/metrics` via [server.go](/workspace/socialpredict/backend/server/server.go)

Important distinction:

- `/health` is an infrastructure/runtime health endpoint
- `/v0/system/metrics` is an application route backed by metrics handlers, not yet a full Prometheus-style infrastructure metrics boundary

So the correct framing is not “create health and metrics from nothing.” The correct framing is “harden and clarify what exists, then extend deliberately.”

### Current logger limitations

Standardizing on `backend/logger` does not mean the logging story is finished.

Current limitations include:

- string-oriented convenience APIs rather than structured event shapes
- per-call convenience construction in [simplelogging.go](/workspace/socialpredict/backend/logger/simplelogging.go)
- incomplete request-correlation and middleware-level logging
- no unified redaction policy documented here yet
- no explicit readiness/liveness split yet
- no tracing boundary yet

That is acceptable for the current note because the first task is not to build everything. The first task is to unify ownership and make future hardening coherent.

## Logging and Observability Taxonomy

### 1. Runtime Diagnostic Logging

This covers backend log emission used to understand process behavior at runtime.

Examples:

- startup and shutdown messages
- request lifecycle logging
- handler/service error logs at the runtime boundary
- migration and seed diagnostics

Ownership:

- runtime and infrastructure boundary
- request middleware and server boundary
- selected handler/application adapters where logging is still necessary

Not owned by:

- core domain policy
- generic debug helper packages

### 2. Health and Readiness Signals

This covers operator-facing signals that determine whether the process is alive, healthy, and ready to receive traffic.

Examples:

- `/health`
- future liveness/readiness separation
- DB readiness and critical dependency checks

These signals must remain simple, deterministic, and orchestration-friendly.

### 3. Metrics

This covers numeric signals intended for aggregation and alerting.

Examples:

- infrastructure request counts and latency
- backend process/runtime metrics
- business-level counters exposed intentionally through owned handlers

Metrics are not a substitute for logs. They answer different questions and should not be collapsed into logging behavior.

### 4. Tracing

This covers cross-cutting request and dependency correlation for deeper production diagnosis.

Tracing is a later observability phase, not a prerequisite for package standardization.

## Logger Package Direction

The package decision is explicit:

- `backend/logger` is the only backend logging package that should survive
- `backend/logging` is legacy duplication and should be deprecated and deleted
- no new backend code should import `socialpredict/logging`
- no rewrite should introduce a second logging abstraction while standardization is incomplete

The required migration path is:

1. move remaining `backend/logging` imports to `backend/logger`
2. port or replace any needed test helpers without preserving package duplication
3. delete `backend/logging`
4. then harden `backend/logger` as the sole backend log-emission adapter

This note treats package unification as a prerequisite to richer observability work.

## Operational Invariants

For production operation, the backend should follow these rules:

- one owned backend logging package only
- never log secrets, tokens, passwords, or API keys
- prefer runtime-boundary logging over domain-policy logging
- keep logs restart-friendly and aggregation-friendly via stdout/stderr-first emission
- keep field vocabulary stable once request/correlation fields are introduced
- treat health, metrics, and tracing as separate operational signals with separate owners
- do not rely on arbitrary reflection-based dumping of runtime values in production code
- do not introduce hot-path logging noise that obscures failures or raises operational cost without diagnostic value

## Current Tree Versus Target Tree

### Current logging-related tree

```text
backend/
├── logger/
│   ├── simplelogging.go
│   ├── simplelogging_test.go
│   └── README_SIMPLELOGGING.md
├── logging/
│   ├── loggingutils.go
│   └── mocklogging.go
├── handlers/
│   ├── markets/
│   │   └── resolvemarket.go
│   ├── metrics/
│   │   ├── getgloballeaderboard.go
│   │   └── getsystemmetrics.go
│   └── users/
│       └── changepassword.go
└── server/
    └── server.go
```

### End-state objective

The objective is not to create a new `observability/` platform tree immediately. The objective is to unify ownership and then evolve incrementally.

```text
backend/
├── logger/
│   ├── logger.go                  # sole owned backend logging adapter
│   ├── logger_test.go
│   ├── middleware.go              # request/correlation logging when introduced
│   └── README.md
├── handlers/
│   ├── metrics/                   # owned application/system metrics handlers
│   └── users/
├── server/
│   └── server.go                  # health/readiness and runtime-boundary wiring
└── logging/                       # deleted
```

The near-term target is one package, one migration path, and one operational vocabulary.

## Design Rules

The intended direction is:

- standardize on `backend/logger`
- deprecate and delete `backend/logging`
- keep logging as a runtime and infrastructure concern
- improve the current `/health` behavior into explicit health/readiness semantics over time
- preserve a distinction between logs, metrics, tracing, and health signals
- add request and correlation context at the runtime boundary in controlled slices
- harden redaction and sensitive-data handling before expanding log volume
- evolve observability incrementally from the existing backend shape

The intended direction is not:

- a new top-level `observability/` package tree before ownership is unified
- a second logging abstraction
- reflection-heavy diagnostic dumping as a durable production strategy
- forcing metrics, tracing, health, and logs into one generic package family
- pretending the backend starts with zero health or metrics support

## Concrete Next Migration Goals

1. Replace the remaining `socialpredict/logging` imports with `socialpredict/logger`.
2. Delete `backend/logging` once remaining usage and tests are migrated.
3. Define the stable backend logging API and message vocabulary for the surviving `backend/logger` package.
4. Add request/correlation logging at the server or middleware boundary rather than scattering it across handlers.
5. Harden `/health` into clearer liveness/readiness semantics without destabilizing the current runtime surface.
6. Decide which metrics belong to runtime/infrastructure versus application/business surfaces before adding more endpoints.
7. Defer tracing until package ownership, request correlation, and health semantics are stable.

## What This Note Replaces

This update replaces the older recommendation to:

- build a broad new observability platform from scratch
- introduce a new `logging/` or `observability/` package family
- assume health and metrics primitives are absent
- select tools first and boundaries later

SocialPredict’s immediate need is operational boundary clarity:

- one backend logging package
- clear deprecation of `backend/logging`
- accurate documentation of the current runtime surface
- staged observability hardening consistent with a high-availability, fault-tolerant backend
