# SocialPredict Load Test Harness

This directory contains the early SocialPredict load-test harness. It lives in this repository while the API contract, fixture shape, and release dossier format are still changing quickly.

The boundary is intentionally portable. If the harness later needs independent versioning or reuse outside SocialPredict, move this top-level `loadtest/` directory to a separate repository.

## Directory Map

- `cli/`: wrapper commands for running scenarios and generating dossier summaries.
- `cli/OPERATING.md`: step-by-step operator/LLM runbook for staging load tests.
- `STAGING_RUNBOOK.md`: focused kconfs.com staging test ladder and dossier workflow.
- `TEMP_DROPLET_RUNBOOK.md`: end-to-end DigitalOcean temporary Droplet workflow for larger raw-IP capacity tests.
- `NEXT_LARGE_HOST_TEST_PLAN_2026-06-10.md`: planned larger-host rerun and mixed website/API workload plan.
- `k6/`: k6 API load-test scenarios.
- `fixtures/`: generated or operator-provided users, credentials, and market IDs. Ignored by default.
- `results/`: raw k6 outputs. Ignored by default.
- `dossier/`: release dossier schema and summarizer.
- `dossier/staging-capacity-2026-06-09.md`: latest kconfs.com staging capacity addendum, including the clean `35/sec for 5m` run.
- `hostops/`: operator-captured host observations from HostOps, SSH, DigitalOcean, and safe Linux commands. Ignored by default.

## Prerequisites

### macOS

Install k6 and Node:

```bash
brew install k6
brew install node
```

Verify:

```bash
k6 version
node --version
```

k6 is a standalone CLI load generator. It runs the JavaScript scenario files in `loadtest/k6/` using its own runtime, so it does not need Node to execute the load tests.

Node is used by this repository for `loadtest/dossier/summarize.mjs`, which converts k6 summary JSON into compact release dossier JSON.

### Load Generator Location

Run k6 from:

- your Mac for smoke and moderate tests
- a separate load-generator droplet for heavier tests

Do not run capacity tests from the same droplet that hosts the app and database. That would consume CPU, memory, disk, and network on the system being measured. Same-droplet runs are only useful for tiny smoke checks.

### Temporary Raw-IP Hosts

For short-lived DigitalOcean capacity tests, prefer creating a temporary Droplet
instead of permanently resizing staging. Use:

- `TEMP_DROPLET_RUNBOOK.md` for the complete `doctl`, GitHub secrets, Ansible deploy, resize, seed, test, dossier, and destroy sequence.
- `cli/OPERATING.md` for load-test CLI command reference and staging-oriented operations.

## Authentication Model

k6 uses normal SocialPredict fake-user credentials and `POST /v0/login` for API traffic. It does not use DigitalOcean credentials and it does not use a betting god key.

HostOps, SSH, and `doctl` are separate infrastructure tools for observing or provisioning hosts.

## Fixture Generation

Use the guarded SocialPredict seed command to create fake users, moderators, published markets, and the CSV files consumed by k6:

```bash
LOAD_TEST_ENABLED=true ./SocialPredict load seed --users 10 --moderators 2 --markets 5 --hot-markets 1 --reset
```

For a larger staging exercise:

```bash
LOAD_TEST_ENABLED=true ./SocialPredict load seed --users 50000 --moderators 100 --markets 1000 --hot-markets 10 --reset
```

Safety rules:

- `LOAD_TEST_ENABLED=true` is required for every seed run.
- `APP_ENV=production` is refused unless `LOAD_TEST_ALLOW_PRODUCTION=true` is also set.
- `--reset` removes only data with the configured load-test prefixes before recreating fixtures.
- Generated fixture CSVs are written to `loadtest/fixtures/` and are ignored by git.

The generated fake users log in through the normal `/v0/login` API and have `must_change_password=false`.
By default, each fake user starts with `1000000` credits so bet traffic is less likely to turn into balance-failure traffic. Override with `--user-balance N` when needed.

## Manual Fixture Files

For early smoke checks, you can still provide fixture CSVs manually.

`fixtures/users.csv`:

```csv
username,password
loaduser000001,loadtest-password
loaduser000002,loadtest-password
```

`fixtures/markets.csv`:

```csv
market_id,kind
1,hot
2,normal
```

## Quick Start

For the current OpenPredictionMarkets staging sequence against `https://kconfs.com`,
start with `STAGING_RUNBOOK.md`. It includes the preflight checks, fixture
seed/reset commands, recommended hot-market ladder, pass/fail criteria, and
dossier update workflow.

Start the app, seed a small local fixture set, and check prerequisites:

