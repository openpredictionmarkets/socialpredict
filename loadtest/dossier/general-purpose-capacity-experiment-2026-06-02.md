# SocialPredict Large Basic AMD Droplet Capacity Experiment

Date opened: 2026-06-02

Status: first Basic AMD `8 vCPU / 32 GiB RAM` discovery and degradation analysis captured; June 3 repeat run added with fresh-zero-bet precondition and Postgres CPU saturation finding

Environment: `temporary-loadtest`

Base URL: `http://161.35.177.38`

Host: `socialpredict-loadtest-20260601`

Droplet ID: `574624493`

Final host lifecycle: destroyed on 2026-06-02 after the initial experiment window. Repeating this experiment requires creating a new temporary Droplet, updating the load-test workflow target IP, redeploying, reseeding fixtures, and pulling fresh fixtures.

Repeat host lifecycle: a second temporary load-test Droplet was created for the June 3 repeat run.

- Droplet ID: `574903216`
- Droplet name: `socialpredict-loadtest-20260602-220232`
- Public IPv4: `161.35.135.167`
- Size slug after resize: `s-8vcpu-32gb-amd`
- CPU model observed: `DO-Premium-AMD`
- vCPUs observed: `8`
- RAM observed: `32095 MiB`
- Root disk observed: approximately `25 GiB`
- Base URL: `http://161.35.135.167`
- Backend image revision verified before repeat: `c42bc34900659bcb5a7cbd251fd7ff021920d0f7`

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

If the temporary Droplet has been destroyed, create a new load-test Droplet and run `deploy_loadtest.yml` with the new raw IP before repeating the test. The `domain_or_ip`, `--base-url`, `--monitor-host`, fixture seed/pull host, and local known-hosts entries must all be updated to the new IP.

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
| `20260602T044554Z` | `300` | `1m` | near-clean post-reset probe | `17997` | `300.0` | `0` | `4` | `51.36ms` | min idle `30.58%`; Docker CPU `437.14%`; Postgres `134.43%` | min available `31179 MiB` | `loadtest/results/hot-market-burst-20260602T044554Z-summary.json`; `loadtest/hostops/hot-market-burst-loadtest-basic-amd-20260602T044554Z-host-summary.json`; raw CSV same prefix |
| `20260602T045057Z` | `300` | intended `5m` | invalid fixture mismatch | `0` valid bets | N/A | `1594` `MARKET_NOT_FOUND` responses | N/A | `39.35ms` | host mostly idle; min idle `96.28%`; Postgres `1.73%` | min available `31151 MiB` | `loadtest/results/hot-market-burst-20260602T045057Z-summary.json`; `loadtest/hostops/hot-market-burst-loadtest-basic-amd-20260602T045057Z-host-summary.json`; raw CSV same prefix |

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

### Post-Reset Probe

After a fixture reset and pull, a `300` bets/sec one-minute probe improved materially:

- p95 HTTP latency dropped to `51.36ms`.
- Postgres max CPU dropped to `134.43%`.
- Host min idle stayed at `30.58%`.
- There were `4` dropped iterations, so this is recorded as near-clean rather than strict-clean.

This supports the data-growth hypothesis: resetting the test state reduced Postgres pressure substantially. However, a following attempted five-minute `300` bets/sec run was invalid because k6 received `MARKET_NOT_FOUND` for market IDs such as `1023` and `1026` while the host was mostly idle. That run indicates a fixture mismatch between local `markets.csv` and the active server database, not application capacity degradation.

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
3. Fixture integrity is now a required precondition for long runs. A smoke test alone is not enough if the local hot-market fixture file can drift from the active server database.

## Future Optimization Analysis

The current codebase already has repository boundaries around market and user persistence, so future Postgres optimizations should live behind those repositories and domain services rather than being scattered through handlers. The main candidates are:

1. Add targeted indexes through timestamped migrations.

The most immediate candidate remains:

```sql
CREATE INDEX CONCURRENTLY idx_bets_market_id_username ON bets (market_id, username);
```

This targets the placement-time check for whether a user has already bet in a market. A second candidate is:

```sql
CREATE INDEX CONCURRENTLY idx_bets_market_id_placed_at_id ON bets (market_id, placed_at, id);
```

This targets market-scoped history reads that need deterministic chronological order.

