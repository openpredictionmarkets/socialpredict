---
title: Long-Term Security Hardening
document_type: production-notes
domain: backend
author: Patrick Delaney
updated_at: 2026-04-30T03:54:41Z
updated_at_display: "Thursday, April 30, 2026 at 3:54 AM UTC"
update_reason: "Record deferred security-hardening gaps, remaining legacy handler error slices, and the current security measurement posture."
status: future
---

# Long-Term Security Hardening

## Purpose

This note is a holding area for deferred security-platform ideas that are not part of the active SocialPredict backend design plan, not part of the current production-note wave sequence, and not part of the current runnable task queue.

Its purpose is to preserve long-term ideas without letting them distort the active near-term architecture in [05-security-hardening.md](/workspace/socialpredict/README/PRODUCTION-NOTES/BACKEND/05-security-hardening.md).

## Current Status

As of 2026-04-30:

- the active backend security note should stay focused on the live request-boundary and runtime hardening surface
- the active design plan is still sequencing configuration, runtime observability, failure recovery, database runtime ownership, legacy-model decoupling, and later auth or API cleanup
- the ideas in this document are explicitly deferred until those nearer-term seams are stable

This document is non-binding on the active design plan and on `TASKS.json`.

## Security Posture And Measurement

This is not a formal security assessment, penetration test, or compliance claim. It is a working posture statement for planning deferred hardening work.

Assuming production runs with updated base images and dependencies, strong secrets, the database not publicly exposed, and TLS correctly terminated in front of the app, the current backend should be treated as:

- reasonable for local development, staging, and controlled beta traffic
- moderate risk for public production until deployment-sensitive controls are explicitly owned
- not yet enterprise-ready or audit-clean against a full web-app verification standard

The current app is not "wildly insecure" on the code paths already hardened. The database runtime ownership work, startup writer gating, readiness checks, atomic bet placement, request validation and sanitization, runtime failure envelopes, security headers, and proxy-header opt-in behavior are all meaningful baseline controls.

The remaining risk is concentrated in places where security behavior depends on deployment assumptions or unfinished boundary cleanup:

- CORS defaults are still broad unless production overrides origins intentionally.
- HSTS and TLS ownership are still a deployment decision across app, ingress, and reverse proxy.
- Rate limiting is still process-local and in-memory, so it is not a complete abuse-control system across replicas or distributed clients.
- Market handlers still contain the remaining raw `http.Error` response slices.
- `resolvemarket.go` still parses JWTs directly and reads `JWT_SIGNING_KEY` outside the centralized auth path.
- JWT key ownership, validation, and rotation are not yet a fully centralized runtime contract.

There is no single standard security score for an entire application. Useful measurement should use different standards for different questions:

- [OWASP ASVS](https://owasp.org/www-project-application-security-verification-standard/) is the best near-term yardstick for application controls. The practical next step is an ASVS Level 1 evidence matrix, then a Level 2 target before treating SocialPredict as ready for higher-trust public production.
- [CVSS](https://www.first.org/cvss/v4.0/) should be used to score individual vulnerabilities, not the whole app.
- [OWASP SAMM](https://owasp.org/www-project-samm/) should be used if SocialPredict wants to measure the maturity of its secure development lifecycle.
- [NIST SSDF](https://csrc.nist.gov/pubs/sp/800/218/final) should be used for secure software development and supply-chain practice alignment.

The intended future metric is not a single percentage score. It should be a short security evidence packet with:

- ASVS Level 1 requirements mapped to code, configuration, tests, or deployment controls
- explicit exceptions for deferred items in this file
- dependency and image scan evidence
- runtime configuration evidence for CORS, TLS/HSTS, trusted proxy headers, JWT secrets, database access, and rate limiting
- a vulnerability register where individual findings receive CVSS scores only when appropriate

## Candidate Future Topics

The following ideas are reasonable future candidates, but they are not current architecture commitments:

### Deferred active-hardening gaps

The following items remain from the active security-hardening note and should be carried forward deliberately rather than treated as completed:

- retire remaining request-shaped auth seams, direct JWT parsing, and route-local auth failure translation
- tighten CORS defaults and document the intended production origin posture
- decide HSTS ownership across application headers, ingress, and reverse proxy policy
- document TLS termination and trusted proxy assumptions, including which deployments may set `TRUST_PROXY_HEADERS=true`, whether proxy IP allowlists are needed, and how forwarded headers are scrubbed before reaching the app
- define whether and when local in-memory rate limiting should become distributed or proxy-owned
- finish migrating the remaining legacy `http.Error` and `PlainTextErrorResponse` route families to stable `ReasonResponse` envelopes
- centralize JWT signing key ownership, validation, and eventual rotation under an explicit runtime contract

### Remaining legacy `http.Error` route slices

As of 2026-04-30, a non-test scan for `http.Error` under `backend/handlers`, `backend/security`, `backend/server`, and `backend/internal` leaves the remaining legacy response work concentrated in market handlers:

- [createmarket.go](/workspace/socialpredict/backend/handlers/markets/createmarket.go): `CreateMarketService.Handle`, `CreateMarketHandlerWithService`, `currentUserOrError`, `writeCreateMarketError`, and the legacy `CreateMarketHandler` bridge still emit plain-text method, auth, decode, validation, service-domain, and temporary-disabled failures.
- [handler.go](/workspace/socialpredict/backend/handlers/markets/handler.go): `UpdateLabels` and `GetMarket` still emit plain-text method, missing ID, invalid ID, and invalid JSON failures.
- [listmarkets.go](/workspace/socialpredict/backend/handlers/markets/listmarkets.go): `handleListMarkets` and `writeListMarketsError` still emit plain-text method, query-parameter, domain-error, and response-encoding failures.
- [getmarkets.go](/workspace/socialpredict/backend/handlers/markets/getmarkets.go): `GetMarketsHandler` still emits plain-text query-parameter, invalid-input, and internal failures.
- [marketdetailshandler.go](/workspace/socialpredict/backend/handlers/markets/marketdetailshandler.go): `MarketDetailsHandler` still emits plain-text invalid-ID, not-found, invalid-input, and internal failures.
- [resolvemarket.go](/workspace/socialpredict/backend/handlers/markets/resolvemarket.go): `ResolveMarketHandler`, `parseResolveRequest`, `extractUsernameFromRequest`, and `writeResolveError` still emit plain-text parse, token, authorization, state, validation, and internal failures; it also reads `JWT_SIGNING_KEY` directly instead of using the centralized auth service path.
- [searchmarkets.go](/workspace/socialpredict/backend/handlers/markets/searchmarkets.go): `handleSearchMarkets`, `writeSearchError`, and the local `httpError` parser path still emit plain-text method, search-parameter, validation, service-domain, and response-build failures.

No remaining non-test `http.Error` calls were found outside the market handler family in that scan.

### Identity and session evolution

- refresh-token architecture
- token revocation or logout invalidation
- stronger session-management posture
- key rotation strategy
- possible migration away from the current simple JWT posture

### Stronger authentication

- MFA
- step-up authentication for sensitive actions
- device or session anomaly handling

### Authorization model evolution

- RBAC
- resource-scoped authorization policies
- admin capability modeling beyond the current lighter-weight checks

### Distributed edge controls

- distributed or shared rate limiting
- clearer proxy-trust and ingress security posture
- IP reputation or abuse-control improvements if real abuse volume justifies them

### Security telemetry and auditability

- structured security event monitoring
- audit trails for high-sensitivity actions
- alerting for suspicious auth or abuse behavior

### External-facing API security posture

- request signing
- anti-replay controls
- stronger partner or machine-to-machine authentication if the API surface evolves in that direction

### Program and compliance work

- formal compliance mapping
- evidence collection for security controls
- program ownership for broader standards work

## Entry Criteria

This note should only become active planning input after the current architecture is substantially more stable.

Reasonable entry criteria are:

- current production notes 01 through 05 are aligned to live code
- the active design-plan waves through at least the current auth and error-alignment work are complete enough to stop rewriting boundary basics
- JWT key ownership, boundary failure handling, rate limiting behavior, and `mustChangePassword` behavior are explicit and stable
- there is an actual business, threat-model, or operational reason to take on one of the future topics

## What Is Explicitly Deferred

The following items are explicitly not part of the current queue unless later reactivated on purpose:

- MFA rollout
- refresh-token or blacklist implementation
- broad RBAC system design
- Redis-backed distributed rate limiting
- security-monitoring platform buildout
- request-signing or anti-replay architecture
- formal compliance implementation work

## Re-Entry Questions

Before pulling any of these topics into the active plan, SocialPredict should answer:

- What concrete threat or operational problem are we solving?
- Is the current boundary/runtime hardening actually complete enough to support a larger security feature safely?
- Does the topic belong in backend runtime behavior, ingress or deployment posture, or a separate platform decision record?
- Is there a measurable need across replicas or across operators, or is the idea still speculative?

## Relationship To The Active Note

The active note at [05-security-hardening.md](/workspace/socialpredict/README/PRODUCTION-NOTES/BACKEND/05-security-hardening.md) is the binding current-state-first architecture note.

This `FUTURE` note exists so that:

- the active note can stay pragmatic
- the task queue can stay focused
- deferred ideas are not lost
- long-term security ambitions do not get mistaken for near-term design commitments
