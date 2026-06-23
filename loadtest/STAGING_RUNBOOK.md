# Staging Load-Test Runbook

This runbook describes how to repeat and extend SocialPredict load tests against the OpenPredictionMarkets staging instance at `https://kconfs.com`.

Use this when preparing a new capacity dossier or when checking whether a staging deploy can still handle the known baseline workload.

## Scope

This is a staging runbook, not a production proof. It assumes:

- target app: `https://kconfs.com`
- API prefix: `/api`
- remote host: `root@kconfs.com`
- remote repo path: `/opt/socialpredict`
- local SSH key: `~/.keys/socialpredict/staging/id_ed25519`
- load generator: your Mac or a separate load-generator host

Do not run heavy k6 tests from the same Droplet that runs the app and Postgres. Same-host tests consume the resources being measured.

For larger one-off capacity tests on a separate raw-IP DigitalOcean Droplet, use
`TEMP_DROPLET_RUNBOOK.md` instead of resizing staging.

## Safety Rules

- Run these tests only against staging unless a separate temporary load-test host has been created.
- Do not run these commands against `brierfoxforecast.com` or any user-facing model-office/prod domain.
- Confirm staging is using the high staging rate-limit profile before running single-source load tests.
- Use `--reset` before proof runs so old fixture bets do not contaminate results.
- Do not delete Docker volumes during cleanup unless you have explicitly confirmed they are disposable.

## Prerequisites

Install the local tools:

```bash
brew install k6
brew install node
```

Verify the load-test CLI:

```bash
./loadtest/cli/loadtest check
```

Confirm SSH access:

```bash
ssh -i ~/.keys/socialpredict/staging/id_ed25519 root@kconfs.com 'hostname && date'
```

`doctl` is not required for the normal staging runbook. Use it only if you also want to inspect DigitalOcean Droplet metadata or graphs.

## Preflight

Confirm the public app is alive and ready:

```bash
curl -sS https://kconfs.com/health
curl -sS https://kconfs.com/readyz
curl -sS https://kconfs.com/api/ops/status
```

Expected readiness body:

```text
ready
```

Capture the host profile and current disk state:

```bash
./loadtest/cli/loadtest host profile staging
./HostOps host disk staging
```

Confirm staging rate limits:

```bash
./loadtest/cli/loadtest host rate-limits staging
```

Expected staging values for single-source external load tests:

```env
RATE_LIMIT_LOGIN_RATE_PER_SECOND=50
RATE_LIMIT_LOGIN_BURST=100
RATE_LIMIT_GENERAL_RATE_PER_SECOND=500
RATE_LIMIT_GENERAL_BURST=1000
RATE_LIMIT_CLEANUP_INTERVAL=5m
```

Stop if staging is using conservative model-office limits. Single-source k6 from one Mac will otherwise test the limiter, not app capacity.

## Optional Cleanup

Use this if disk pressure or stale images are visible before a test:

```bash
./HostOps host cleanup staging
./HostOps host cleanup staging --all-images
./HostOps host disk staging
```

Do not use volume cleanup for routine staging tests. Database and certificate volumes can be important depending on the host state.

## Seed And Pull Fixtures

For a small staging confirmation:

```bash
./loadtest/cli/loadtest fixtures seed staging \
  --users 100 \
  --moderators 5 \
  --markets 20 \
  --hot-markets 2 \
  --reset

./loadtest/cli/loadtest fixtures pull staging
```

For a stronger run with more pre-authenticated users:

```bash
./loadtest/cli/loadtest fixtures seed staging \
  --users 500 \
  --moderators 10 \
  --markets 50 \
  --hot-markets 5 \
  --reset

./loadtest/cli/loadtest fixtures pull staging
```

Use the stronger seed before any run with `--preauth-users` above `100`.

The remote seed wrapper runs `./SocialPredict load seed` with:

```env
LOAD_TEST_ENABLED=true
LOAD_TEST_ALLOW_PRODUCTION=true
```

Fixture CSVs are pulled into `loadtest/fixtures/` and are ignored by git.

If seed fails, do not continue with `fixtures pull`. A pull after a failed seed
can copy stale CSVs from an older run and make the next k6 command look valid
while targeting old fixture IDs.

