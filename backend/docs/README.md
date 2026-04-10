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

## How to Update

1. Add or adjust DTO structs in the relevant handler package (`backend/handlers/<service>/dto`).
2. Mirror those shapes under `components/schemas` and update the relevant path
   entry in `openapi.yaml`.
3. Keep responses consistent with handlers (e.g. all errors use the JSON
   wrapper `{ "error": "…" }`).
4. Run the OpenAPI linter once we wire one into CI (placeholder `make
   lint-openapi`) before committing.

When we spin a service into its own repo, copy the tagged section and any
referenced schemas or convert them into standalone files referenced via `$ref`.
