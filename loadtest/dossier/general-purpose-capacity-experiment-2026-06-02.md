# SocialPredict Large Basic AMD Droplet Capacity Experiment

Date opened: 2026-06-02

Status: first Basic AMD `8 vCPU / 32 GiB RAM` discovery and degradation analysis captured; fresh-reset confirmation still in progress

Environment: `temporary-loadtest`

Base URL: `http://161.35.177.38`

Host: `socialpredict-loadtest-20260601`

Droplet ID: `574624493`

## Research Question

What hot-market betting throughput can SocialPredict sustain on a larger single-node DigitalOcean Basic AMD shared-CPU Droplet when app, Traefik, nginx, frontend, backend, and Postgres remain colocated in Docker?

## Hypothesis

A Basic AMD shared-CPU `8 vCPU / 32 GiB RAM` Droplet should sustain materially higher concentrated hot-market betting throughput than the previous Basic `1 vCPU / 1 GiB RAM` staging result. Because this is a shared-CPU class, the result should be reported as observed Basic/shared-CPU evidence rather than dedicated General Purpose evidence.

The prior staging dossier identified `50` bets/second as the cleanest supported pre-authenticated hot-market datapoint on the Basic `1 vCPU / 1 GiB` host, while `75` bets/second showed one-minute warning signs and five-minute degradation.

Expected outcome before testing:

- `100` bets/second should pass cleanly.
- `200-300` bets/second may be plausible but must be proven.
- `500` bets/second is an ambitious event-window target, not assumed capacity.

## Experimental Setup

### System Under Test

- Provider: DigitalOcean
- Region: `nyc3`
- Droplet name: `socialpredict-loadtest-20260601`
- Droplet ID: `574624493`
- Public IPv4: `161.35.177.38`
- Public edge mode: HTTP-only raw IP
- Base URL: `http://161.35.177.38`
- Application environment: production topology with load-test rate limits
- App checkout path: `/opt/socialpredict`
- Database: local Docker Postgres
- Runtime topology: single Droplet running Traefik, nginx, frontend, backend, backend startup writer, and Postgres

### Current Host State Before Resize

Observed before this experiment:

- Size slug: `s-1vcpu-1gb`
- Memory: `1024 MiB`
- vCPUs: `1`
- Disk reported by DigitalOcean: `25 GiB`
- OS: Ubuntu 24.04 LTS
- Docker: installed by cloud-init
- Explicit Docker container CPU limits: none observed
- Explicit Docker container memory limits: none observed

### Target Host State

Observed target after resize:

- Size slug: `s-8vcpu-32gb-amd`
- Class: DigitalOcean Basic AMD shared CPU
- Memory: `32768 MiB`
- vCPUs: `8`
- Plan disk if resized with disk: `400 GiB`
- Resize mode used: CPU/RAM-only resize, without `--resize-disk`
- Root disk after resize: remained approximately `25 GiB`
- Price observed from `doctl` on 2026-06-02: `$0.250000/hr`, `$168/mo`
- Approximate 48-hour test cost at target size: `$12.00`

Resize command:

```bash
doctl compute droplet-action resize 574624493 --size s-8vcpu-32gb-amd --wait
```

Do not pass `--resize-disk`. DigitalOcean CPU/RAM-only resize is reversible; disk increases are not.

### Deployment Source

- SocialPredict repository: `openpredictionmarkets/socialpredict`
- Expected ref: `main`
- Ansible repository: `openpredictionmarkets/ansible_playbooks`
- Temporary load-test workflow: `deploy_loadtest.yml`
- Image source: GitHub Container Registry images
- Backend image: `ghcr.io/openpredictionmarkets/socialpredict-backend:latest`
- Frontend image: `ghcr.io/openpredictionmarkets/socialpredict-frontend:latest`

If containers do not recover after resize, rerun:

```bash
gh workflow run deploy_loadtest.yml \
  --repo openpredictionmarkets/ansible_playbooks \
  -f socialpredict_ref=main \
  -f tls_mode=http \
  -f domain_or_ip=161.35.177.38
```

### Interpretation Caveat

The target host is a Basic AMD shared-CPU Droplet. A successful run can prove that this observed host, during this test window, sustained the measured traffic. It should not be presented as dedicated-CPU or General Purpose evidence. If the final result is used for production planning, the dossier should explicitly state that CPU scheduling may be less predictable than a dedicated General Purpose or CPU-Optimized class.

## Constants

These should remain fixed throughout this experiment unless a deviation is recorded.