2. Replace full-row GORM loads with narrow query projections where the domain only needs a subset of columns.

Current repository paths such as `ListBetsForMarket` and `loadMarketData` load full bet rows before converting them into boundary/domain structs. That is appropriate for correctness-first implementation, but it can be wasteful under hot-market load. Candidate follow-up work:

- use GORM `Select(...)` for known domain projections
- use custom SQL through the repository for high-traffic read paths
- avoid loading fields that are not consumed by probability, volume, position, or display calculations
- add tests that preserve domain output while changing the persistence query shape

3. Move common aggregates into explicit repository methods.

`CalculateMarketVolume` currently asks the repository for all bets and then sums in Go. For larger histories, the repository can expose a database-backed aggregate such as:

```sql
SELECT COALESCE(SUM(amount), 0) FROM bets WHERE market_id = $1;
```

That keeps the domain service API clean while letting Postgres do the aggregate without transferring every bet row.

4. Introduce market summary state for write-heavy hot paths if indexes and narrow reads are not enough.

If sustained tests still show Postgres CPU saturation after indexing, add transactionally maintained summary tables or columns for:

- current market probability
- market volume
- per-market bet count
- per-user per-market position

This is a larger design change because it shifts some values from computed-on-read to updated-on-write. It should only happen after index and query-shape changes are measured.

5. Revisit the game/probability engine boundary.

The probability engine currently operates over bet history. That is simple and auditable, but hot markets expose the cost of repeatedly replaying history. A future game-engine design can define whether probability is:

- replayed from the event ledger for audit/debug paths
- incrementally updated from the prior market state for hot write paths
- periodically reconciled by a background consistency check

This should be designed at the game-engine boundary rather than as an ad hoc Postgres shortcut.

6. Separate Postgres or use managed Postgres before treating bigger app hosts as the only answer.

The June 3 runs show CPU pressure concentrated in Postgres while RAM remains plentiful. A CPU-optimized or dedicated-CPU Droplet is a useful next hardware comparison, but a production-oriented architecture should also evaluate:

- app/database separation
- managed Postgres sizing
- connection pooling
- Postgres parameter tuning
- backups and recovery posture

The right sequencing is: index migration, query-shape optimization, fresh five-minute load tests, then hardware/architecture comparison.

7. Use Redis or another cache selectively for read-heavy, slightly stale views.

Redis is an in-memory key-value store. In this context, "using Redis to cache Postgres" should not mean putting Redis in front of every database call or treating Redis as the source of truth. It means caching selected read results that are expensive to compute and safe to serve for a short time.

Good candidates:

| Candidate | Cache Shape | Likelihood | Notes |
| --- | --- | --- | --- |
| Global leaderboard | snapshot keyed by leaderboard version/query/page with short TTL | High | The analytics code already has a `GlobalLeaderboardSnapshot` seam, which is an appropriate cache boundary. |
| System financial metrics | snapshot keyed by config/version with short TTL | Medium-High | Useful because these are diagnostic aggregates, not per-request transactional writes. |
| Market summary cards | per-market summary with short TTL or write-through invalidation | Medium | Useful for list pages if cards repeatedly show probability, volume, bet count, and status. |
| Market bet/position page counts | count/summary cache with short TTL | Medium | Can avoid repeated count queries on hot pages; detail rows should still be paginated. |
| Public setup/config frontend policy | in-process or Redis cache with long TTL/version key | Medium | Config is mostly static after install; cache mainly reduces repeated serialization, not DB pressure. |

Poor candidates:

| Candidate | Why Not |
| --- | --- |
| Bet placement correctness | Bets must remain transactional in Postgres. Redis should not decide balances, duplicate-bet checks, or final financial correctness unless there is a much larger event-sourcing design. |
| User balances during writes | Balance changes must be atomic with bet writes. A cache can display a recent balance, but Postgres must own the ledger. |
| Full unpaginated histories | Caching huge payloads can move pressure from Postgres to Redis/network/memory rather than solving the shape problem. Paginate first. |

Cache invalidation policy should be explicit:

- Prefer short TTLs for analytics snapshots.
- Include page/query parameters in cache keys.
- Invalidate or version market summary keys after bet placement, market resolution, or market approval/rejection.
- Keep cache failures non-fatal: if Redis is down, fall back to Postgres and log the miss/error.
- Never let cached financial data become the ledger of record.

