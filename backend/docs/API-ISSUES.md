# Backend API Follow-Ups

This file tracks the remaining API-shaping decisions after Waves 1-4.

Source of truth order for the current backend remains:

1. `backend/server/server.go`
2. touched handlers and DTOs under `backend/handlers/**`
3. `backend/docs/openapi.yaml`
4. this file

This document is intentionally narrow. It should reflect the code that exists
today, not earlier uncertainty and not aspirational redesign work that has not
been scheduled.

## Completed In Waves 1-4

The following items were the main concerns in the earlier draft of this file
and should now be treated as closed for documentation purposes:

- The backend route inventory is no longer being reconstructed from guesswork.
  `backend/server/server.go` is the route source of truth and
  `backend/docs/openapi.yaml` has been updated around that implementation.
- The previously tracked error-contract cleanup hotspots were normalized onto
  the current response helpers instead of ad hoc raw error writes. This applies
  to the stats handler, homepage CMS handler, bets buy/sell handlers, positions
  handlers, display-name/profile helpers, and the authenticated user-position
  handler.
- The extracted auth/config/domain wiring introduced in the earlier waves is
  now the implementation baseline used by the current handlers and OpenAPI
  document. `API-ISSUES.md` should no longer describe that work as unresolved.

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

Why it remains deferred:

- The backend now documents and enforces the current contract consistently
  enough to avoid treating this as an unknown.
- The product/security redesign question is still whether first-login users
  should receive a limited-scope or short-lived token instead of the normal
  bearer token.

Decision for this wave:

- Keep the existing login contract and its current OpenAPI shape.
- Do not redesign token issuance in this documentation-only task.

### 2. Public Route Reorganization

Current implementation:

- `backend/server/server.go` makes the existing public and protected routes
  explicit.
- The OpenAPI document reflects the current monolith route layout rather than a
  future CRUD-style reorganization.
- Legacy and service-shaped paths still coexist where the code already exposes
  them.

Why it remains deferred:

- The remaining work is a route-design decision, not a documentation
  reconstruction problem.
- Reorganizing public resource paths would be broader than this task and would
  require coordinated code, OpenAPI, and client updates.

Decision for this wave:

- Keep documenting the live route structure exactly as implemented.
- Do not revive the earlier CRUD rewrite proposal in this file.

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
- rewriting the API into a new CRUD path taxonomy
- bundling implementation changes into this documentation cleanup

If future work changes the backend or OpenAPI contract, update that code/spec
first and then revise this file to match the new implementation.