| Constant | Value |
| --- | --- |
| Base URL | `http://161.35.177.38` |
| API prefix | `/api` |
| Scenario | `hot-market-burst` |
| Load generator | macOS local k6 |
| k6 scenario executor | `constant-arrival-rate` |
| Betting endpoint | `POST /api/v0/bet` |
| Authentication pattern | setup-time pre-authentication, then bearer-token reuse |
| Bet amount | default k6 `BET_AMOUNT`, currently `1` unless overridden |
| Outcome selection | randomized `YES`/`NO` |
| Market selection | hot markets only |
| Monitoring interval | `5s` |
| Host telemetry | CPU, RAM, disk, disk IO, network IO, Docker aggregate, backend/Postgres/Traefik CPU |
| Pass preference | zero failed bets |
| Final proof duration | `5m` |

## Independent Variables

These are intentionally varied.

| Variable | Planned Values |
| --- | --- |
| Droplet size | Current `s-1vcpu-1gb`, target `s-8vcpu-32gb-amd` |
| Target betting rate | Discovery ladder: `100`, `150`, `200`, `250`, `300`, `400`, `500` bets/sec |
| Confirmation target rate | Highest clean discovery rate, then adjusted downward if needed |
| Confirmation duration | `5m` |
| Pre-authenticated user pool | `2000` for one-minute discovery, `5000` for final confirmation |

## Dependent Variables

These are measured outcomes.

| Measurement | Source |
| --- | --- |
| Successful bets/sec | k6 summary, `sp_bets_succeeded` |
| Failed bets | k6 summary, `sp_bets_failed` |
| HTTP failure rate | k6 summary, `http_req_failed` |
| HTTP p50/p95/max latency | k6 summary |
| Dropped iterations | k6 summary |
| Host CPU user/system/idle | host telemetry CSV/summary |
| Available RAM | host telemetry CSV/summary |
| Disk used and disk IO | host telemetry CSV/summary |
| Network RX/TX | host telemetry CSV/summary |
| Docker CPU/RAM aggregate | host telemetry CSV/summary |
| Backend/Postgres/Traefik CPU slices | host telemetry CSV/summary |
| DB pool status | `/api/ops/status` before/after or during spot checks |

## Controlled Variables

These are held constant to isolate the host-size and target-rate effects.

- Same temporary raw-IP host identity and IP.
- Same single-node Docker topology.
- Same local Docker Postgres topology.
- Same HTTP-only edge mode.
- Same k6 scenario implementation.
- Same fixture prefixes and password policy.
- Same load-test rate-limit profile.
- Same local Mac load-generator unless a later deviation is recorded.
- Same Basic/shared-CPU caveat recorded in the final interpretation.

## Rate-Limit Configuration

Expected load-test profile:

```env
RATE_LIMIT_LOGIN_RATE_PER_SECOND=100
RATE_LIMIT_LOGIN_BURST=200
RATE_LIMIT_GENERAL_RATE_PER_SECOND=1000
RATE_LIMIT_GENERAL_BURST=2000
RATE_LIMIT_CLEANUP_INTERVAL=5m
```

These limits are intentionally permissive for single-source capacity testing. They are not model-office or production recommendations.

Normal/model-office comparison profile:

```env
RATE_LIMIT_LOGIN_RATE_PER_SECOND=0.1
RATE_LIMIT_LOGIN_BURST=3
RATE_LIMIT_GENERAL_RATE_PER_SECOND=1
RATE_LIMIT_GENERAL_BURST=10
RATE_LIMIT_CLEANUP_INTERVAL=5m
```

## Fixture Plan

Planned seed:

- Regular users: `10000`
- Moderators: `100`
- Markets: `1000`
- Hot markets: `10`
- `must_change_password`: `false`
- User balance: load-test seed default unless overridden

Seed and pull commands:

```bash
./loadtest/cli/loadtest fixtures seed loadtest \
  --host root@161.35.177.38 \
  --key ~/.keys/socialpredict/loadtest/id_ed25519 \
  --repo-path /opt/socialpredict \
  --users 10000 \
  --moderators 100 \
  --markets 1000 \
  --hot-markets 10 \
  --reset

./loadtest/cli/loadtest fixtures pull loadtest \
  --host root@161.35.177.38 \
  --key ~/.keys/socialpredict/loadtest/id_ed25519 \
  --remote-path /opt/socialpredict/loadtest/fixtures \
  --local-path loadtest/fixtures
```

## User-Load Interpretation

Bet rate is translated into active-user equivalents with:

```text
active_users = bets_per_second * seconds_between_bets_per_user
```

For `500` bets/sec:

| Assumed user behavior | Active-user equivalent |
| --- | ---: |
| `1` bet/sec/user | `500` users |
| `1` bet every `10s`/user | `5000` users |
| `1` bet every `25s`/user | `12500` users |
| `1` bet every `60s`/user | `30000` users |

If `25%` of a `50000` user platform is active in hot markets, that is `12500` active users. `500` bets/sec corresponds to those `12500` active users averaging one bet every `25s`.

## Procedure

### 1. Resize And Verify Host

```bash
doctl compute droplet-action resize 574624493 --size s-8vcpu-32gb-amd --wait

doctl compute droplet get 574624493 \
  --format ID,Name,PublicIPv4,Status,Memory,VCPUs,Disk,Region,Tags

ssh -i ~/.keys/socialpredict/loadtest/id_ed25519 root@161.35.177.38 '
  nproc
  free -h
  df -h /
  docker ps --format "table {{.Names}}\t{{.Status}}"
'

curl -fsS http://161.35.177.38/readyz
```

### 2. Confirm Runtime Configuration

```bash
ssh -i ~/.keys/socialpredict/loadtest/id_ed25519 root@161.35.177.38 '
  cd /opt/socialpredict
  grep -E "^(DOMAIN_URL|API_URL|TLS_MODE|RATE_LIMIT_)" .env
'
```

### 3. Seed And Pull Fixtures

Use the fixture commands in the Fixture Plan section.

### 4. Run Smoke Test

```bash
./loadtest/cli/loadtest run smoke \
  --base-url http://161.35.177.38 \
  --api-prefix /api
```

Smoke must pass before any capacity run.

### 5. Run One-Minute Discovery Ladder

```bash
for rate in 100 150 200 250 300 400 500; do
  ./loadtest/cli/loadtest run hot-market-burst \
    --base-url http://161.35.177.38 \
    --api-prefix /api \
    --duration 1m \
    --target-rate "$rate" \
    --preauth-users 2000 \
    --setup-timeout 5m \
    --monitor-env loadtest-basic-amd \
    --monitor-host root@161.35.177.38 \
    --monitor-key ~/.keys/socialpredict/loadtest/id_ed25519 \
    --monitor-interval 5
done
```

If the first failure/degradation point appears, bisect around the boundary in `25` or `50` bets/sec increments.

### 6. Run Five-Minute Confirmation

```bash
./loadtest/cli/loadtest run hot-market-burst \
  --base-url http://161.35.177.38 \
  --api-prefix /api \
  --duration 5m \
  --target-rate <highest-clean-rate> \
  --preauth-users 5000 \
  --setup-timeout 10m \
  --monitor-env loadtest-basic-amd \
  --monitor-host root@161.35.177.38 \
  --monitor-key ~/.keys/socialpredict/loadtest/id_ed25519 \
  --monitor-interval 5
```

## Pass / Degraded / Fail Criteria

### Pass

- `0` failed bets preferred.
- `http_req_failed` is effectively `0`.
- No dropped iterations in final confirmation run.
- HTTP p95 remains below `1s`.
- Host CPU idle is not pinned near `0%` for most samples.
- Available RAM remains comfortably above `500 MiB`.
- No obvious sustained Postgres or proxy saturation.

### Degraded

- No failed bets, but HTTP p95 exceeds `1s`.
- Dropped iterations appear during one-minute discovery.
- Host CPU or memory shows concerning but not catastrophic pressure.

### Fail

- Any sustained failed bets.
- HTTP failure rate above noise.
- Repeated rate-limit failures.
- Dropped iterations in the five-minute confirmation run.
- Test aborts or the site becomes unreachable.
- Host telemetry shows severe CPU/RAM exhaustion.

## Planned Analysis

After each meaningful run:

1. Pair the k6 summary with the host telemetry summary.
2. Generate dossier JSON using `.context/loadtest/metadata-basic-amd-loadtest.json`.
3. Classify the run as `pass`, `degraded`, or `fail`.
4. Plot or tabulate:
   - target bets/sec
   - successful bets/sec
   - failed bets
   - dropped iterations
   - HTTP p95/max
   - max Docker CPU
   - backend/Postgres/Traefik CPU
   - minimum available RAM
5. Compare the clean five-minute rate to the Basic `1 vCPU / 1 GiB` staging dossier.
6. Translate the clean rate into active-user equivalents using the user-load formula.

Dossier generation command shape:

```bash
./loadtest/cli/loadtest dossier \
  --summary loadtest/results/<summary>.json \
  --host-summary loadtest/hostops/<run>-host-summary.json \
  --metadata .context/loadtest/metadata-basic-amd-loadtest.json \
  --decision <pass|degraded|fail> \
  --out loadtest/dossier/runs/<run>-g4.json
```

