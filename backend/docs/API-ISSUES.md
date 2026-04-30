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

## Deferred Decisions

Only the following API decisions remain deferred at this stage.

### 1. Limited-Scope Token Login Redesign

Current implementation:

- `POST /v0/login` returns a normal bearer token plus `mustChangePassword`.
- Protected handlers typically enforce the password-change gate through
  `auth.ValidateUserAndEnforcePasswordChangeGetUser(...)` or
  `AuthService.CurrentUser(...)`.
- `POST /v0/changepassword` still accepts an authenticated request using the
  current token-validation path.
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
- treating security-boundary cleanup as an API-design backlog after WAVE05; the
  remaining raw auth and validation failures are tracked as backend auth or
  market-handler boundary migration seams
- rewriting the API into a new CRUD path taxonomy
- bundling unrelated implementation changes into this documentation cleanup

If future work changes the backend or OpenAPI contract, update that code/spec
first and then revise this file to match the new implementation.
