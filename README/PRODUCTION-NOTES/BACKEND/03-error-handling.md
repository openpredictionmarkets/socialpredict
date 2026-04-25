---
title: Error Handling
document_type: production-notes
domain: backend
author: Patrick Delaney
updated_at: 2026-04-25T19:15:00-05:00
updated_at_display: "Saturday, April 25, 2026 at 7:15 PM Central (CDT)"
update_reason: "Replace the greenfield error-framework plan with guidance aligned to the live backend, the active design plan, and the HA/fault-tolerance objective."
status: active
---

# Error Handling

## Update Summary

This note was updated on Saturday, April 25, 2026 to replace an older greenfield error-framework plan with guidance that matches the current SocialPredict backend, the active design-plan posture, and the high-availability/fault-tolerance objective.

| Topic | Prior to April 25, 2026 | After April 25, 2026 |
| --- | --- | --- |
| Core framing | Treated error handling as a new application-wide framework to build from scratch | Treats error handling as failure classification, containment, and boundary translation to harden incrementally |
| Main proposal | Build a new `AppError` hierarchy, handler base classes, and monitoring stack first | Start from the live mixed backend and converge runtime, middleware, auth, and handler failure behavior deliberately |
| Current-state accuracy | Assumed `backend/errors` was the real backbone | Recognizes that the live contract already centers on `handlers/envelope.go`, while `backend/errors` is only thin legacy overlap |
| Contract model | Assumed one new public `error.code/message` response model | Recognizes a live split between `ReasonResponse`, legacy plain-text responses, and route-family migration state |
| Recovery stance | Proposed generic retries, circuit breakers, and recovery strategies | Rejects generic write-path retry for accounting-sensitive flows and prioritizes server-owned recovery, sanitization, and diagnosis |
| HA posture | Optimized for framework breadth first | Optimizes for deterministic fault containment, clear operator diagnosis, stable client semantics, and safe migration |

## Executive Direction

SocialPredict should treat error handling as a boundary and runtime concern, not as a new generic error framework.

The backend direction is:

1. Keep typed internal failures close to their owned seams: domain, repository, auth, and runtime helpers.
2. Own request correlation, panic recovery, and middleware-generated failures at the server/runtime boundary.
3. Use shared failure translation at the HTTP boundary to map internal outcomes to HTTP status plus stable public `reason` values for routes that advertise the envelope.
4. Treat raw `http.Error`, status-only failures, and `PlainTextErrorResponse` as explicit migration state, not the target architecture.
5. Treat `backend/errors` as deletion-only or compatibility-only scope, not as the future error platform.
6. Do not introduce generic retry or circuit-breaker behavior for market, bet, account, or auth writes without separate idempotency and accounting design.

For a high-availability, fault-tolerant, enterprise-ready system, the backend should prefer:

- controlled fault containment at the server boundary
- request-correlated diagnostics for unexpected failures
- sanitized logs and sanitized client-visible failures
- stable route-visible failure semantics for migrated API routes
- fail-fast startup behavior and safe request-path recovery behavior

This note explicitly rejects building a large new universal `AppError` hierarchy as the primary design move.

## Why This Matters

Error handling is not only about user-facing messages. For a high-availability and fault-tolerant backend, it also determines:

- whether one bad request or panic turns into diagnosable, bounded failure behavior
- whether clients can depend on stable negative outcomes instead of ad hoc strings
- whether middleware and handler failures behave coherently across replicas
- whether operator logs are useful without leaking secrets or raw internal details
- whether accounting-sensitive flows avoid dangerous hidden retries

The older note was useful as a generic wishlist, but it no longer matches the codebase. SocialPredict already has a partial public error contract, partial typed internal errors, and partial handler-level failure translation. The job now is not to start over. The job is to make that mixed surface architecturally consistent.

## Current Code Snapshot

As of 2026-04-25, the backend already has meaningful error-handling structure, but it is split across several competing patterns.

### The live public failure contract is mixed

The backend already has a shared failure envelope and reason vocabulary in [envelope.go](/workspace/socialpredict/backend/handlers/envelope.go):

```go
type FailureEnvelope struct {
    OK     bool   `json:"ok"`
    Reason string `json:"reason"`
}

func WriteFailure(w http.ResponseWriter, statusCode int, reason FailureReason) error
```

Newer handlers already use this contract. Examples include:

- [failure_helpers.go](/workspace/socialpredict/backend/handlers/markets/failure_helpers.go)
- [sellpositionhandler.go](/workspace/socialpredict/backend/handlers/bets/selling/sellpositionhandler.go)
- [profile_helpers.go](/workspace/socialpredict/backend/handlers/users/profile_helpers.go)
- [loggin.go](/workspace/socialpredict/backend/internal/service/auth/loggin.go)

