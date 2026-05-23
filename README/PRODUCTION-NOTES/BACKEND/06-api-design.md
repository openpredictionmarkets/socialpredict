---
title: API Design
document_type: production-notes
domain: backend
author: Patrick Delaney
updated_at: 2026-04-30T13:45:00Z
updated_at_display: "Thursday, April 30, 2026 at 1:45 PM UTC"
update_reason: "Finish the WAVE06 stop-and-review inventory for remaining plain-text/auth boundary seams."
status: active
---

# API Design

## Update Summary

On Sunday, May 10, 2026, this note was touched during API documentation consolidation. The deferred follow-up pointer now references `backend/docs/README.md` because the old `backend/docs/API-ISSUES.md` file was folded into the canonical backend docs README.

This note was updated on Sunday, April 26, 2026 to replace an older API standardization plan with guidance that matches the live SocialPredict backend, the current OpenAPI contract, and the active design-plan posture.

On Thursday, April 30, 2026, the runtime-boundary API behavior changed in a few important ways: `/health` now returns `live`, `/readyz` returns readiness state, runtime-owned `405` and `429` failures use JSON `reason` envelopes, and the private action routes enforce `PASSWORD_CHANGE_REQUIRED` before their domain handlers run.

WAVE06 also makes the mixed contract explicit in the OpenAPI artifact instead
of treating it as cleanup folklore. The source-of-truth order remains
`server.go` -> touched handlers or DTOs -> `backend/docs/openapi.yaml` -> `backend/docs/README.md`.
The route-family migration matrix below is mirrored by
`x-route-family-migration-matrix` in `backend/docs/openapi.yaml` and by focused
contract tests.

This slice also makes docs publishing explicit: local, dev, and production
surfaces use the backend-owned root paths `/swagger/` and `/openapi.yaml`.
The nginx templates publish those paths before the frontend catch-all route;
`/api/` does not own the docs contract.

The WAVE06 stop-and-review outcome is intentionally narrower than a platform
backlog: auth transport mapping now has a shared HTTP seam, touched private-user
and private-action paths use route-visible `ReasonResponse` failures,
`/v0/markets/search` retired its plain-text application failures while
preserving raw `SearchResponse` success, and the remaining plain-text boundary
inventory is concentrated in the markets route family plus intentional probe
transport.

| Topic | Prior to April 26, 2026 | After April 26, 2026 |
| --- | --- | --- |
| Core framing | Treated API design as a new platform to build from scratch | Treats API design as route-family contract governance and migration of a live monolith |
| Current-state accuracy | Assumed there was no OpenAPI or Swagger surface yet | Recognizes the live `openapi.yaml`, embedded Swagger UI, route/spec parity tests, and backend docs README |
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
   - human overview and deferred follow-ups in `backend/docs/README.md`
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

### Proxy docs publishing is explicit

The backend-served docs work directly when the backend is exposed at its own
root, and the nginx proxy templates now publish the same paths explicitly in
dev and production. Staging uses the same contract when it is deployed from the
production nginx template.

The chosen publishing path is primary-domain root publishing:

- `/swagger/`
- `/openapi.yaml`

The related behavior is:

- the backend serves docs at those root paths
- `GET /swagger` redirects to `/swagger/`
- the embedded Swagger UI requests `/openapi.yaml` as an absolute root path
- the dev and production nginx templates route `/openapi.yaml`, `/swagger`,
  and `/swagger/` to the backend before `/api/` and `/`; staging inherits this
  only when it uses the production template
- the Traefik template owns the public host and TLS edge; nginx owns the
  path-level docs publishing decision

This keeps the external docs contract identical to the backend contract while
leaving `/api/` as the application API prefix bridge.

The active rule is:

- keep the backend as the source of truth for Swagger and OpenAPI
- expose docs explicitly at `/swagger/` and `/openapi.yaml` through the
  reverse-proxy layer
- keep access open in this slice; if operations later requires restriction,
  apply it at the proxy or hostname layer without moving docs ownership
- do not maintain a second frontend-owned Swagger copy

This is a deployment and contract-publishing issue, not a reason to create a separate API documentation platform inside the application.

