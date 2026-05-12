---
title: Error Handling
document_type: production-notes
domain: backend
author: Patrick Delaney
updated_at: 2026-04-30T03:33:18Z
updated_at_display: "Thursday, April 30, 2026 at 3:33 AM UTC"
update_reason: "Record migration of public user read failures from PlainTextErrorResponse to ReasonResponse."
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
| Telemetry model | Blurred monitoring and public error-contract concerns | Separates public `reason` responses from OpenTelemetry-aligned runtime failure telemetry |
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
7. Use OpenTelemetry-aligned telemetry for unexpected failures and correlation, but keep the public API failure contract separate from telemetry internals.

For a high-availability, fault-tolerant, enterprise-ready system, the backend should prefer:

- controlled fault containment at the server boundary
- request-correlated diagnostics for unexpected failures
- sanitized logs and sanitized client-visible failures
- stable route-visible failure semantics for migrated API routes
- fail-fast startup behavior and safe request-path recovery behavior
- one telemetry story for runtime failures that does not redefine the public HTTP contract

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

### Auth now exposes typed outcomes for migrated routes

`internal/service/auth` now returns typed `AuthError` outcomes from [authutils.go](/workspace/socialpredict/backend/internal/service/auth/authutils.go):

```go
type AuthError struct {
    Kind    ErrorKind
    Message string
}
```

The auth helpers in [auth.go](/workspace/socialpredict/backend/internal/service/auth/auth.go) no longer own public status and response text directly:

```go
func ValidateUserAndEnforcePasswordChangeGetUser(
    r *http.Request,
    svc dusers.ServiceInterface,
) (*dusers.User, *AuthError)
```

Mature route families map those internal outcomes at the HTTP boundary through [authhttp.go](/workspace/socialpredict/backend/handlers/authhttp/authhttp.go). Login also maps internal auth outcomes to the public envelope at the HTTP edge in [loggin.go](/workspace/socialpredict/backend/internal/service/auth/loggin.go).

### Middleware and router failures are runtime-owned

Some failures happen before handlers run at all.

Rate limiting now emits shared JSON failure envelopes in [ratelimit.go](/workspace/socialpredict/backend/security/ratelimit.go):

```go
if !limiter.GetLimiter(ip).Allow() {
    WriteRateLimited(w, RuntimeReasonRateLimited)
    return
}
```

The router-level method-not-allowed handler in [server.go](/workspace/socialpredict/backend/server/server.go) writes the `Allow` header and the shared JSON `METHOD_NOT_ALLOWED` envelope.

This matters because these failures are part of the public runtime behavior even though they do not originate in a route handler.

### Application-owned recovery now starts at the request boundary

`buildHandler` in [server.go](/workspace/socialpredict/backend/server/server.go) now wraps the router with [requestboundary.go](/workspace/socialpredict/backend/security/requestboundary.go), which establishes request correlation, sanitized panic recovery, and runtime status classification.

That gives unexpected request-path panics a SocialPredict-owned recovery path with:

- consistent client-visible failure semantics
- request correlation
- controlled runtime logging vocabulary

The remaining work is to keep extending that boundary deliberately instead of reintroducing scattered runtime failure writers.

### Runtime fault telemetry is not standardized yet

The live backend also lacks a shared OpenTelemetry-aligned runtime failure telemetry story.

Current code does not define one owned approach for:

- attaching request or trace correlation to unexpected failures
- assigning stable failure classes such as `error.type`
- deciding which handled negative outcomes should or should not count as telemetry errors
- avoiding duplicate exception recording across middleware, handlers, and runtime helpers

That gap matters because high availability requires both bounded client-visible behavior and bounded operator-visible diagnosis.

### OpenAPI already documents a migration state

The OpenAPI document is more current than the old note. At the top of [openapi.yaml](/workspace/socialpredict/backend/docs/openapi.yaml), it documents shared middleware `429` JSON envelope behavior. It also defines `ReasonResponse` and `PlainTextErrorResponse` separately.

That is the right current-state framing:

- some route families already expose stable `reason` values
- some paths still return plain-text or other legacy error forms
- the transition needs to be managed explicitly, not hidden behind a new framework

### Protected user profile route-family migration state

The protected user profile family has converged on the shared JSON envelope and stable public `reason` values:

