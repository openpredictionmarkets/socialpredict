---
title: Data Validation
document_type: production-notes
domain: backend
author: Patrick Delaney
updated_at: 2026-05-02T23:30:00Z
updated_at_display: "Saturday, May 2, 2026 at 11:30 PM UTC"
update_reason: "Close WAVE10 with the landed markets boundary-validation consolidation, remaining route-family bypass inventory, and explicit deferral of framework-heavy validation ideas."
status: active
---

# Data Validation

## Update Summary

This note was updated on Saturday, May 2, 2026 to close WAVE10 with the validation-specific stop-and-review point. The wave consolidated the markets request-boundary seam without introducing a new validation framework, and this note now records the remaining route-family bypasses that should drive the next slice.

| Topic | Prior to April 27, 2026 | After April 27, 2026 |
| --- | --- | --- |
| Core framing | Treated validation as a new framework and middleware subsystem to build | Treats validation as consolidation and hardening of the boundary and domain validation that already exists |
| Current-state accuracy | Assumed validation and sanitization were thin and mostly absent | Recognizes the live `backend/security` validator and sanitizer, route usage in handlers and auth, and existing domain-level validators |
| Main proposal | Build a top-level validation engine, rule registry, and request-body middleware | Keep shared request-boundary validation in `backend/security`, keep use-case and business invariants in domain services, and converge failure shaping incrementally |
| Architecture posture | Proposed a second validation architecture | Extends the live `security`, handlers, auth, and domain seams |
| Output strategy | Mixed output sanitization and broad framework design into the same plan | Focuses on boundary input validation, sanitization, and stable failure behavior first |
| HA posture | Optimized for framework breadth | Optimizes for predictable request rejection and safe failure shaping across replicas |
| Future ideas | Mixed registry, middleware, and platform ideas into the active note | Defers broader framework ideas instead of letting them drive the active slice |

## Executive Direction

SocialPredict should treat data validation as a shared request-boundary and domain-policy concern in the backend that already exists, not as a greenfield validation platform.

The active direction is:

1. Keep request parsing, validation, and sanitization close to the boundary through [backend/security](/workspace/socialpredict/backend/security/security.go) and touched handler helpers.
2. Keep business or use-case invariants in domain services instead of moving them into generic middleware or a database-backed rules engine.
3. Continue using sanitization and validation to reject unsafe input predictably rather than letting raw parser or DB failures shape the public contract.
4. Converge validation failure behavior toward stable client-visible outcomes instead of growing more ad hoc strings.
5. Reduce one-off sanitizer construction and mixed failure shaping where practical, but do not build a second architecture just to centralize logic that already has a reasonable owner.
6. Defer generic validation registries, request-body rewriting middleware, broad output-sanitization programs, and other framework-heavy ideas.

For a high-availability and fault-tolerant backend, validation should prefer:

- deterministic rejection of bad input
- shared boundary helpers over copy-pasted checks
- domain-owned business invariants
- sanitized failure messages
- no new platform layer unless the existing seams are proven insufficient

This note explicitly rejects building a large new `validation/` or `sanitization/` subsystem as the immediate main move.

## WAVE10 Stop-And-Review Outcome

WAVE10 landed a narrow boundary-validation consolidation in the markets route family:

- [server.go](/workspace/socialpredict/backend/server/server.go) and [container.go](/workspace/socialpredict/backend/internal/app/container.go) now pass the runtime-owned `*security.SecurityService` into markets handlers instead of letting production handlers construct fresh security services or sanitizers.
- [createmarket.go](/workspace/socialpredict/backend/handlers/markets/createmarket.go) and [searchmarkets.go](/workspace/socialpredict/backend/handlers/markets/searchmarkets.go) use the shared security seam for create/search sanitization and map malformed, sanitizer-rejected, and domain-validation failures to stable `ReasonResponse` values.
- [queryparams.go](/workspace/socialpredict/backend/security/queryparams.go) now owns the shared bounded integer query parsing helper used by markets list/search pagination.
- [docs/API-ISSUES.md](/workspace/socialpredict/backend/docs/API-ISSUES.md) and [openapi.yaml](/workspace/socialpredict/backend/docs/openapi.yaml) document the create/search validation failure split: malformed or sanitizer-rejected input is `INVALID_REQUEST`, while syntactically valid domain validation failures are `VALIDATION_FAILED`.

That work was intentionally not a universal envelope migration and did not move business invariants out of `internal/domain/markets`.

## Remaining Boundary-Validation Bypasses

The remaining bypasses are route-family seams, not proof that SocialPredict needs a new validation platform:

| Route family | Remaining bypass | Why it matters |
| --- | --- | --- |
| Markets legacy/details/actions | `handlers/markets/marketdetailshandler.go`, `resolvemarket.go`, `marketprojectedprobability.go`, and legacy methods on `handlers/markets/handler.go` still parse path/action input locally and some still use plain `http.Error` failure shaping | This is the closest continuation of WAVE10 because it keeps the work in the already-touched markets boundary and can reuse shared market ID, amount, outcome, and failure helpers |
| Market bets and positions | `handlers/bets/betshandler.go` and `handlers/positions/positionshandler.go` have package-local `marketId` parsing helpers | Their failures already use `ReasonResponse`, but the parsing rule is duplicated instead of living in a shared request-boundary helper |
| Private actions | `/v0/bet` and `/v0/sell` decode payloads locally and rely on domain services for amount/outcome validation | This preserves domain invariants, but there is still a boundary-helper gap for DTO-level malformed versus domain-invalid input if the route family is revisited |
| Admin and content | Admin create-user now uses injected security validation, while CMS content remains a domain-specific HTML rendering/sanitization seam | These should not drive the next validation slice unless a concrete route behavior issue appears |

