---
title: Deployment Infrastructure
document_type: production-notes
domain: backend
author: Patrick Delaney
updated_at: 2026-05-02T16:52:00Z
updated_at_display: "Saturday, May 02, 2026 at 04:52 PM UTC"
update_reason: "Close WAVE08 with a deployment stop-and-review inventory for current health, writer-role, docs-publishing, and graceful-shutdown behavior."
status: active
---

# Deployment Infrastructure

## Update Summary

This note was updated on Monday, April 27, 2026 to replace an older Kubernetes-heavy deployment plan with guidance that matches the live SocialPredict deployment topology and the current design-plan priority on runtime safety first.

On Thursday, April 30, 2026, the first deployment-facing health problem was finished for the backend serving path: `/health` now reports liveness, `/readyz` checks readiness and database availability, and Docker black-box checks confirmed both endpoints on `http://localhost:8080`. On Saturday, May 02, 2026, the backend image and production compose topology were wired to that contract: the image-level Docker `HEALTHCHECK` consumes `/health` as a process liveness probe, while production compose overrides backend service healthchecks to consume `/readyz` before starting dependents. The nginx production template also publishes backend-owned infra probes, Swagger UI, and the OpenAPI document explicitly at `/health`, `/readyz`, `/swagger/`, and `/openapi.yaml`. As of upstream `main` at `051aac6b2fefa5634b8c98cc38caf52acf0043a9`, startup mutation mode is explicit: the `backend-startup-writer` compose service runs the same backend image with `STARTUP_WRITER=true` for migrations and startup-owned seeds, while the request-serving `backend` service sets `STARTUP_WRITER=false` and verifies applied migrations before serving. The backend now closes readiness, waits `BACKEND_READINESS_DRAIN_SECONDS`, and then lets HTTP shutdown drain in-flight requests for `BACKEND_SHUTDOWN_TIMEOUT_SECONDS`.

| Topic | Prior to April 27, 2026 | After April 27, 2026 |
| --- | --- | --- |
| Core framing | Treated deployment infrastructure as a large new platform buildout | Treats deployment infrastructure as hardening the runtime and publish path the repo already uses |
| Current-state accuracy | Claimed automated deployment and deployment structure were mostly absent | Recognizes the live Docker image build, production compose topology, nginx and Traefik edge, and staging or production dispatch workflows |
| Main proposal | Build Kubernetes, Helm, ingress manifests, autoscaling, and broad infra scaffolding first | Focus on health semantics, graceful startup and shutdown, startup-writer safety, proxy publishing, and runtime env or secret ownership first |
| Architecture posture | Assumed a new deployment stack should define the architecture | Keeps deployment hardening tied to the existing backend binary, compose stack, nginx proxy, and repo workflows |
| Docs publishing | Did not reflect the current Swagger and OpenAPI publishing caveat behind `/api/` proxying | Keeps docs publishing explicit as part of deployment ownership |
| HA posture | Optimized for orchestration features first | Optimizes for fail-closed startup, truthful readiness, and safer replica behavior first |
| Future ideas | Mixed Kubernetes, Helm, Terraform, autoscaling, and broader platform ideas into the active note | Defers larger platform ideas to [FUTURE/04-long-term-deployment-platform.md](/workspace/socialpredict/README/PRODUCTION-NOTES/BACKEND/FUTURE/04-long-term-deployment-platform.md) |

## Executive Direction

SocialPredict should treat deployment infrastructure as runtime and publish-path hardening of the existing backend, not as a greenfield platform migration.

The active direction is:

1. Keep the existing deployment truth explicit:
   - backend Docker image
   - production compose stack
   - nginx proxy
   - Traefik entrypoint
   - GitHub build and deploy workflows
2. Treat health, readiness, startup ownership, and graceful shutdown as the main deployment problems to solve first.
3. Publish backend docs and infra routes intentionally through the proxy layer rather than assuming `/api/` prefixing covers them.
4. Treat replica safety as a deployment and runtime contract issue first, because every replica still bootstraps DB and runtime state while only an explicitly selected startup writer may mutate shared startup state.
5. Keep runtime env and secret ownership explicit rather than spreading deployment-sensitive behavior into ad hoc application helpers.
6. Defer Kubernetes, Helm, Terraform, autoscaling, and larger infra-as-code ambitions until the live runtime contract is materially stronger.

For a high-availability and fault-tolerant backend, deployment infrastructure should prefer:

- fail-closed startup over warning-and-continue boot behavior
- truthful readiness over placeholder health endpoints
- explicit proxy publishing over hidden root-path assumptions
- minimal shared-state startup mutation per replica
- deployment contracts that match the live binary and topology

This note explicitly rejects treating Kubernetes manifests or a platform rewrite as the main immediate move for the active slice.

