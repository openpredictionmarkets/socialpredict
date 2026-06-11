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

The write-only hot-market burst is still useful because it isolates the transaction path. The mixed workload is a separate test because it answers a different question: "can the same host keep accepting trades while ordinary users are also clicking around the website?"

### Test Types

| Test type | Script | Purpose | Suggested rates | Duration | Endpoints covered | Pass signal |
| --- | --- | --- | --- | --- | --- | --- |
| Smoke | `loadtest/k6/smoke.js` | Prove deployment, fixtures, login, and one trade path work before capacity testing. | `3` total iterations | Under `1m` | `/health`, `/readyz`, `/ops/status`, `/v0/markets`, `/v0/markets/{id}`, `/v0/login`, `/v0/bet` | `0` failures, all checks pass |
| Hot-market discovery ladder | `loadtest/k6/hot-market-burst.js` | Find the short-run write ceiling on a clean host. | `100`, `200`, `250`, `300` bets/sec | `1m` each | `/v0/login` during setup, `/v0/bet` during run, probes | no failed bets; no fixture mismatch |
| Hot-market sustained confirmation | `loadtest/k6/hot-market-burst.js` | Confirm the strict clean write rate after the short ladder. | Start at `250` bets/sec; only try `300` if `250` is clean | `5m` | same as hot-market discovery | `0` failed bets, p95 preferably below `1s`, no dropped iterations |
| Existing narrow baseline | `loadtest/k6/baseline.js` | Quick read/write sanity check using current simple browse traffic. | Example: `10` bets/sec + `20` browse/sec | `2m` | `/v0/markets`, `/v0/markets/{id}`, `/v0/bet`, probes | should pass before richer mixed test |
| New site-mix low | `loadtest/k6/site-mix.js` | First realistic "click around while trading" API simulation. | `25` bets/sec + `250` reads/sec | `5m` | cached discovery, market summary, positions, leaderboards, user financial summaries, stats, trades | `0` failed bets; low read failure rate |
| New site-mix medium | `loadtest/k6/site-mix.js` | Midpoint mixed workload after low passes. | `50` bets/sec + `500` reads/sec | `5m` | same as site-mix low | same, with p95 and CPU watched closely |
| New site-mix high | `loadtest/k6/site-mix.js` | Stress mixed workload. | `100` bets/sec + `1000` reads/sec | `5m` | same as site-mix low | likely warning/fail if CPU or Postgres saturates |
| Optional read-only storm | `loadtest/k6/site-mix.js` with writes disabled | Isolate read-model/page pressure without transaction writes. | `500`, `1000`, `2000` reads/sec | `2m` to `5m` | same read categories, no `/v0/bet` | helps separate read bottlenecks from write bottlenecks |

### Endpoint Categories For Site-Mix

The first mixed test should hit API endpoints directly rather than drive a browser. This is not a full React rendering test. It is a backend/API/database load test that approximates users clicking through the site.

For this application, browser static-asset pressure should be small after first load because the site is mostly text, CSS, JavaScript, and emoji rather than large image feeds. Treat cold browser/static traffic as an optional `<=5%` slice unless the goal is specifically testing first-visit asset delivery. The primary capacity signal should come from the API/database paths below.

Prefer public display/read-model endpoints when they exist. Do not intentionally hit raw recomputation endpoints during the normal site-mix test unless the purpose is to measure worst-case cache-miss behavior.

| User action being modeled | Preferred display endpoint family | Cache/read-model posture in current API | Auth? | Default `site-mix` share | Requests/sec at `250` browse/sec | Notes |
| --- | --- | --- | --- | ---: | --- |
| Main markets page | `GET /v0/read/market-discovery/markets?limit=21&offset=0`, `GET /v0/read/market-discovery/markets?status=active&limit=21&offset=0` | Snapshot-backed discovery read model | Public | `25%` | `62.5/sec` | Use `/v0/read/market-discovery/...` for normal page traffic. Avoid raw `/v0/markets` except in control tests. |
| Topic/category pages | `GET /v0/read/market-discovery/{tagSlug}?limit=21&offset=0`, `GET /v0/read/market-discovery/{tagSlug}?status=active&limit=21&offset=0` | Snapshot-backed discovery read model with tag filtering | Public | `15%` | `37.5/sec` | Models `/markets/topic/:slug` and persistent topic navigation. |
| Market detail display widgets | `GET /v0/read/markets/{id}/summary` | Display read-model/accounting freshness path, target `1m` | Public | `30%` | `75/sec` | This is the default market-page/card/pinned-market display path in `site-mix`. |
| Market tabs: positions | `GET /v0/markets/positions/{id}?limit=21&offset=0` | Paginated read-model snapshot when `limit/offset` are present, target `10m` | Public | `10%` | `25/sec` | Always include pagination to stay on the read-model path. |
| Market tabs: leaderboard | `GET /v0/markets/{id}/leaderboard?limit=21&offset=0` | Paginated read-model snapshot, target `10m` | Public | `7%` | `17.5/sec` | Always include pagination to stay on the read-model path. |
| User/profile financial display | `GET /v0/read/users/{username}/financial-summary` | User financial summary read model | Logged-in | `8%` | `20/sec` | Uses pre-authenticated users so logged-in-only visibility is represented without raw portfolio recomputation. |
| Stats pages | `GET /v0/system/metrics`, `GET /v0/global/leaderboard?limit=21&offset=0` | Cached reporting read models: system metrics target `1h`, global leaderboard target `15m` | Public or logged-in depending CMS visibility | `5%` | `12.5/sec` | These are now wired to cached read models. |
| Optional cold browser/static/config overhead | frontend route HTML/JS/CSS and low-frequency CMS/config reads | Browser cache should make repeat traffic cheap | Public | `0-5%` | not in default `site-mix` | Only include if testing first-load behavior. It should not dominate SocialPredict capacity planning right now. |
| Trade execution | `POST /v0/bet` | Authoritative transaction path, never cached | Logged-in | separate write rate | This remains the main write-pressure signal. |

