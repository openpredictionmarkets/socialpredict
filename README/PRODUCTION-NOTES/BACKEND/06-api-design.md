---
title: API Design
document_type: production-notes
domain: backend
author: Patrick Delaney
updated_at: 2026-04-30T11:55:00Z
updated_at_display: "Thursday, April 30, 2026 at 11:55 AM UTC"
update_reason: "Align the API note with the April 30 runtime-boundary completion for health, readiness, 405, 429, and auth-gate behavior."
status: active
---

# API Design

## Update Summary

This note was updated on Sunday, April 26, 2026 to replace an older API standardization plan with guidance that matches the live SocialPredict backend, the current OpenAPI contract, and the active design-plan posture.

On Thursday, April 30, 2026, the runtime-boundary API behavior changed in a few important ways: `/health` now returns `live`, `/readyz` returns readiness state, runtime-owned `405` and `429` failures use JSON `reason` envelopes, and the private action routes enforce `PASSWORD_CHANGE_REQUIRED` before their domain handlers run.

| Topic | Prior to April 26, 2026 | After April 26, 2026 |
| --- | --- | --- |
| Core framing | Treated API design as a new platform to build from scratch | Treats API design as route-family contract governance and migration of a live monolith |
| Current-state accuracy | Assumed there was no OpenAPI or Swagger surface yet | Recognizes the live `openapi.yaml`, embedded Swagger UI, route/spec parity tests, and existing deferred follow-up file |
| Main proposal | Build `api/standards.go`, version managers, response middleware, code generation, and broad versioning/platform features | Focus on truthful contract documentation, response-family convergence, auth semantics, route-family migration, and OpenAPI parity |
| Architecture posture | Proposed a new top-level `api/` subsystem | Extends the live `server`, `handlers`, `security`, `internal/service/auth`, and `backend/docs/openapi.yaml` seams |
| Response strategy | Assumed a universal `APIResponse` wrapper should become the standard immediately | Documents the current mixed route families honestly and shrinks that mixed state incrementally |
| Versioning strategy | Proposed header-based versioning and `/v1` rollout planning | Documents the current truth: `/v0` application routes plus unversioned infra routes |
| Docs publishing | Assumed API docs were a future generation problem | Recognizes that Swagger/OpenAPI already work at the backend directly but need explicit proxy publishing in staging/production |
| Future ideas | Mixed HATEOAS, content negotiation, version-platform work, and client-generation ambitions into the active note | Defers broader platform ideas to [FUTURE/02-long-term-api-design.md](/workspace/socialpredict/README/PRODUCTION-NOTES/BACKEND/FUTURE/02-long-term-api-design.md) |

## Executive Direction

SocialPredict should treat API design as contract governance for the backend that already exists, not as a greenfield API-platform build.

The active direction is:

1. Keep the source-of-truth order explicit:
   - route wiring in `backend/server/server.go`
   - touched handlers and DTOs
   - `backend/docs/openapi.yaml`
   - deferred follow-ups in `backend/docs/API-ISSUES.md`
2. Describe the live route and response families honestly before trying to normalize them.
3. Use the existing envelope and public `reason` direction where it already exists, but treat mixed `PlainTextErrorResponse` and legacy raw-response paths as migration state rather than pretending the whole API already converged.
4. Keep API cleanup aligned with the active design-plan sequencing:
   - runtime/failure ownership first
   - then scoped API/auth/OpenAPI alignment
5. Preserve current public route structure and `/v0` versioning posture unless a later explicit redesign says otherwise.
6. Keep auth contract behavior, especially `mustChangePassword`, explicit and route-family-aware instead of hiding it inside vague “API standards” language.
7. Make externally visible docs paths explicit in staging and production instead of assuming that a frontend proxy prefix will automatically expose backend-served Swagger and OpenAPI correctly.
8. Defer API-platform features such as HATEOAS, XML/content negotiation, version managers, generator-first workflows, or a dedicated API service layer until the real current contract inconsistencies are materially smaller.

For a high-availability, fault-tolerant backend, API design should prefer:

- deterministic route behavior across handlers, middleware, and replicas
- explicit boundary ownership of failures
- truthful OpenAPI documentation of live behavior
- route-family migration over flag-day rewrites
- contract clarity without inventing a second framework

This note explicitly rejects building a broad new `api/` subsystem as the main move for the active slice.

## Why This Matters

The active API problem in SocialPredict is not lack of theory. The active problem is a mixed but already partially-governed contract.

That means:

- some routes already use shared `{ "ok": true, "result": ... }` and `{ "ok": false, "reason": ... }` envelopes
- some routes still return raw JSON DTOs
- some middleware and legacy paths still emit plain text
- the OpenAPI document already exists and already documents part of this mixed state
- auth semantics such as `mustChangePassword` are already real and should be documented precisely rather than redesigned casually

For a high-availability and fault-tolerant backend, this matters because clients, operators, and future migrations need predictable contract behavior at the boundary. The older note blurred that need with platform ambitions. The current job is to make the live HTTP contract legible, stable, and incrementally safer to evolve.

