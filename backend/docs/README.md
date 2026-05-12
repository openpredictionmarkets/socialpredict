# SocialPredict Backend API

`backend/docs` is the canonical API documentation directory for the backend.
Keep the machine-readable contract and the human overview here so contributors
and operators do not have to choose between competing API docs.

## Source Of Truth

The current API source-of-truth order is:

1. `backend/server/server.go`
2. Handler and DTO code under `backend/handlers/**`
3. `backend/docs/openapi.yaml`
4. This README for human context and deferred follow-ups

`openapi.yaml` is the only OpenAPI document maintained by the backend. Do not
add shadow copies under `README/BACKEND/API`, `backend/README/BACKEND/API`, or
other documentation trees.

## Published Docs

The backend serves the canonical API docs directly:

- `GET /openapi.yaml` returns `backend/docs/openapi.yaml`.
- `GET /swagger/` serves the embedded Swagger UI.
- `GET /swagger` redirects to `/swagger/`.

In local development, use:

```bash
curl http://localhost:8080/openapi.yaml
open http://localhost:8080/swagger/
```

The dev and production nginx templates intentionally route `/openapi.yaml`,
`/swagger`, and `/swagger/` to the backend before frontend catch-all routing.
Do not publish a separate frontend-owned Swagger copy.

## Contract Shape

The API is still in a staged migration from older raw/partial responses toward
more consistent JSON envelopes. The OpenAPI file records this honestly instead
of pretending the migration is complete.

Current route-family notes:

- Infra probes use plain text: `/health` and `/readyz`.
- Operator status uses cache-disabled JSON: `/ops/status`.
- Runtime middleware can return JSON `{ ok: false, reason }` for router-owned
  `405` and middleware-owned `429` failures.
- Many newer or touched handlers return envelope-shaped success or failure
  responses, and legacy market list/get/details/update failures now use shared
  failure envelopes.
- Some older market and public-user success contracts intentionally remain raw
  DTOs for compatibility.

Public failure `reason` values are owned by `handlers.PublicFailureReasons` and
mirrored in `openapi.yaml` under `x-route-family-migration-matrix`.

## Updating The API Contract

When changing API behavior:

1. Update `backend/server/server.go` if routes or methods change.
2. Update handler DTOs and tests with the runtime behavior.
3. Update `backend/docs/openapi.yaml` in the same change.
4. Update this README only when the documentation model, migration notes, or
   deferred follow-ups change.
5. Run the focused backend OpenAPI tests before merging.

Useful validation commands:

```bash
cd backend
go test ./...
```

If a full backend run is too broad for the change, at minimum run the OpenAPI
and server contract tests that cover route/spec parity and docs publishing.

## Deferred API Follow-Ups

These are intentionally deferred API-shaping decisions, not contradictions in
the current contract.

### Limited-Scope Token Login

`POST /v0/login` currently returns a normal bearer token plus
`mustChangePassword`. Protected handlers enforce the password-change gate at the
HTTP boundary, while `POST /v0/changepassword` intentionally accepts the current
token path so first-login users can complete the password change.

A future redesign may issue limited-scope or short-lived first-login tokens, but
that would change auth semantics across multiple routes and should be handled as
a separate product/security decision.

### Public Route Organization

`backend/server/server.go` remains the live route source of truth. The OpenAPI
document mirrors the current monolith layout, including public aliases and
legacy service-shaped paths that still exist in code.

Do not rewrite the API docs into a new REST taxonomy unless the route code and
clients are changing in the same work.

### Bets-To-Trades Naming

The API still uses bets terminology in routes, tags, and schema names including
`/v0/markets/bets/{marketId}`, `/v0/bet`, and `/v0/sell`.

A future rename to trades should be done as a coordinated code, client, and docs
change. Avoid partial route/tag/schema renames in unrelated documentation work.

### Remaining Markets Boundary Cleanup

Markets create/search now use shared security validation and bounded query
parsing, but older market detail, resolve, projection, and compatibility methods
still own some local path/action parsing and failure shaping.

The next narrow migration seam is the remaining markets path/action helper gap:
market ID, projection amount, resolution outcome, and related failure responses.
Do not turn that into a generic validation registry or broad middleware rewrite.
