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
- `cases/sell-shares-overcashout.md`
- `cases/sell-shares-two-share-cap.md`
- `reports/multiple-choice-binary-markets-2026-06-15.md`
- `reports/schemathesis-read-2026-06-16.md`
- `scripts/multiple-choice-binary-api.mjs`
- `scripts/sell-shares-overcashout.mjs`
- `scripts/sell-shares-two-share-cap.mjs`
- `scripts/schemathesis-read.sh`
- `scripts/schemathesis-grouped-market.mjs`

These files are the current scenario/test-case source of truth for the multiple-choice binary market work. No root-level `TEST_SCENARIO.md`, `scenario.md`, or `TESTCASE.md` file is currently present in this repo.

## Current Passing Baseline

Last updated: 2026-06-17.

| Suite | Command | Current result |
|-------|---------|----------------|
| Multiple-choice binary API scenario runner | `node integrationtest/scripts/multiple-choice-binary-api.mjs --base-url http://localhost:8080 --api-prefix /v0` | 17/17 checks passing |
| Sell shares over-cashout API scenario runner | `node integrationtest/scripts/sell-shares-overcashout.mjs --base-url http://localhost:8080 --api-prefix /v0` | Covers valid quote/sell, dust accounting, rejected over-cashout, and projection-inexecutable sell error details |
| Sell shares two-share backend cap runner | `node integrationtest/scripts/sell-shares-two-share-cap.mjs --base-url http://localhost:8080 --api-prefix /v0` | Covers locked initial buys, different-user unlock, backend sell message, and oversized quote/sell cap |
| Read-only Schemathesis contract smoke | `MAX_EXAMPLES=1 integrationtest/scripts/schemathesis-read.sh` | 4/4 operations passing |
| Grouped-market Schemathesis contract runner | `node integrationtest/scripts/schemathesis-grouped-market.mjs --base-url http://localhost:8080 --api-prefix /v0` | Covers grouped read schemas with a real group ID |
| Backend tests | `JWT_SIGNING_KEY=test-secret-key-for-testing go test ./...` | Passing |
| Frontend build | `npm run build:report` | Passing with existing warnings |

## Quick Start

Run against a local dev stack after `./SocialPredict dev-bootstrap-users`:

```bash
node integrationtest/scripts/multiple-choice-binary-api.mjs \
  --base-url http://localhost:8080 \
  --api-prefix /v0
```

Run the sell-shares over-cashout regression:

```bash
node integrationtest/scripts/sell-shares-overcashout.mjs \
  --base-url http://localhost:8080 \
  --api-prefix /v0
```

Run the focused sell unlock and two-share backend cap regression:

```bash
node integrationtest/scripts/sell-shares-two-share-cap.mjs \
  --base-url http://localhost:8080 \
  --api-prefix /v0
```

Defaults assume seeded users `admin`, `testuser01`, `testuser02`, `testuser03`,
and `testuser04` all use password `password`.

Run read-only OpenAPI fuzz/contract checks with Schemathesis:

```bash
integrationtest/scripts/schemathesis-read.sh
```

Run grouped-market OpenAPI contract checks after the local stack is running and
dev users are bootstrapped:

```bash
node integrationtest/scripts/schemathesis-grouped-market.mjs \
  --base-url http://localhost:8080 \
  --api-prefix /v0
```

This runner seeds real grouped markets, asserts grouped activity freshness and
group `N/A` behavior, writes a temporary OpenAPI schema with the real group ID,
and then runs Schemathesis fuzzing against:

- `/v0/market-groups/{id}`
- `/v0/market-groups/{id}/bets`
- `/v0/market-groups/{id}/positions`
- `/v0/market-groups/{id}/leaderboard`

The grouped runner defaults to `PHASES=fuzzing`, `MAX_EXAMPLES=5`, a fixed
`SCHEMATHESIS_SEED`, `SCHEMATHESIS_RATE_LIMIT=1/s`, and a
`SCHEMATHESIS_DELAY_MS=5000` cooldown before fuzzing. The fuzzing phase validates
the concrete seeded group paths without coverage-mode path probing, while the
lower rate and cooldown avoid local development rate-limit noise after setup
traffic. The direct setup and mutation assertions use authenticated users;
Schemathesis itself runs the grouped read contract unauthenticated by default
because these grouped read routes are public. It writes its own `summary.json`
artifact; Schemathesis-native reports are opt-in via `REPORT=junit`.

Grouped runner environment variables:

```bash
BASE_URL=http://localhost:8080
MAX_EXAMPLES=5
SCHEMATHESIS_DELAY_MS=5000
REPORT=junit
PHASES=coverage
```

Read-only `schemathesis-read.sh` environment variables:

```bash
AUTH_TOKEN=<bearer-token-for-authenticated-read-routes>
READ_PATHS='/v0/setup /v0/setup/frontend /v0/stats /v0/market-tags'
```

The script writes ignored Schemathesis artifacts under `integrationtest/artifacts/`.

## Operating Notes

- Prefer testing through public HTTP/API routes when possible.
- Record the branch, date, environment, seed users, and setup values in each report.
- Keep raw outputs out of the repo unless they are small and useful for review.
- Put large or generated outputs in `artifacts/` with a short index file if they need to be preserved.
