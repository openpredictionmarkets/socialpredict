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

## Published Operational Signals

The current backend boundary exposes a small operator-facing signal inventory:

- `GET /health` is a process liveness probe. It returns `text/plain` body
  `live` with `200` while the HTTP process is serving.
- `GET /readyz` is a readiness probe. It returns `text/plain` body `ready`
  with `200` only after startup mutation or verification has completed and the
  database remains reachable; otherwise it returns `not ready` with `503`.
- `GET /ops/status` is the minimal operator-facing status export. It returns
  JSON `{ live, ready, requestFailuresTotal, dbPool }`, uses `503` when the
  backend is unready, and keeps the non-probe request-failure counter plus SQL
  pool saturation/wait counters process-local for spike and pool-tuning alerts.
- Startup configuration load, database initialization, database readiness,
  security configuration, startup mutation mode, shutdown configuration,
  migration/verification, and seed failures are fatal startup failures. The
  process exits through the startup logger before readiness opens.
- Runtime-owned request failures currently share JSON `{ ok: false, reason }`
  responses for router `405`, security middleware `429`, and recovered
  unhandled panics.

The first supported alert set is intentionally small: backend down or unready
via `/health`, `/readyz`, or `/ops/status`; fatal startup failure via process
exit before readiness opens plus startup logger events; and severe request
failure spikes via the monotonic `/ops/status.requestFailuresTotal` counter.
`/ops/status.dbPool.waitCount` and `waitDurationNanoseconds` provide the first
pool saturation and wait-latency seam for DB pool tuning checks. These counters
are process-local and reset when the process restarts, so they are first spike
signals rather than fleet metrics.

These HTTP signals are reachable once the backend HTTP server is listening. The
current startup path completes migration or verification and opens readiness
before starting the listener, so `/ops/status` does not yet expose early startup
progress while migrations or startup-owned seeds are still running.

`GET /v0/system/metrics` remains an application reporting route for
economics/accounting output such as money creation and utilization. It is not
the operational monitoring surface and should not be treated as a liveness,
readiness, latency, or telemetry-export endpoint.

Remaining monitoring gaps should stay scoped to the next app-owned signal
seams: request latency/duration at the request boundary, route or reason
classification for server-side failures, process start/reset metadata for local
counters, and a clear backend-versus-proxy ownership line for traffic and
edge-failure signals. Prometheus exposition, Grafana dashboards, Alertmanager
routing, paging policy, centralized log-platform rollout, SLOs, and
error-budget programs remain deferred outside the current backend OpenAPI
contract.

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