If seed fails with a schema error such as:

```text
ERROR: column "steward_username" of relation "markets" does not exist
```

staging is running with a database schema older than the source checkout used
by the load-test seeder. The normal fix is:

1. Run the `Create and publish Docker images` workflow from `main` so GHCR
   `latest` contains the current backend migrations.
2. Run the `Deploy To Staging` workflow from `main` so Ansible pulls and
   restarts the current image.
3. Wait for public readiness.
4. Verify the expected migration/column before reseeding:

```bash
ssh -i ~/.keys/socialpredict/staging/id_ed25519 root@kconfs.com '
  cd /opt/socialpredict &&
  set -a && . ./.env && set +a &&
  docker exec socialpredict-postgres-container psql \
    -U "$POSTGRES_USER" \
    -d "$POSTGRES_DATABASE" \
    -Atc "select exists (
      select 1
      from information_schema.columns
      where table_name = '\''markets'\''
        and column_name = '\''steward_username'\''
    );"
'
```

Expected output:

```text
t
```

## Smoke Test

Run smoke first after every seed/reset:

```bash
./loadtest/cli/loadtest run smoke \
  --base-url https://kconfs.com \
  --api-prefix /api \
  --monitor-env staging \
  --monitor-interval 5
```

Smoke should have:

- `0` failed HTTP requests
- `0` failed bets
- health, readiness, status, login, market detail, and bet checks passing

Do not proceed to capacity tests if smoke fails.

## Baseline Tests

Low baseline:

```bash
./loadtest/cli/loadtest run baseline \
  --base-url https://kconfs.com \
  --api-prefix /api \
  --duration 2m \
  --browse-rate 5 \
  --bet-rate 2 \
  --monitor-env staging \
  --monitor-interval 5
```

Moderate baseline:

```bash
./loadtest/cli/loadtest run baseline \
  --base-url https://kconfs.com \
  --api-prefix /api \
  --duration 2m \
  --browse-rate 20 \
  --bet-rate 10 \
  --monitor-env staging \
  --monitor-interval 5
```

These runs mix browsing and betting. Use them to confirm normal public API behavior before isolating hot-market write pressure.

## Hot-Market Ladder

Before running any command in this ladder, load users/markets into staging and
pull the generated CSVs locally:

```bash
./loadtest/cli/loadtest fixtures seed staging \
  --users 100 \
  --moderators 5 \
  --markets 20 \
  --hot-markets 2 \
  --reset

./loadtest/cli/loadtest fixtures pull staging
```

For the `50/sec` sustained `5m` command, seed more users first:

```bash
./loadtest/cli/loadtest fixtures seed staging \
  --users 500 \
  --moderators 10 \
  --markets 50 \
  --hot-markets 5 \
  --reset

./loadtest/cli/loadtest fixtures pull staging
```

Previous staging evidence on the Basic `1 vCPU / 1 GiB RAM / 25 GiB Disk` host showed:

- `50` pre-authenticated bets/sec for `1m` passed cleanly.
- `75` bets/sec for `1m` showed warning signs.
- `75` bets/sec for `5m` failed/degraded.
- `100` bets/sec for `1m` was beyond the host.

Those are historical ceiling notes, not the safest place to restart testing.
For routine staging checks, begin smaller and prove the fixture/setup path
before trying to confirm the old `50/sec` result.

Recommended next staging ladder:

1. `5` bets/sec for `1m` with `25` pre-authenticated users as the setup sanity check.
2. `10` bets/sec for `1m` with `25` pre-authenticated users as the first staging write check.
3. `20` bets/sec for `1m` with `50` pre-authenticated users.
4. `35` bets/sec for `1m` with `100` pre-authenticated users.
5. `50` bets/sec for `1m` with `100` pre-authenticated users only after the smaller runs are clean.
6. `50` bets/sec for `5m` with `500` pre-authenticated users only when intentionally collecting sustained capacity evidence.
7. Do not run `75/sec` for `5m` unless you are intentionally collecting a failure boundary.

Small first hot-market command:

```bash
./loadtest/cli/loadtest run hot-market-burst \
  --base-url https://kconfs.com \
  --api-prefix /api \
  --duration 1m \
  --target-rate 10 \
  --preauth-users 25 \
  --setup-timeout 3m \
  --monitor-env staging \
  --monitor-interval 5
```

Known historical confirmation command:

```bash
./loadtest/cli/loadtest run hot-market-burst \
  --base-url https://kconfs.com \
  --api-prefix /api \
  --duration 1m \
  --target-rate 50 \
  --preauth-users 100 \
  --setup-timeout 5m \
  --monitor-env staging \
  --monitor-interval 5
```

Sustained confirmation command:

```bash
./loadtest/cli/loadtest run hot-market-burst \
  --base-url https://kconfs.com \
  --api-prefix /api \
  --duration 5m \
  --target-rate 50 \
  --preauth-users 500 \
  --setup-timeout 8m \
  --monitor-env staging \
  --monitor-interval 5
```

Seed at least as many regular users as `--preauth-users` before running a
pre-authenticated burst. For the `5m` sustained command above, run the stronger
`--users 500` seed first.

If k6 shows only `Run setup()` and no hot-market traffic yet, the scenario is
still pre-authenticating users. That is not capacity evidence. Let setup finish,
or restart with a smaller `--preauth-users` value such as `25`.

## Pass And Fail Criteria

A clean proof run should have:

- `http_req_failed = 0`
- `sp_bets_failed = 0`
- `dropped_iterations = 0`
- p95 HTTP latency below roughly `1s` for a `1m` run
- p95 HTTP latency preferably below `1-2s` for a `5m` run
- no readiness failure during or immediately after the run
- available RAM not falling below roughly `150 MiB` on the `1 GiB` staging host

Treat the run as a warning or boundary if:

- any bets fail
- k6 drops iterations
- p95 latency crosses `1s` on a short run
- host CPU idle stays near `0%`
- Postgres or Traefik CPU dominates the host
- RAM headroom collapses
- the app becomes unavailable from a browser during the test

For release claims, prefer the highest clean `5m` run over a higher `1m` burst.

## After-Run Artifacts

Each monitored run writes:

- k6 summary JSON under `loadtest/results/`
- host telemetry CSV under `loadtest/hostops/`
- host summary JSON under `loadtest/hostops/`
- host profile JSON under `loadtest/hostops/`

Find the newest files:

```bash
ls -t loadtest/results/*summary.json | head
ls -t loadtest/hostops/*host-summary.json | head
ls -t loadtest/hostops/*host-profile.json | head
```

Generate a compact dossier JSON:

```bash
./loadtest/cli/loadtest dossier \
  --summary loadtest/results/<summary>.json \
  --host-summary loadtest/hostops/<host-summary>.json \
  --out loadtest/dossier/runs/<run>.json
```

The raw `results/`, `hostops/`, `fixtures/`, and `dossier/runs/` outputs are ignored by git. Commit only curated markdown conclusions unless a raw artifact is intentionally promoted.

## Dossier Update Workflow

After a meaningful run:

1. Record the command, target, duration, rate, fixture size, and timestamp.
2. Record whether the run started from a fresh `--reset` fixture state.
3. Summarize k6 pass/fail, successful bets/sec, failed bets, p95, max latency, and dropped iterations.
4. Summarize host CPU, RAM, disk, Postgres CPU, backend CPU, and Traefik CPU.
5. Compare the result to the previous staging dossier.
6. Update or create a markdown dossier under `loadtest/dossier/`.

Suggested next staging dossier name:

```text
loadtest/dossier/staging-capacity-2026-06-08.md
```

## Interpretation Notes

- Staging uses higher per-IP rate limits than model-office/prod so one Mac can create pressure.
- The normal rate limiter is per client identity/IP, not a global server cap.
- A single-source k6 run does not prove behavior for many NATs, proxies, or geographically distributed clients.
- Staging colocates app, Traefik, nginx, frontend, backend, and Postgres on one host.
- Old accumulated bets can slow later tests. Use `--reset` before proof runs.
- Host telemetry is sampled, not continuous tracing. It is sufficient for capacity dossiers but not a replacement for production observability.
