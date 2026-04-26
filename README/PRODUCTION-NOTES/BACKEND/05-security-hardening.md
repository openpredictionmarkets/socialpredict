---
title: Security Hardening
document_type: production-notes
domain: backend
author: Patrick Delaney
updated_at: 2026-04-26T14:07:06Z
updated_at_display: "Sunday, April 26, 2026 at 2:07 PM UTC"
update_reason: "Replace the older greenfield security-platform plan with current-state-first security hardening guidance grounded in the live backend."
status: active
---

# Security Hardening

## Update Summary

This note was updated on Sunday, April 26, 2026 to replace an older greenfield security-platform plan with guidance that matches the live SocialPredict backend, the active design-plan posture, and the high-availability and fault-tolerance objective.

| Topic | Prior to April 26, 2026 | After April 26, 2026 |
| --- | --- | --- |
| Core framing | Treated security hardening as a new platform to build from scratch | Treats security hardening as incremental boundary and runtime hardening of the backend that already exists |
| Current-state accuracy | Assumed JWT auth lived in `middleware/auth.go` and that major security building blocks were still mostly ahead | Recognizes the live `security` package, auth helpers, request validation, rate limiting, headers, and CORS wiring already in production code |
| Main proposal | Build refresh tokens, MFA, RBAC, Redis rate limiting, monitoring, and API signing as the primary move | Focus on rate limiting, validation, sanitization, headers, CORS, auth transport cleanup, and password-change enforcement first |
| Architecture posture | Proposed new `security/auth`, `security/monitoring`, `security/api`, and `middleware/*` trees | Extends the live `backend/security`, `internal/service/auth`, `server`, and handler-boundary seams |
| Failure posture | Blurred security behavior with generic middleware/platform expansion | Aligns security middleware and auth failures with the existing boundary-owned error-handling direction |
| HA posture | Optimized for feature breadth first | Optimizes for deterministic request rejection, sanitized failures, explicit proxy and runtime assumptions, and replica-safe behavior |
| Future ideas | Mixed long-term identity/security-platform ideas into the active note | Defers long-term ideas to [FUTURE/01-long-term-security-hardening.md](/workspace/socialpredict/README/PRODUCTION-NOTES/BACKEND/FUTURE/01-long-term-security-hardening.md) |

## Executive Direction

SocialPredict should treat security hardening as request-boundary and runtime hardening of the existing backend, not as a greenfield security platform.

The backend direction is:

1. Keep shared security primitives in [backend/security](/workspace/socialpredict/backend/security/security.go) and tighten them incrementally rather than replacing them with a second subsystem.
2. Treat rate limiting, validation, sanitization, security headers, CORS posture, and request-path auth checks as the current hardening surface.
3. Keep identity and authorization outcomes close to internal auth and domain seams, while moving route-visible HTTP policy to handlers or middleware boundaries over time.
4. Align middleware-generated `429`, `405`, auth rejection, and other boundary failures with the shared error-handling direction instead of preserving scattered plain-text `http.Error` behavior.
5. Keep JWT key ownership and other security-sensitive runtime settings under runtime/bootstrap ownership rather than hiding them in ad hoc helpers or application-policy config.
6. Preserve the current `mustChangePassword` policy and make its route-boundary behavior explicit before attempting larger auth redesigns.
7. Defer distributed rate limiting, MFA, RBAC, token/session redesign, request signing, and compliance-program work until the active notes and design-plan waves land.

For a high-availability, fault-tolerant, enterprise-ready system, the backend should prefer:

- deterministic request acceptance and rejection behavior across replicas
- sanitized and eventually correlated runtime failures at the request boundary
- explicit runtime assumptions for CORS, trusted proxy headers, and JWT key presence
- minimal secret exposure in logs and client responses
- security features that harden current behavior before adding platform complexity

This note explicitly rejects building a large new security subsystem tree as the main design move for the active slice.

## Why This Matters

Security hardening in SocialPredict is not only about adding more controls. It is also about making the controls that already exist behave predictably in production.

For a high-availability and fault-tolerant backend, that means:

- rate limiting should reject traffic consistently rather than through ad hoc strings
- auth failures should be boundary-owned and predictable rather than leaking transport policy from internal seams
- validation and sanitization should be reusable and explicit instead of being copied into handlers
- headers and CORS should reflect an intentional deployment posture rather than permissive defaults carried forward indefinitely
- sensitive runtime conditions such as missing JWT keys should fail safely

