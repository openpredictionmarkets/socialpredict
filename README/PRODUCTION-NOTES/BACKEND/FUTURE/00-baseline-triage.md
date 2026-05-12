---
title: Future Backend Baseline Triage
document_type: production-notes
domain: backend
author: Patrick Delaney
updated_at: 2026-05-11T00:00:00Z
updated_at_display: "Monday, May 11, 2026"
update_reason: "Add a baseline triage document to rank future backend work after release and deployment hardening."
status: draft
---

# Future Backend Baseline Triage

## Purpose

This note is a triage layer over the backend `FUTURE` notes. It does not activate
all deferred work. It exists to answer: if there is not yet direct customer input,
what baseline production work is most important next, and why?

The decision rule is simple: prefer small, evidence-producing seams that harden
the live deployment and request boundary before adding larger platform systems.
That matches the active design posture: preserve behavior, clarify runtime
boundaries, keep deployment behavior explicit, and align API/auth behavior only
after the underlying seams are stable.

## Current Baseline

The backend is no longer starting from zero on production operations:

- `/health` reports process liveness.
- `/readyz` reports traffic readiness and database availability.
- `/ops/status` reports a small JSON operator status body with liveness,
  readiness, request-failure counting, and DB pool counters.
- staging and production workflows dispatch downstream Ansible deploys and then
  verify public readiness externally from GitHub Actions.
- outside the codebase, DigitalOcean host-level disk, memory, and CPU visibility
  is available as an operational fallback for the current single-host topology.

That baseline means the next move should not be Kubernetes, Prometheus/Grafana,
OAuth, API keys, broad load testing, or background jobs by default. Those may
become appropriate later, but they should not displace the remaining baseline
hardening work.

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
- [resolvemarket.go](/workspace/socialpredict/backend/handlers/markets/resolvemarket.go)
  still parses bearer JWTs directly, reads `JWT_SIGNING_KEY` in the handler
  path, and emits raw `http.Error` failures.
- The remaining non-test `http.Error` calls under
  [handlers/markets](/workspace/socialpredict/backend/handlers/markets) are
  concentrated in legacy market handlers, while newer market helper code already
  uses shared failure writers.

DigitalOcean host metrics are not repository code. They are operator-provided
fallback visibility for the current deployment environment and should not be
treated as a substitute for app-owned status or future time-series monitoring.

## Highest-Priority Baseline Work

### 1. Stabilize and operationalize `/ops/status`

This is the most important next baseline slice.

The active monitoring note already treats `/ops/status` as the first
operator-facing JSON status surface. The next step is not to invent a new
monitoring platform; it is to make this contract boring, documented, tested,
and useful enough that operators and GitHub workflows can trust it.

Why this matters:

- [05-long-term-monitoring-platform.md](05-long-term-monitoring-platform.md)
  explicitly defers Prometheus, Grafana, Alertmanager, centralized logs, OTel,
  and SLO programs until the backend has a clearer app-owned signal contract.
- [08-early-startup-operational-status.md](08-early-startup-operational-status.md)
  explains that early-startup HTTP status remains deferred because opening HTTP
  before startup completion is a runtime safety change.
- [04-long-term-deployment-platform.md](04-long-term-deployment-platform.md)
  says broader deployment-platform work should wait until liveness/readiness
  policy, startup-writer posture, graceful shutdown, and proxy publishing have
  been proven.
- [06-long-term-performance-optimization.md](06-long-term-performance-optimization.md)
  points to DB pool and operational signals as prerequisites before broader
  performance programs.

Recommended scope:

- keep `/health` and `/readyz` small and stable
- keep `/ops/status` JSON, cache-disabled, and safe for public proxy exposure
- document the exact response shape and which fields are process-local
- test that no secrets, env values, usernames, tokens, or sensitive config leak
- verify staging and production expose `/ops/status` through the proxy
- decide whether GitHub deploy verification should also fetch `/ops/status`
  after `/health` and `/readyz`

Do not include:

- early-startup HTTP listener redesign
- Prometheus format
- dashboards
- alert routing
- distributed/fleet-wide aggregation

