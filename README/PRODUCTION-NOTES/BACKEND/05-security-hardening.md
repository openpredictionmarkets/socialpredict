---
title: Security Hardening
document_type: production-notes
domain: backend
author: Patrick Delaney
updated_at: 2026-04-30T05:38:00Z
updated_at_display: "Thursday, April 30, 2026 at 5:38 AM UTC"
update_reason: "Close WAVE05 with a security stop-and-review inventory and the next precise auth-boundary seam."
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
- per-client-identity bucketing through the explicit runtime proxy-trust contract

Important current limitations:

- rate limiting is process-local and in-memory; each backend replica enforces its own bucket state
- forwarded IP headers are only trusted when `TRUST_PROXY_HEADERS=true`
- production still needs an explicit ingress contract for forwarded-header scrubbing and any trusted proxy allowlist
- distributed or proxy-owned rate limiting is still deferred

This means the active problem is not “build advanced rate limiting from scratch.” The active problem is to keep the current limiter safe at the boundary while the future deployment model decides whether rate limiting remains application-local, moves to a proxy, or becomes distributed.

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

### Auth still has request-boundary cleanup seams

```go
type AuthError struct {
    Kind    ErrorKind
    Message string
}
```

The current auth helper layer now returns `AuthError` values from [auth.go](/workspace/socialpredict/backend/internal/service/auth/auth.go), and migrated handlers translate those through [authhttp.go](/workspace/socialpredict/backend/handlers/authhttp/authhttp.go).

That is cleaner than the earlier transport-shaped `HTTPError` seam, but auth is not fully centralized yet:

- auth helpers still take `*http.Request`
- some legacy market paths still translate auth failures into raw `http.Error` responses
- [resolvemarket.go](/workspace/socialpredict/backend/handlers/markets/resolvemarket.go) still parses JWTs directly and reads `JWT_SIGNING_KEY` instead of going through the auth service contract

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

### Runtime/bootstrap now owns the first security posture slice

As of WAVE05-SEC-001, [internal/app/runtime/security.go](/workspace/socialpredict/backend/internal/app/runtime/security.go) owns the first explicit runtime security snapshot for serving:

- `JWT_SIGNING_KEY` is required at startup for the auth-serving path; a missing or blank key is a fatal bootstrap error, and server handler construction also fails closed if no signing key is injected.
- `TRUST_PROXY_HEADERS` is read once by runtime bootstrap and passed into the shared client-identity extractor used by request-boundary logging and rate limiting. The default is `false`, so `X-Forwarded-For` and `X-Real-IP` are ignored unless deployment explicitly opts in.
- CORS is read once by runtime bootstrap and passed into server wiring. Current defaults intentionally preserve existing behavior: CORS enabled, wildcard origins, broad API methods, `Content-Type` plus `Authorization` headers, no credentials, and `600` second max age.
- Security headers are still code-defined defaults from `backend/security`, but server wiring now receives them through the runtime security snapshot. Application HSTS is disabled by default because TLS termination and HSTS ownership may belong at ingress or proxy; operators can enable app HSTS with `SECURITY_HSTS_ENABLED=true`, `SECURITY_HSTS_MAX_AGE`, `SECURITY_HSTS_INCLUDE_SUBDOMAINS`, and `SECURITY_HSTS_PRELOAD`.

This slice deliberately keeps deployment-sensitive runtime posture separate from application-policy configuration. `setup` and `internal/service/config` should not become the home for JWT signing material, proxy-header trust, CORS deployment posture, or TLS/HSTS ownership.

### WAVE05 stop-and-review inventory

WAVE05 hardened the active request-boundary security surface without reopening the deferred security-platform backlog:

- runtime/bootstrap now owns JWT signing-key presence, CORS posture, proxy-header trust, and HSTS configuration before route registration
- request-boundary logging and rate limiting share the same explicit client-identity extractor, so spoofable forwarded headers are ignored unless `TRUST_PROXY_HEADERS=true`
- rate-limit and method-not-allowed middleware failures now use shared runtime JSON failure envelopes instead of plain-text middleware responses
- login and admin-user validation receive the shared security service from application wiring instead of constructing ad hoc security helpers inside each request path
- the auth facade can receive the runtime-owned JWT signing key, and the private action routes `/v0/bet`, `/v0/sell`, and `/v0/userposition/{marketId}` enforce `mustChangePassword` with the shared `PASSWORD_CHANGE_REQUIRED` failure reason before domain handlers run

The remaining security-specific exceptions after this wave are intentionally narrow:

- [resolvemarket.go](/workspace/socialpredict/backend/handlers/markets/resolvemarket.go) is the next auth-boundary migration seam because it still parses bearer tokens directly, reads `JWT_SIGNING_KEY` in the handler path, and emits raw `http.Error` auth and authorization failures instead of going through `AuthService` plus `authhttp`.
- The remaining raw `http.Error` responses are concentrated in market handlers. They are a route-family boundary migration, not evidence that middleware or runtime security still lacks a shared failure contract.
- Forwarded client identity remains spoofable unless production ingress scrubs untrusted `X-Forwarded-For` and `X-Real-IP` before enabling `TRUST_PROXY_HEADERS=true`. Until that deployment contract exists, the safe app default remains to ignore forwarded headers.
- CORS still defaults to wildcard origins to preserve current behavior. Production must set explicit origins before treating the API as hardened for public traffic.

