---
title: Load Testing And Release Dossier Evidence
document_type: production-notes
domain: features
author: Patrick Delaney
updated_at: 2026-05-28T00:00:00Z
updated_at_display: "Thursday, May 28, 2026"
update_reason: "Start the feature spec for staging load testing, hot-market betting simulations, and release dossier evidence capture."
status: draft
---

# Load Testing And Release Dossier Evidence

## Purpose

SocialPredict needs evidence about how much market and betting traffic a given deployment can handle before a release is treated as production-ready for larger public use.

This feature defines a safe load-testing program for staging and future production-like environments. It focuses on realistic API traffic, database contention, hot-market betting bursts, and a release dossier artifact that records what was tested and what the result means.

The release dossier target is compatible with the release dossier dashboard concept at <https://pwdel.github.io/socialpredict-release-dossier-dashboard/>.

## Feature Artifact Map

This directory keeps load-testing work together:

- [04-load-testing.md](./04-load-testing.md): feature overview, product/ops goals, expected scenarios, and acceptance criteria.
- [DESIGN.md](./DESIGN.md): architecture and boundary design aligned with the canonical design plan.
- [PLAN.md](./PLAN.md): implementation sequencing and PR slicing plan derived from the design.

## Product And Operations Questions

The first load-testing program should answer these questions:

- Given a specific DigitalOcean deployment size, how many normal users can browse, inspect markets, and place bets before latency or error rate becomes unacceptable?
- How does the system behave with 100 moderators, 1,000 created markets, and 50,000 fictional users?
- How does the system behave when 25-50% of the user base concentrates activity on one or a few hot markets?
- Does the app fail safely under load, or do errors create financial/accounting ambiguity?
- Which bottleneck appears first: app CPU, database CPU, database locks, disk I/O, memory, connection pool exhaustion, reverse proxy limits, rate limits, or application code?
- What deployment size or topology is needed for the next release target?

## Load Model

Baseline population target:

- 100 fictional moderators.
- 1,000 fictional markets.
- 50,000 fictional regular users.
- 10 hot markets.
- 1 extreme hot market.

Traffic target examples:

| Scenario | Example Shape | Approximate Write Load |
| --- | --- | --- |
| Wide one-hour activity | 50,000 users place one bet over one hour | 14 bets/second |
| Hot-minute activity | 25,000 users place one bet over one minute | 417 bets/second |
| Hot-ten-second burst | 25,000 users place one bet over ten seconds | 2,500 bets/second |
| Sub-second panic burst | thousands of users hit one market at once | must be measured, not guessed |

These numbers are intentionally directional. The implementation should make scenario parameters configurable rather than hard-code one traffic pattern.

## Scope

In scope:

- Staging-first load generation against normal public and authenticated API routes.
- Safe fictional users and markets created only in environments explicitly marked for load testing.
- Normal login and bearer-token behavior for test users.
- Betting load through `/v0/bet` using normal application rules.
- Read load against public market, position, credit, leaderboard, health, readiness, and operational-status routes.
- Dossier output that records environment, scenario, rates, latency, error rate, and infrastructure observations.
- Manual or scripted DigitalOcean metric capture for CPU, memory, disk, and network observations.

Out of scope for the first feature slice:

- Running destructive load tests against production.
- A permanent god key that bypasses normal application behavior.
- A new authorization model for load testing.
- A custom performance platform before k6 or another standard tool proves insufficient.
- Solving scaling problems before the first evidence pass shows the bottleneck.
- Adding Redis, queues, read replicas, or caching in the same PR as the test harness.

## Safety Rules

Load testing must be opt-in and environment-scoped.

Required safety posture:

- Load-test seed/reset commands must refuse to run unless an environment flag such as `LOAD_TEST_ENABLED=true` is present.
- `./SocialPredict load seed` is the current seed/reset entrypoint; it writes generated credentials and market IDs under `loadtest/fixtures/`.
- A representative staging seed command is `LOAD_TEST_ENABLED=true ./SocialPredict load seed --users 50000 --moderators 100 --markets 1000 --hot-markets 10 --reset`.
- The default production path must not expose test seeding or reset endpoints.
- Fictional users should authenticate through the normal login path unless a short-lived staging-only fixture export is explicitly implemented.
- Any admin or bootstrap token should be used only for seed/reset orchestration, not for placing bets.
- Tests should target staging or a dedicated load-test droplet, not production.
- Test data should be easy to wipe and should not be mixed with durable production data.

## API Surface Under Test

Initial API surfaces:

- `GET /health`
- `GET /readyz`
- `GET /ops/status`
- `GET /openapi.yaml`
- `GET /swagger/`
- `POST /v0/login`
- `GET /v0/setup/frontend`
- `GET /v0/markets`
- `GET /v0/markets/{id}`
- `GET /v0/markets/{id}/projection`
- `POST /v0/markets`
- `POST /v0/bet`
- `GET /v0/markets/bets/{marketId}`
- `GET /v0/markets/positions/{marketId}`
- `GET /v0/usercredit/{username}`
- `GET /v0/portfolio/{username}`
- `GET /v0/global/leaderboard`

