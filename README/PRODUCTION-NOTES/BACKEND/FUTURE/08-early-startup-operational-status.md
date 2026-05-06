---
title: Early Startup Operational Status
document_type: production-notes
domain: backend
author: Patrick Delaney
updated_at: 2026-05-03T14:12:53Z
updated_at_display: "Sunday, May 3, 2026 at 2:12 PM UTC"
update_reason: "Defer pre-readiness HTTP status visibility until the startup serving model is redesigned safely."
status: draft
---

# Early Startup Operational Status

## Purpose

This note holds the deferred idea of exposing operational status while the
backend is still running startup work.

The active monitoring and alerting contract remains in
[09-monitoring-alerting.md](/workspace/socialpredict/README/PRODUCTION-NOTES/BACKEND/09-monitoring-alerting.md).

## Deferred Idea

The current backend completes environment loading, DB initialization, DB
readiness checks, configuration loading, startup mutation or verification, and
seed work before it calls `readiness.MarkReady()` and starts the HTTP server.
That fail-closed order is production-safe, but it means `/ops/status` cannot
report `live: true, ready: false` during early startup work because the HTTP
listener is not open yet.

A future design could start a small HTTP surface earlier with readiness closed,
then open readiness only after startup mutation or verification completes.

## Why Deferred

Starting HTTP earlier is a runtime design change, not a documentation tweak. It
must prove that user-facing application routes cannot serve traffic before the
backend has completed migrations, verified schema state, loaded security
configuration, and finished startup-owned seeds.

## Preconditions

Do not activate this work until the current WAVE08 and WAVE09 contracts are
proven in staging:

- `/health`, `/readyz`, `/ops/status`, `/openapi.yaml`, `/swagger`, and
  `/swagger/` are published through the production proxy.
- production compose starts the non-writer backend only after the startup writer
  becomes healthy.
- public host checks confirm that readiness controls traffic admission.
- shutdown closes readiness before the HTTP drain window.

## Possible Future Shape

A future implementation could:

- split the server into an early infra-only listener and application route
  activation, or
- start the full router early but gate all application routes until startup
  completion, or
- keep the current startup order and expose startup progress only through logs
  and process supervisor state.

The active production-readiness path should keep the current fail-closed startup
behavior until one of these designs has a concrete safety argument and tests.
