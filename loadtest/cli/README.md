# Load Test CLI

`loadtest/cli/loadtest` is a small wrapper around k6 and the dossier summarizer.

Seed fixture CSVs with `./SocialPredict load seed`, or provide files under `loadtest/fixtures/` or with `--users-file` and `--markets-file`.

## Commands

```bash
LOAD_TEST_ENABLED=true ./SocialPredict load seed --users 10 --moderators 2 --markets 5 --reset
./loadtest/cli/loadtest check
./loadtest/cli/loadtest host rate-limits staging
./loadtest/cli/loadtest host monitor staging --duration 2m --interval 5
./loadtest/cli/loadtest host summarize loadtest/hostops/<run>-host.csv
./loadtest/cli/loadtest fixtures seed staging --users 100 --moderators 5 --markets 20 --hot-markets 2 --reset
./loadtest/cli/loadtest fixtures pull staging
./loadtest/cli/loadtest run smoke --base-url https://kconfs.com --api-prefix /api
./loadtest/cli/loadtest run baseline --base-url https://kconfs.com --api-prefix /api --duration 5m --browse-rate 1 --browse-time-unit 5s --bet-rate 1 --bet-time-unit 20s
./loadtest/cli/loadtest run hot-market-burst --base-url https://kconfs.com --api-prefix /api --target-rate 50 --duration 60s --preauth-users 100
./loadtest/cli/loadtest run hot-market-burst --base-url https://kconfs.com --api-prefix /api --target-rate 50 --duration 60s --preauth-users 100 --monitor-env staging
./loadtest/cli/loadtest dossier --summary loadtest/results/<summary>.json --host-summary loadtest/hostops/<run>-host-summary.json --metadata loadtest/dossier/metadata.example.json --out loadtest/dossier/runs/<run>.json
```

## Authentication

k6 logs in with normal fake SocialPredict users from `users.csv` and uses normal bearer tokens for `/v0/bet`.

No DigitalOcean credentials or betting god token are used by this CLI.

## Operator Runbook

For a step-by-step staging sequence, including SSH key expectations and what an
LLM agent should run in order, see [`OPERATING.md`](./OPERATING.md).

## Remote Host Checks

Show the active remote `RATE_LIMIT_*` values:

```bash
./loadtest/cli/loadtest host rate-limits staging
```

Override the SSH target when needed:

```bash
./loadtest/cli/loadtest host rate-limits staging \
  --host root@45.55.227.1 \
  --key ~/.keys/socialpredict/staging/id_ed25519 \
  --repo-path /opt/socialpredict
```

Capture CPU, RAM, disk, and Docker stats as CSV while you run a test:

```bash
./loadtest/cli/loadtest host monitor staging --duration 2m --interval 5
```

Or attach monitoring directly to a k6 run:

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

The CSV is written to `loadtest/hostops/` by default. It includes host load,
CPU user/system/idle percentages, memory, swap, root disk usage, disk read/write
rates, network receive/transmit rates, summed Docker CPU/memory, and
backend/Postgres/Traefik CPU slices.

Summarize a telemetry CSV after the fact:

```bash
./loadtest/cli/loadtest host summarize loadtest/hostops/<run>-host.csv
```

When `--monitor-env` is attached to a k6 run, the CLI automatically writes a
sibling `<run>-host-summary.json` and prints the same summary after the k6
output.

## Remote Fixture Seed And Pull

Seed staging over SSH, then pull the generated fixture CSVs back to your
load-generator machine:

```bash
./loadtest/cli/loadtest fixtures seed staging \
  --users 100 \
  --moderators 5 \
  --markets 20 \
  --hot-markets 2 \
  --reset

./loadtest/cli/loadtest fixtures pull staging
```

The staging defaults are `root@kconfs.com`,
`~/.keys/socialpredict/staging/id_ed25519`, `/opt/socialpredict`, and
`/opt/socialpredict/loadtest/fixtures`.

Override values when needed:

```bash
./loadtest/cli/loadtest fixtures seed staging \
  --host root@45.55.227.1 \
  --key ~/.keys/socialpredict/staging/id_ed25519 \
  --repo-path /opt/socialpredict \
  --users 100 \
  --moderators 5 \
  --markets 20 \
  --hot-markets 2 \
  --reset

./loadtest/cli/loadtest fixtures pull staging \
  --host root@45.55.227.1 \
  --key ~/.keys/socialpredict/staging/id_ed25519 \
  --remote-path /opt/socialpredict/loadtest/fixtures \
  --local-path loadtest/fixtures
```

Public staging/prod Nginx publishes API routes under `/api`, so use `--api-prefix /api` when running against `https://kconfs.com` or `https://brierfoxforecast.com`. Direct backend targets such as `http://localhost:8080` should omit the prefix.

## Hot-Market Burst

`hot-market-burst` pre-authenticates users during k6 `setup()` and then reuses
those bearer tokens during the measured betting scenario. This keeps high-rate
hot-market tests focused on betting throughput instead of measuring login churn.

```bash
./loadtest/cli/loadtest run hot-market-burst \
  --base-url https://kconfs.com \
  --api-prefix /api \
  --duration 1m \
  --target-rate 50 \
  --preauth-users 100
```

## Low-Rate Runs

k6 `constant-arrival-rate` values must be positive integers. For sub-1/sec traffic, increase the time unit instead of passing a fractional rate:

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