The older note was useful as a wishlist, but it no longer matches the live backend. The current job is not to invent a new security platform. The current job is to make the existing security surface architecturally consistent and safer to operate.

## Current Code Snapshot

As of 2026-04-26, the backend already has meaningful security structure, but it is split between good direction and transitional behavior.

### Shared security package already exists

The backend already has a concrete security package:

- [security.go](/workspace/socialpredict/backend/security/security.go)
- [ratelimit.go](/workspace/socialpredict/backend/security/ratelimit.go)
- [headers.go](/workspace/socialpredict/backend/security/headers.go)
- [validator.go](/workspace/socialpredict/backend/security/validator.go)
- [sanitizer.go](/workspace/socialpredict/backend/security/sanitizer.go)

This means the active backend is not missing a security layer. It already has one. The work now is to harden and clarify ownership.

### Rate limiting exists, but it is local and transport-rough

Rate limiting already exists in [ratelimit.go](/workspace/socialpredict/backend/security/ratelimit.go).

The live behavior already provides:

- a general in-memory limiter
- a stricter login limiter
- per-client-IP bucketing based on request headers

Important current limitations:

- rate limiting is process-local and in-memory
- `429` responses still use plain-text `http.Error`
- client-IP extraction trusts forwarded headers without an explicit proxy-trust model

This means the active problem is not “build advanced rate limiting from scratch.” The active problem is to make the current limiter safer and more consistent at the boundary.

### Validation and sanitization already exist and are used directly

Validation and sanitization are already live in:

- [validator.go](/workspace/socialpredict/backend/security/validator.go)
- [sanitizer.go](/workspace/socialpredict/backend/security/sanitizer.go)
- [security.go](/workspace/socialpredict/backend/security/security.go)

These helpers already enforce:

- username rules
- password rules
- safe-string checks
- market outcome validation
- betting amount validation
- market-title, description, emoji, and personal-link sanitization

They are already used in real request paths, including login in [loggin.go](/workspace/socialpredict/backend/internal/service/auth/loggin.go), which sanitizes the username and validates the login payload before authentication.

The current opportunity is to keep these helpers as shared request-boundary support, not to rebuild a new schema engine or validation platform for the active slice.

### Auth still leaks transport policy from an internal seam

The current auth helper layer still defines a transport-shaped error in [authutils.go](/workspace/socialpredict/backend/internal/service/auth/authutils.go):

```go
type HTTPError struct {
    StatusCode int
    Message    string
}
```

And request-path auth helpers still return that type directly from [auth.go](/workspace/socialpredict/backend/internal/service/auth/auth.go):

```go
func ValidateTokenAndGetUser(r *http.Request, svc dusers.ServiceInterface) (*dusers.User, *HTTPError)
```

That keeps route-visible HTTP status and message policy inside an internal seam, which is inconsistent with the current error-handling and design-plan direction.

At the same time, login already uses the shared envelope path in [loggin.go](/workspace/socialpredict/backend/internal/service/auth/loggin.go) through `handlers.WriteFailure`, so the live system is mixed rather than empty.

### Header and CORS posture already exists, but it is only partly owned

Security headers are already applied by [headers.go](/workspace/socialpredict/backend/security/headers.go), including:

- `Content-Security-Policy`
- `X-Content-Type-Options`
- `X-Frame-Options`
- `Referrer-Policy`
- `Permissions-Policy`
- cross-origin embedder/opener/resource policy headers

But the live header posture is still transitional:

- defaults are static and code-defined
- HSTS is not currently part of the runtime header policy
- the note should treat TLS termination and HSTS ownership as an explicit deployment question, not assume application-only ownership

CORS is already runtime-configured in [server.go](/workspace/socialpredict/backend/server/server.go) through environment variables, but current defaults remain broad:

- `CORS_ENABLED` defaults to enabled
- `CORS_ALLOW_ORIGINS` defaults to `*`
- allowed methods and headers are broad by default

That is current production posture, not future theory, and the note should describe it honestly.

### `mustChangePassword` is already part of the live security policy

The backend already enforces a locked password-change policy in [auth.go](/workspace/socialpredict/backend/internal/service/auth/auth.go) through `CheckMustChangePasswordFlag`.

