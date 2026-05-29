# SocialPredict Load Test Harness

This directory contains the early SocialPredict load-test harness. It lives in this repository while the API contract, fixture shape, and release dossier format are still changing quickly.

The boundary is intentionally portable. If the harness later needs independent versioning or reuse outside SocialPredict, move this top-level `loadtest/` directory to a separate repository.

## Directory Map

- `cli/`: wrapper commands for running scenarios and generating dossier summaries.
- `k6/`: k6 API load-test scenarios.
- `fixtures/`: generated or operator-provided users, credentials, and market IDs. Ignored by default.
- `results/`: raw k6 outputs. Ignored by default.
- `dossier/`: release dossier schema and summarizer.
- `hostops/`: operator-captured host observations from HostOps, SSH, DigitalOcean, and safe Linux commands. Ignored by default.

## Authentication Model

k6 uses normal SocialPredict fake-user credentials and `POST /v0/login` for API traffic. It does not use DigitalOcean credentials and it does not use a betting god key.

HostOps, SSH, and `doctl` are separate infrastructure tools for observing or provisioning hosts.

## Minimal Fixture Files

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

```bash
./loadtest/cli/loadtest check
./loadtest/cli/loadtest run smoke --base-url https://kconfs.com
./loadtest/cli/loadtest dossier --summary loadtest/results/<summary>.json --out loadtest/dossier/runs/<run>.json
```

Run from a Mac or a separate load-generator droplet for meaningful capacity evidence. Running k6 on the app/database droplet is only valid for tiny smoke checks.
