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

## Authentication Model

k6 uses normal SocialPredict fake-user credentials and `POST /v0/login` for API traffic. It does not use DigitalOcean credentials and it does not use a betting god key.

HostOps, SSH, and `doctl` are separate infrastructure tools for observing or provisioning hosts.

## Minimal Fixture Files

The initial harness expects fixture CSVs to already exist. The guarded seed/reset command is a later implementation slice.

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

Copy or generate fixture files:

```bash
cp loadtest/fixtures/users.example.csv loadtest/fixtures/users.csv
cp loadtest/fixtures/markets.example.csv loadtest/fixtures/markets.csv
```

Then check prerequisites:

```bash
./loadtest/cli/loadtest check
```

Run a smoke scenario:

```bash
./loadtest/cli/loadtest run smoke --base-url https://kconfs.com
```

Generate a release dossier from the k6 summary output:

```bash
./loadtest/cli/loadtest dossier --summary loadtest/results/<summary>.json --out loadtest/dossier/runs/<run>.json
```

Run from a Mac or a separate load-generator droplet for meaningful capacity evidence. Running k6 on the app/database droplet is only valid for tiny smoke checks.
