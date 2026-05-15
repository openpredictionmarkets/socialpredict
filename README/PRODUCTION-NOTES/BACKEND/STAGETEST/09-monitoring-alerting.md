---
title: Monitoring and Alerting Stage Test
document_type: stage-test-notes
domain: backend
author: Patrick Delaney
updated_at: 2026-05-14T10:23:25Z
updated_at_display: "Thursday, May 14, 2026 at 10:23 AM UTC"
update_reason: "Refresh ops status shape and keep it separate from required release readiness."
status: active
---

# Monitoring and Alerting Stage Test

This staging checklist supports
[09-monitoring-alerting.md](../09-monitoring-alerting.md). It captures what
should be verified once WAVE09 is deployed to a staging host.

## Target Contract

The staging host should prove that the backend and production proxy expose the
first operational signal set:

- `/health` reports backend HTTP process liveness.
- `/readyz` reports traffic readiness after startup and database availability.
- `/ops/status` reports JSON `{ live, ready, requestFailuresTotal, dbPool }`.
- `/ops/status` is published at the public host root by nginx, not hidden behind
  `/api/` or direct `:8080` backend access.
- `/ops/status` is a runtime signal for operator troubleshooting or future
  smoke checks, not the required release-to-readiness gate.
- non-probe backend `5xx` responses increment the process-local
  `requestFailuresTotal` counter.
- `/v0/system/metrics` remains a business/economics reporting route, not the
  operational monitoring surface.
- Swagger UI and OpenAPI remain published and usable after the WAVE08/WAVE09
  merge.

## Checklist

- [ ] Confirm the deployed branch includes WAVE09 and the WAVE08 Swagger CSP fix.
- [ ] Confirm production nginx has exact root locations for `/health`, `/readyz`,
  and `/ops/status`.
- [ ] Confirm public `https://STAGING_HOST/health` returns `200` and body `live`.
- [ ] Confirm public `https://STAGING_HOST/readyz` returns `200` and body `ready`
  after startup completes.
- [ ] Confirm public `https://STAGING_HOST/ops/status` returns `200` JSON with
  `live: true`, `ready: true`, numeric `requestFailuresTotal`, and a `dbPool`
  object.
- [ ] Confirm `dbPool` fields are process-local SQL pool counters, not
  fleet-wide metrics.
- [ ] Confirm `/ops/status` response includes `Cache-Control: no-store`.
- [ ] Confirm `/ops/status` does not expose business metrics such as money
  creation, utilization, user balances, or market accounting values.
- [ ] Confirm public `https://STAGING_HOST/openapi.yaml` documents `/ops/status`.
- [ ] Confirm public `https://STAGING_HOST/swagger/` renders Swagger UI and can
  load the OpenAPI document.
- [ ] Confirm public `https://STAGING_HOST/swagger/swagger-ui-bundle.js` returns
  the Swagger UI JavaScript asset.
- [ ] Confirm `GET /v0/system/metrics` still behaves as the economics/accounting
  route and is not used as a health or readiness check.
- [ ] Confirm backend logs include readiness transition events when readiness
  opens or closes.
- [ ] Confirm startup failures before readiness opens are visible as process
  exits plus startup logger events such as `startup.incompatibility` or
  `startup.migration_failed`.
- [ ] Confirm `/ops/status` is not treated as an early startup-progress endpoint;
  current startup still completes mutation or verification before HTTP starts
  listening.
- [ ] Confirm any future requirement for `live: true, ready: false` during early
  startup is tracked in
  [08-early-startup-operational-status.md](../FUTURE/08-early-startup-operational-status.md).

## Suggested Commands

Replace `STAGING_HOST` with the actual staging domain.

```bash
curl -i https://STAGING_HOST/health
curl -i https://STAGING_HOST/readyz
curl -i https://STAGING_HOST/ops/status
curl -i https://STAGING_HOST/openapi.yaml
curl -I https://STAGING_HOST/swagger
curl -i https://STAGING_HOST/swagger/
curl -i https://STAGING_HOST/swagger/swagger-ui-bundle.js
```

Optional JSON-focused check:

```bash
curl -fsS https://STAGING_HOST/ops/status
```

Expected shape:

```json
{
  "live": true,
  "ready": true,
  "requestFailuresTotal": 0,
  "dbPool": {
    "maxOpenConnections": 25,
    "openConnections": 1,
    "inUseConnections": 0,
    "idleConnections": 1,
    "waitCount": 0,
    "waitDurationNanoseconds": 0,
    "maxIdleClosedConnections": 0,
    "maxLifetimeClosedConnections": 0
  }
}
```

On the staging server:

```bash
docker compose ps
docker compose logs backend
docker compose exec backend printenv STARTUP_WRITER
```

## Pass Criteria

Staging passes WAVE09 monitoring verification when the public host root exposes
`/health`, `/readyz`, and `/ops/status`; `/ops/status` returns the expected
runtime JSON without business metrics; Swagger and OpenAPI still work; and the
current limitation around early-startup status visibility is understood as a
deferred design issue rather than a WAVE09 production-readiness blocker.
The required release-to-readiness gate remains `/health` body `live` and
`/readyz` body `ready`; `/ops/status` stays outside that gate unless later
promoted by design-plan review.