The betting path is the first critical write path because it touches authentication, market state, account balance, transaction boundaries, and bet persistence.

## Tooling Direction

The first implementation should prefer `k6` for API load generation because it supports scenario-based request-rate testing, JSON result output, and scripting realistic HTTP flows from a local Mac or separate load-generator host.

k6 is not an in-server test runner for this feature. It is the external traffic generator. For end-to-end evidence, k6 should call the public staging URL, such as `https://kconfs.com`, through the same DNS, TLS, reverse proxy, backend, and Postgres path a real user would use.

Valid runner locations:

- a developer Mac for smoke and moderate tests
- a separate DigitalOcean load-generator droplet for heavier tests
- a GitHub workflow only for very small smoke checks, not heavy load

Avoid running the load generator on the same app/database droplet when measuring capacity. That would spend CPU, memory, disk, and network on the machine being measured and make the result misleading.

Locust can be considered later if Python-based user-behavior modeling becomes more valuable than k6's simpler API-load workflow.

Recommended first tool layout:

```text
loadtest/
  README.md
  cli/
    README.md
  k6/
    smoke.js
    baseline.js
    hot-market-burst.js
    soak.js
  fixtures/
    .gitkeep
  results/
    .gitignore
  dossier/
    schema.example.json
    runs/
      .gitignore
  hostops/
    README.md
```

This should start inside the SocialPredict repository only to keep early development close to the API, setup, and release-dossier contracts it depends on. The top-level `loadtest/` directory should be treated as a portable tool boundary so it can move to a separate repository later if the harness needs its own release cycle or is reused across multiple projects.

Directory intent:

- `loadtest/cli`: wrapper commands for seeding, running scenarios, and generating dossier summaries.
- `loadtest/k6`: k6 scenario scripts.
- `loadtest/fixtures`: generated users, credentials, token caches, and market ID files; ignored by default.
- `loadtest/results`: raw k6 outputs; ignored by default unless explicitly promoted.
- `loadtest/dossier`: compact release dossier schema and curated run summaries.
- `loadtest/hostops`: notes or captured observations from HostOps, SSH, DigitalOcean metrics, and safe host commands.

Runner authentication is split by concern:

- k6 uses normal SocialPredict fake-user credentials and `POST /v0/login` for API traffic.
- HostOps or SSH can be used separately for host observation.
- `doctl` can be used separately in the future to provision a load-generator droplet.
- DigitalOcean credentials are never used to bypass SocialPredict betting authorization.

The tool should run from a developer machine or from a separate DigitalOcean load-generator droplet. The load generator should not run on the same droplet as the app/database when measuring server capacity, because it would contaminate CPU/network results.

## Release Dossier Evidence

Each meaningful test run should produce a dossier artifact.

Minimum dossier fields:

```json
{
  "release": "vX.Y.Z or commit SHA",
  "environment": "staging",
  "baseUrl": "https://kconfs.com",
  "appTopology": "single droplet app+postgres",
  "databaseTopology": "local docker postgres",
  "dropletSize": "example: 2 vCPU / 4 GB",
  "scenario": "hot-market-burst",
  "seed": {
    "users": 50000,
    "moderators": 100,
    "markets": 1000,
    "hotMarkets": 10
  },
  "traffic": {
    "duration": "60s",
    "targetBetRatePerSecond": 1000,
    "readWriteMix": "70/30"
  },
  "results": {
    "requestsTotal": 0,
    "betsAttempted": 0,
    "betsSucceeded": 0,
    "errorRate": 0,
    "httpReqDurationP95Ms": 0,
    "httpReqDurationP99Ms": 0
  },
  "infrastructureObservations": {
    "appCpuPeakPercent": null,
    "memoryPeakPercent": null,
    "diskIoNotes": "",
    "dbConnectionNotes": ""
  },
  "decision": "pass | fail | inconclusive",
  "knownRisks": []
}
```

The dossier should be committed only when it is intended as a durable release artifact. Raw high-volume result files should usually stay out of git unless they are summarized.

## Observability Baseline

Minimum first-pass observations:

- k6 request count, throughput, error rate, p50, p95, and p99 latency.
- `/health`, `/readyz`, and `/ops/status` before, during, and after the test.
- DigitalOcean CPU, memory, disk I/O, and network graphs for the droplet.
- Basic Linux checks during load, such as `docker stats`, `df -h`, and selected Postgres connection views when safe.

Future observations may include:

- Postgres slow-query logs.
- `pg_stat_activity` and `pg_stat_statements` snapshots.
- App-level request metrics.
- Dedicated dashboards.
- Load-generator resource telemetry.

## Acceptance Criteria

The first load-testing feature is acceptable when:

- There is a documented, staging-safe way to seed fictional users and markets.
- There is a documented, repeatable way to run smoke, baseline, hot-market burst, and soak tests.
- The load generator uses normal API behavior for betting traffic.
- Results can be summarized into a release dossier artifact.
- The feature identifies bottlenecks without prematurely changing architecture.
- Operators can clearly tell whether a test is safe for staging, unsafe for production, or intended for a dedicated load-test environment.
