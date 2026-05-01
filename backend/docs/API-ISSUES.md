# Backend API Follow-Ups

This file tracks only intentionally deferred API-shaping decisions after the
final API sweep.

Source of truth order for the current backend remains:

1. `backend/server/server.go`
2. touched handlers and DTOs under `backend/handlers/**`
3. `backend/docs/openapi.yaml`
4. this file

The final sweep reconciled the live backend, embedded OpenAPI document, and
Swagger UI surface for the currently implemented contract. If future work
introduces route/spec drift, update the code and `backend/docs/openapi.yaml`
first and only then revise this file if a decision is intentionally deferred.

## WAVE06 Route-Family Migration Matrix

`backend/docs/openapi.yaml` now carries the canonical machine-checkable matrix
as `x-route-family-migration-matrix`. This file mirrors the deferred-decision
parts of that contract so follow-up API work does not need to rediscover the
mixed state from handler code.

WAVE06 stop-and-review outcome:

- Touched private-user routes no longer expose raw auth transport messages in
  their response bodies. They route auth outcomes through `handlers/authhttp`
  and return stable `ReasonResponse` values.
- `/v0/changepassword` intentionally uses token-only auth so first-login users
  can complete the password change, but its auth failures still use
  `ReasonResponse` rather than raw auth messages.
- `/v0/markets/search` keeps its raw `SearchResponse` success body while its
  touched validation, method, service, and write failures now use
  `ReasonResponse`; its plain-text application failures are retired.
- Remaining documented plain-text application failures are limited to older
  markets compatibility entry points and are the next route-family migration
  seam, not an API-platform backlog.

The source-of-truth order remains:

1. `backend/server/server.go`
2. touched handlers and DTOs under `backend/handlers/**`
3. `backend/docs/openapi.yaml`
4. this file

| Route family | Paths | Success contract | Failure contract | Migration state |
| --- | --- | --- | --- | --- |
| Infra probes | `/health`, `/readyz` | `text/plain` probe body | `text/plain` probe body where applicable | Intentional infra transport |
| Infra docs | `/openapi.yaml`, `/swagger`, `/swagger/` | YAML, redirect, or embedded Swagger UI asset | Router/runtime failure outside application envelope | Intentional unversioned docs transport published at the proxy root, not under `/api/` |
| Runtime middleware | all registered routes | Not applicable | JSON `{ ok: false, reason }` for router-owned `405` and middleware-owned `429` | Runtime-boundary envelope |
| Auth | `/v0/login` | JSON `{ ok: true, result }` | `ReasonResponse` | Envelope-based |
| Setup | `/v0/setup`, `/v0/setup/frontend` | Raw JSON DTO | `ReasonResponse` plus middleware `429` | Mixed raw success plus `ReasonResponse` |
| Reporting | `/v0/home`, `/v0/stats`, `/v0/system/metrics`, `/v0/global/leaderboard` | JSON `{ ok: true, result }` | `ReasonResponse` plus middleware `429` | Envelope-based |
| Markets | `/v0/markets`, status aliases, detail, resolve, leaderboard, projection routes, and both legacy market-projection slash variants | Mixed raw JSON DTO, no-content action, and selected envelope results | `ReasonResponse` plus middleware `429` | Mixed raw success plus `ReasonResponse` |
| Market search | `/v0/markets/search` | Raw JSON `SearchResponse` | `ReasonResponse` plus middleware `429` | Raw success with plain-text failures retired |
| Market bets and positions | market bet and position routes under `/v0/markets/...` | JSON `{ ok: true, result }` | `ReasonResponse` plus middleware `429` where wrapped | Envelope-based |
| Public users | `/v0/userinfo/{username}`, `/v0/usercredit/{username}`, `/v0/portfolio/{username}`, `/v0/users/{username}/financial` | Raw JSON DTO | `ReasonResponse` plus middleware `429` | Mixed raw success plus `ReasonResponse` |
| Private users | `/v0/privateprofile`, `/v0/changepassword`, profile-change routes | JSON `{ ok: true, result }` | `ReasonResponse` plus middleware `429` | Envelope-based |
| Private actions | `/v0/bet`, `/v0/userposition/{marketId}`, `/v0/sell` | JSON `{ ok: true, result }` | `ReasonResponse` plus middleware `429` and `PASSWORD_CHANGE_REQUIRED` gate | Envelope-based |
| Admin and content | `/v0/admin/createuser`, `/v0/content/home`, `/v0/admin/content/home` | Mixed raw JSON DTO and JSON `{ ok: true, result }` | `ReasonResponse` plus middleware `429` where wrapped | Mixed raw success plus `ReasonResponse` |

The public `reason` values in this matrix are client-facing contract values
only. They must stay separate from runtime or telemetry vocabulary such as
`error.type`.

The current public `reason` vocabulary is:

`METHOD_NOT_ALLOWED`, `RATE_LIMITED`, `LOGIN_RATE_LIMITED`,
`INVALID_REQUEST`, `INVALID_TOKEN`, `AUTHORIZATION_DENIED`,
`PASSWORD_CHANGE_REQUIRED`, `NOT_FOUND`, `USER_NOT_FOUND`,
`MARKET_NOT_FOUND`, `VALIDATION_FAILED`, `MARKET_CLOSED`,
`INSUFFICIENT_BALANCE`, `NO_POSITION`, `INSUFFICIENT_SHARES`,
`DUST_CAP_EXCEEDED`, and `INTERNAL_ERROR`.

`PlainTextErrorResponse` is retained only as explicit migration state for infra
probe failures and untouched handler-owned plain-text failures that have not yet
been converted. At this checkpoint the explicit inventory is:

- `/health` and `/readyz` probe transport, which intentionally remains
  `text/plain`.
- The older market handler entry points in `handlers/markets/getmarkets.go`,
  `handlers/markets/listmarkets.go`,
  `handlers/markets/marketdetailshandler.go`,
  `handlers/markets/resolvemarket.go`, and the legacy update/get methods on
  `handlers/markets/handler.go`.
- The remaining market-create compatibility helpers in
  `handlers/markets/createmarket.go` that still use `http.Error` while the
  canonical registered `POST /v0/markets` path now goes through
  `Handler.CreateMarket` and `ReasonResponse` failures.

The precise next migration seam is the markets route family: remove or retire
the legacy compatibility entry points where they are no longer routed, and then
convert any still-registered market route failures to shared `ReasonResponse`
helpers without introducing a universal wrapper or changing successful raw DTO
contracts in the same slice.

## Publishing Decision

The canonical backend docs endpoints are `GET /openapi.yaml`, `GET /swagger/`,
and the `GET /swagger` redirect. The dev and production nginx templates publish
those paths at the primary domain root and proxy them to the backend before
frontend routing; staging inherits this contract only when deployed from the
production template. The `/api/` proxy remains for application API traffic only
and does not own the docs contract. In production, Traefik owns the public host
and TLS edge; nginx owns the docs path publishing behind that edge.

Access restriction is not added in this slice. If operations later needs to
limit public docs access, keep the backend endpoints and Swagger assets
backend-owned and add the restriction at the proxy or host layer without
creating a frontend-maintained Swagger copy.

## Deferred Decisions

Only the following API decisions remain deferred at this stage.

### 1. Limited-Scope Token Login Redesign

Current implementation:

- `POST /v0/login` returns a normal bearer token plus `mustChangePassword`.
- Protected handlers enforce the password-change gate through HTTP-boundary
  helpers such as `authhttp.CurrentUser(...)` or the shared auth service path,
  with raw auth messages translated before they reach the response body.
- `POST /v0/changepassword` still accepts an authenticated request using the
  current token-validation path.
- The private-users route family now routes auth checks through the HTTP
  boundary helper: profile and profile-change routes enforce the
  `mustChangePassword` gate, while `/v0/changepassword` intentionally uses the
  token-only auth path so first-login users can complete the password change.
- Admin-only routes use `AuthService.RequireAdmin(...)`, which enforces the
  same password-change gate before the admin role check.

Why it remains deferred:

- The product/security redesign question is still whether first-login users
  should receive a limited-scope or short-lived token instead of the normal
  bearer token.
- That redesign would change auth semantics across multiple routes and does not
  belong in this validation-and-reconciliation task.

Decision for this wave:

- Keep the existing login contract and its current OpenAPI shape.
- Do not redesign token issuance in this task.

### 2. Public Route Reorganization

Current implementation:

- `backend/server/server.go` remains the route source of truth.
- The OpenAPI document mirrors the current monolith route layout, including the
  existing public aliases and legacy service-shaped paths that still exist in
  code.
- Swagger UI is served from `/swagger/`, while the live contract document is
  served from `/openapi.yaml`.

Why it remains deferred:

- The remaining work is a route-design decision, not a documentation
  reconstruction problem.
- Reorganizing public resource paths would be broader than this task and would
  require coordinated code, OpenAPI, and client updates.

Decision for this wave:

- Keep documenting the live route structure exactly as implemented.
- Do not revive the earlier CRUD-style rewrite proposal in this file.

### 3. Bets-To-Trades Rename

Current implementation:

- The API still uses bets terminology in both routes and tags, including
  `/v0/markets/bets/{marketId}`, `/v0/bet`, `/v0/sell`, and the related
  OpenAPI schema/tag naming.

Why it remains deferred:

- The naming change is still a cross-cutting rename with client and
  documentation impact.
- Nothing in the current code state narrows that work beyond a future rename
  decision.

Decision for this wave:

- Keep the existing bets terminology as the canonical current contract.
- Do not partially rename routes, tags, or schemas in this task.

## Non-Goals For This File

The following ideas were discussed historically but are not active issues for
this task:

- forcing a universal `ok/result/reason` response envelope across the API
- treating security-boundary cleanup as an API-platform backlog; the remaining
  auth call-site cleanup and market-handler failures are tracked as bounded
  backend boundary migration seams
- rewriting the API into a new CRUD path taxonomy
- bundling unrelated implementation changes into this documentation cleanup

If future work changes the backend or OpenAPI contract, update that code/spec
first and then revise this file to match the new implementation.