The normal read mix should total about `10` reads per `1` trade. Static/browser overhead can be modeled as part of the app shell/config slice or added separately at `<=5%` if we want a cold-visitor test.

The default `site-mix` intentionally does not include the raw full market detail endpoint (`/v0/markets/{id}`) or the live bets table endpoint (`/v0/markets/bets/{id}`). Those remain useful for separate control or worst-case tests, but the default mixed browsing run should measure the intended cached/read-model display paths.

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

The existing scripts are sufficient for:

- deployment smoke testing
- hot-market write-path capacity
- narrow list/detail browsing

The mixed workload is now covered by:

```text
loadtest/k6/site-mix.js
```

Scenario behavior:

- one `bet` scenario at the chosen write rate
- one `browse` scenario at roughly `10x` the write rate
- setup-time pre-auth for users where private/profile endpoints are included
- random mix of public and logged-in reads
- cached/read-model endpoints are preferred where available

If an endpoint is not stable or not public, exclude it from the first version rather than blocking the test.

The helper functions should live in `loadtest/k6/lib/common.js` so the same endpoint reads can be reused by `baseline`, `soak`, and future scenarios.

### Mixed Workload Ladder

Start conservative:

```bash
./loadtest/cli/loadtest run site-mix \
  --base-url "http://$LOADTEST_DROPLET_IP" \
  --api-prefix /api \
  --duration 5m \
  --bet-rate 25 \
  --browse-rate 250 \
  --preauth-users 2000 \
  --setup-timeout 10m \
  --monitor-env loadtest-basic-amd \
  --monitor-host root@"$LOADTEST_DROPLET_IP" \
  --monitor-key ~/.keys/socialpredict/loadtest/id_ed25519 \
  --monitor-interval 5
```

Then:

```bash
./loadtest/cli/loadtest run site-mix \
  --base-url "http://$LOADTEST_DROPLET_IP" \
  --api-prefix /api \
  --duration 5m \
  --bet-rate 50 \
  --browse-rate 500 \
  --preauth-users 5000 \
  --setup-timeout 10m \
  --monitor-env loadtest-basic-amd \
  --monitor-host root@"$LOADTEST_DROPLET_IP" \
  --monitor-key ~/.keys/socialpredict/loadtest/id_ed25519 \
  --monitor-interval 5

./loadtest/cli/loadtest run site-mix \
  --base-url "http://$LOADTEST_DROPLET_IP" \
  --api-prefix /api \
  --duration 5m \
  --bet-rate 100 \
  --browse-rate 1000 \
  --preauth-users 10000 \
  --setup-timeout 15m \
  --monitor-env loadtest-basic-amd \
  --monitor-host root@"$LOADTEST_DROPLET_IP" \
  --monitor-key ~/.keys/socialpredict/loadtest/id_ed25519 \
  --monitor-interval 5
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

- [x] Add `loadtest/k6/site-mix.js` or expand `baseline.js` to support a 10:1 read/write mix.
- [x] Confirm exact endpoint paths for topic discovery, paginated bets, paginated positions, market leaderboard, global leaderboard, and user financial read models.
- [x] Use pre-authenticated users for private/logged-in read paths in `site-mix.js`.
- [x] Add CLI examples to `TEMP_DROPLET_RUNBOOK.md` after the mixed scenario exists.
- [ ] Add dossier table fields for read throughput, write throughput, and combined HTTP request rate.