### Source-of-truth order already exists

The repo already has a practical source-of-truth order in [backend/docs/README.md](/workspace/socialpredict/backend/docs/README.md):

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

Several older compatibility handlers still return raw JSON on success and
`http.Error` style plain-text failures where they have not yet been migrated to
the canonical registered `/v0` route boundary.

Examples include:

- [getmarkets.go](/workspace/socialpredict/backend/handlers/markets/getmarkets.go)
- [listmarkets.go](/workspace/socialpredict/backend/handlers/markets/listmarkets.go)
- [marketdetailshandler.go](/workspace/socialpredict/backend/handlers/markets/marketdetailshandler.go)

#### Middleware and infra transport responses

Some failures and infra routes intentionally remain outside application envelopes today:

- `/health` returns plain text `live`
- `/readyz` returns plain text `ready` or `not ready`
- `/openapi.yaml` returns YAML
- `/swagger/*` serves static assets
- shared security middleware emits JSON `429` envelopes with stable `RATE_LIMITED` or `LOGIN_RATE_LIMITED` reasons
- router-level method handling emits JSON `405` envelopes and sets `Allow`

This mixed state is real and should be documented honestly rather than hidden behind a universal response-wrapper proposal.

### WAVE06 route-family migration matrix

The current contract slice is intentionally mixed. This matrix is the starting
point for route-family migration, not a promise that every route already uses
one target shape.

| Route family | Paths | Success contract | Failure contract | Migration state |
| --- | --- | --- | --- | --- |
| Infra probes | `/health`, `/readyz` | `text/plain` probe body (`live`, `ready`) | `text/plain` probe body (`not ready` where applicable) | Intentional infra transport |
| Infra docs | `/openapi.yaml`, `/swagger`, `/swagger/` | OpenAPI YAML, redirect, or embedded Swagger UI asset | Router/runtime failure outside application envelope | Intentional unversioned docs transport |
| Runtime middleware | all registered routes | Not applicable | JSON `{ ok: false, reason }` for router-owned `405` and middleware-owned `429` | Runtime-boundary envelope |
| Auth | `/v0/login` | JSON `{ ok: true, result }` | `ReasonResponse` | Envelope-based |
| Setup | `/v0/setup`, `/v0/setup/frontend` | Raw JSON DTO | `ReasonResponse` plus middleware `429` | Mixed raw success plus `ReasonResponse` |
| Reporting | `/v0/home`, `/v0/stats`, `/v0/system/metrics`, `/v0/global/leaderboard` | JSON `{ ok: true, result }` | `ReasonResponse` plus middleware `429` | Envelope-based |
| Markets | `/v0/markets`, `/v0/markets/status`, `/v0/markets/status/{status}`, legacy status aliases, market detail/resolve/leaderboard/projection routes, and both legacy market-projection slash variants | Mixed raw JSON DTO, no-content action, and selected envelope results | `ReasonResponse` plus middleware `429` | Mixed raw success plus `ReasonResponse` |
| Market search | `/v0/markets/search` | Raw JSON `SearchResponse` | `ReasonResponse` plus middleware `429` | Raw success with plain-text failures retired |
| Market bets and positions | `/v0/markets/bets/{marketId}`, `/v0/markets/positions/{marketId}`, `/v0/markets/positions/{marketId}/{username}` | JSON `{ ok: true, result }` | `ReasonResponse` plus middleware `429` where wrapped | Envelope-based |
| Public users | `/v0/userinfo/{username}`, `/v0/usercredit/{username}`, `/v0/portfolio/{username}`, `/v0/users/{username}/financial` | Raw JSON DTO | `ReasonResponse` plus middleware `429` | Mixed raw success plus `ReasonResponse` |
| Private users | `/v0/privateprofile`, `/v0/changepassword`, profile-change routes | JSON `{ ok: true, result }` | `ReasonResponse` plus middleware `429` | Envelope-based |
| Private actions | `/v0/bet`, `/v0/userposition/{marketId}`, `/v0/sell` | JSON `{ ok: true, result }` | `ReasonResponse` plus middleware `429` and the route-boundary `PASSWORD_CHANGE_REQUIRED` gate | Envelope-based |
| Admin and content | `/v0/admin/createuser`, `/v0/content/home`, `/v0/admin/content/home` | Mixed raw JSON DTO and JSON `{ ok: true, result }` | `ReasonResponse` plus middleware `429` where wrapped | Mixed raw success plus `ReasonResponse` |