The current server behavior is also tested in [server_contract_test.go](/workspace/socialpredict/backend/server/server_contract_test.go):

- users flagged with `MustChangePassword` may still use `/v0/changepassword`
- other authenticated actions are intended to be blocked once the enforcement path is touched

This means password-change enforcement is not a speculative future feature. It is current security behavior that needs cleaner boundary ownership and clearer route-family consistency.

### Wider route-family cleanup is still a boundary-migration problem

Many handlers still emit raw `http.Error` strings across the backend, including security-adjacent request paths such as:

- [createmarket.go](/workspace/socialpredict/backend/handlers/markets/createmarket.go)
- [publicuser.go](/workspace/socialpredict/backend/handlers/users/publicuser.go)
- [portfolio.go](/workspace/socialpredict/backend/handlers/users/publicuser/portfolio.go)
- [financial.go](/workspace/socialpredict/backend/handlers/users/financial.go)

That overlap matters, but this note should not pretend that all route-family response cleanup belongs to a standalone security-platform initiative. It is tied directly to the active error-handling and auth-alignment work.

## What Security Hardening Should Own

### Request-boundary security ownership

Security hardening should own:

- rate limiting posture
- trusted-client-IP extraction assumptions
- security headers
- CORS posture
- request-body validation and sanitization at the boundary
- JWT key presence and basic auth bootstrap expectations

### Auth-boundary cleanup direction

This note should also own the direction that:

- internal auth code should move away from transport-shaped `HTTPError`
- route-visible auth failures should be translated at handlers or middleware
- `mustChangePassword` remains a server-side enforcement concern
- login stays usable while broader token redesign is deferred

### Deployment-sensitive security posture

This note should be explicit about the runtime assumptions that affect safe operation:

- whether forwarded IP headers are trusted
- where TLS termination occurs
- whether HSTS belongs in app headers, ingress, or both
- how missing JWT runtime config should fail

## What This Note Should Not Own

This note should not become the home for every long-term security idea.

It should explicitly defer:

- MFA
- refresh-token or revocation or blacklist architecture
- broader session-management redesign
- RBAC or fine-grained authorization framework design
- Redis or distributed rate limiting
- request signing or anti-replay systems
- encryption-platform design
- security-event monitoring platform design
- compliance-program promises such as SOC 2, GDPR, or PCI mapping

Those topics now belong in [FUTURE/01-long-term-security-hardening.md](/workspace/socialpredict/README/PRODUCTION-NOTES/BACKEND/FUTURE/01-long-term-security-hardening.md), not in the active production note.

## Near-Term Sequencing

The near-term security direction should align with the current design-plan waves rather than invent a separate security-platform track.

1. Keep configuration and runtime ownership explicit so JWT key presence, CORS posture, and runtime-sensitive security settings are not hidden in ad hoc helpers.
2. Use the active error-handling wave to converge boundary-owned `429`, `405`, and sanitized failure behavior where security middleware currently bypasses the shared contract.
3. Use the later auth and API alignment wave to retire `internal/service/auth.HTTPError` and clean up route-visible auth failure mapping.
4. Tighten CORS, proxy-trust, and header posture once deployment assumptions are explicit.
5. Keep long-term identity and security-platform work deferred until the active production notes and current design-plan waves are complete.

## Open Questions

- Should HSTS be owned in application headers, ingress or proxy policy, or both?
- What exact proxy-trust model should govern `X-Forwarded-For` and `X-Real-IP` handling in rate limiting?
- When does SocialPredict actually need distributed rate limiting rather than the current per-process limiter?
- What is the intended long-term runtime contract for JWT signing key management and rotation?
- Which middleware-generated security failures should eventually use stable public `reason` values across route families?

## Explicit Do-Not-Do List

- Do not create a new top-level `security/auth`, `security/monitoring`, `security/api`, or `middleware` platform tree as part of the active slice.
- Do not treat RBAC, MFA, or refresh-token/session redesign as current-wave requirements.
- Do not add Redis or distributed rate limiting to the active wave queue by default.
- Do not blur security hardening with the broader error-contract migration by claiming universal envelope coverage before the touched route families actually converge.
- Do not make compliance claims in the active production note unless SocialPredict has a real program, owner, and implementation path for them.
