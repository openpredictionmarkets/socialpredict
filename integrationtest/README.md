# Integration Testing

This directory holds end-to-end integration testing material for SocialPredict.

Use this top-level area for tests that exercise the application through real
HTTP/API flows, Docker services, and database state. It is intentionally separate
from `loadtest/`, which is for capacity and performance testing.

## Structure

- `cases/`: scenario definitions and expected behavior.
- `reports/`: dated test run reports and findings.
- `scripts/`: future executable helpers for repeatable integration runs.
- `artifacts/`: future raw outputs, API captures, screenshots, or exported data.

## Current Coverage

- `cases/multiple-choice-binary-markets.md`
- `reports/multiple-choice-binary-markets-2026-06-15.md`
- `reports/schemathesis-read-2026-06-16.md`
- `scripts/multiple-choice-binary-api.mjs`
- `scripts/schemathesis-read.sh`

These files are the current scenario/test-case source of truth for the multiple-choice binary market work. No root-level `TEST_SCENARIO.md`, `scenario.md`, or `TESTCASE.md` file is currently present in this repo.

## Current Passing Baseline

Last updated: 2026-06-16.

| Suite | Command | Current result |
|-------|---------|----------------|
| Multiple-choice binary API scenario runner | `node integrationtest/scripts/multiple-choice-binary-api.mjs --base-url http://localhost:8080 --api-prefix /v0` | 17/17 checks passing |
| Read-only Schemathesis contract smoke | `MAX_EXAMPLES=1 integrationtest/scripts/schemathesis-read.sh` | 4/4 operations passing |
| Backend tests | `JWT_SIGNING_KEY=test-secret-key-for-testing go test ./...` | Passing |
| Frontend build | `npm run build:report` | Passing with existing warnings |

## Quick Start

Run against a local dev stack after `./SocialPredict dev-bootstrap-users`:

```bash
node integrationtest/scripts/multiple-choice-binary-api.mjs \
  --base-url http://localhost:8080 \
  --api-prefix /v0
```

Defaults assume seeded users `admin`, `testuser01`, and `testuser02` all use
password `password`.

Run read-only OpenAPI fuzz/contract checks with Schemathesis:

```bash
integrationtest/scripts/schemathesis-read.sh
```

Optional environment variables:

```bash
BASE_URL=http://localhost:8080
MAX_EXAMPLES=5
REPORT=junit
PHASES=coverage
AUTH_TOKEN=<bearer-token-for-authenticated-read-routes>
READ_PATHS='/v0/setup /v0/setup/frontend /v0/stats /v0/market-tags'
```

The script writes ignored Schemathesis artifacts under `integrationtest/artifacts/`.

## Operating Notes

- Prefer testing through public HTTP/API routes when possible.
- Record the branch, date, environment, seed users, and setup values in each report.
- Keep raw outputs out of the repo unless they are small and useful for review.
- Put large or generated outputs in `artifacts/` with a short index file if they need to be preserved.