Those remain separate future decisions.

### 2. Improve deploy evidence visibility, not deploy verification semantics

The external GitHub deploy verification is baseline-stable: the workflow runs
outside the host, waits for the downstream Ansible workflow, then checks the
public host readiness endpoints. That is the right boundary: Ansible says "I
ran the deploy," and the SocialPredict workflow says "the public service became
reachable and ready."

The remaining improvement is presentation, not core stability.

Why this matters:

- [04-long-term-deployment-platform.md](04-long-term-deployment-platform.md)
  defers larger platform migration until current deployment policy consumes
  liveness/readiness intentionally.
- [05-long-term-monitoring-platform.md](05-long-term-monitoring-platform.md)
  defers broader monitoring until deployment environments can consume app-owned
  signals reliably.

Recommended scope:

- write the verified URLs and final probe results to the GitHub Actions job
  summary after deploy
- include only the current public environments for now:
  `https://kconfs.com` and `https://brierfoxforecast.com`
- optionally list the public environment roots near the top of `README.md`
- avoid treating README links as health evidence; the workflow summary is the
  audit trail for a particular deploy

Do not include:

- adding extra public endpoints just to have more checks
- moving Ansible success/failure ownership into the app repo
- replacing external verification with host-internal checks

### 3. Clean up market route request-boundary drift

This is the most important code-cleanup slice after the operational signal work.

The problem is not that the backend has no auth or error system. It does. The
problem is that the remaining legacy market route family still has raw
plain-text failure responses and one direct JWT parsing path.

Why this matters:

- [01-long-term-security-hardening.md](01-long-term-security-hardening.md)
  identifies the remaining security risk as deployment assumptions and
  unfinished boundary cleanup, especially raw market-handler failures and direct
  JWT parsing in `resolvemarket.go`.
- [02-long-term-api-design.md](02-long-term-api-design.md) says the active API
  problem is route-family migration and OpenAPI parity, not HATEOAS, universal
  wrappers, or new versioning infrastructure.
- The active security note states that `resolvemarket.go` is the next precise
  auth-boundary migration seam because it still parses bearer tokens directly,
  reads `JWT_SIGNING_KEY` in the handler path, and emits raw `http.Error`
  failures.

Recommended scope:

- migrate `backend/handlers/markets/resolvemarket.go` to the centralized auth
  service path
- remove handler-local direct `JWT_SIGNING_KEY` reads from resolve-market auth
- translate resolve-market failures through the shared failure-envelope path
- then move through remaining market read-route raw-error slices before
  write-sensitive routes
- keep `backend/docs/openapi.yaml` aligned with changed response behavior

Do not include:

- MFA
- RBAC
- refresh tokens
- API keys
- broad `/v1` planning
- universal response-wrapper middleware

Those are explicitly deferred until a concrete product, integration, or threat
model requires them.

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

This should stay behind the current `/ops/status` stabilization work.

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

1. Does this reduce the chance of a deploy looking successful while the public
   service is broken?
2. Does this clarify an existing runtime, request-boundary, or API contract
   seam rather than create a second architecture?
3. Does this produce operational evidence that helps the next decision?
4. Is the problem present in current staging or production behavior?
5. Is there a customer, operator, security, or integration need that makes the
   work concrete?

If the answer is no, keep the item in `FUTURE`.

## Design Plan Alignment

The read-only design-plan reference emphasizes runtime boundary cleanup,
configuration/service ownership, legacy-model decoupling, and API/auth contract
alignment in small reversible steps.

This triage follows that posture:

- `/ops/status` stabilization is runtime and operational-boundary work
- deploy evidence visibility is external verification of the runtime contract
- market route cleanup is API/auth boundary alignment
- Grafana, Prometheus, Kubernetes, Terraform, OAuth, and queues remain deferred
  until smaller seams prove the need

The design plan does not need a full rewrite before this triage document. If
the next branch starts implementing `/ops/status` changes or market route auth
cleanup, the active design plan should receive a short update naming that
specific next slice.