## Current Code Snapshot

As of 2026-04-26, the backend already has an API contract surface with meaningful structure, but it is still mixed across route families.

### OpenAPI and Swagger already exist

The backend already exposes:

- [openapi.yaml](/workspace/socialpredict/backend/docs/openapi.yaml)
- [Swagger UI at `/swagger/`](/workspace/socialpredict/backend/swagger-ui/swagger-initializer.js)
- [route/spec parity tests](/workspace/socialpredict/backend/openapi_test.go)
- [server contract coverage for `/openapi.yaml` and `/swagger/`](/workspace/socialpredict/backend/server/server_contract_test.go)

That means the active backend does not need a first Swagger/OpenAPI generator project. It already has a live contract artifact and embedded docs surface.

If the backend is running on the default port, the docs surface is:

- `http://localhost:8080/swagger/`
- `http://localhost:8080/openapi.yaml`

If `BACKEND_PORT` is overridden, replace `8080` with that port.

### Staging and production docs publishing is still a deployment caveat

The backend-served docs work directly when the backend is exposed at its own root, but the current nginx proxy templates do not yet publish them explicitly in staging or production.

Today:

- the backend serves docs at root paths:
  - `/swagger/`
  - `/openapi.yaml`
- the embedded Swagger UI requests `/openapi.yaml` as an absolute root path
- the current nginx templates proxy:
  - `/api/` to the backend
  - `/` to the frontend

That means the current staging/production topology is likely to hide or break the docs surface unless nginx or ingress publishes those routes explicitly.

The active recommendation is:

- keep the backend as the source of truth for Swagger and OpenAPI
- expose docs explicitly at the reverse-proxy layer, such as:
  - `/swagger/`
  - `/openapi.yaml`
- or use a dedicated docs/admin hostname if operations prefers stricter isolation
- do not maintain a second frontend-owned Swagger copy

This is a deployment and contract-publishing issue, not a reason to create a separate API documentation platform inside the application.

### Source-of-truth order already exists

The repo already has a practical source-of-truth order in [API-ISSUES.md](/workspace/socialpredict/backend/docs/API-ISSUES.md):

1. `backend/server/server.go`
2. touched handlers and DTOs
3. `backend/docs/openapi.yaml`
4. deferred API follow-up notes

That should remain the active rule for this slice.

### The live route surface is `/v0` plus unversioned infra routes

The application routes remain under `/v0`, while infra/documentation routes are unversioned:

- `/health`
- `/openapi.yaml`
- `/swagger/*`

The backend does not currently have:

- header-based version negotiation
- `/v1` or `/v2`
- deprecation headers
- sunset management

The active note should document that truth rather than invent a versioning platform ahead of need.

### Response families are still mixed

The live backend currently has several route families:

#### Shared envelope routes

Many routes already use the shared helpers in [envelope.go](/workspace/socialpredict/backend/handlers/envelope.go):

- `ok/result` success envelopes
- `ok/reason` failure envelopes

Examples include:

- login
- many private/profile-change handlers
- bets buy/sell flows
- stats and metrics
- CMS homepage admin/public handlers

#### Bare JSON success plus `ReasonResponse` failure routes

Some route families still return raw JSON DTO success while using shared `reason` failures where touched.

That mixed state is visible in parts of the markets surface and some setup/config flows.

#### Legacy raw JSON plus plain-text failure routes

Several older public/reporting routes still return raw JSON on success and `http.Error` style plain-text failures.

Examples include:

- [publicuser.go](/workspace/socialpredict/backend/handlers/users/publicuser.go)
- [portfolio.go](/workspace/socialpredict/backend/handlers/users/publicuser/portfolio.go)
- [financial.go](/workspace/socialpredict/backend/handlers/users/financial.go)
- [usercredit.go](/workspace/socialpredict/backend/handlers/users/credit/usercredit.go)
- [searchmarkets.go](/workspace/socialpredict/backend/handlers/markets/searchmarkets.go)

#### Middleware and infra transport responses

Some failures and infra routes intentionally remain outside application envelopes today:

- `/health` returns plain text `live`
- `/readyz` returns plain text `ready` or `not ready`
- `/openapi.yaml` returns YAML
- `/swagger/*` serves static assets
- shared security middleware emits JSON `429` envelopes with stable `RATE_LIMITED` or `LOGIN_RATE_LIMITED` reasons
- router-level method handling emits JSON `405` envelopes and sets `Allow`

This mixed state is real and should be documented honestly rather than hidden behind a universal response-wrapper proposal.

### OpenAPI already documents migration state

The current OpenAPI document already carries:

- [ReasonResponse](/workspace/socialpredict/backend/docs/openapi.yaml)
- [PlainTextErrorResponse](/workspace/socialpredict/backend/docs/openapi.yaml)

That is important. It means the contract already acknowledges that some route families have converged while others are still in migration.

The active goal is not to erase that truth in documentation. The active goal is to keep the spec aligned while shrinking the legacy surface.

### Auth semantics are already specific

