# API Documentation

This directory holds the OpenAPI contract for the monolith.  The document is
written so each “service” slice (Markets, Users, Bets, …) can be lifted into its
own microservice spec without rewriting definitions.

## Layout

- `openapi.yaml` – master document. Paths are grouped by tag. As we backfill
  more routes, keep each service’s paths together and scope the shared schemas
  in `components/schemas`.
- Future service fragments can live under `services/<service>.yaml` and be
  `$ref`’d from `openapi.yaml` when we need more modularity.

## Published Endpoints

The backend is the single source of truth for API docs. It serves the canonical
contract document at `GET /openapi.yaml` and the embedded Swagger UI at
`GET /swagger/`; `GET /swagger` redirects to `/swagger/`.

Local backend access uses those root paths directly, for example
`http://localhost:8080/openapi.yaml` and `http://localhost:8080/swagger/`.
The dev and production nginx templates deliberately publish the same root paths
to the backend before the frontend catch-all route; staging inherits this
contract when it is deployed from the production template. In production,
Traefik owns the public host and TLS edge while nginx owns these path routes.
Do not publish these docs by copying Swagger assets into the frontend or by
relying on `/api/` prefix routing.

## How to Update

1. Add or adjust DTO structs in the relevant handler package (`backend/handlers/<service>/dto`).
2. Mirror those shapes under `components/schemas` and update the relevant path
   entry in `openapi.yaml`.
3. Keep responses consistent with handlers and the route-family migration
   matrix in `openapi.yaml`.
4. Run the OpenAPI linter once we wire one into CI (placeholder `make
   lint-openapi`) before committing.

When we spin a service into its own repo, copy the tagged section and any
referenced schemas or convert them into standalone files referenced via `$ref`.