## Why This Matters

The current backend already has a deployable topology, but it does not yet have a trustworthy runtime contract for orchestration.

That matters because:

- [main.go](/workspace/socialpredict/backend/main.go) still performs env load, DB init, readiness wait, config load, security config load, startup-mode selection, startup mutation or verification, and server startup in one process path
- [server.go](/workspace/socialpredict/backend/server/server.go) now exposes `/health` and `/readyz`, and deployment healthcheck policy consumes both endpoints intentionally
- the production compose stack in [docker-compose-prod.yaml](/workspace/socialpredict/scripts/docker-compose-prod.yaml) has one explicit `backend-startup-writer` service and one explicit non-writer `backend` service, and health-gates frontend and nginx startup on request-serving backend `/readyz`
- the backend Docker image in [Dockerfile](/workspace/socialpredict/docker/backend/Dockerfile) declares a liveness `HEALTHCHECK` against `/health`
- the nginx production template in [default.conf.template](/workspace/socialpredict/data/nginx/vhosts/prod/default.conf.template) now explicitly proxies `/health`, `/readyz`, `/swagger/`, and `/openapi.yaml` to the backend before `/api/` and `/`

So the active deployment problem is not “invent a new cluster platform.” The active problem is to make the current deployment topology safer and more truthful.

## Current Code Snapshot

### The live deployment topology already exists

The repo already has:

- backend image build in [docker.yml](/workspace/socialpredict/.github/workflows/docker.yml)
- production compose wiring in [docker-compose-prod.yaml](/workspace/socialpredict/scripts/docker-compose-prod.yaml)
- nginx proxying in [default.conf.template](/workspace/socialpredict/data/nginx/vhosts/prod/default.conf.template)
- Traefik edge wiring in [docker-compose-prod.yaml](/workspace/socialpredict/scripts/docker-compose-prod.yaml)
- staging dispatch in [deploy-to-staging.yml](/workspace/socialpredict/.github/workflows/deploy-to-staging.yml)
- production dispatch in [deploy-to-production.yml](/workspace/socialpredict/.github/workflows/deploy-to-production.yml)

This means the active note should not pretend deployment is a blank page.

### Startup mutation mode is explicit in production compose

The current process startup in [main.go](/workspace/socialpredict/backend/main.go) still does all of the following in every backend process:

- load local env overrides
- load DB config
- open the DB
- wait for DB readiness
- load config service
- load security config
- load startup mutation mode
- start serving

Shared startup DB writes are gated explicitly through [startup_mutation.go](/workspace/socialpredict/backend/startup_mutation.go) and [runtime/startup_mutation.go](/workspace/socialpredict/backend/internal/app/runtime/startup_mutation.go):

- `STARTUP_WRITER=true` runs migrations plus user and homepage seeds
- non-writer replicas call `migration.VerifyApplied` before serving

Production compose now makes that operational contract concrete without adding a new binary, advisory lock, leader election path, or platform control plane:

- `backend-startup-writer` uses `BACKEND_IMAGE_NAME`, sets `STARTUP_WRITER=true`, and is the only production compose service allowed to run startup-owned migrations and seeds
- `backend` uses the same `BACKEND_IMAGE_NAME`, has no fixed `container_name`, sets `STARTUP_WRITER=false`, and is the request-serving backend target for frontend and nginx
- the non-writer `backend` waits for `backend-startup-writer` to become ready, then verifies applied migrations before opening its own readiness gate
- frontend and nginx depend on the non-writer `backend` health state, so public request serving stays attached to the explicit non-writer path

Operators should preserve exactly one `STARTUP_WRITER=true` runtime path in this topology. Additional request-serving backend replicas must inherit the non-writer posture and set `STARTUP_WRITER=false`; they should not be introduced by scaling the startup-writer service.

### Health semantics now have a serving-path baseline

The current infra route registration in [server.go](/workspace/socialpredict/backend/server/server.go) exposes:

- `GET /health`
- `GET /readyz`

As of April 30, 2026:

- `/health` returns plain-text `live` for liveness
- `/readyz` returns `ready` only after the readiness gate is open and database availability passes
- `/readyz` returns `not ready` with `503` when the readiness gate is closed or the database check fails

That problem is finished for the backend serving path. Deployment infrastructure now consumes those endpoints explicitly in the backend image, production compose health policy, and nginx proxy publishing.

### The backend image is simple, production-usable, and liveness-aware

The current backend Dockerfile in [Dockerfile](/workspace/socialpredict/docker/backend/Dockerfile):

- builds a static backend binary
- runs as a non-root user
- exposes port `8080`