```bash
./SocialPredict up
LOAD_TEST_ENABLED=true ./SocialPredict load seed --users 10 --moderators 2 --markets 5 --hot-markets 1 --reset
./loadtest/cli/loadtest check
```

Or copy example fixture files without touching a database:

```bash
cp loadtest/fixtures/users.example.csv loadtest/fixtures/users.csv
cp loadtest/fixtures/markets.example.csv loadtest/fixtures/markets.csv
```

Run a smoke scenario:

```bash
./loadtest/cli/loadtest fixtures seed staging \
  --users 100 \
  --moderators 5 \
  --markets 20 \
  --hot-markets 2 \
  --reset
./loadtest/cli/loadtest fixtures pull staging
./loadtest/cli/loadtest run smoke --base-url https://kconfs.com --api-prefix /api
```

Run a low baseline without fractional k6 rates:

```bash
./loadtest/cli/loadtest run baseline \
  --base-url https://kconfs.com \
  --api-prefix /api \
  --duration 1m \
  --browse-rate 1 \
  --browse-time-unit 5s \
  --bet-rate 1 \
  --bet-time-unit 20s
```

Run a hot-market burst with setup-time pre-authentication so the measured
scenario focuses on concentrated betting throughput:

```bash
./loadtest/cli/loadtest run hot-market-burst \
  --base-url https://kconfs.com \
  --api-prefix /api \
  --duration 1m \
  --target-rate 10 \
  --preauth-users 25
```

Capture host CPU, RAM, disk, and Docker stats during a run:

```bash
./loadtest/cli/loadtest run hot-market-burst \
  --base-url https://kconfs.com \
  --api-prefix /api \
  --duration 1m \
  --target-rate 50 \
  --preauth-users 100 \
  --monitor-env staging \
  --monitor-interval 5
```

The monitor writes a CSV under `loadtest/hostops/` by default. This is ignored
by git and should be promoted into a curated dossier only after the run is
interpreted.

The CLI also writes a sibling host summary JSON and prints the key after-run
stats, including CPU, RAM, disk usage, disk IO, network IO, Docker aggregate
CPU/RAM, and backend/Postgres/Traefik CPU slices.

When `--monitor-env` is used, the CLI also captures a static host profile JSON
before the run. That profile records the control variables for interpreting
capacity evidence: OS/kernel, CPU count/model, RAM/swap, root disk size, Docker
server/storage/cgroup settings, Docker-visible CPU/RAM, and whether running
containers have explicit CPU or memory limits.

Interpretation:

- `cpu_user_pct`, `cpu_system_pct`, and `cpu_idle_pct` are whole-machine host CPU samples from `/proc/stat`.
- `docker_cpu_pct_sum` is the sum of `docker stats` CPU percentages across containers.
- On a `1 vCPU` host, Docker CPU near `100%` plus host idle near `0%` means the containers are effectively consuming the whole machine.
- If the host profile shows zero explicit container CPU/memory limits, there is no Docker Compose cap to raise; more headroom requires a larger host, fewer services on the host, or architectural changes.

Generate a release dossier from the k6 summary output:

```bash
./loadtest/cli/loadtest dossier --summary loadtest/results/<summary>.json --out loadtest/dossier/runs/<run>.json
```

Attach the host summary JSON to a generated dossier when available:

```bash
./loadtest/cli/loadtest dossier \
  --summary loadtest/results/<summary>.json \
  --host-summary loadtest/hostops/<run>-host-summary.json \
  --out loadtest/dossier/runs/<run>.json
```

Generated dossier JSON also includes `rateLimitEquivalents`, which converts the
measured successful bet rate into equivalent normal-limit client identities:

```text
ceil(successful_bets_per_second / normal_general_rate_per_second)
```

For the current `secure-default`/model-office policy, the normal sustained
general API limit is `1` request/second per client identity with burst `10`.
These are client identity/IP equivalents for the current limiter, not guaranteed
unique human users if many users share one NAT or proxy identity.

For capacity evidence, copy `loadtest/dossier/metadata.example.json` and record
the active `RATE_LIMIT_*` values under `rateLimitPolicy`. For
OpenPredictionMarkets staging, those values are expected to come from
`deploy/env/.env.staging` during single-source load tests. The summarizer also
copies k6 counters for `sp_rate_limited` and `sp_login_rate_limited` into the
dossier so rate-limit failures can be separated from app, database, or host
pressure.

Run from a Mac or a separate load-generator droplet for meaningful capacity evidence. Running k6 on the app/database droplet is only valid for tiny smoke checks.