| Route | Handler boundary behavior | Stable failure reasons | Plain-text exception |
| --- | --- | --- | --- |
| `GET /v0/privateprofile` | Uses `authhttp.WriteFailure` for auth outcomes and `handlers.WriteFailure` for user lookup/runtime failures. Shared route dispatch can also emit JSON 405. | `INVALID_TOKEN`, `PASSWORD_CHANGE_REQUIRED`, `USER_NOT_FOUND`, `METHOD_NOT_ALLOWED`, `INTERNAL_ERROR` | None. |
| `POST /v0/profilechange/description` | Uses `authhttp.WriteFailure` for auth outcomes, `INVALID_REQUEST` for malformed JSON, `METHOD_NOT_ALLOWED` for wrong methods, and `writeProfileError` for service failures. | `INVALID_TOKEN`, `PASSWORD_CHANGE_REQUIRED`, `INVALID_REQUEST`, `VALIDATION_FAILED`, `USER_NOT_FOUND`, `AUTHORIZATION_DENIED`, `METHOD_NOT_ALLOWED`, `INTERNAL_ERROR` | None. |
| `POST /v0/profilechange/displayname` | Same protected profile update boundary as description. | `INVALID_TOKEN`, `PASSWORD_CHANGE_REQUIRED`, `INVALID_REQUEST`, `VALIDATION_FAILED`, `USER_NOT_FOUND`, `AUTHORIZATION_DENIED`, `METHOD_NOT_ALLOWED`, `INTERNAL_ERROR` | None. |
| `POST /v0/profilechange/emoji` | Same protected profile update boundary as description. | `INVALID_TOKEN`, `PASSWORD_CHANGE_REQUIRED`, `INVALID_REQUEST`, `VALIDATION_FAILED`, `USER_NOT_FOUND`, `AUTHORIZATION_DENIED`, `METHOD_NOT_ALLOWED`, `INTERNAL_ERROR` | None. |
| `POST /v0/profilechange/links` | Same protected profile update boundary as description. | `INVALID_TOKEN`, `PASSWORD_CHANGE_REQUIRED`, `INVALID_REQUEST`, `VALIDATION_FAILED`, `USER_NOT_FOUND`, `AUTHORIZATION_DENIED`, `METHOD_NOT_ALLOWED`, `INTERNAL_ERROR` | None. |
| `POST /v0/changepassword` | Uses `authhttp.WriteFailure` for token outcomes, `INVALID_REQUEST` for malformed JSON, `METHOD_NOT_ALLOWED` for wrong methods, and `writeChangePasswordError` for password service failures. | `INVALID_TOKEN`, `INVALID_REQUEST`, `VALIDATION_FAILED`, `AUTHORIZATION_DENIED`, `USER_NOT_FOUND`, `METHOD_NOT_ALLOWED`, `INTERNAL_ERROR` | None. |

The public user read/credit/portfolio/financial routes have now moved their handler-owned failures to `ReasonResponse`. Their successful payloads remain route-specific legacy JSON shapes.

### WAVE03 stop-and-review inventory

WAVE03 migrated the protected profile/password slice and runtime-owned `405`, `429`, and panic recovery responses to JSON failure envelopes. Those request boundaries no longer import or depend on `backend/errors`; runtime-owned helpers live in [failures.go](/workspace/socialpredict/backend/security/failures.go), and auth HTTP mapping for migrated handlers lives in [authhttp.go](/workspace/socialpredict/backend/handlers/authhttp/authhttp.go).

The remaining live migration surface is explicit:

| Surface | Current live behavior | Examples | Next-wave seam |
| --- | --- | --- | --- |
| Legacy markets reads and writes | The read-route failure slice for list/get/details and legacy label/update helpers now uses shared failure envelopes; remaining raw `http.Error` behavior is limited to the old resolve handler and disabled create compatibility bridge on this branch. Some successful delete/resolve paths still intentionally write status-only `204` responses. | `resolvemarket.go`, `createmarket.go` compatibility bridge | Finish or retire the remaining compatibility entry points without changing successful status-only action semantics. |
| Legacy logger fallback | `logger.RequestLoggingMiddleware` still has a raw `http.Error` fallback for a nil handler, but the main server build path now uses `security.RequestBoundaryMiddleware`. | `logger/middleware.go` | Retire or convert this fallback after confirming no live route wiring still wraps requests with the old logger middleware. |
| `backend/errors` package | No live request-boundary import remains after WAVE03; the package is compatibility-only test-covered code. | `errors/httperror.go`, `errors/normalerror.go` | Package deletion is a separate cleanup after route-family migrations confirm no generated docs, tests, or legacy adapters still depend on it. |

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