## Data

Initial data collection was performed against the resized Basic AMD host. The one-minute runs show clean burst behavior through `300` bets/sec. The five-minute runs show sustained degradation once the hot-market bet history has grown, with Postgres saturating the host CPU.

The k6 rate columns in the raw summaries include setup/pre-auth time. The table below computes achieved bet throughput from successful bets divided by the scenario duration.

| Run timestamp | Target bets/sec | Duration | Decision | Successful bets | Achieved bets/sec | Failed bets | Dropped iterations | HTTP p95 | Host CPU notes | RAM notes | Artifact paths |
| --- | ---: | --- | --- | ---: | ---: | ---: | ---: | ---: | --- | --- | --- |
| `20260602T033940Z` | `100` | `1m` | pass | `6001` | `100.0` | `0` | `0` | `39.15ms` | min idle `87.58%`; Docker CPU `94.42%`; Postgres `35.39%` | min available `31238 MiB` | `loadtest/results/hot-market-burst-20260602T033940Z-summary.json`; `loadtest/hostops/hot-market-burst-loadtest-basic-amd-20260602T033940Z-host-summary.json`; raw CSV same prefix |
| `20260602T034343Z` | `200` | `1m` | pass | `12001` | `200.0` | `0` | `0` | `40.76ms` | min idle `71.45%`; Docker CPU `218.38%`; Postgres `108.75%` | min available `31235 MiB` | `loadtest/results/hot-market-burst-20260602T034343Z-summary.json`; `loadtest/hostops/hot-market-burst-loadtest-basic-amd-20260602T034343Z-host-summary.json`; raw CSV same prefix |
| `20260602T034654Z` | `300` | `1m` | pass | `18001` | `300.0` | `0` | `0` | `72.65ms` | min idle `6.68%`; Docker CPU `633.07%`; Postgres `272.4%` | min available `31207 MiB` | `loadtest/results/hot-market-burst-20260602T034654Z-summary.json`; `loadtest/hostops/hot-market-burst-loadtest-basic-amd-20260602T034654Z-host-summary.json`; raw CSV same prefix |
| `20260602T035651Z` | `350` | `1m` | degraded | `20583` | `343.1` | `0` | `418` | `1.3s` | min idle `0%`; Docker CPU `783.4%`; Postgres `632.09%` | min available `31040 MiB` | `loadtest/results/hot-market-burst-20260602T035651Z-summary.json`; `loadtest/hostops/hot-market-burst-loadtest-basic-amd-20260602T035651Z-host-summary.json`; raw CSV same prefix |
| `20260602T035223Z` | `400` | `1m` | degraded | `23143` | `385.7` | `0` | `858` | `2.3s` | min idle `0%`; Docker CPU `786.93%`; Postgres `540.31%` | min available `31020 MiB` | `loadtest/results/hot-market-burst-20260602T035223Z-summary.json`; `loadtest/hostops/hot-market-burst-loadtest-basic-amd-20260602T035223Z-host-summary.json`; raw CSV same prefix |
| `20260602T040136Z` | `300` | `5m` | degraded | `85511` | `285.0` | `0` | `4490` | `4.96s` | min idle `0%`; Docker CPU `790.59%`; Postgres `664.07%` | min available `30753 MiB` | `loadtest/results/hot-market-burst-20260602T040136Z-summary.json`; `loadtest/hostops/hot-market-burst-loadtest-basic-amd-20260602T040136Z-host-summary.json`; raw CSV same prefix |
| `20260602T041807Z` | `200` | `5m` | degraded | `57542` | `191.8` | `0` | `2459` | `7.66s` | min idle `0%`; Docker CPU `795%`; Postgres `692.12%` | min available `30742 MiB` | `loadtest/results/hot-market-burst-20260602T041807Z-summary.json`; `loadtest/hostops/hot-market-burst-loadtest-basic-amd-20260602T041807Z-host-summary.json`; raw CSV same prefix |
| `20260602T042645Z` | `100` | `5m` | fail | `29952` | `99.8` | `48` | `0` | `88.05ms` expected responses; failed requests timed out at `1m0s` | min idle `2.68%`; Docker CPU `765.03%`; Postgres `696.55%` | min available `30925 MiB` | `loadtest/results/hot-market-burst-20260602T042645Z-summary.json`; `loadtest/hostops/hot-market-burst-loadtest-basic-amd-20260602T042645Z-host-summary.json`; raw CSV same prefix |

### Database Diagnostics After Cumulative Runs

