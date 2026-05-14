---
title: Future Backend Baseline Triage
document_type: production-notes
domain: backend
author: Patrick Delaney
updated_at: 2026-05-11T21:15:00Z
updated_at_display: "Monday, May 11, 2026 at 9:15 PM CDT"
update_reason: "Align backend future triage with the canonical v3.0.1 design-plan refresh from spec-socialpredict-tasks PR #30."
status: draft
---

# Future Backend Baseline Triage

## Purpose

This note is a triage layer over the backend `FUTURE` notes. It does not activate
all deferred work. It exists to answer: if there is not yet direct customer input,
what baseline production work is most important next, and why?

The decision rule is simple: prefer small, evidence-producing seams that harden
the live release/deploy boundary, runtime boundary, accounting correctness, and
request boundary before adding larger platform systems.

This note is aligned to the canonical `spec-socialpredict-tasks` design-plan
refresh from [`pwdel-auto/spec-socialpredict-tasks` PR #30](https://github.com/pwdel-auto/spec-socialpredict-tasks/pull/30), commit
`cfc3849c8405ced945f1a11b217572dd89c3f868`, merged as `3ef031c`. That refresh
makes SocialPredict `v3.0.1` the current design baseline and treats
release-to-readiness feedback as an architecture boundary.

## Current Baseline

The backend is no longer starting from zero on production operations:

- SocialPredict `v3.0.0` landed major runtime, security, domain, repository,
  OpenAPI, and verification hardening.
- SocialPredict `v3.0.1` corrected production deployment/runtime defaults,
  including Traefik `v3.6.1` and packaged local Docker Postgres using
  `DB_REQUIRE_TLS=false`.
- `/health` reports process liveness.
- `/readyz` reports traffic readiness and database availability.
- `/ops/status` reports a small JSON operator status body with liveness,
  readiness, request-failure counting, and DB pool counters.
- staging and production workflows dispatch downstream Ansible deploys and then
  verify public readiness externally from GitHub Actions.
- packaged production compose defines a startup-writer service and a separate
  request-serving backend service.
- outside the codebase, DigitalOcean host-level disk, memory, and CPU visibility
  is available as an operational fallback for the current single-host topology.

That baseline means the next move should not be Kubernetes, Prometheus/Grafana,
OAuth, API keys, broad load testing, Redis, caching, or background jobs by
default. Those may become appropriate later, but they should not displace the
remaining baseline hardening work.

## Code-Grounded Evidence

The assumptions above are grounded in the current repo as follows:

- [server.go](/workspace/socialpredict/backend/server/server.go) registers
  `/health`, `/readyz`, and `/ops/status` as unversioned infrastructure routes.
- [operational_status.go](/workspace/socialpredict/backend/internal/app/runtime/operational_status.go)
  owns the process-local request-failure counter used by `/ops/status`.
- [openapi.yaml](/workspace/socialpredict/backend/docs/openapi.yaml) documents
  `/ops/status` and the `OperationalStatus` response shape.
- [default.conf.template](/workspace/socialpredict/data/nginx/vhosts/prod/default.conf.template)
  publishes `/ops/status` through the production nginx proxy alongside
  `/health` and `/readyz`.
- [deploy-to-staging.yml](/workspace/socialpredict/.github/workflows/deploy-to-staging.yml)
  and [deploy-to-production.yml](/workspace/socialpredict/.github/workflows/deploy-to-production.yml)
  dispatch the downstream Ansible deploy and then run the public readiness
  verification job.
- [verify-public-readiness/action.yml](/workspace/socialpredict/.github/actions/verify-public-readiness/action.yml)
  polls the public `/health` and `/readyz` endpoints and requires exact `live`
  and `ready` bodies.
- [docker-compose-prod.yaml](/workspace/socialpredict/scripts/docker-compose-prod.yaml)
  has `backend-startup-writer` with `STARTUP_WRITER=true` and the request-serving
  `backend` with `STARTUP_WRITER=false`.
- [install.sh](/workspace/socialpredict/scripts/install.sh) writes
  `DB_REQUIRE_TLS=false` for the packaged local Docker Postgres production
  topology.
- [db.go](/workspace/socialpredict/backend/internal/app/runtime/db.go) owns DB
  config, TLS/SSL posture, pool/lifecycle settings, readiness checks, and legacy
  `SetDB`/`GetDB` compatibility access.
- [auth.go](/workspace/socialpredict/backend/internal/service/auth/auth.go) and
  [auth_service.go](/workspace/socialpredict/backend/internal/service/auth/auth_service.go)
  still expose HTTP request-shaped auth APIs and keep a JWT signing-key fallback
  path.
- [resolvemarket.go](/workspace/socialpredict/backend/handlers/markets/resolvemarket.go)
  still parses bearer JWTs directly, reads `JWT_SIGNING_KEY` in the handler
  path, and emits raw `http.Error` failures.
- The remaining non-test `http.Error` calls under
  [handlers/markets](/workspace/socialpredict/backend/handlers/markets) are
  concentrated in legacy market handlers, while newer market helper code already
  uses shared failure writers.
- [handlers/cms/homepage](/workspace/socialpredict/backend/handlers/cms/homepage)
  still contains homepage service and GORM persistence code, and
  [seed.go](/workspace/socialpredict/backend/seed/seed.go) imports that handler
  package.
- [uniqueness.go](/workspace/socialpredict/backend/internal/domain/users/uniqueness.go)
  and [repository.go](/workspace/socialpredict/backend/internal/domain/analytics/repository.go)
  are examples of remaining misplaced persistence/model seams called out by the
  refreshed design plan.
- [place_transaction.go](/workspace/socialpredict/backend/internal/repository/bets/place_transaction.go)
  is the landed place-bet unit-of-work baseline, while sell-position and
  market-resolution style flows still need explicit transaction/concurrency
  policy.

DigitalOcean host metrics are not repository code. They are operator-provided
fallback visibility for the current deployment environment and should not be
treated as a substitute for app-owned status, release-to-readiness verification,
or future time-series monitoring.

## Highest-Priority Baseline Work

### 1. Formalize release-to-readiness feedback and probe semantics

This is the highest-priority baseline slice after the v3.0.1 design refresh.

The design plan now treats tagged image publication, Ansible dispatch/wait,
packaged compose topology, and public `/health` plus `/readyz` verification as a
first-class release/deployment control boundary. The next work is not merely to
make workflow output prettier. It is to document what the release pipeline is
allowed to conclude from those probes and what changes require design review.

Why this matters:

- [04-long-term-deployment-platform.md](04-long-term-deployment-platform.md)
  defers broader deployment-platform work until the current liveness/readiness,
  startup-writer, graceful-shutdown, and proxy-publishing contract is proven.
- [05-long-term-monitoring-platform.md](05-long-term-monitoring-platform.md)
  defers broader monitoring until deployment environments can consume app-owned
  signals reliably.
- The refreshed design plan explicitly says release/deploy workflows and public
  readiness verification are operational architecture, not incidental CI detail.
- The refreshed design plan asks what exact guarantees `/readyz` should make
  beyond the current readiness gate plus DB reachability.

Recommended scope:

- add an ADR or production-note update for release-to-readiness feedback policy
- document what `/health`, `/readyz`, and `/ops/status` each do and do not prove
- write verified URLs, probe bodies, and final probe results to the GitHub
  Actions job summary after deploy
- decide whether deploy verification should also fetch `/ops/status` after
  `/health` and `/readyz`
- define which release/deploy changes require design-plan review because they
  alter public readiness, startup-writer roles, DB TLS topology, or production
  verification semantics
- include only current public environments for now:
  `https://kconfs.com` and `https://brierfoxforecast.com`

Do not include:

- adding arbitrary extra public endpoints just to have more checks
- replacing external verification with host-internal checks
- moving downstream Ansible implementation ownership into this repo
- adding Kubernetes, blue-green deploys, or a monitoring platform

### 2. Govern production topology, DB TLS, and startup-writer guarantees

The refreshed design plan makes packaged production topology part of the current
architecture baseline. The next baseline work should preserve the v3.0.1 fix
without accidentally turning its local-Docker assumption into a universal
production database policy.

Why this matters:

- [04-long-term-deployment-platform.md](04-long-term-deployment-platform.md)
  names startup-writer posture and deployment policy as preconditions for larger
  platform work.
- [03-long-term-test-infrastructure.md](03-long-term-test-infrastructure.md)
  keeps source-of-truth DB checks targeted rather than platform-first.
- The refreshed design plan calls out packaged local Docker Postgres with
  `DB_REQUIRE_TLS=false` as topology-specific, while external production DBs
  must make `DB_REQUIRE_TLS` and `DB_SSLMODE` explicit operator choices.
- The refreshed design plan asks whether non-packaged or scaled-out topologies
  need deployment policy, a dedicated migration/seed job, or a DB-backed
  advisory lock for exactly-one-writer safety.

Recommended scope:

- document the implemented `internal/app/runtime/db.go` production contract for
  pool sizing, connection lifetime, SSL mode validation, readiness ping, close
  behavior, readiness drain, and graceful shutdown
- clarify that `DB_REQUIRE_TLS=false` belongs to packaged local Docker Postgres,
  not external production databases
- document the current packaged startup-failure posture: writer mode runs
  migrations/seeds, startup mutation failure is fatal, non-writer mode verifies
  migrations before serving
- decide the next non-packaged topology rule: deployment-only exactly-one-writer
  policy, dedicated migration job, or DB-backed advisory lock
- keep deployment docs and HostOps wording aligned with that boundary

Do not include:

- Terraform or Kubernetes migration
- generic database abstraction layers
- Redis/caching rollout
- broad testcontainers-first infrastructure

### 3. Extend accounting correctness beyond the landed place-bet unit of work

The place-bet transaction baseline is landed. The refreshed design plan says the
next correctness work should target the remaining high-risk accounting workflows,
not restart transaction analysis from scratch.

Why this matters:

- [06-long-term-performance-optimization.md](06-long-term-performance-optimization.md)
  says performance work should wait for correctness, DB ownership, and real
  hotspot evidence.
- [07-long-term-background-jobs.md](07-long-term-background-jobs.md) defers async
  processing until transaction boundaries, idempotency, retry ownership, and
  monitoring are clearer.
- The refreshed design plan calls out sell-position and market-resolution style
  workflows as transaction/concurrency-policy candidates.
- The refreshed design plan says Postgres-backed checks exist for startup and
  place-bet behavior, but remain opt-in and incomplete for other DB-truthful
  invariants.

Recommended scope:

- treat place-bet as the reference unit-of-work pattern
- define transaction and concurrency policy for sell-position flows
- define transaction and concurrency policy for market-resolution flows
- decide the narrowest next Postgres companion check for a DB-specific invariant
- keep SQLite tests for broad fast coverage, but do not claim they prove locking
  or transaction behavior that only Postgres can prove

Do not include:

- broad load testing
- response caching
- background queues
- generic retry middleware around accounting-sensitive writes

### 4. Clean up request-boundary and dependency-direction drift

The refreshed design plan broadens the next cleanup beyond only market route raw
errors. Market route convergence is still real, but it is one part of a wider
set of current boundary violations.

Why this matters:

- [01-long-term-security-hardening.md](01-long-term-security-hardening.md)
  identifies remaining security risk around unfinished boundary cleanup,
  especially raw market-handler failures and direct JWT parsing.
- [02-long-term-api-design.md](02-long-term-api-design.md) says the active API
  problem is route-family migration and OpenAPI parity, not HATEOAS, universal
  wrappers, or new versioning infrastructure.
- The refreshed design plan identifies concrete current seams: HTTP-shaped auth
  APIs, login/token extraction inside internal auth, JWT env fallback,
  CMS homepage service/repository placement under handlers, seed-to-handler
  coupling, misplaced GORM adapters, and mixed legacy handler error responses.

Recommended scope:

- split auth token extraction and login HTTP handling into HTTP adapters so
  internal auth accepts context plus token/claims and injected key, not
  `*http.Request` or env fallback
- migrate `backend/handlers/markets/resolvemarket.go` away from direct JWT
  parsing and raw `http.Error` failures
- draft the public `reason` vocabulary and route-family migration matrix before
  broad route convergence
- move homepage content service/rendering policy to an internal content/domain
  boundary, move GORM persistence to an internal repository adapter, and leave
  handlers with HTTP mapping
- inventory remaining raw GORM and misplaced adapter seams, especially
  `internal/domain/users/uniqueness.go`, `internal/domain/analytics/repository.go`,
  and global DB compatibility access
- keep `backend/docs/openapi.yaml` aligned with changed response behavior

Do not include:

- MFA
- RBAC
- refresh tokens
- API keys
- broad `/v1` planning
- universal response-wrapper middleware
- a new top-level security platform tree

### 5. Stabilize `/ops/status` as a runtime signal, not a monitoring platform

`/ops/status` remains important, but the refreshed design plan makes it a
supporting runtime signal within release-to-readiness and observability
ownership. It should be stabilized without turning it into Prometheus or Grafana
by accident.

Why this matters:

- [05-long-term-monitoring-platform.md](05-long-term-monitoring-platform.md)
  explicitly defers Prometheus, Grafana, Alertmanager, centralized logs, OTel,
  and SLO programs until the backend has a clearer app-owned signal contract.
- [08-early-startup-operational-status.md](08-early-startup-operational-status.md)
  explains that early-startup HTTP status remains deferred because opening HTTP
  before startup completion is a runtime safety change.
- The refreshed design plan treats runtime logging, health/readiness signals,
  metrics, and tracing as separate operational concerns with separate owners and
  sequencing.

Recommended scope:

- keep `/health` and `/readyz` small and stable
- keep `/ops/status` JSON, cache-disabled, and safe for public proxy exposure
- document the exact response shape and which fields are process-local
- test that no secrets, env values, usernames, tokens, or sensitive config leak
- verify staging and production expose `/ops/status` through the proxy
- decide whether `/ops/status` belongs in deploy verification, a separate smoke
  check, or operator-only troubleshooting docs

Do not include:

- early-startup HTTP listener redesign
- Prometheus format
- dashboards
- alert routing
- distributed/fleet-wide aggregation

Those remain separate future decisions.

## Deferred For Now

### Full monitoring stack

Prometheus and Grafana are reasonable future candidates, but not the next
baseline step unless current operations prove that DigitalOcean host metrics,
GitHub deploy verification, `/health`, `/readyz`, and `/ops/status` are
insufficient.

Entry criteria for reactivating this:

- operators need history, not only current status
- request failures, DB pool pressure, latency, or container restarts need to be
  correlated across time
- DigitalOcean host-level CPU, memory, and disk views are not enough
- there is a clear decision about whether Grafana is private, authenticated,
  or reachable only through SSH/VPN/internal network access

Likely future shape:

- backend exports a narrow Prometheus-compatible `/metrics` endpoint
- Prometheus scrapes backend and maybe node/container exporters
- Grafana reads Prometheus
- Traefik keeps Grafana private or protects it with explicit authentication

This should stay behind release-to-readiness, DB topology, accounting
correctness, and `/ops/status` stabilization work.

### Deployment platform migration

Terraform, Kubernetes, Helm, autoscaling, service mesh, and blue-green/canary
rollouts stay deferred under [04-long-term-deployment-platform.md](04-long-term-deployment-platform.md).
They should require operational evidence that the current Docker Compose,
Traefik, GitHub Actions, Ansible, and DigitalOcean topology has become the
limiting factor.

### API automation authentication

API keys, OAuth, device flow, and service accounts stay deferred under
[08-long-term-api-automation-auth.md](08-long-term-api-automation-auth.md).
They should require a real non-browser use case such as scripts, bots,
classroom imports, partner integrations, or CLI workflows.

### Background jobs

Queues, schedulers, retries, and dead-letter handling stay deferred under
[07-long-term-background-jobs.md](07-long-term-background-jobs.md). They should
require a concrete async use case and stronger idempotency, transaction, and
monitoring ownership first.

### Broad performance platform

Load testing, `pprof`, response caching, CDN strategy, and query-plan programs
stay deferred under [06-long-term-performance-optimization.md](06-long-term-performance-optimization.md).
They should follow evidence from `/ops/status`, real traffic, DB pool pressure,
latency, or specific user-visible slowness.

## Decision Framework

Use this order when pulling future work into the active plan:

1. Does this preserve or improve release-to-readiness confidence for staging or
   production?
2. Does this clarify a current runtime, deployment, DB, request-boundary, or API
   contract seam rather than create a second architecture?
3. Does this protect accounting correctness or avoid replica/deployment drift?
4. Does this produce operational evidence that helps the next decision?
5. Is the problem present in current staging or production behavior?
6. Is there a customer, operator, security, or integration need that makes the
   work concrete?

If the answer is no, keep the item in `FUTURE`.

## Design Plan Alignment

The read-only `spec-socialpredict-tasks` reference was checked from
[`pwdel-auto/spec-socialpredict-tasks` PR #30](https://github.com/pwdel-auto/spec-socialpredict-tasks/pull/30), which refreshes
`lib/design/design-plan.json` against SocialPredict `v3.0.0` and `v3.0.1`.
That PR is the relevant canonical design-plan update for this triage.

Important source note: use the local `/Users/patrick/Projects/spec-socialpredict-tasks-auto`
checkout for this design-plan reference. It tracks
`git@github.com:pwdel-auto/spec-socialpredict-tasks.git` and is current at
`3ef031c` with tag `v3.3.0`. The similarly named
`/Users/patrick/Projects/spec-socialpredict-tasks` checkout tracks the non-auto
repo and may not contain the same canonical `lib/design/design-plan.json`
content.

The refreshed design plan says:

- SocialPredict `v3.0.1` is the current design baseline.
- Release-triggered deploy workflows and public readiness verification are
  operational architecture, not incidental CI detail.
- Packaged local Docker Postgres uses `DB_REQUIRE_TLS=false`, while external
  production databases need explicit `DB_REQUIRE_TLS` and `DB_SSLMODE` review.
- Runtime/bootstrap owns DB pool, SSL, lifecycle, readiness, shutdown, and
  deployment-sensitive security posture.
- Release/deployment control owns tagged-image publication, Ansible
  dispatch/wait behavior, packaged compose topology, HostOps documentation, and
  public readiness verification.
- Remaining work should focus on auth transport leakage, CMS content ownership,
  misplaced persistence adapters, legacy route error shapes, setup/runtime
  compatibility bridges, sell/resolve transaction policy, release-to-readiness
  feedback, and opt-in Postgres verification coverage.

This triage follows that posture:

- release-to-readiness feedback is the top baseline governance slice
- `/ops/status` stabilization is runtime-signal work within that boundary, not a
  monitoring-platform rollout
- DB TLS and startup-writer guarantees are production topology policy
- sell/resolve transaction work is accounting correctness work
- auth/CMS/market error cleanup is request-boundary and dependency-direction
  alignment
- Grafana, Prometheus, Kubernetes, Terraform, OAuth, Redis, caching, and queues
  remain deferred until smaller seams prove the need
