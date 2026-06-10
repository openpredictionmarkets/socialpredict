# Next Large-Host Load-Test Plan

Planned date: 2026-06-10

Target environment: temporary DigitalOcean raw-IP load-test Droplet

Primary references:

- `TEMP_DROPLET_RUNBOOK.md`
- `dossier/general-purpose-capacity-experiment-2026-06-02.md`
- `dossier/capacity-forecast-2026-06-02.md`
- `dossier/staging-capacity-2026-06-09.md`

## Goals

1. Recreate the larger temporary-host experiment without resizing `kconfs.com`.
2. Prove that the current `main` build still reaches the same general capacity range as the prior larger-host dossier.
3. Add a mixed website/API workload that runs alongside trading pressure, because ordinary site usage is expected to create substantially more reads than writes.
4. Capture enough host telemetry and k6 output to update the capacity dossier with a direct before/after comparison.

## Hypothesis

The current code should still reproduce the prior larger-host result range on a temporary Basic AMD `8 vCPU / 32 GiB` host:

- `250` hot-market bets/sec for `5m` should remain the strict clean sustained target.
- `300` hot-market bets/sec for `1m` should remain a clean or near-clean burst target.
- `300` hot-market bets/sec for `5m` should remain warning-zone or fail unless recent read-model/cache/index work materially reduced write-path pressure.

The mixed workload may reduce the clean betting ceiling because it adds market-list, market-detail, topic-list, leaderboard, positions, portfolio, and user/profile reads while the write path is active.

## Test Host Plan

Use `TEMP_DROPLET_RUNBOOK.md`.

Recommended sequence:

1. Create a fresh temporary Droplet at `s-1vcpu-1gb`.
2. Attach the shared `port80-access` firewall.
3. Verify SSH, Docker, and Docker Compose.
4. Set/update `LOADTEST_*` secrets in `openpredictionmarkets/ansible_playbooks`.
5. Deploy via `deploy_loadtest.yml` with:

```text
tls_mode=http
domain_or_ip=<temporary_droplet_ip>
socialpredict_ref=main
```

6. Verify raw-IP health:

```bash
curl -fsS "http://$LOADTEST_DROPLET_IP/health"
curl -fsS "http://$LOADTEST_DROPLET_IP/readyz"
curl -fsS "http://$LOADTEST_DROPLET_IP/api/ops/status"
```

7. Seed small fixtures and run smoke.
8. Resize CPU/RAM only to:

```text
s-8vcpu-32gb-amd
```

Do not use `--resize-disk`.

9. Re-verify health after resize.
10. Seed larger fixtures.

## Fixture Plan

Use the same fixture shape as the prior larger-host experiment unless the goal changes:

```bash
./loadtest/cli/loadtest fixtures seed loadtest \
  --host root@"$LOADTEST_DROPLET_IP" \
  --key ~/.keys/socialpredict/loadtest/id_ed25519 \
  --repo-path /opt/socialpredict \
  --users 10000 \
  --moderators 100 \
  --markets 1000 \
  --hot-markets 10 \
  --reset

./loadtest/cli/loadtest fixtures pull loadtest \
  --host root@"$LOADTEST_DROPLET_IP" \
  --key ~/.keys/socialpredict/loadtest/id_ed25519 \
  --remote-path /opt/socialpredict/loadtest/fixtures \
  --local-path loadtest/fixtures
```

Before any capacity run, verify fixture integrity:

```bash
ssh -i ~/.keys/socialpredict/loadtest/id_ed25519 root@"$LOADTEST_DROPLET_IP" <<'REMOTE'
cd /opt/socialpredict || exit 1
set -a && . ./.env && set +a
docker exec socialpredict-postgres-container psql \
  -U "$POSTGRES_USER" \
  -d "$POSTGRES_DATABASE" \
  -v ON_ERROR_STOP=1 <<SQL
SELECT min(id) AS min_market_id, max(id) AS max_market_id, count(*) AS markets FROM markets;
SELECT count(*) AS bets FROM bets;
SQL
REMOTE

head -20 loadtest/fixtures/markets.csv
tail -20 loadtest/fixtures/markets.csv
```

Stop and reseed if local fixture IDs do not match server market IDs.

## Phase 1: Smoke And Control Checks

Run smoke:

```bash
./loadtest/cli/loadtest run smoke \
  --base-url "http://$LOADTEST_DROPLET_IP" \
  --api-prefix /api
```

Capture host profile:

```bash
./loadtest/cli/loadtest host profile loadtest \
  --host root@"$LOADTEST_DROPLET_IP" \
  --key ~/.keys/socialpredict/loadtest/id_ed25519
```

Confirm rate limits:

```bash
./loadtest/cli/loadtest host rate-limits loadtest \
  --host root@"$LOADTEST_DROPLET_IP" \
  --key ~/.keys/socialpredict/loadtest/id_ed25519
```

Expected load-test profile:

```env
RATE_LIMIT_LOGIN_RATE_PER_SECOND=100
RATE_LIMIT_LOGIN_BURST=200
RATE_LIMIT_GENERAL_RATE_PER_SECOND=1000
RATE_LIMIT_GENERAL_BURST=2000
RATE_LIMIT_CLEANUP_INTERVAL=5m
```

## Phase 2: Reproduce Prior Hot-Market Write Path

Run one-minute discovery ladder first:

```bash
for rate in 100 200 250 300; do
  ./loadtest/cli/loadtest run hot-market-burst \
    --base-url "http://$LOADTEST_DROPLET_IP" \
    --api-prefix /api \
    --duration 1m \
    --target-rate "$rate" \
    --preauth-users 2000 \
    --setup-timeout 5m \
    --monitor-env loadtest-basic-amd \
    --monitor-host root@"$LOADTEST_DROPLET_IP" \
    --monitor-key ~/.keys/socialpredict/loadtest/id_ed25519 \
    --monitor-interval 5
done
```