The matrix deliberately tracks middleware and handler-owned responses together
because clients see both at the same HTTP boundary.

### Public reason vocabulary

The touched public `reason` values are client-facing HTTP contract values, not
runtime telemetry vocabulary and not the internal domain error taxonomy. They
are owned in code by `handlers.PublicFailureReasons()` and mirrored in
`ReasonResponse`:

`METHOD_NOT_ALLOWED`, `INVALID_REQUEST`, `INVALID_TOKEN`,
`AUTHORIZATION_DENIED`, `PASSWORD_CHANGE_REQUIRED`, `NOT_FOUND`,
`RATE_LIMITED`, `LOGIN_RATE_LIMITED`, `USER_NOT_FOUND`, `MARKET_NOT_FOUND`,
`VALIDATION_FAILED`, `MARKET_CLOSED`, `INSUFFICIENT_BALANCE`, `NO_POSITION`,
`INSUFFICIENT_SHARES`, `DUST_CAP_EXCEEDED`, `INTERNAL_ERROR`.

Runtime telemetry classifications, including values such as `error.type`, stay
separate from this public vocabulary. The public `reason` value is what clients
can branch on; telemetry fields are operator diagnostics.

### PlainTextErrorResponse migration state

`PlainTextErrorResponse` remains in the OpenAPI component set as an explicit
migration-state marker for infra probe failures and untouched handler-owned
plain-text failures still being retired. At this checkpoint that means
intentional `/health` and `/readyz` probe transport plus the remaining markets
compatibility seams listed in [backend/docs/README.md](/workspace/socialpredict/backend/docs/README.md):
`getmarkets.go`, `listmarkets.go`, `marketdetailshandler.go`,
legacy update/get methods on `handler.go`, and market-create compatibility
helpers in `createmarket.go`.

It is not the target application contract. Touched `/v0` route families should
converge toward `ReasonResponse` or documented infra transport behavior rather
than adding new plain-text application failures. The concrete next API slice is
the markets route family: retire unused compatibility entry points and convert
any still-routed market handler failures to shared `ReasonResponse` helpers
without a universal wrapper or success-body flag day.

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
- touched private-user auth failures are translated at the HTTP boundary and do
  not expose raw auth service messages in response bodies

Grounding:

- [loggin.go](/workspace/socialpredict/backend/internal/service/auth/loggin.go)
- [auth.go](/workspace/socialpredict/backend/internal/service/auth/auth.go)
- [changepassword.go](/workspace/socialpredict/backend/handlers/users/changepassword.go)
- [server_contract_test.go](/workspace/socialpredict/backend/server/server_contract_test.go)
- [backend/docs/README.md](/workspace/socialpredict/backend/docs/README.md)

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
3. Keep the current auth HTTP mapping seam in `handlers/authhttp` and retire
   direct auth-error-to-transport call sites as they are touched.
4. Keep parameter naming stable within current route families before attempting broader normalization.
5. Keep the explicit root publishing path for backend-served docs intact so Swagger and `openapi.yaml` remain reachable behind reverse proxies without a duplicate frontend copy.
6. Prioritize the highest-value legacy route families for convergence rather than attempting a full API flag day.
7. Keep route reorganization, token redesign, and API-platform ambitions deferred until the current route-family and boundary work is substantially further along.

## Open Questions

- Which markets compatibility entry points are still exercised by tests or
  callers, and which can be deleted before converting remaining market failures?
- Should all `/v0` application routes ultimately converge on `ReasonResponse`, or are there route families that should intentionally retain a non-envelope contract?
- Which current plain-text failures are acceptable migration state, and which should be prioritized for removal first?
- If operations later restricts docs access, should that restriction be host-based, network-based, or proxy-auth-based?
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
