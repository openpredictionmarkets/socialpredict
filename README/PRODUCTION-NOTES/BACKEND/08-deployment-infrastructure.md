---
title: Deployment Infrastructure
document_type: production-notes
domain: backend
author: Patrick Delaney
updated_at: 2026-05-01T11:50:49Z
updated_at_display: "Friday, May 01, 2026 at 11:50 AM UTC"
update_reason: "Refresh deployment guidance against upstream main at 051aac6 and record landed startup-writer behavior."
status: active
---

# Deployment Infrastructure

## Update Summary

This note was updated on Monday, April 27, 2026 to replace an older Kubernetes-heavy deployment plan with guidance that matches the live SocialPredict deployment topology and the current design-plan priority on runtime safety first.

On Thursday, April 30, 2026, the first deployment-facing health problem was finished for the backend serving path: `/health` now reports liveness, `/readyz` checks readiness and database availability, and Docker black-box checks confirmed both endpoints on `http://localhost:8080`. The nginx proxy templates also now publish backend-owned docs explicitly at `/swagger/` and `/openapi.yaml`. As of upstream `main` at `051aac6b2fefa5634b8c98cc38caf52acf0043a9`, startup mutation mode is also explicit: `STARTUP_WRITER=true` replicas run migrations and seeds, while non-writers verify applied migrations before serving. Deployment work still needs image healthcheck policy, operational enforcement of exactly one startup writer, and graceful rollout/shutdown discipline.

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
- [server.go](/workspace/socialpredict/backend/server/server.go) now exposes `/health` and `/readyz`, but deployment healthcheck policy is not yet wired into the image or compose topology
- the production compose stack in [docker-compose-prod.yaml](/workspace/socialpredict/scripts/docker-compose-prod.yaml) uses `depends_on`, not real health-gated orchestration
- the backend Docker image in [Dockerfile](/workspace/socialpredict/docker/backend/Dockerfile) has no `HEALTHCHECK`
- the nginx production template in [default.conf.template](/workspace/socialpredict/data/nginx/vhosts/prod/default.conf.template) now explicitly proxies `/swagger/` and `/openapi.yaml` to the backend before `/api/` and `/`

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

### Startup mutation mode exists, but deployment still has to enforce it

The current process startup in [main.go](/workspace/socialpredict/backend/main.go) still does all of the following in every replica:

- load local env overrides
- load DB config
- open the DB
- wait for DB readiness
- load config service
- load security config
- load startup mutation mode
- start serving

Shared startup DB writes are now gated explicitly through [startup_mutation.go](/workspace/socialpredict/backend/startup_mutation.go) and [runtime/startup_mutation.go](/workspace/socialpredict/backend/internal/app/runtime/startup_mutation.go):

- `STARTUP_WRITER=true` runs migrations plus user and homepage seeds
- non-writer replicas call `migration.VerifyApplied` before serving

That is a meaningful improvement over the older "every replica mutates startup state" posture, but deployment still has to guarantee that exactly one replica is allowed to act as the writer.

### Health semantics now have a serving-path baseline

The current infra route registration in [server.go](/workspace/socialpredict/backend/server/server.go) exposes:

- `GET /health`
- `GET /readyz`

As of April 30, 2026:

- `/health` returns plain-text `live` for liveness
- `/readyz` returns `ready` only after the readiness gate is open and database availability passes
- `/readyz` returns `not ready` with `503` when the readiness gate is closed or the database check fails

That problem is finished for the backend serving path. Deployment infrastructure still needs to decide how Docker, compose, nginx, Traefik, and future orchestrators should consume those endpoints.

### The backend image is simple and production-usable, but not health-aware yet

The current backend Dockerfile in [Dockerfile](/workspace/socialpredict/docker/backend/Dockerfile):

- builds a static backend binary
- runs as a non-root user
- exposes port `8080`

But it does not yet declare:

- a `HEALTHCHECK`
- explicit startup or shutdown expectations
- readiness semantics

That is the right scale of current deployment work for this note.

### The proxy topology is real, and docs publishing is part of it

The current production nginx template in [default.conf.template](/workspace/socialpredict/data/nginx/vhosts/prod/default.conf.template) proxies:

- `/openapi.yaml` to the backend
- `/swagger` and `/swagger/` to the backend
- `/api/` to the backend
- `/` to the frontend

But the backend itself serves:

- `/health`
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

## Near-Term Sequencing

The design-plan-aligned deployment direction is:

1. Keep the current deployment topology truthful in the note.
2. Operationalize the current runtime startup contract:
   - fail closed on incompatible DB startup conditions
   - ensure exactly one replica runs with `STARTUP_WRITER=true`
   - wire liveness and readiness into image and deployment health policy
3. Add graceful shutdown and explicit health ownership to the backend runtime.
4. Add remaining explicit proxy or ingress publishing for infra routes that should remain visible beyond the docs paths.
5. Revisit broader replica or orchestration strategy only after those runtime contracts are materially stronger.

## Open Questions

- If docs access is restricted later, should the restriction be host-based, network-based, or proxy-auth-based
- Which startup actions remain inside the backend process and which later move to a separate startup-writer path
- Whether readiness should fail on any migration incompatibility or only on specific hard failures
- Which deployment-sensitive settings should be normalized into runtime bootstrap rather than read ad hoc
