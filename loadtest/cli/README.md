# Load Test CLI

`loadtest/cli/loadtest` is a small wrapper around k6 and the dossier summarizer.

Seed fixture CSVs with `./SocialPredict load seed`, or provide files under `loadtest/fixtures/` or with `--users-file` and `--markets-file`.

## Commands

```bash
LOAD_TEST_ENABLED=true ./SocialPredict load seed --users 10 --moderators 2 --markets 5 --reset
./loadtest/cli/loadtest check
./loadtest/cli/loadtest fixtures pull staging
./loadtest/cli/loadtest run smoke --base-url https://kconfs.com --api-prefix /api
./loadtest/cli/loadtest run baseline --base-url https://kconfs.com --api-prefix /api --duration 5m --browse-rate 1 --browse-time-unit 5s --bet-rate 1 --bet-time-unit 20s
./loadtest/cli/loadtest run hot-market-burst --base-url https://kconfs.com --api-prefix /api --target-rate 100 --duration 60s
./loadtest/cli/loadtest dossier --summary loadtest/results/<summary>.json --metadata loadtest/dossier/metadata.example.json --out loadtest/dossier/runs/<run>.json
```

## Authentication

k6 logs in with normal fake SocialPredict users from `users.csv` and uses normal bearer tokens for `/v0/bet`.

No DigitalOcean credentials or betting god token are used by this CLI.

## Remote Fixture Pull

After running `./SocialPredict load seed` on a remote staging host, pull the generated fixture CSVs back to your load-generator machine:

```bash
./loadtest/cli/loadtest fixtures pull staging
```

The staging defaults are `root@kconfs.com`, `~/.keys/socialpredict/staging/id_ed25519`, and `/opt/socialpredict/loadtest/fixtures`.

Override values when needed:

```bash
./loadtest/cli/loadtest fixtures pull staging \
  --host root@45.55.227.1 \
  --key ~/.keys/socialpredict/staging/id_ed25519 \
  --remote-path /opt/socialpredict/loadtest/fixtures \
  --local-path loadtest/fixtures
```

Public staging/prod Nginx publishes API routes under `/api`, so use `--api-prefix /api` when running against `https://kconfs.com` or `https://brierfoxforecast.com`. Direct backend targets such as `http://localhost:8080` should omit the prefix.

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