It now declares an image-level Docker `HEALTHCHECK` against
`http://127.0.0.1:${BACKEND_PORT:-8080}/health` and requires the exact `live`
body returned by the backend liveness contract. That image check answers only
"is the backend process serving HTTP" and deliberately does not require
database readiness.

Production compose owns the traffic-readiness policy by overriding backend
service healthchecks to call
`http://127.0.0.1:${BACKEND_PORT:-8080}/readyz` and require the exact `ready`
body. The startup-writer service must become ready before the explicit
non-writer backend starts, and frontend and nginx startup are gated on the
non-writer backend service health state. This keeps liveness, readiness, and
startup mutation ownership separate while still using the single backend binary.

For migrations and startup-owned seeds, the production rule is intentionally
simple: the writer owns the mutation attempt, and non-writers only verify that
registered migrations are already applied. If the writer cannot complete
migrations, user seed writes, or homepage seed writes, readiness remains closed
and request-serving replicas do not start from this compose file.

On termination, the backend closes the readiness gate before calling HTTP server
shutdown. The readiness-drain window is `BACKEND_READINESS_DRAIN_SECONDS` and
defaults to 5 seconds. The HTTP shutdown timeout is
`BACKEND_SHUTDOWN_TIMEOUT_SECONDS` and defaults to 10 seconds. Production
compose sets both values explicitly for the startup writer and request-serving
backend, and gives Docker a 20 second `stop_grace_period` so the backend-owned
drain and shutdown windows can complete before the container receives a forced
kill.

### The proxy topology is real, and docs publishing is part of it

The current production nginx template in [default.conf.template](/workspace/socialpredict/data/nginx/vhosts/prod/default.conf.template) proxies:

- `/health` to the backend
- `/readyz` to the backend
- `/openapi.yaml` to the backend
- `/swagger` and `/swagger/` to the backend
- `/api/` to the backend
- `/` to the frontend

But the backend itself serves:

- `/health`
- `/readyz`
- `/openapi.yaml`
- `/swagger/`

So deployment ownership now makes those backend-root docs paths visible
intentionally, rather than assuming the `/api/` proxy prefix will make them
work. Access remains open in this slice; any future restriction should be
applied at the proxy or host layer while keeping backend-served docs as the
single contract surface.

## What Deployment Infrastructure Should Own

This note should own:

- the current Docker and proxy topology
- runtime env and secret ownership expectations
- health, readiness, and liveness direction
- startup-writer and migration-safety deployment posture
- graceful startup and shutdown expectations
- explicit docs and infra route publishing through the proxy layer

## What This Note Should Not Own

This note should not become the home for every future platform ambition.

It should explicitly defer:

- Kubernetes rollout
- Helm charts
- Terraform
- service mesh
- advanced autoscaling programs
- multi-cluster deployment patterns
- large infra-as-code redesigns

Those topics now belong in [FUTURE/04-long-term-deployment-platform.md](/workspace/socialpredict/README/PRODUCTION-NOTES/BACKEND/FUTURE/04-long-term-deployment-platform.md), not in the active production note.

## WAVE08 Stop and Review

The active deployment slice is now at a stop-and-review point. The touched
runtime and deployment paths agree on the current contract:

- health is split into `/health` liveness and `/readyz` readiness
- the backend image checks `/health`, while production compose gates startup on
  `/readyz`
- exactly one production compose service, `backend-startup-writer`, runs with
  `STARTUP_WRITER=true`
- request-serving backend containers run with `STARTUP_WRITER=false` and remain
  the only backend target for frontend and nginx traffic
- nginx publishes `/health`, `/readyz`, `/openapi.yaml`, `/swagger`, and
  `/swagger/` explicitly at the public host root
- staging and production dispatch workflows pass the startup-writer and serving
  backend service names to the downstream deploy repository
- backend shutdown closes readiness first, waits the configured readiness-drain
  window, and then lets HTTP shutdown drain for the configured timeout
- production compose sets the drain and shutdown windows explicitly and gives
  Docker a longer `stop_grace_period`

The remaining deployment gap is intentionally narrow: the downstream deployment
runner and proxy rollout procedure still need to prove that an update applies
the nginx template and compose service split together, then verifies the public
host returns the expected root-path probe and docs responses after the serving
backend is healthy.

The next seam should be one concrete rollout check in the existing
compose/nginx/Traefik topology: after dispatch, run a host-level verification
that confirms `/health`, `/readyz`, `/openapi.yaml`, `/swagger`, and `/swagger/`
are routed through the production proxy to the non-writer `backend` service, and
that only `backend-startup-writer` runs with `STARTUP_WRITER=true`.

Larger deployment-platform ambitions remain deferred. Do not use this stop point
to reopen Kubernetes, Helm, Terraform, service mesh, autoscaling, multi-cluster,
or broad infra-as-code work in the active deployment note.
