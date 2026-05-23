---
title: ASVS Level 1 Evidence Starter
document_type: reading-notes
domain: backend
scope: security-evidence
author: Patrick Delaney
updated_at: 2026-05-12T00:00:00Z
updated_at_display: "Tuesday, May 12, 2026"
status: active
---

# ASVS Level 1 Evidence Starter

## Purpose

This is a lightweight evidence packet for evaluating SocialPredict against
[OWASP ASVS](https://owasp.org/www-project-application-security-verification-standard/)
Level 1 style expectations.

It is not a certification claim, penetration test, SOC 2 control set, or complete ASVS assessment. Its job is to keep near-term security work evidence-driven and to prevent broad security-platform ideas from replacing concrete controls already visible in the codebase.

The active security note remains [05-security-hardening.md](../05-security-hardening.md). Long-term security ideas remain in [FUTURE/01-long-term-security-hardening.md](../FUTURE/01-long-term-security-hardening.md).

## How To Use This

Use this as a working checklist when deciding whether SocialPredict is ready for wider public production traffic:

1. Mark each topic as `Covered`, `Partial`, `Deferred`, or `Unknown`.
2. Link evidence to code, tests, deployment docs, or operator configuration.
3. Keep exceptions explicit instead of converting them into vague security debt.
4. Do not treat one green row as proof that adjacent rows are also covered.
5. Revisit the matrix after major auth, deployment, or request-boundary changes.

## Current Evidence Matrix

| Area | Current status | Evidence | Exception or follow-up |
| --- | --- | --- | --- |
| Application starts with required secrets | Partial | Runtime security config requires `JWT_SIGNING_KEY` for serving-path construction; install generates a signing key for local/deploy env files. | Key rotation and multi-key validation are deferred. Secret-manager ownership is not yet defined. |
| Password-change gate | Covered for touched private action routes | `mustChangePassword` is enforced before selected protected handlers run; `/v0/changepassword` intentionally remains usable for first-login completion. | Route-family consistency should continue to be checked as auth wrappers are retired. |
| Auth failure response shape | Partial | Migrated handlers map typed auth outcomes to shared public `reason` values through `authhttp` and `handlers.WriteFailure`. | Some compatibility paths remain. Internal auth still has request-shaped wrapper APIs, though token-string seams now exist. |
| JWT validation seam | Partial | `internal/service/auth` has token-string plus `context.Context` validation entry points with injected signing-key support. | Some handlers still call request-shaped compatibility wrappers. `resolvemarket.go` on this branch still parses JWT directly unless PR #685 is also merged. |
| Request-boundary failure containment | Partial | Runtime middleware owns panic recovery, request IDs, JSON `405`, and JSON rate-limit failures. Market read/get/detail/update failure paths now use shared envelopes on this stack. | Remaining compatibility paths should keep shrinking; do not claim universal envelope coverage yet. |
| Rate limiting | Partial | App-level in-memory rate limiting exists, with trusted-proxy behavior gated by `TRUST_PROXY_HEADERS`. | Process-local limits are not distributed. Proxy-owned or distributed rate limiting remains deferred until operational need is clear. |
| CORS posture | Partial | Runtime owns CORS configuration before server construction. | Defaults are broad for compatibility. Production must set explicit origins before treating CORS as hardened. |
| Security headers | Partial | Runtime wires configured security headers; HSTS can be enabled through env. | HSTS ownership across app/proxy/ingress remains a deployment decision. |
| Trusted proxy handling | Partial | Forwarded headers are ignored unless `TRUST_PROXY_HEADERS=true`. | Production ingress must scrub forwarded headers before enabling trust. |
| Liveness/readiness | Covered for serving path | `/health`, `/readyz`, and `/ops/status` are published and tested; `/ops/status` is cache-disabled JSON. | Early startup progress over HTTP is intentionally deferred. |
| Database connection posture | Partial | Runtime owns DB config, readiness ping, pool settings, TLS validation, and deploy topology docs. | Packaged local Postgres permits `DB_REQUIRE_TLS=false`; external production DBs require explicit TLS review. |
| Transaction-sensitive accounting | Partial | Place-bet and sell-position have explicit transaction seams in current PR stack history. Resolve-market payout accounting now has DSN-gated Postgres coverage. | Market resolution transaction/concurrency policy still needs a dedicated slice. |
| Input validation and sanitization | Partial | `backend/security` owns reusable validation/sanitization helpers; market create/search and several user/admin paths use them. | Remaining DTO/body/path validation should keep moving route-family by route-family. |
| Error information exposure | Partial | Shared failure envelopes expose stable public reasons rather than raw internal errors on touched routes. | Legacy compatibility paths and some raw DTO success families remain migration state. |
| OpenAPI accuracy | Partial | `backend/docs/openapi.yaml` documents current mixed route-family behavior and operation surfaces. | Keep updating OpenAPI with each route-contract change. This is not a generated source of truth. |
| Dependency and image hygiene | Unknown | GitHub reports at least one low Dependabot finding on `main` during push output. | Add a short vulnerability register and decide what evidence is required before public production claims. |
| Audit trail and security telemetry | Deferred | Request IDs and runtime logs exist. | Dedicated audit logging, security-event taxonomy, and tamper-resistant audit storage are not current baseline. |
| MFA, refresh tokens, API keys, RBAC | Deferred | Future notes explicitly hold these as long-term identity/platform topics. | Do not start these without a concrete product, threat-model, or automation use case. |

## Immediate Evidence Gaps Worth Closing

These are small enough to consider before larger platform work:

1. Create a vulnerability register that records known dependency/image findings, owner, severity, and disposition.
2. Document production runtime values for CORS, HSTS/TLS ownership, trusted proxy handling, and JWT secret storage.
3. Add a specific market-resolution transaction/concurrency note and test target before expanding more accounting workflows.
4. Continue route-family failure migration until `PlainTextErrorResponse` only covers intentional infra probes.
5. Record whether current DigitalOcean firewall/proxy settings match the assumptions in the infra docs.

## Explicit Non-Claims

This document does not claim:

- ASVS Level 1 compliance
- ASVS Level 2 readiness
- SOC 2 readiness
- penetration-test completion
- full public-production security approval
- full automation-client security support

It is a working evidence map for deciding what to harden next.