The next validation seam should be the remaining markets path/action helper gap: consolidate market ID, projection amount, resolution outcome, and related failure shaping for market detail, resolve, projection, and legacy markets handler methods. That is a concrete route-family continuation of WAVE10 and should happen before any discussion of registries, generic request-body middleware, or a platform-wide validation subsystem.

## Why This Matters

Validation is one of the places where architectural drift happens quietly.

If validation logic spreads randomly:

- handlers diverge in request rejection behavior
- clients see inconsistent failure formats
- domain rules leak upward into transport code
- sanitization becomes ad hoc and harder to reason about

The current backend already has a meaningful validation and sanitization seam. The active job is to make that seam consistent and boundary-safe, not to replace it with a generalized framework.

## Current Code Snapshot

### Shared boundary validation and sanitization already exists

The live security seam already includes:

- [validator.go](/workspace/socialpredict/backend/security/validator.go)
- [sanitizer.go](/workspace/socialpredict/backend/security/sanitizer.go)
- [security.go](/workspace/socialpredict/backend/security/security.go)

That seam already owns reusable checks for:

- usernames
- passwords
- safe strings
- market titles
- personal links
- emojis
- market outcomes
- positive amounts

This is already a real validation layer. It just needs clearer consolidation and failure-shaping discipline.

### Handlers already use the shared security seam

Representative live usage already exists in:

- [adduser.go](/workspace/socialpredict/backend/handlers/admin/adduser.go)
- [createmarket.go](/workspace/socialpredict/backend/handlers/markets/createmarket.go)
- [searchmarkets.go](/workspace/socialpredict/backend/handlers/markets/searchmarkets.go)
- [loggin.go](/workspace/socialpredict/backend/internal/service/auth/loggin.go)

That means the active note should not pretend validation is still hypothetical.

### The users domain already owns sanitizer-dependent profile rules

The users domain service already accepts a sanitizer dependency and applies it in [service.go](/workspace/socialpredict/backend/internal/domain/users/service.go) for:

- profile description updates
- display-name updates
- personal-link updates
- password changes

This is the right direction. Boundary helpers sanitize request data, while the domain still owns whether the change is allowed and how it mutates state.

### Domain validators already exist for use-case rules

The backend also already has domain-specific validation seams such as:

- bet place and sell validators in [internal/domain/bets/service.go](/workspace/socialpredict/backend/internal/domain/bets/service.go) and [bet_support.go](/workspace/socialpredict/backend/internal/domain/bets/bet_support.go)
- market probability validation in [internal/domain/markets/service.go](/workspace/socialpredict/backend/internal/domain/markets/service.go) and [service_policies.go](/workspace/socialpredict/backend/internal/domain/markets/service_policies.go)

That is important. It means "validation" in SocialPredict is not only about request-body field syntax. It also includes domain-level eligibility and rule enforcement.

### Failure shaping is better in some places than others

Some touched handler families already sanitize failure behavior, such as the profile helpers around [writeProfileError](/workspace/socialpredict/backend/handlers/users/profile_helpers.go) and their tests in [profile_handlers_test.go](/workspace/socialpredict/backend/handlers/users/profile_handlers_test.go).

But the backend still has mixed response families and mixed legacy error shaping elsewhere.

So the active validation work is as much about consistent boundary failure behavior as it is about the validation rules themselves.

## What Data Validation Should Own

This note should own:

- request-boundary validation and sanitization ownership
- shared helper ownership in `backend/security`
- stable public handling of validation failures
- explicit separation between boundary validation and domain invariants
- no-secret and no-unsafe-content expectations for rejection paths

## What This Note Should Not Own

This note should not become the home for every future framework idea.

It should explicitly defer:

- a top-level `validation/` platform
- middleware that rewrites arbitrary request bodies generically
- a DB-backed business-rule registry
- a versioned `/v1` validation framework
- a broad output-sanitization campaign across every response type

Those ideas are not the current priority for the active slice.

## Near-Term Sequencing

The design-plan-aligned validation direction is:

1. Keep `backend/security` as the shared request-boundary seam.
2. Reduce ad hoc one-off validation or sanitizer construction where touched.
3. Converge validation failures toward the shared boundary failure posture already established in the security and API notes.
4. Keep business invariants in domain validators rather than moving them into middleware.
5. Continue with the remaining markets path/action helper gap before moving to bets, positions, or private-action DTO validation.
6. Revisit broader framework ideas only if the current seam proves structurally insufficient.

## Open Questions

- Whether the remaining markets path/action helper gap should live in `backend/security` or in a markets-boundary helper that delegates only generic parsing to `security`.
- Whether market bets and positions should share the same market ID boundary helper after the markets slice lands.
- Whether any output-sanitization follow-up is actually required beyond the current boundary and domain posture.

WAVE10 stops here. A top-level validation framework, rule registry, generic request-body rewriting middleware, broad output-sanitization campaign, validation platform backlog, or DB-backed business-rule registry remains explicitly deferred until the existing route-family seams prove insufficient.