The concrete follow-on queue slice should be: migrate `handlers/markets/resolvemarket.go` to the injected auth service and shared failure-envelope path, then carry the same pattern through the remaining market handler raw-error slices. That is a precise auth and boundary-hardening seam. It is not a prompt to activate MFA, RBAC, distributed rate limiting, request signing, session redesign, or compliance work.

### `mustChangePassword` is already part of the live security policy

The backend already enforces a locked password-change policy in [auth.go](/workspace/socialpredict/backend/internal/service/auth/auth.go) through `CheckMustChangePasswordFlag`.

The current server behavior is also tested in [server_contract_test.go](/workspace/socialpredict/backend/server/server_contract_test.go):

- users flagged with `MustChangePassword` may still use `/v0/changepassword`
- other authenticated actions are intended to be blocked once the enforcement path is touched
- the private action route family `/v0/bet`, `/v0/sell`, and `/v0/userposition/{marketId}` is explicitly blocked at the route boundary with the shared `PASSWORD_CHANGE_REQUIRED` failure reason

This means password-change enforcement is not a speculative future feature. It is current security behavior that needs cleaner boundary ownership and clearer route-family consistency.

### Wider route-family cleanup is still a boundary-migration problem

The remaining non-test raw `http.Error` calls are concentrated in market handlers, including security-adjacent request paths such as:

- [createmarket.go](/workspace/socialpredict/backend/handlers/markets/createmarket.go)
- [handler.go](/workspace/socialpredict/backend/handlers/markets/handler.go)
- [listmarkets.go](/workspace/socialpredict/backend/handlers/markets/listmarkets.go)
- [getmarkets.go](/workspace/socialpredict/backend/handlers/markets/getmarkets.go)
- [marketdetailshandler.go](/workspace/socialpredict/backend/handlers/markets/marketdetailshandler.go)
- [resolvemarket.go](/workspace/socialpredict/backend/handlers/markets/resolvemarket.go)
- [searchmarkets.go](/workspace/socialpredict/backend/handlers/markets/searchmarkets.go)

That overlap matters, but this note should not pretend that all route-family response cleanup belongs to a standalone security-platform initiative. It is tied directly to the active error-handling and auth-alignment work. The explicit remaining slice list is carried in [FUTURE/01-long-term-security-hardening.md](/workspace/socialpredict/README/PRODUCTION-NOTES/BACKEND/FUTURE/01-long-term-security-hardening.md).

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

- internal auth code should keep route-visible status and response shape out of core auth decisions
- route-visible auth failures should be translated at handlers or middleware
- `mustChangePassword` remains a server-side enforcement concern
- login stays usable while broader token redesign is deferred

### Deployment-sensitive security posture

This note should be explicit about the runtime assumptions that affect safe operation:

- whether forwarded IP headers are trusted
- where TLS termination occurs
- whether HSTS belongs in app headers, ingress, or both
- how missing JWT runtime config should fail

### Runtime/bootstrap ownership added in WAVE05-SEC-001

Runtime/bootstrap now validates and freezes the deployment-sensitive security posture before the HTTP server is built:

- `JWT_SIGNING_KEY` is required at startup; a missing or blank value is a fatal runtime configuration error rather than a deferred login-path failure.
- `TRUST_PROXY_HEADERS` defaults to `false`; when enabled, the shared client-identity extractor may use `X-Forwarded-For` or `X-Real-IP` for request logging and rate limiting, so production ingress must scrub untrusted forwarded headers before they reach the app.
- `CORS_ENABLED` defaults to enabled, `CORS_ALLOW_ORIGINS` defaults to `*`, `CORS_ALLOW_METHODS` defaults to `GET,POST,PUT,PATCH,DELETE,OPTIONS`, `CORS_ALLOW_HEADERS` defaults to `Content-Type,Authorization`, and production should override origins explicitly.
- Security headers continue to be applied by the existing `backend/security` package, but runtime owns HSTS posture through `SECURITY_HSTS_ENABLED`, `SECURITY_HSTS_MAX_AGE`, `SECURITY_HSTS_INCLUDE_SUBDOMAINS`, and `SECURITY_HSTS_PRELOAD`. HSTS is disabled by default until TLS termination ownership is explicit.

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
2. Use the next auth-boundary slice to migrate [resolvemarket.go](/workspace/socialpredict/backend/handlers/markets/resolvemarket.go) away from direct JWT parsing and raw auth failures.
3. Use the active error-handling wave to converge remaining market handler failures and other sanitized boundary behavior on shared envelopes.
4. Tighten CORS, proxy-trust, and header posture once deployment assumptions are explicit.
5. Keep long-term identity and security-platform work deferred until the active production notes and current design-plan waves are complete.

## Open Questions

- Should HSTS be owned in application headers, ingress or proxy policy, or both?
- What exact proxy-trust model should govern `X-Forwarded-For` and `X-Real-IP` handling beyond the current `TRUST_PROXY_HEADERS` opt-in?
- When does SocialPredict actually need distributed rate limiting rather than the current per-process limiter?
- What is the intended long-term runtime contract for JWT signing key management and rotation?
- Which middleware-generated security failures should eventually use stable public `reason` values across route families?

## Explicit Do-Not-Do List

- Do not create a new top-level `security/auth`, `security/monitoring`, `security/api`, or `middleware` platform tree as part of the active slice.
- Do not treat RBAC, MFA, or refresh-token/session redesign as current-wave requirements.
- Do not add Redis or distributed rate limiting to the active wave queue by default.
- Do not blur security hardening with the broader error-contract migration by claiming universal envelope coverage before the touched route families actually converge.
- Do not make compliance claims in the active production note unless SocialPredict has a real program, owner, and implementation path for them.