But many legacy handlers still return raw `http.Error` strings, for example:

- [createmarket.go](/workspace/socialpredict/backend/handlers/markets/createmarket.go)
- [handler.go](/workspace/socialpredict/backend/handlers/markets/handler.go)
- [publicuser.go](/workspace/socialpredict/backend/handlers/users/publicuser.go)
- [portfolio.go](/workspace/socialpredict/backend/handlers/users/publicuser/portfolio.go)

So the live system is not “missing error handling.” It already has multiple error dialects.

### `backend/errors` is not the real production backbone

The old note treated `backend/errors` as the starting point for a new system. In reality, it is only thin legacy overlap:

- [httperror.go](/workspace/socialpredict/backend/errors/httperror.go) logs an error and writes `{ "error": "..." }`
- [normalerror.go](/workspace/socialpredict/backend/errors/normalerror.go) only logs a string plus an error

Those files are not the main public contract for live routes, and they should not dictate the future architecture.

### Auth still leaks transport concerns into an internal seam

`internal/service/auth` currently defines its own transport-shaped error type in [authutils.go](/workspace/socialpredict/backend/internal/service/auth/authutils.go):

```go
type HTTPError struct {
    StatusCode int
    Message    string
}
```

And auth helpers return that type directly from [auth.go](/workspace/socialpredict/backend/internal/service/auth/auth.go):

```go
func ValidateUserAndEnforcePasswordChangeGetUser(
    r *http.Request,
    svc dusers.ServiceInterface,
) (*dusers.User, *HTTPError)
```

That means an internal service seam still owns route-visible status and message policy. Login already shows the better direction: it maps internal auth outcomes to the public envelope at the HTTP edge in [loggin.go](/workspace/socialpredict/backend/internal/service/auth/loggin.go).

### Middleware and router failures bypass the shared failure contract

Some failures happen before handlers run at all.

Rate limiting currently emits plain-text `429` responses in [ratelimit.go](/workspace/socialpredict/backend/security/ratelimit.go):

```go
if !limiter.GetLimiter(ip).Allow() {
    http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
    return
}
```

The router-level method-not-allowed handler in [server.go](/workspace/socialpredict/backend/server/server.go) writes status and `Allow` but no shared envelope body.

This matters because these failures are part of the public runtime behavior even though they do not originate in a route handler.

### There is no application-owned recovery boundary yet

`buildHandler` in [server.go](/workspace/socialpredict/backend/server/server.go) currently builds the router and wraps it with CORS, but it does not establish request correlation or panic-recovery middleware owned by the application.

That means unexpected request-path panics currently fall back to the standard library's default recovery behavior rather than to a SocialPredict-owned recovery path with:

- consistent client-visible failure semantics
- request correlation
- controlled runtime logging vocabulary

For HA and fault tolerance, that is too weak to be the long-term design.

### OpenAPI already documents a migration state

The OpenAPI document is more current than the old note. At the top of [openapi.yaml](/workspace/socialpredict/backend/docs/openapi.yaml), it already documents shared middleware `429 text/plain` behavior. It also defines `ReasonResponse` and `PlainTextErrorResponse` separately.

That is the right current-state framing:

- some route families already expose stable `reason` values
- some paths still return plain-text or other legacy error forms
- the transition needs to be managed explicitly, not hidden behind a new framework

## Failure Taxonomy

SocialPredict should use a boundary-first vocabulary for failures.

### 1. Business Rejection

This is a processed request where the backend understood the request and applied business policy, but the outcome is negative.

Examples:

- market is closed
- insufficient balance
- no position
- insufficient shares
- password change required as an access policy outcome

These outcomes should map to stable public `reason` values on routes that have adopted the envelope contract.

### 2. Access Failure

This is a request rejected because authentication or authorization requirements were not satisfied.

Examples:

- missing or invalid bearer token
- caller lacks admin privileges
- must-change-password restriction blocks an authenticated action

Auth may use typed internal outcomes, but the transport boundary should own the final public status and reason semantics.

### 3. Transport Failure

This is a request that cannot be processed correctly because the HTTP interaction itself is malformed or rejected before business logic completes.

Examples:

- unsupported method
- malformed JSON
- missing path or query parameter
- rate limiting before the handler runs

Transport failures belong to middleware, router, and HTTP-boundary ownership, not domain ownership.

### 4. Infrastructure Failure

This is an unexpected runtime or dependency failure.

Examples:

- DB unavailable
- response encoding failure
- unexpected panic
- missing critical runtime dependency

Infrastructure failures must be sanitized for clients and logged for operators with correlation and no secret leakage.

## Boundary Direction

### Domain, repository, and auth internal seams

These seams should own typed internal failures and business-rule rejections. They should not be the long-term owners of:

- HTTP status codes
- route-visible reason strings
- response envelopes
- middleware behavior

This is why `internal/service/auth.HTTPError` is transitional rather than target architecture.

### Runtime failure containment

The server/runtime boundary should own:

- request correlation
- panic recovery
- middleware-generated failures such as shared 405 and 429 behavior
- sanitized unexpected-failure handling
- operator-facing logging of runtime faults

This is where HA/fault-tolerant behavior begins. SocialPredict should not rely on scattered handler logic or default library behavior for this responsibility.

### HTTP failure translation

The handler and route boundary should own translation from typed internal outcomes to:

- HTTP status
- stable public `reason`
- legacy plain-text response where migration has not finished yet
- OpenAPI contract updates that match the route family's actual behavior

This is not a demand for one giant centralized mapper. Route-family helpers such as [failure_helpers.go](/workspace/socialpredict/backend/handlers/markets/failure_helpers.go) are a valid migration shape when they stay boundary-owned and consistent.

### Legacy compatibility scope

Two current pieces are explicitly transitional:

- `backend/errors`
- auth-local `HTTPError`

Neither should become the future architectural center of error handling.

## Current Tree Versus Target Tree

### Current error-related tree

```text
backend/
├── errors/
│   ├── httperror.go
│   └── normalerror.go
├── handlers/
│   ├── envelope.go
│   ├── markets/
│   ├── users/
│   ├── bets/
│   ├── setup/
│   ├── stats/
│   └── metrics/
├── internal/
│   └── service/
│       └── auth/
│           ├── auth.go
│           ├── authutils.go
│           └── loggin.go
├── security/
│   └── ratelimit.go
└── server/
    └── server.go
```

### End-state objective

The objective is not a new top-level error platform. The objective is one coherent set of failure boundaries.

An illustrative target shape is:

```text
backend/
├── server/
│   ├── server.go                 # route wiring
│   ├── recovery.go               # panic recovery and request correlation
│   └── failures.go               # shared middleware/runtime failure writers
├── handlers/
│   ├── envelope.go               # stable public reason vocabulary for migrated routes
│   ├── markets/
│   ├── users/
│   ├── bets/
│   └── ...
├── internal/
│   └── service/
│       └── auth/                 # typed internal auth outcomes, not public wire policy
└── errors/                       # deleted or compatibility-only
```

This tree is directional, not mandatory. The architectural point is ownership:

- runtime recovery and middleware failures belong near `server`
- route-visible failure translation belongs at the HTTP boundary
- typed business and auth outcomes belong inside owned internal seams

## Design Rules

The intended direction is:

- treat error handling as failure translation and recovery, not a new universal hierarchy
- use stable public `reason` values where routes have adopted the envelope
- keep `PlainTextErrorResponse` explicit and temporary where migration is incomplete
- recover unexpected request-path panics at the server boundary with sanitized responses and correlated logs
- make middleware and handler failure behavior coherent for protected JSON routes
- keep secrets, passwords, tokens, and API keys out of logs and client-visible failures
- reject generic retry or circuit-breaker behavior for accounting-sensitive writes unless a separate design proves idempotency and safety

The intended direction is not:

- a new framework-first `AppError` platform
- a mandatory base handler class
- expanding `backend/errors` into the central architecture
- hidden automatic recovery logic for write-heavy financial flows
- assuming middleware failures do not matter because they happen before handlers

## Concrete Next Migration Goals

1. Add application-owned request-correlation and panic-recovery middleware at the server boundary.
2. Add shared failure writers for runtime-owned `405`, `429`, and sanitized internal `500` behavior.
3. Decide whether protected `/v0` routes should converge fully on `ReasonResponse` or retain explicit plain-text exceptions in some route families.
4. Replace `internal/service/auth.HTTPError` with cleaner typed internal auth outcomes before HTTP mapping occurs.
5. Inventory route families that still emit raw `http.Error`, status-only failures, or ad hoc strings, then migrate them by highest-value groups.
6. Retire `backend/errors` once no live request path depends on it.
7. Keep OpenAPI and production notes route-family accurate throughout the migration instead of claiming universal envelope coverage too early.

## What This Note Replaces

This update replaces the older recommendation to:

- build a universal `AppError` hierarchy first
- create a new handler base-class pattern
- add generic retry, circuit-breaker, and recovery strategies as default architecture
- assume `backend/errors` should become the core package for all failure handling
- define a new public `error.code/message` schema as the main target contract

SocialPredict’s immediate need is more pragmatic:

- one owned runtime recovery path
- one clear migration away from ad hoc handler and middleware failures
- one route-visible contract story that matches the actual backend
- one error-handling direction consistent with high availability, fault tolerance, and accounting safety