## OpenTelemetry and Errors

OpenTelemetry applies to runtime observability of failures. It should not replace the public HTTP failure contract.

### Public contract versus telemetry

For API clients, the contract remains:

- HTTP status
- stable public `reason` values for migrated envelope routes
- explicit legacy plain-text or other route-family exceptions while migration is incomplete

For operators and runtime tooling, the contract should move toward OpenTelemetry-aligned telemetry:

- trace correlation
- stable failure classification such as `error.type`
- one-time exception or failure recording
- metrics and traces that can be correlated with logs

### What should count as a telemetry error

Unexpected runtime failures should be treated as telemetry errors. Examples include:

- panics
- DB or dependency failures
- encoding failures
- unexpected 5xx responses
- auth or middleware failures when the operation genuinely failed at runtime

Handled business rejections should not automatically be marked as telemetry errors just because the user-visible outcome is negative. Examples include:

- `MARKET_CLOSED`
- `INSUFFICIENT_BALANCE`
- `NO_POSITION`
- `PASSWORD_CHANGE_REQUIRED`

Those outcomes are often expected policy results rather than infrastructure faults. The runtime boundary may still log or count them deliberately, but it should not blindly mark every such case as a failed span.

### Recording rules

When an operation truly fails at runtime, the backend should:

- set error status on the relevant span
- attach a stable `error.type`
- record the exception or failure once
- avoid sensitive data in span status, attributes, or logs

When an error is handled and the operation completes gracefully, the backend should avoid recording it as the final operation error signal.

Telemetry failure must never change application behavior. If tracing, export, or correlation machinery misbehaves, the request path must still preserve the same client-visible semantics.

## Boundary Direction

### Domain, repository, and auth internal seams

These seams should own typed internal failures and business-rule rejections. They should not be the long-term owners of:

- HTTP status codes
- route-visible reason strings
- response envelopes
- middleware behavior

This is why auth now exposes typed internal outcomes and leaves HTTP status and public reason mapping to `handlers/authhttp`.

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

The remaining compatibility scope is explicitly transitional:

- `backend/errors`

It should not become the future architectural center of error handling. Auth-local `HTTPError` has already been replaced by typed internal auth outcomes for the migrated protected profile/password slice.

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
- align unexpected runtime failure telemetry with OpenTelemetry-compatible trace correlation and stable failure classification
- make middleware and handler failure behavior coherent for protected JSON routes
- keep secrets, passwords, tokens, and API keys out of logs and client-visible failures
- keep public `reason` values separate from internal telemetry fields such as `error.type`
- avoid recording the same exception or failure at multiple layers unless the signals are intentionally different
- reject generic retry or circuit-breaker behavior for accounting-sensitive writes unless a separate design proves idempotency and safety

The intended direction is not:

- a new framework-first `AppError` platform
- a mandatory base handler class
- expanding `backend/errors` into the central architecture
- treating every negative business outcome as an OpenTelemetry error by default
- using OpenTelemetry telemetry fields as the public API error contract
- hidden automatic recovery logic for write-heavy financial flows
- assuming middleware failures do not matter because they happen before handlers

## Concrete Next Migration Goals

1. Finish or retire the remaining market compatibility entry points that still emit raw `http.Error`, starting with the old resolve handler and disabled create bridge.
2. Keep successful status-only `204` responses separate from failure migration work; only status-only failures should be converted to an explicit failure contract.
3. Retire or convert the legacy logger middleware fallback after confirming no live server route build still uses it.
4. Delete `backend/errors` only as a later cleanup after compatibility tests and adapters are removed; do not expand it into the central architecture.
5. Keep OpenAPI and production notes route-family accurate throughout the migration instead of claiming universal envelope coverage too early.

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
- one OpenTelemetry-aligned runtime failure telemetry story that stays separate from public `reason` values
- one error-handling direction consistent with high availability, fault tolerance, and accounting safety
