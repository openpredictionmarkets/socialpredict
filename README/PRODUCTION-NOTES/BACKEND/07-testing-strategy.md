---
title: Testing Strategy
document_type: production-notes
domain: backend
author: Patrick Delaney
updated_at: 2026-04-30T11:55:00Z
updated_at_display: "Thursday, April 30, 2026 at 11:55 AM UTC"
update_reason: "Record the April 30 WAVE05 backend security smoke evidence while keeping the testing strategy package-local and evidence-driven."
status: active
---

# Testing Strategy

## Update Summary

This note was updated on Monday, April 27, 2026 to replace an older greenfield test-platform plan with guidance that matches the live SocialPredict backend, idiomatic Go test layout, and the active design-plan posture.

On Thursday, April 30, 2026, the WAVE05 backend security slice was verified with both package-level Go tests and Docker-backed black-box HTTP checks. That evidence confirms the current testing direction: keep most checks close to the owning packages, then add a small number of black-box smoke checks when runtime wiring, CORS, rate limiting, or auth gates need end-to-end confirmation.

| Topic | Prior to April 27, 2026 | After April 27, 2026 |
| --- | --- | --- |
| Core framing | Treated testing as a large new subsystem to build | Treats testing as evidence for the backend that already exists |
| Current-state accuracy | Claimed testing was limited and that integration/API testing were mostly absent | Recognizes the live suite across handlers, auth, security, runtime, migrations, seed behavior, repositories, domains, OpenAPI, and server contracts |
| Test layout | Proposed a centralized `testing/` tree and shared `TestSuite` | Follows idiomatic Go posture: tests stay near the package they verify, with only a small number of multi-boundary tests centralized later if needed |
| Database-testing posture | Leaned toward a testcontainers-first infrastructure rollout | Keeps fast helpers for broad coverage but treats Postgres-backed checks as the source of truth for locking, transaction, and HA-sensitive runtime behavior |
| CI posture | Assumed a larger new test platform and coverage/performance program | Builds on the live smoke plus `go test ./...` CI path already present in the repo |
| Main risk model | Optimized for framework breadth | Optimizes for readiness, startup discipline, failure handling, accounting correctness, and contract safety under HA pressure |
| Future ideas | Mixed long-term containerized and larger test-platform ideas into the active note | Defers broader infrastructure ideas to [FUTURE/03-long-term-test-infrastructure.md](/workspace/socialpredict/README/PRODUCTION-NOTES/BACKEND/FUTURE/03-long-term-test-infrastructure.md) |

## Executive Direction

SocialPredict should treat testing as a boundary-aware verification strategy for a live backend, not as a greenfield testing platform project.

The active direction is:

1. Keep the default Go posture: tests live beside the code they verify, in `*_test.go` files within the same package directory.
2. Prefer package-local tests for handlers, services, repositories, runtime helpers, migrations, and security behavior rather than a centralized `testing/` subsystem.
3. Use small, explicit test helpers where they reduce repetition, but do not turn helpers into a second architecture.
4. Allow a small number of broader integration tests only when they are genuinely cross-boundary and do not fit cleanly inside one owning package.
5. Treat SQLite-backed helpers as convenience for broad fast coverage, not as the source of truth for Postgres locking, transaction isolation, migration coordination, or HA-sensitive runtime behavior.
6. Add Postgres-backed verification only where the behavior under test is meaningfully database-specific.
7. Keep the current CI posture simple and deterministic, then expand only where the active modernization waves expose real verification gaps.
8. Defer testcontainers-first rollout, shared `TestSuite` design, performance/load suites, and broad coverage-gating programs until there is a concrete need.

For a high-availability, fault-tolerant backend, testing should prefer:

- fast feedback near the owning package
- boundary and contract verification where behavior crosses packages
- source-of-truth verification for DB/runtime behavior that SQLite cannot model credibly
- deterministic tests over infrastructure-heavy test theater
- targeted integration coverage for the real architectural risks

This note explicitly rejects building a large new `testing/` subsystem as the main move for the active slice.

