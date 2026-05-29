---
title: Load Testing And Release Dossier Design
document_type: feature-design
domain: features
author: Patrick Delaney
updated_at: 2026-05-28T00:00:00Z
updated_at_display: "Thursday, May 28, 2026"
update_reason: "Add architecture and boundary design for load testing, hot-market betting simulations, and release dossier evidence capture."
status: draft
---

# Load Testing And Release Dossier Design

## Purpose

This document translates [04-load-testing.md](./04-load-testing.md) into an architecture and boundary design before implementation begins.

It is not the implementation plan. The implementation sequence lives in [PLAN.md](./PLAN.md).

## Design Inputs

Primary inputs:

- [04-load-testing.md](./04-load-testing.md)
- Canonical design plan: `spec-socialpredict-tasks-auto/lib/design/design-plan.json`
- Current backend API source of truth: `backend/docs/openapi.yaml`
- Release dossier dashboard concept: <https://pwdel.github.io/socialpredict-release-dossier-dashboard/>

The canonical design plan remains the repository-level design source of truth. This feature design must conform to it rather than create a competing architecture.

## Designer Lens Review

Evans/domain lens:

- Load testing should use SocialPredict language: users, moderators, markets, hot markets, bets, positions, balances, readiness, release evidence, and deployment topology.
- The feature should distinguish test orchestration from application policy. Fictional users and markets are setup data, not a new production domain.
- A release dossier is an operational evidence artifact, not an analytics product feature.

Fowler/evolutionary lens:

- Start with a simple external load generator and JSON evidence before introducing a performance platform.
- Keep test scenarios parameterized and repeatable so results can evolve across droplet sizes and database topologies.
- Optimize only from measured evidence. Do not add Redis, queues, replicas, or caching before the first bottleneck is known.

Martin/clean-architecture lens:

- Application load tests should call public/API boundaries rather than reach directly into repositories.
- Seed/reset helpers must be isolated from normal production behavior and guarded by runtime configuration.
- Load-test implementation should not introduce god-mode application paths that bypass the real auth, account, or betting flows.

## Problem Framing

SocialPredict's most important high-load risk is not generic page traffic. It is a concentrated betting event where many authenticated users place bets on the same market or small set of markets inside a short time window.

This creates pressure on:

- HTTP request handling.
- JWT/auth validation.
- User balance mutation.
- Bet insertion.
- Market state checks.
- Postgres transaction and row-lock behavior.
- Read-model fan-out for positions, market details, and leaderboards.
- Docker host CPU, memory, disk I/O, and network capacity.

The design must produce evidence about this pressure without making production less safe.

## Bounded Context Alignment

| Design-plan boundary | Load-testing ownership |
| --- | --- |
| Release and Deployment Control | Owns when load tests are run, which environment is targeted, and how evidence is attached to release decisions. |
| API and Auth Contract Boundary | Owns the route-visible contracts used by load tests and OpenAPI alignment for generated clients or scripts. |
| Runtime Bootstrap and Infrastructure | Owns `LOAD_TEST_ENABLED`, deployment topology notes, readiness, DB pool settings, and safe startup behavior. |
| Prediction Market Context | Owns market creation, publication state, hot-market selection, and market visibility rules exercised by tests. |
| Betting and Position Ledger Context | Owns buy/sell correctness, transaction boundaries, account balance effects, and contention behavior under load. |
| Participant Account Context | Owns fictional user credentials, balances, password-change state, and normal login behavior used by tests. |
| Observability Boundary | Owns logs, request correlation, health/readiness/status observations, and future metric export decisions. |
| Documentation and Release Evidence | Owns release dossier shape, run notes, result summaries, and operator guidance. |

## Context Map

The load-testing feature should sit outside the app runtime for traffic generation and inside guarded app/runtime seams only for seed/reset support.

- External load generator: runs k6 scripts from a Mac or separate load-generator host and calls the public staging URL through DNS, TLS, ingress/reverse proxy, backend, and database.
- Seed/reset command: creates fictional users, moderators, and markets in staging only.
- API traffic: uses normal HTTP routes and bearer tokens.
- Dossier builder: summarizes k6 output plus environment metadata into a stable JSON artifact.
- Operator documentation: explains safe targets, commands, expected outputs, and how to read results.

Running k6 on the app/database droplet is valid only for tiny local smoke checks. It is not valid capacity evidence because the load generator would compete with the app and Postgres for the same host resources.

## Boundary Decisions

### Do Not Use A Betting God Key

A god key would invalidate the test because it would skip or distort the real traffic path.

Betting load must use normal application behavior:

- `POST /v0/login` or pre-generated normal user credentials.
- Bearer token auth.
- Normal `/v0/bet` payloads.
- Normal user balances and market state rules.
- Normal failure reasons and rate limits unless the test explicitly evaluates rate-limit behavior.

A staging-only bootstrap token may be used to create/reset fixtures, but not to place bets.

### Keep HostOps Separate

HostOps is for host and infrastructure operations: SSH, environment inspection, log access, and future cloud orchestration.

Load testing is application/API verification. It should live in a top-level `loadtest/` tree while the harness is young. HostOps can help an operator connect to a host for observation, but it should not own k6 scenarios or application seed semantics.

### Keep The First CLI In This Repository

The first load-test CLI should live in this repository, not a separate repository, because early development needs tight feedback with SocialPredict's API and fixture behavior.

Reasons:

- The CLI depends on this app's OpenAPI contract, route names, fixture semantics, and release dossier shape.
- Reviewing load-test scripts beside the API they exercise makes contract drift easier to catch.
- Keeping the first implementation under `loadtest/` avoids a second repository before the harness proves it needs independent release/versioning.
- A top-level `loadtest/` directory is easier to move later than scripts scattered through app, backend, or HostOps directories.

A separate repository can be reconsidered later if the load generator becomes a reusable product, needs a separate release cycle, or must be shared across multiple applications.

Proposed portable boundary:

```text
loadtest/
  cli/       wrapper commands for seed/run/dossier workflows
  k6/        API scenario scripts
  fixtures/ generated local fixture files, ignored by default
  results/  raw k6 outputs, ignored by default
  dossier/  compact release dossier schema and curated summaries
  hostops/  host observation notes and captured metrics
```

`loadtest/hostops` should store observations gathered via HostOps, SSH, DigitalOcean screenshots/exports, or safe Linux commands. It should not become HostOps itself.

### Runner Authentication And Operation

The runner machine authenticates in two different ways depending on what it is doing.

For application traffic:

- k6 authenticates by logging in as fictional users through `POST /v0/login`.
- k6 stores normal bearer tokens in memory for the test run.
- k6 places bets through `/v0/bet` using those normal bearer tokens.
- k6 does not need DigitalOcean credentials for normal API traffic.

For host observation or load-generator setup:

- an operator may use HostOps or SSH to inspect the staging host
- an operator may use `doctl` to create a separate load-generator droplet in the future
- DigitalOcean credentials are not part of the betting API path

This preserves a clean split: DigitalOcean auth operates infrastructure, while SocialPredict fake-user auth exercises the product API.

### Keep Dossier Output Small And Durable

Raw load-test output can be large. The release dossier should be a summarized evidence artifact with links or paths to raw results when needed.

Durable dossier fields should be stable enough for the dashboard to render:

- release or commit SHA
- environment
- topology
- seed counts
- scenario
- latency and error metrics
- infrastructure observations
- decision
- known risks

### Keep Database Truth Explicit

Postgres is the system of record for betting and account state. Load tests should treat database behavior as central evidence, not an implementation detail.

The first feature should measure before redesigning. Later architecture decisions might include larger droplets, app/database separation, managed Postgres, connection pool tuning, query/index optimization, cached read models, or event-driven projections. Those are follow-up decisions, not prerequisites for the first test harness.

## Scenario Design

Initial scenario families:

| Scenario | Purpose | Primary risk exposed |
| --- | --- | --- |
| Smoke | Prove credentials, market IDs, and API paths work. | Broken setup or auth. |
| Baseline browse | Measure normal read behavior across market lists/details/portfolio. | Read latency and fan-out. |
| Baseline bet | Measure low-to-moderate write traffic. | Basic transaction throughput. |
| Hot-market burst | Concentrate bet writes on one market. | DB locks, connection pool pressure, latency spikes. |
| Mixed hot markets | Spread bursts across several markets. | Whether contention is market-local or global. |
| Soak | Keep steady traffic for a longer period. | Memory, disk, logs, and slow degradation. |

## Seed Design

Seed/reset should support:

- deterministic username prefixes
- deterministic password convention for fictional users
- configurable user counts
- configurable moderator counts
- configurable market counts
- configurable hot-market selection
- option to export a credential/token fixture for load generation
- clear refusal outside `LOAD_TEST_ENABLED=true`

Seeded users should have `must_change_password=false` so they can place bets without frontend intervention. Passwords must be fictional and environment-specific.

## Metrics And Thresholds

Initial metrics:

- HTTP request rate.
- HTTP error rate.
- bet success count.
- bet rejection count by status/reason.
- p50/p95/p99 request duration.
- max request duration.
- dropped iterations or client-side failures.
- `/health`, `/readyz`, and `/ops/status` status before/during/after.
- app CPU, memory, disk I/O, network.
- Postgres connection count and lock notes when available.

Initial pass/fail gates should be intentionally conservative and adjustable:

- smoke scenario: zero unexpected failures.
- baseline scenario: p95 below a chosen target and low error rate.
- hot-market scenario: no data corruption or accounting ambiguity; latency/error thresholds can be exploratory until a capacity target is chosen.
- soak scenario: no sustained memory/disk growth that threatens the host.

## Deployment Topology Notes

The dossier must record the topology because results are meaningless without it.

Examples:

- single DigitalOcean droplet running app, frontend, Traefik/nginx, and local Docker Postgres
- app droplet plus managed Postgres
- staging droplet size
- database volume type
- Docker image tags or commit SHA
- DB pool settings
- load-generator location and machine type

The load generator should normally run outside the target host so it does not steal CPU from the system under test.

## Failure Modes To Preserve

A load test failure is useful if it is truthful.

The implementation should preserve evidence for:

- HTTP 401/403 auth failures.
- HTTP 409 market-state failures.
- HTTP 422 insufficient-balance failures.
- HTTP 429 rate-limit failures.
- HTTP 5xx server failures.
- timeouts.
- database connection exhaustion.
- readiness failure during load.

Do not hide these under a generic success/failure counter. The dossier should summarize failure categories.

## Open Questions

- What exact latency/error SLO should gate a release for staging and model-office/prod?
- Should load-test seed/reset live in `./SocialPredict`, a Go command under `backend/cmd`, or both?
- Should the first load-generator scripts store user credentials or obtain tokens dynamically per run?
- Should DigitalOcean metrics be manually recorded in the first version or collected through API later?
- When should Postgres-specific telemetry become a required part of the dossier?
- Should load tests run only manually, or should a small smoke load test be available as a workflow dispatch?