If `300/sec for 1m` is clean, run sustained confirmation:

```bash
./loadtest/cli/loadtest run hot-market-burst \
  --base-url "http://$LOADTEST_DROPLET_IP" \
  --api-prefix /api \
  --duration 5m \
  --target-rate 250 \
  --preauth-users 5000 \
  --setup-timeout 10m \
  --monitor-env loadtest-basic-amd \
  --monitor-host root@"$LOADTEST_DROPLET_IP" \
  --monitor-key ~/.keys/socialpredict/loadtest/id_ed25519 \
  --monitor-interval 5
```

Only attempt `300/sec for 5m` if `250/sec for 5m` is strict-clean and host telemetry has meaningful headroom.

## Phase 3: Mixed Website/API Workload

### Why This Matters

Real users do not only place bets. They browse markets, refresh market pages, inspect positions, visit profile/user pages, view topic pages, and sometimes open statistics/leaderboards. A reasonable first model is at least `10` read actions for every `1` trade execution.

### Browser Simulation Versus API Simulation

There are two possible levels:

| Level | Tool | Use Tomorrow? | Notes |
| --- | --- | --- | --- |
| API-level simulation | k6 HTTP scenario | Yes | Fast, repeatable, already fits the current harness and host telemetry workflow. |
| Browser-level simulation | Playwright or k6 browser | Not first | More realistic frontend rendering and asset behavior, but heavier to set up and harder to interpret. Use later after API mix is stable. |

Tomorrow's practical plan should start with API-level simulation. This tests the backend/API and database pressure caused by browsing behavior. It does not measure browser rendering cost or client-side React performance.

### Current Gap

The current k6 `baseline` and `soak` scenarios read only:

- `/v0/markets`
- `/v0/markets/:id`
- health/readiness/status probes
- plus optional betting

That is not enough to model real website navigation after the recent feature work.

### Scenario To Add Before Or During Tomorrow's Test

Add a new k6 scenario, tentatively:

```text
loadtest/k6/site-mix.js
```

Target behavior:

- one `bet` scenario at the chosen write rate
- one `browse` scenario at roughly `10x` the write rate
- setup-time pre-auth for users where private/profile endpoints are included
- random mix of public and logged-in reads

Recommended first endpoint mix:

| Category | Example endpoint family | Suggested share of read traffic |
| --- | --- | ---: |
| Market discovery | `/v0/read/market-discovery/markets`, `/v0/markets`, `/v0/markets?status=...` | `30%` |
| Market detail | `/v0/markets/:id` and probability/history/detail read paths used by market pages | `25%` |
| Topic pages | `/v0/read/market-discovery/topic/:slug` or equivalent topic route once confirmed | `10%` |
| Market activity | paginated bets, positions, market leaderboard endpoints | `15%` |
| User/profile views | public user, portfolio, owned markets, financial summaries where logged-in access is required | `10%` |
| Stats/read models | global leaderboard/system stats only if visibility and auth permit | `5%` |
| CMS/static config reads | frontend setup/config/content reads | `5%` |

If an endpoint is not stable or not public, exclude it from the first version rather than blocking the test.

### Mixed Workload Ladder

Start conservative:

```text
25 bets/sec + 250 reads/sec for 5m
```

Then:

```text
50 bets/sec + 500 reads/sec for 5m
100 bets/sec + 1000 reads/sec for 5m
```

Only run the higher mixed workloads if the prior step is clean:

- no failed bets
- HTTP failure rate effectively `0`
- no dropped iterations in confirmation runs
- p95 below service target
- host CPU idle not pinned near `0%`
- Postgres CPU not saturating the host

## Phase 4: Dossier Update

Create a new markdown addendum under:

```text
loadtest/dossier/
```

Suggested filename:

```text
loadtest/dossier/large-host-capacity-rerun-2026-06-10.md
```

Include:

- temporary Droplet ID, size slug, IP, region, lifecycle start/end
- backend/frontend image tag or SocialPredict commit deployed
- fixture counts
- rate-limit profile
- raw artifact paths
- hot-market-only reproduction results
- mixed read/write workload results
- comparison against `general-purpose-capacity-experiment-2026-06-02.md`
- whether the new run matches, improves, or regresses from prior evidence
- explicit caveat if the result is Basic/shared-CPU rather than dedicated CPU

## Pass / Warning / Fail Criteria

### Pass

- `0` failed bets
- `0` or near-zero HTTP failures
- no dropped iterations during `5m` confirmation
- HTTP p95 below `1s` for write-path confirmation
- host CPU idle not pinned near `0%` for most samples
- no memory pressure

### Warning

- no failed bets, but p95 exceeds `1s`
- dropped iterations appear
- CPU idle repeatedly hits `0%`
- Postgres CPU dominates the box

### Fail

- any sustained failed bets
- HTTP failures above noise
- site unreachable
- repeated rate-limit failures after load-test rate limits are confirmed
- fixture mismatch such as `MARKET_NOT_FOUND`

## Open Implementation Tasks Before Mixed Test

- [ ] Add `loadtest/k6/site-mix.js` or expand `baseline.js` to support a 10:1 read/write mix.
- [ ] Confirm exact endpoint paths for topic discovery, paginated bets, paginated positions, global leaderboard, and user financial read models.
- [ ] Decide whether private/logged-in reads should use pre-authenticated users from setup.
- [ ] Add CLI examples to `TEMP_DROPLET_RUNBOOK.md` after the mixed scenario exists.
- [ ] Add dossier table fields for read throughput, write throughput, and combined HTTP request rate.