## Why This Matters

The active testing problem in SocialPredict is not the absence of tests. The active problem is making the test strategy match the real codebase and the current architectural risks.

For the current backend, that means:

- handlers should be tested as request-boundary translators
- repositories should be tested as persistence-edge translators
- runtime/bootstrap should be tested for startup ownership and failure posture
- OpenAPI and server wiring should be tested as public contract guards
- accounting-sensitive flows should be tested for correctness under real transaction assumptions

The older note was too greenfield. It assumed the next move was building a generalized framework. The actual next move is to strengthen verification around the modernization seams that matter most for HA and fault tolerance.

## Current Code Snapshot

As of 2026-04-27, the backend already has a meaningful test surface. The issue is not that testing is absent. The issue is that the old note does not describe the testing posture that is already present.

### The backend already has broad package-local coverage

The live backend currently has:

- `94` Go test files
- `343` `Test*` functions
- `35` test files using `httptest`
- `34` test files using `modelstesting.NewFakeDB`

Coverage already exists across:

- handlers
- auth middleware and login behavior
- security middleware and validators
- runtime/bootstrap helpers
- migrations
- seed/bootstrap behavior
- repositories
- domain services and math
- OpenAPI validation and route/spec parity
- server-level contract behavior

This is not a repo with “limited testing.” It is a repo with an existing testing shape that now needs a better architectural description.

### Go-style package-local testing is already the dominant pattern

The current repo already follows the normal Go testing posture:

- tests live in the package directory they verify
- test files are `*_test.go`
- many tests use either the same package or package-level black-box seams in the same directory

Examples:

- [openapi_test.go](/workspace/socialpredict/backend/openapi_test.go)
- [server_contract_test.go](/workspace/socialpredict/backend/server/server_contract_test.go)
- [migrate_test.go](/workspace/socialpredict/backend/migration/migrate_test.go)
- [ratelimit_test.go](/workspace/socialpredict/backend/security/ratelimit_test.go)
- [handler_contract_test.go](/workspace/socialpredict/backend/handlers/markets/handler_contract_test.go)

That package-local posture should remain the default.

### Shared helpers already exist, but they are narrow

The repo already has helper packages such as:

- [modelstesting](/workspace/socialpredict/backend/models/modelstesting/modelstesting.go)
- [setuptesting](/workspace/socialpredict/backend/setup/setuptesting/setuptesting.go)

These helpers are useful because they reduce repetition without requiring a centralized test framework.

That is the right scale for the active slice:

- narrow helpers are good
- helper packages are fine
- a large shared `TestSuite` base architecture is not needed

### Contract and boundary tests already exist

The backend already protects important public seams:

- OpenAPI document validity and route/spec parity in [openapi_test.go](/workspace/socialpredict/backend/openapi_test.go)
- docs/auth/public-route behavior in [server_contract_test.go](/workspace/socialpredict/backend/server/server_contract_test.go)
- handler boundary behavior in [handler_contract_test.go](/workspace/socialpredict/backend/handlers/markets/handler_contract_test.go)

This means the current job is not to invent “API testing” from scratch. The current job is to extend these checks where the active modernization waves need stronger proof.

### CI already runs a backend test path

The repo already has backend CI in [backend.yml](/workspace/socialpredict/.github/workflows/backend.yml).

That workflow already runs:

- a smoke startup path
- `go test ./...`

The current note should build on that reality rather than assuming CI is missing.

### SQLite-backed helpers are useful but limited

Much of the current suite uses [modelstesting.NewFakeDB](/workspace/socialpredict/backend/models/modelstesting/modelstesting.go), which provides:

- in-memory SQLite
- quick migration setup
- fast execution

This is useful for broad package-local verification, but it is not the source of truth for:

- Postgres locking behavior
- transaction-isolation semantics
- multi-replica startup coordination
- migration serialization under real deployment conditions
- runtime readiness against a real Postgres dependency

The active strategy should be explicit: SQLite-backed tests are convenience coverage, not authoritative HA database verification.

