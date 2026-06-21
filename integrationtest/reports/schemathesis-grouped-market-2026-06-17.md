# Schemathesis Report - Grouped Market Contract Runner

**Date:** 2026-06-17 America/Chicago  
**UTC artifact stamp:** 20260618T025320Z  
**Branch:** `feature/multiple-choice-binary-markets`  
**Base URL:** `http://localhost:8080`  
**Schema:** `backend/docs/openapi.yaml`

## Command

```bash
node integrationtest/scripts/schemathesis-grouped-market.mjs \
  --base-url http://localhost:8080 \
  --api-prefix /v0
```

Prerequisites used for this run:

```bash
docker restart socialpredict-backend-container
./SocialPredict dev-bootstrap-users
```

## Coverage

The runner seeds real grouped markets, asserts the new grouped-market behavior directly, writes a temporary OpenAPI schema with the seeded group ID constrained into path parameters, then runs Schemathesis against grouped read endpoints.

| Area | Result |
| --- | --- |
| Seeded admin/moderator/bettor login | PASS |
| OpenAPI grouped answer cap is `50` | PASS |
| OpenAPI grouped resolve supports `na` mode | PASS |
| Grouped bets freshness metadata | PASS |
| Grouped positions freshness metadata | PASS |
| Grouped leaderboard freshness metadata | PASS |
| Group `N/A` resolves every child `N/A` | PASS |
| Group `N/A` marks parent resolved | PASS |
| Schemathesis grouped read contract | PASS |

## Schemathesis Result

| Metric | Value |
| --- | --- |
| Operations selected | 4 / 96 |
| Operations tested | 4 |
| Test cases generated | 16 |
| Test cases passed | 16 |
| Failures | 0 |
| Warnings | 0 |

Selected paths:

- `/v0/market-groups/{id}`
- `/v0/market-groups/{id}/bets`
- `/v0/market-groups/{id}/positions`
- `/v0/market-groups/{id}/leaderboard`

Artifacts:

- `integrationtest/artifacts/schemathesis-grouped-20260618025320/summary.json`
- `integrationtest/artifacts/schemathesis-grouped-20260618025320/openapi-grouped-25.yaml`

## Warning Status

No Schemathesis warnings were emitted in the final run. The runner now pauses
briefly between authenticated setup/assertions and unauthenticated grouped-read
fuzzing so local rate-limit buckets do not create misleading warning noise.