Recommended Redis sequencing:

1. Add indexes and pagination first.
2. Add a cache interface behind analytics/repository boundaries.
3. Start with global leaderboard and system stats snapshots.
4. Add market summary caching only after summary contracts are explicit.
5. Load test read-heavy browsing scenarios with cache on/off.

## UX And API Demand-Shaping Checklist

Not every capacity improvement needs to come from a faster database query. Some load can be avoided by changing default UI behavior so expensive data paths are opt-in, paginated, or progressively disclosed. These changes should be treated as product/API design work, not only frontend polish, because the backend endpoints must support the lighter access pattern.

Likelihood estimates below mean "likelihood of reducing avoidable backend/database load during normal browsing." They are not a substitute for measurement.

| Done | Idea | Likelihood | Reason |
| --- | --- | --- | --- |
| [ ] | Paginate market bet history on each market page. | High | `BetsActivity` currently fetches market bets as a full list. Hot markets can accumulate tens of thousands of rows, so pagination directly limits transfer, JSON encoding, sorting, and probability display work. |
| [ ] | Paginate market positions on each market page. | High | `PositionsActivity` currently requests all market positions and filters/sorts client-side. Large markets can make this expensive even when users only inspect the first page. |
| [ ] | Change market activity default tab away from heavy data. Prefer a lightweight `Comments`/overview tab once comments are real, or another cheap summary tab until then. | High | The current activity tab order starts with `Positions`, causing position fetches on market detail load. Defaulting to a cheap tab makes bets/positions user-initiated. |
| [ ] | Lazy-load Bets, Positions, and Leaderboard tab contents only when the tab is opened. | High | Avoids paying for all market-adjacent views when a visitor only wants the market question/trade panel. |
| [ ] | Add "Show bets" / "Show positions" buttons for expensive market detail sections. | High | Explicit disclosure is clearer than hidden automatic fetches and protects hot market pages during traffic spikes. |
| [ ] | Make user financial positions opt-in behind a button on profile/user pages. | Medium-High | Public profile/portfolio/financial views can trigger multi-market position calculations; requiring a click prevents casual profile views from doing heavy accounting work. |
| [ ] | Split simple user financial summary from complex financial statements. | Medium-High | Lightweight balances and totals can load first; high-cost derived financial/accounting views can be requested only by users who need them. |
| [ ] | Paginate public portfolio positions. | Medium-High | User portfolios can grow with market count; pagination caps response size and client rendering cost. |
| [ ] | Reorder Stats tabs to `Setup Configuration`, `System Financial Metrics`, then `Global Leaderboard`. | Medium | `Stats.jsx` currently defaults to Global Leaderboard, though it already requires a button click. Reordering communicates that setup config is the cheap/default view. |
| [ ] | Keep Global Leaderboard behind an explicit "Calculate" button and add backend pagination. | Medium-High | The button already prevents automatic calculation, but once loaded the leaderboard should not require all rows at once. |
| [ ] | Paginate global leaderboard results. | Medium-High | Global leaderboards can touch many users/markets and produce large result sets; pagination bounds response and render size. |
| [ ] | Simplify System Financial Metrics by default and hide complex/derived metrics behind "Show advanced metrics." | Medium | Some system metrics are likely aggregate-heavy. Progressive disclosure prevents expensive diagnostics from being part of the default stats page. |
| [ ] | Cache or snapshot system statistics with a short TTL. | Medium | System stats are read-style diagnostics and may not need exact real-time recomputation for every click. |
| [ ] | Add endpoint-level `limit`, `cursor`/`offset`, and `sort` contracts to OpenAPI for bets, positions, and leaderboards. | High | UI pagination only reduces DB load if the API and repository actually page at the database boundary. |
| [ ] | Return summary counts separately from full detail rows. | Medium | Pages often need "there are N bets/positions" before they need every row. Count/summary endpoints can be cheaper and cacheable. |
| [ ] | Add load-test scenarios for browsing market detail pages with and without optional panels open. | Medium | Current hot-market tests focus on betting. A separate browsing scenario should measure whether pagination/lazy loading reduces read pressure. |

Suggested implementation order:

1. Add backend pagination contracts for market bets, market positions, and global leaderboard.
2. Update market detail tabs so heavy tabs fetch only after user action.
3. Default market detail activity to a cheap tab or placeholder until comments exist.
4. Split profile/user financial views into summary-first and advanced-on-click.
5. Reorder Stats tabs and keep leaderboard/system metrics explicit.
6. Add load-test browsing scenarios to quantify read-path improvement.

## Conclusion

Preliminary conclusion:

- Do not claim `500` bets/sec on this host.
- Do not claim sustained `300` or `200` bets/sec on the current schema.
- The best current burst claim is `300` bets/sec for `1m` on a reset or low-history dataset.
- Sustained claims should wait for an indexed bet-history migration and fresh five-minute confirmation runs.
- Any future five-minute run must first verify that local fixture market IDs exist on the server.

Recommended schema investigation:

```sql
CREATE INDEX CONCURRENTLY idx_bets_market_id_username ON bets (market_id, username);
```

Likely follow-up index for market-history reads:

```sql
CREATE INDEX CONCURRENTLY idx_bets_market_id_placed_at_id ON bets (market_id, placed_at, id);
```

The first index directly targets `UserHasBet`, which runs during bet placement. The second targets market-scoped ordered bet reads observed in analytics and position paths. These should be implemented through the project migration system before using them as production evidence.

Recommended fixture-integrity check before capacity runs:

```bash
ssh -i ~/.keys/socialpredict/loadtest/id_ed25519 root@<LOADTEST_IP> <<'REMOTE'
PG=$(docker ps --format '{{.Names}}' | grep -E 'postgres' | head -1)
docker exec -i "$PG" sh -lc '
DB=${POSTGRES_DB:-${POSTGRES_DATABASE:-postgres}}
U=${POSTGRES_USER:-postgres}
psql -U "$U" -d "$DB" -v ON_ERROR_STOP=1 <<SQL
SELECT min(id) AS min_market_id, max(id) AS max_market_id, count(*) AS markets FROM markets;
SELECT count(*) AS bets FROM bets;
SQL
'
REMOTE

head -20 loadtest/fixtures/markets.csv
tail -20 loadtest/fixtures/markets.csv
```

If local fixture IDs are not within the server market range, rerun fixture seed with `--reset`, then immediately pull fixtures again before running k6.

### June 3 Repeat Run On Fresh Fixtures

The June 3 repeat used a new temporary load-test host at `161.35.135.167`. Before the run, fixtures were reseeded with `--reset`, pulled locally, and verified directly in Postgres:

| Precondition | Value |
| --- | ---: |
| Regular load users | `10000` |
| Load moderators | `100` |
| Markets | `1000` |
| Hot markets | `10` |
| Bets before smoke | `0` |

Smoke then passed against `http://161.35.135.167` with `3/3` smoke bets succeeding. The following sustained run was attempted:

```bash
./loadtest/cli/loadtest run hot-market-burst \
  --base-url http://161.35.135.167 \
  --api-prefix /api \
  --duration 5m \
  --target-rate 300 \
  --preauth-users 100 \
  --setup-timeout 3m \
  --monitor-env loadtest-basic-amd \
  --monitor-host root@161.35.135.167 \
  --monitor-key ~/.keys/socialpredict/loadtest/id_ed25519 \
  --monitor-interval 5
```

Result:

| Run timestamp | Target bets/sec | Duration | Decision | Successful bets | Failed bets | Dropped iterations | HTTP p95 | Host CPU notes | RAM notes | Artifact paths |
| --- | ---: | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- |
| `20260603T034323Z` | `300` | interrupted at `4m30.5s` measured scenario time | fail for strict clean-run standard | `80266` | `288` | `563` | `226.83ms`; max included stalled failed requests | min idle `0.12%`; Docker CPU `775.15%`; backend `134.43%`; Postgres `568.27%`; Traefik `35.73%` | min available `31012 MiB` | `loadtest/results/hot-market-burst-20260603T034323Z-summary.json`; `loadtest/hostops/hot-market-burst-loadtest-basic-amd-20260603T034323Z-host-summary.json`; raw CSV same prefix |
| `20260603T035954Z` | `250` | `5m` | pass | `75001` | `0` | `0` | `153.72ms`; max `550.66ms` | min idle `0.86%`; Docker CPU `767.69%`; backend `166.81%`; Postgres `407.78%`; Traefik `40.69%` | min available `31092 MiB` | `loadtest/results/hot-market-burst-20260603T035954Z-summary.json`; `loadtest/hostops/hot-market-burst-loadtest-basic-amd-20260603T035954Z-host-summary.json`; raw CSV same prefix |