The live API already has important auth contract rules:

- `POST /v0/login` returns a normal bearer token plus `mustChangePassword`
- most authenticated actions enforce password-change gating
- `POST /v0/changepassword` intentionally remains usable when `mustChangePassword` is set

Grounding:

- [loggin.go](/workspace/socialpredict/backend/internal/service/auth/loggin.go)
- [auth.go](/workspace/socialpredict/backend/internal/service/auth/auth.go)
- [changepassword.go](/workspace/socialpredict/backend/handlers/users/changepassword.go)
- [server_contract_test.go](/workspace/socialpredict/backend/server/server_contract_test.go)
- [API-ISSUES.md](/workspace/socialpredict/backend/docs/API-ISSUES.md)

The note should describe that explicitly and treat broader token redesign as deferred.

### Parameter families are route-specific today

The live API does not currently have universal pagination or filtering standards.

Instead:

- list markets uses `status`, `created_by`, `limit`, `offset`
- search markets uses `query`, legacy `q`, `status`, `limit`, `offset`

This is already documented in [openapi.yaml](/workspace/socialpredict/backend/docs/openapi.yaml) and implemented in handler/domain paths like [searchmarkets.go](/workspace/socialpredict/backend/handlers/markets/searchmarkets.go).

The active note should document the real parameter families, not propose `page/per_page/sort/order/meta/links` as if they already exist.

## What API Design Should Own

### Contract governance

API design in the active slice should own:

- truthful source-of-truth ordering
- route-family contract documentation
- response-family inventory
- public failure vocabulary where envelope routes exist
- OpenAPI parity with live route behavior
- auth-visible route semantics such as `mustChangePassword`

### Route-family migration direction

This note should also own the direction that:

- `PlainTextErrorResponse` is migration state, not target state
- route families should converge deliberately rather than through universal wrapper middleware
- legacy alias paths and action-style routes should be documented honestly until explicitly redesigned
- handler/middleware failure behavior should be tracked together

### Boundary clarity

This note should be explicit that:

- HTTP contract shaping is a boundary concern
- `API and Auth Contract Boundary` is the right concept
- API is not a business subdomain and should not become a new service/domain layer inside the monolith

## What This Note Should Not Own

This note should not become the home for every long-term API-platform idea.

It should explicitly defer:

- HATEOAS
- XML or broader content negotiation
- a universal `APIResponse` wrapper with timestamps, meta, links, and version fields
- header-based versioning
- `/v1` rollout and sunset-manager work
- generated-client-first workflow
- a top-level `api/` subsystem or API service domain
- route reorganization into a new CRUD taxonomy

Those topics now belong in [FUTURE/02-long-term-api-design.md](/workspace/socialpredict/README/PRODUCTION-NOTES/BACKEND/FUTURE/02-long-term-api-design.md), not in the active production note.

## Near-Term Sequencing

The near-term API direction should align with the current design-plan waves rather than opening a parallel API platform track.

1. Keep runtime and boundary failure ownership explicit so middleware-generated `429`, router-level `405`, and auth/runtime failures can be converged intentionally.
2. Keep `openapi.yaml` aligned route-by-route with live code and document mixed response families honestly during the migration.
3. Use the later scoped API/auth alignment wave to retire `internal/service/auth.HTTPError` and improve route-visible auth failure translation.
4. Keep parameter naming stable within current route families before attempting broader normalization.
5. Decide and document the explicit staging/production publishing path for backend-served docs so Swagger and `openapi.yaml` remain reachable behind reverse proxies without a duplicate frontend copy.
6. Prioritize the highest-value legacy route families for convergence rather than attempting a full API flag day.
7. Keep route reorganization, token redesign, and API-platform ambitions deferred until the current route-family and boundary work is substantially further along.

## Open Questions

- Which route families should converge on envelope success/failure first after the current security and auth cleanup work?
- Should all `/v0` application routes ultimately converge on `ReasonResponse`, or are there route families that should intentionally retain a non-envelope contract?
- Which current plain-text failures are acceptable migration state, and which should be prioritized for removal first?
- Should staging and production expose backend-served docs at `/swagger/` and `/openapi.yaml`, through a dedicated docs hostname, or only behind internal/admin access?
- When does the backend actually have enough stable contract surface to justify broader versioning policy or generated clients?
- Are there any route families where the current `limit/offset` posture is no longer sufficient and a broader pagination standard is justified?

## Explicit Do-Not-Do List

- Do not create a new top-level `api/` platform tree as part of the active slice.
- Do not add body-rewriting response middleware that captures successful handler output and re-emits it inside a universal wrapper.
- Do not pretend the backend already has one uniform success/error contract.
- Do not start `/v1` or header-based versioning before the `/v0` contract is materially more stable.
- Do not solve staging/production docs publishing by maintaining a second frontend-owned Swagger copy.
- Do not add HATEOAS, XML content negotiation, or generated clients as current-wave requirements.
- Do not collapse public API `reason` vocabulary into runtime telemetry or internal domain error taxonomy.
