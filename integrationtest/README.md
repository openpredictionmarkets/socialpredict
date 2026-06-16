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
- `scripts/multiple-choice-binary-api.mjs`

## Quick Start

Run against a local dev stack after `./SocialPredict dev-bootstrap-users`:

```bash
node integrationtest/scripts/multiple-choice-binary-api.mjs \
  --base-url http://localhost:8080 \
  --api-prefix /v0
```

Defaults assume seeded users `admin`, `testuser01`, and `testuser02` all use
password `password`.

## Operating Notes

- Prefer testing through public HTTP/API routes when possible.
- Record the branch, date, environment, seed users, and setup values in each report.
- Keep raw outputs out of the repo unless they are small and useful for review.
- Put large or generated outputs in `artifacts/` with a short index file if they need to be preserved.
