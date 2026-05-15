---
title: Deployment Infrastructure Stage Test
document_type: stage-test-notes
domain: backend
author: Patrick Delaney
updated_at: 2026-05-14T10:23:25Z
updated_at_display: "Thursday, May 14, 2026 at 10:23 AM UTC"
update_reason: "Align staging checks with release-to-readiness feedback and current public environment examples."
status: active
---

# Deployment Infrastructure Stage Test

This staging checklist supports
[08-deployment-infrastructure.md](../08-deployment-infrastructure.md). It
captures what should be verified once WAVE08 is deployed to a staging host.

## Target Contract

The staging host should prove that the production compose and nginx topology
were applied together:

- one startup-writer backend service runs with `STARTUP_WRITER=true`
- request-serving backend service runs with `STARTUP_WRITER=false`
- nginx and frontend traffic target the non-writer `backend` service
- `/health` and `/readyz` are publicly reachable at the host root
- public release-to-readiness checks use `https://kconfs.com/health` and
  `https://kconfs.com/readyz`
- production uses the same public-boundary contract at
  `https://brierfoxforecast.com`
- Swagger UI and OpenAPI are publicly reachable at the host root
- graceful shutdown closes readiness before the backend exits

## Checklist

- [ ] Confirm the deployment used the WAVE08 compose file and nginx template.
- [ ] Confirm exactly one `backend-startup-writer` service exists.
- [ ] Confirm `backend-startup-writer` has `STARTUP_WRITER=true`.
- [ ] Confirm request-serving `backend` has `STARTUP_WRITER=false`.
- [ ] Confirm no public proxy target points at `backend-startup-writer`.
- [ ] Confirm `backend` depends on `backend-startup-writer` becoming healthy.
- [ ] Confirm `frontend` depends on request-serving `backend` health.
- [ ] Confirm `webserver` depends on request-serving `backend` health.
- [ ] Confirm the backend image healthcheck calls `/health`.
- [ ] Confirm production compose service healthchecks call `/readyz`.
- [ ] Confirm public `https://STAGING_HOST/health` returns `200` and `live`.
- [ ] Confirm public `https://STAGING_HOST/readyz` returns `200` and `ready`.
- [ ] Confirm the public probe result is treated as release-to-readiness
  evidence only, not a zero-downtime or business-correctness proof.
- [ ] Confirm `/ops/status` is not required for the deploy gate unless a later
  design-plan review promotes it.
- [ ] Confirm public `https://STAGING_HOST/openapi.yaml` returns the OpenAPI document.
- [ ] Confirm public `https://STAGING_HOST/swagger` redirects to `/swagger/`.
- [ ] Confirm public `https://STAGING_HOST/swagger/` renders Swagger UI, not only a blank HTML shell.
- [ ] Confirm Swagger UI loads `/swagger/swagger-ui-bundle.js` and `/swagger/swagger-initializer.js`.
- [ ] Confirm public `/api/` routes still proxy to the backend.
- [ ] Confirm public `/` routes still proxy to the frontend.
- [ ] Confirm backend logs show startup writer mode for `backend-startup-writer`.
- [ ] Confirm backend logs show non-writer verification mode for `backend`.
- [ ] Confirm failed migrations or seed failures prevent request-serving backend readiness.
- [ ] Confirm stopping `backend` marks `/readyz` unavailable before the container exits.
- [ ] Confirm `BACKEND_READINESS_DRAIN_SECONDS` is set to `5` in staging compose.
- [ ] Confirm `BACKEND_SHUTDOWN_TIMEOUT_SECONDS` is set to `10` in staging compose.
- [ ] Confirm backend `stop_grace_period` is long enough for drain plus shutdown.

## Suggested Commands

Replace `STAGING_HOST` with the actual staging domain.

```bash
curl -i https://STAGING_HOST/health
curl -i https://STAGING_HOST/readyz
curl -i https://STAGING_HOST/ops/status
curl -i https://STAGING_HOST/openapi.yaml
curl -I https://STAGING_HOST/swagger
curl -i https://STAGING_HOST/swagger/
```

On the staging server:

```bash
docker compose ps
docker compose logs backend-startup-writer backend webserver
docker compose exec backend-startup-writer printenv STARTUP_WRITER
docker compose exec backend printenv STARTUP_WRITER
docker compose exec backend printenv BACKEND_READINESS_DRAIN_SECONDS
docker compose exec backend printenv BACKEND_SHUTDOWN_TIMEOUT_SECONDS
```

## Pass Criteria

Staging passes WAVE08 deployment verification when the public host root exposes
the expected probe and docs routes, public traffic is routed only to the
non-writer backend service, the startup writer is the only service with
`STARTUP_WRITER=true`, and shutdown readiness behavior is visible in logs or
host-level checks. The required release gate remains `/health` body `live` and
`/readyz` body `ready`.