Read-only diagnostics after the degraded `100` bets/sec five-minute run showed:

| Diagnostic | Observation |
| --- | --- |
| Total `bets` rows | `258785` |
| Hot-market distribution | ten hot markets each had approximately `25603-26141` bets |
| `bets` table total size | `35 MB` |
| `bets` indexes observed | `bets_pkey`, `idx_bets_deleted_at` |
| Missing write-path index | no composite index on `(market_id, username)` |
| Autovacuum/analyze | `bets` had recent autovacuum and autoanalyze |
| Active DB state after settling | no persistent active backlog observed |

The placement path calls `UserHasBet` on every bet:

```sql
WHERE market_id = ? AND username = ?
```

Without a composite index on that predicate, the cost of each bet can grow as the `bets` table grows. This is the leading explanation for why the host could pass one-minute bursts on a fresher dataset but later timed out at only `100` bets/sec after the cumulative bet table reached roughly `259k` rows.

## Analysis

The first one-minute ladder established a clear burst profile:

- `100` and `200` bets/sec were clean with low latency and substantial CPU headroom.
- `300` bets/sec was still a clean functional pass, but CPU idle dropped to `6.68%`, so it was already near the top of the clean burst envelope.
- `350` and `400` bets/sec remained correct at the application level but degraded through dropped iterations, multi-second p95 latency, and host CPU saturation.

The five-minute confirmation runs changed the interpretation:

- `300` bets/sec for `5m` degraded: `4490` dropped iterations and HTTP p95 `4.96s`.
- `200` bets/sec for `5m` also degraded: `2459` dropped iterations and HTTP p95 `7.66s`.
- `100` bets/sec for `5m` then produced `48` timeout failures despite low p95 among successful responses.

This pattern is not simply "the host cannot do 100 bets/sec." The earlier `100` and `200` one-minute runs were clean. The later sustained failures happened after cumulative test state had inserted hundreds of thousands of hot-market bets. The database diagnostic points to a data-growth-sensitive write-path query rather than RAM pressure:

- RAM remained plentiful throughout the experiment.
- Postgres CPU repeatedly saturated multiple cores during degraded runs.
- The `bets` table was not huge in disk terms, but the hot predicate lacked an index.
- The hot-market workload concentrates repeated writes and lookups against the same small set of markets, making it intentionally adversarial.

The current evidence supports two separate claims:

1. On this resized Basic AMD shared-CPU Droplet, fresh-ish one-minute burst capacity is clean up to `300` bets/sec and degraded by `350-400` bets/sec.
2. Sustained capacity cannot be claimed yet because cumulative bet-history growth degrades the write path. The next engineering step is to add and test the missing bet-history indexes, then rerun the sustained ladder from a reset fixture state.

## Conclusion

Preliminary conclusion:

- Do not claim `500` bets/sec on this host.
- Do not claim sustained `300` or `200` bets/sec on the current schema.
- The best current burst claim is `300` bets/sec for `1m` on a reset or low-history dataset.
- Sustained claims should wait for an indexed bet-history migration and fresh five-minute confirmation runs.

Recommended schema investigation:

```sql
CREATE INDEX CONCURRENTLY idx_bets_market_id_username ON bets (market_id, username);
```

Likely follow-up index for market-history reads:

```sql
CREATE INDEX CONCURRENTLY idx_bets_market_id_placed_at_id ON bets (market_id, placed_at, id);
```

The first index directly targets `UserHasBet`, which runs during bet placement. The second targets market-scoped ordered bet reads observed in analytics and position paths. These should be implemented through the project migration system before using them as production evidence.

## Deviations

Record deviations here as they happen.

| Time UTC | Deviation | Reason | Impact |
| --- | --- | --- | --- |
| `2026-06-02T03:28Z` | Initial `100` bets/sec run used `--preauth-users 2000` without enough k6 setup timeout | k6 default setup timeout is `60s` | Run timed out during setup and was not counted as capacity evidence |
| `2026-06-02T03:37Z` | First post-timeout rerun captured insufficient host telemetry | Monitor duration did not include setup/pre-auth time | k6 result was useful, but host telemetry did not cover the full burst window |
| `2026-06-02T03:39Z` | Host monitor duration was corrected to include setup timeout | Needed telemetry covering setup plus scenario execution | Later runs have usable host telemetry summaries |
| `2026-06-02T04:26Z` | Later sustained runs were cumulative-state tests, not fresh-reset tests | Prior runs inserted hundreds of thousands of hot-market bets | Sustained degradation may reflect schema/data-growth behavior, not only raw host capacity |