### The biggest remaining gaps are architectural, not framework-related

The live backend still has important verification gaps around the same seams the design plan already highlights:

- DB-backed readiness and health semantics
- migration failure posture
- single-writer startup behavior
- proxy-sensitive rate-limiting behavior
- middleware-owned failure convergence
- Postgres-specific transaction behavior
- accounting-sensitive atomicity for place/sell/resolve workflows

Those are the areas where new testing effort should go first.

## What Testing Strategy Should Own

### Default test posture

This note should own the default posture that:

- tests stay near the owning package
- `*_test.go` is the standard file shape
- package-local tests are preferred for most work
- broader cross-package tests are used only where truly justified

### Boundary-aware verification

This note should also own the expectation that tests prove boundaries hold:

- handlers map transport correctly
- repositories translate persistence correctly
- runtime/bootstrap owns startup and readiness behavior
- server and OpenAPI stay aligned
- security middleware behaves consistently
- accounting-sensitive flows preserve correctness

### Database-verification posture

This note should state clearly that:

- fast in-memory helpers are acceptable for broad coverage
- they are not authoritative for Postgres-specific runtime or transaction behavior
- a small amount of Postgres-backed verification is justified when the behavior truly depends on Postgres semantics

## What This Note Should Not Own

This note should not become the home for every ambitious testing-platform idea.

It should explicitly defer:

- a top-level `testing/` subsystem
- a shared `TestSuite` base architecture
- testcontainers-by-default for all integration work
- dedicated performance/load/stress programs
- hard coverage quotas as the main active goal
- benchmark gating
- broad CI redesign
- a separate helper repository unless multi-repo sharing becomes a concrete need

Those topics now belong in [FUTURE/03-long-term-test-infrastructure.md](/workspace/socialpredict/README/PRODUCTION-NOTES/BACKEND/FUTURE/03-long-term-test-infrastructure.md), not in the active production note.

## Near-Term Sequencing

The near-term testing direction should align with the current design-plan waves rather than invent a separate testing platform track.

In practice, that means:

1. Preserve the current package-local test posture and existing helper packages.
2. Expand verification around runtime and DB ownership:
   - readiness semantics, completed for the serving path on April 30, 2026
   - migration failure behavior
   - startup coordination
3. Expand verification around request-boundary convergence:
   - auth failure behavior
   - middleware failure behavior
   - rate limiting and proxy assumptions, with the first spoofed-forwarded-header smoke check completed on April 30, 2026
4. Strengthen accounting-sensitive verification for:
   - place bet
   - sell position
   - resolve market
5. Add a small number of Postgres-backed checks only where SQLite is a poor proxy.
6. Revisit broader infrastructure only after the active runtime, DB, security, and API seams are more stable.

## Open Questions

The current production note should keep the following questions visible:

- Which exact runtime and DB behaviors now require real-Postgres verification rather than SQLite-backed convenience coverage?
- Where should the small number of truly cross-boundary integration tests live if package-local placement becomes awkward?
- Do we want a clear fast/slow split later, and if so, should it be based on runtime cost rather than a large taxonomy of suite types?
- Which current helper packages should remain repo-local, and which would only justify a separate shared helper repo if multiple repos genuinely need them?

## Relationship To Future Test Infrastructure

The active note should stay focused on the current backend and current modernization risks.

Longer-term ideas such as:

- `testcontainers-go`
- centralized Postgres-backed integration suites
- broader performance/load test programs
- more elaborate CI/test orchestration
- multi-repo shared test-helper distribution

are intentionally deferred to [FUTURE/03-long-term-test-infrastructure.md](/workspace/socialpredict/README/PRODUCTION-NOTES/BACKEND/FUTURE/03-long-term-test-infrastructure.md).

## Bottom Line

SocialPredict does not need a testing platform rebuild first.

It needs a testing strategy that matches idiomatic Go layout, preserves fast package-local verification, treats SQLite as convenience rather than truth, and adds heavier Postgres-backed or broader integration checks only where the active HA-focused architecture actually requires them.
