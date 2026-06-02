# SocialPredict Large Basic AMD Droplet Capacity Experiment

Date opened: 2026-06-02

Status: experimental setup drafted; data collection pending

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

Planned target:

- Size slug: `s-8vcpu-32gb-amd`
- Class: DigitalOcean Basic AMD shared CPU
- Memory: `32768 MiB`
- vCPUs: `8`
- Plan disk: `400 GiB`
- Resize mode: CPU/RAM-only resize, without `--resize-disk`
- Expected root disk after resize: remains approximately `25 GiB`
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
    --monitor-env loadtest-g4 \
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
  --monitor-env loadtest-g4 \
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
2. Generate dossier JSON using `.context/loadtest/metadata-g4-loadtest.json`.
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
  --metadata .context/loadtest/metadata-g4-loadtest.json \
  --decision <pass|degraded|fail> \
  --out loadtest/dossier/runs/<run>-g4.json
```

## Data

Data collection pending.

| Run | Target bets/sec | Duration | Decision | Successful bets/sec | Failed bets | Dropped iterations | HTTP p95 | Host CPU notes | RAM notes |
| --- | ---: | --- | --- | ---: | ---: | ---: | ---: | --- | --- |
| TBD | TBD | TBD | TBD | TBD | TBD | TBD | TBD | TBD | TBD |

## Analysis

Analysis pending data collection.

## Conclusion

Conclusion pending data collection.

## Deviations

Record deviations here as they happen.

| Time UTC | Deviation | Reason | Impact |
| --- | --- | --- | --- |
| TBD | TBD | TBD | TBD |