Post-run database state:

| Measurement | Value |
| --- | ---: |
| Total `bets` rows after `300/sec` attempt | `80313` |
| Markets with bets after `300/sec` attempt | `11` |
| Hot-market distribution after `300/sec` attempt | ten hot markets each had approximately `7904-8179` bets |
| Total `bets` rows after fresh `250/sec` confirmation | `75001` |
| Markets with bets after fresh `250/sec` confirmation | `10` |
| Hot-market distribution after fresh `250/sec` confirmation | ten hot markets each had approximately `7382-7642` bets |

Interpretation:

- This run started from a zero-bet load-test state, so it is not merely a cumulative-history artifact.
- The `300/sec` repeat still failed the project's preferred strict standard because there were `288` failed bet requests and `563` dropped iterations.
- The following fresh `250/sec` repeat passed the strict standard: `0` failed bets and `0` dropped iterations for the full `5m` scenario.
- RAM was not stressed. The minimum available RAM stayed above `31 GiB`.
- CPU saturated the host, with Postgres consuming the largest slice.
- This establishes `250/sec for 5m` as the current best clean sustained datapoint on this Basic AMD `8 vCPU / 32 GiB` shared-CPU single-node host.
- This also strengthens the conclusion that the next scaling constraint is CPU-bound Postgres/write-path behavior, not memory.
- A CPU-optimized or dedicated-CPU host is a more relevant next hardware experiment than a RAM-optimized host, but schema/index work should still happen before relying on bigger hardware as the solution.

## Deviations

Record deviations here as they happen.

| Time UTC | Deviation | Reason | Impact |
| --- | --- | --- | --- |
| `2026-06-02T03:28Z` | Initial `100` bets/sec run used `--preauth-users 2000` without enough k6 setup timeout | k6 default setup timeout is `60s` | Run timed out during setup and was not counted as capacity evidence |
| `2026-06-02T03:37Z` | First post-timeout rerun captured insufficient host telemetry | Monitor duration did not include setup/pre-auth time | k6 result was useful, but host telemetry did not cover the full burst window |
| `2026-06-02T03:39Z` | Host monitor duration was corrected to include setup timeout | Needed telemetry covering setup plus scenario execution | Later runs have usable host telemetry summaries |
| `2026-06-02T04:26Z` | Later sustained runs were cumulative-state tests, not fresh-reset tests | Prior runs inserted hundreds of thousands of hot-market bets | Sustained degradation may reflect schema/data-growth behavior, not only raw host capacity |
| `2026-06-02T04:45Z` | Post-reset `300` bets/sec one-minute probe was run | Needed to distinguish cumulative-state degradation from fresh-state host capacity | Improved latency and CPU data strengthened the bet-history growth hypothesis |
| `2026-06-03T03:43Z` | Repeated `300` bets/sec five-minute test on a new host after zero-bet fixture reset | Needed to separate fresh-state host ceiling from cumulative bet-history degradation | Run still failed strict clean criteria with Postgres CPU saturation, indicating a CPU/write-path ceiling even before large cumulative history |
| `2026-06-03T03:59Z` | Repeated `250` bets/sec five-minute test after another zero-bet fixture reset | Needed to find the clean sustained boundary below failed `300/sec` | Run passed cleanly with `75001` successful bets, `0` failed bets, and `0` dropped iterations, but CPU still approached saturation |
| `2026-06-02T04:50Z` | Attempted post-reset `300` bets/sec five-minute run was invalid | k6 used market IDs not found by the server, producing `MARKET_NOT_FOUND` while host was idle | Not capacity evidence; fixture integrity check added as required precondition |
| `2026-06-02T04:56Z` | Temporary load-test Droplet was powered off and destroyed | Avoid ongoing DigitalOcean hourly billing after experiment window | Future testing requires new Droplet/IP and workflow target update |
