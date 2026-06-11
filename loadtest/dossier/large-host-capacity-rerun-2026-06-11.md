# SocialPredict Large-Host Capacity Rerun Notebook

Date opened: 2026-06-11

Status: active experiment; temporary large host is deployed, seeded, smoke-tested, and has one clean hot-market baseline datapoint. The first mixed workload series was run before `v3.4.0`; the next mixed workload series should be labeled as `v3.4.0` after the load-test deploy completes.

Base URL: `http://159.65.37.166`

## Research Question

Can the current release build, including cached reporting/read-model improvements and the `v3.4.0` market-summary read-model fix, reproduce or improve the previous larger-host SocialPredict capacity evidence on a temporary DigitalOcean Basic AMD host?

This rerun also adds a mixed cached-read workload to model ordinary site browsing while hot-market betting is active.

## Host

| Field | Value |
| --- | --- |
| Provider | DigitalOcean |
| Droplet name | `socialpredict-loadtest-20260611-023140` |
| Droplet ID | `576807920` |
| Public IPv4 | `159.65.37.166` |
| Region | `nyc3` |
| Size slug | `s-8vcpu-32gb-amd` |
| CPU model observed | `DO-Premium-AMD` |
| Host CPU count | `8` |
| Host RAM total | `32095 MiB` |
| Root disk observed | `395700 MiB` |
| Root disk used at profile | `4133 MiB` |
| Docker CPU count | `8` |
| Docker RAM total | `32095 MiB` |
| Explicit container CPU limits | `0/6` |
| Explicit container memory limits | `0/6` |

Interpretation caveat: this is a DigitalOcean Basic AMD shared-CPU host. Results should be presented as observed shared-CPU evidence, not dedicated-CPU evidence.

## Deployment

The temporary host was created with Ubuntu 24.04 and Docker cloud-init, then deployed through the `openpredictionmarkets/ansible_playbooks` load-test workflow.

Workflow:

```text
https://github.com/openpredictionmarkets/ansible_playbooks/actions/runs/27319946755
```

Version boundary:

| Time UTC | Workflow / release | Deployed ref | Interpretation |
| --- | --- | --- | --- |
| `2026-06-11T02:33:54Z` | Load-test deploy `27319946755` | `main` before `v3.4.0` | Used for smoke, `100/sec` hot-market baseline, and the first mixed workload attempts below. This is effectively `v3.3.0`-era code plus then-current `main`, but without PR #749's pure market-summary read-model fix. |
| `2026-06-11T03:59:21Z` | SocialPredict release `v3.4.0` | `v3.4.0` | Adds cached market read-model behavior, including `/v0/read/markets/{id}/summary` avoiding synchronous full market accounting refresh. |
| `2026-06-11T04:03:31Z` | Load-test deploy `27322972380` | `v3.4.0` | This deploy was dispatched after the release. Mixed workload results after this deploy should be recorded as the `v3.4.0` series. |

External readiness passed:

```text
GET http://159.65.37.166/health -> live
GET http://159.65.37.166/readyz -> ready
GET http://159.65.37.166/api/ops/status -> ready DB pool
```

## Fixtures

Seed command:

```bash
./loadtest/cli/loadtest fixtures seed loadtest \
  --host root@159.65.37.166 \
  --key ~/.keys/socialpredict/loadtest/id_ed25519 \
  --repo-path /opt/socialpredict \
  --users 10000 \
  --moderators 100 \
  --markets 1000 \
  --hot-markets 10 \
  --user-balance 1000000 \
  --reset
```

Seed output:

| Fixture | Count |
| --- | ---: |
| Regular users | `10000` |
| Moderators | `100` |
| Markets | `1000` |
| Hot markets | `10` |
| Password change required | `false` |

Local fixture pull confirmed `markets.csv` contains IDs `1-1000`, with `1-10` marked `hot`.

## Smoke

Smoke test passed against the raw-IP deployment.

| Metric | Value |
| --- | ---: |
| Bets attempted | `3` |
| Bets succeeded | `3` |
| HTTP failures | `0%` |
| HTTP p95 | `170.79ms` |

Raw artifact:

```text
loadtest/results/smoke-20260611T023734Z-summary.json
```

## Hot-Market Baseline: 100/sec For 1m

Command:

```bash
./loadtest/cli/loadtest run hot-market-burst \
  --base-url http://159.65.37.166 \
  --api-prefix /api \
  --duration 1m \
  --target-rate 100 \
  --preauth-users 2000 \
  --setup-timeout 5m \
  --monitor-env loadtest-basic-amd \
  --monitor-host root@159.65.37.166 \
  --monitor-key ~/.keys/socialpredict/loadtest/id_ed25519 \
  --monitor-interval 5
```

Result:

| Metric | Value |
| --- | ---: |
| Result | Pass |
| Measured scenario duration | `1m` |
| Target betting rate | `100/sec` |
| Bets attempted | `6001` |
| Bets succeeded | `6001` |
| Failed bets | `0` |
| HTTP failures | `0/8004` |
| HTTP p95 | `66.45ms` |
| HTTP max | `178.62ms` |
| Iteration p95 | `80.40ms` |
| Interrupted iterations | `0` |

Important interpretation note: k6 custom metric rates include the setup phase, so `sp_bets_succeeded` reported around `33/sec` over total wall time. The measured scenario itself completed `6001` bets over `60s`, which is effectively `100/sec`.

Host telemetry:

| Metric | Value |
| --- | ---: |
| Samples | `22` |
| Window | `2026-06-11T02:38:53Z` to `2026-06-11T02:41:44Z` |
| Max CPU user | `8.38%` |
| Max CPU system | `8.63%` |
| Min CPU idle | `82.99%` |
| Min RAM available | `31128 MiB` |
| Max RAM used | `967 MiB` |
| Max disk used | `2%` |
| Max disk write | `42784 KiB/s` |
| Max Docker CPU sum | `121.82%` |
| Max Docker RAM sum | `133.5 MiB` |
| Max backend CPU | `66.88%` |
| Max Postgres CPU | `46.77%` |
| Max Traefik CPU | `10.75%` |

Raw artifacts:

```text
loadtest/results/hot-market-burst-20260611T023847Z-summary.json
loadtest/hostops/hot-market-burst-loadtest-basic-amd-20260611T023847Z-host.csv
loadtest/hostops/hot-market-burst-loadtest-basic-amd-20260611T023847Z-host-summary.json
loadtest/hostops/hot-market-burst-loadtest-basic-amd-20260611T023847Z-host-profile.json
```

## Pre-v3.4.0 Mixed Cached-Read Workload Plan

The first mixed workload attempts used `site-mix`, which combines hot-market betting with cached/read-model browsing paths. These runs happened before the `v3.4.0` load-test deploy, so they should not be used as final evidence for the `v3.4.0` cache behavior.

Default read distribution at `250` reads/sec:

| Share | Approx/sec | Endpoint family |
| ---: | ---: | --- |
| `25%` | `62.5/sec` | `/v0/read/market-discovery/markets?...` |
| `15%` | `37.5/sec` | `/v0/read/market-discovery/{tagSlug}?...` |
| `43%` | `107.5/sec` | `/v0/read/markets/{id}/summary` |
| `10%` | `25/sec` | `/v0/markets/positions/{id}?limit=21&offset=0` |
| `7%` | `17.5/sec` | `/v0/markets/{id}/leaderboard?limit=21&offset=0` |

The default mixed run intentionally excludes the raw full market detail route, the live bets table route, global reporting endpoints, and user financial summaries. Those should be measured separately as control or worst-case paths after the harness can warm the relevant snapshots.

Pre-`v3.4.0`, market positions and market leaderboard requests used cached snapshots only when pagination parameters were present, and could still refresh synchronously when stale. Global reporting endpoints were excluded because active betting marked analytics snapshots stale, which could cause synchronous global recomputation on read. The `v3.4.0` release changes the read model behavior so existing stale display snapshots are served without synchronous refresh.

### Aborted Setup Attempt

An initial `site-mix` attempt started at `2026-06-11T02:58:01Z` but did not reach the measured scenarios. k6 was still in `setup()` pre-authenticating `2000` users when one `/api/v0/login` request timed out. The helper then attempted to parse the timeout response's null body as JSON and aborted the script.

This is classified as a load-test harness failure, not an application capacity failure:

| Metric | Value |
| --- | ---: |
| Authenticated logins before abort | `774` |
| Failed login requests | `1` |
| HTTP failure rate during setup | `0.12%` |
| Max host CPU user during setup | `1.24%` |
| Max Postgres CPU during setup | `1.41%` |
| Min CPU idle during setup | `97.64%` |

Raw artifacts:

```text
loadtest/results/site-mix-20260611T025801Z-summary.json
loadtest/hostops/site-mix-loadtest-basic-amd-20260611T025801Z-host.csv
loadtest/hostops/site-mix-loadtest-basic-amd-20260611T025801Z-host-summary.json
loadtest/hostops/site-mix-loadtest-basic-amd-20260611T025801Z-host-profile.json
```

Harness adjustment: login token parsing now tolerates null/timeout bodies, and setup logins retry transient failures instead of aborting the entire run.

### Interrupted Fixture-Shape Attempt

A second `site-mix` attempt started at `2026-06-11T03:09:34Z` and reached the measured scenarios, but it was manually interrupted after repeated `404 NOT_FOUND` responses from `/api/v0/read/users/{username}/financial-summary`.

This is also classified as a load-test fixture-shape issue, not a host-capacity failure. The endpoint requires an existing user financial read-model snapshot. Random load-test users did not have those snapshots precomputed, so the endpoint consistently returned `404`.

Partial run observations before interruption:

| Metric | Value |
| --- | ---: |
| Runtime before interruption | `~2m02s` wall time, measured scenarios about `6s` |
| Bets attempted | `159` |
| Bets succeeded | `152` |
| Site reads attempted | `1549` |
| Site read failures | `117` |
| Failed read family | `user financial read model` |
| HTTP p95 | `804.28ms` |
| Dropped iterations | `33` |
| Max CPU user | `1.61%` |
| Max Postgres CPU | `2.05%` |
| Min CPU idle | `97.03%` |

Raw artifacts:

```text
loadtest/results/site-mix-20260611T030934Z-summary.json
loadtest/hostops/site-mix-loadtest-basic-amd-20260611T030934Z-host.csv
loadtest/hostops/site-mix-loadtest-basic-amd-20260611T030934Z-host-summary.json
loadtest/hostops/site-mix-loadtest-basic-amd-20260611T030934Z-host-profile.json
```

Harness adjustment: user financial summaries were removed from the default `site-mix` distribution. They should be added back only after a snapshot warmup command exists or the endpoint returns a safe zero/default read model for existing users without snapshots.

### Mixed Workload Failure: 25 Bets/sec Plus 250 Reads/sec

A third `site-mix` attempt started at `2026-06-11T03:13:46Z` after user financial summaries were removed from the default mix. This run reached the measured scenario but was manually interrupted after the host saturated and k6 began reporting request timeouts.

Command:

```bash
./loadtest/cli/loadtest run site-mix \
  --base-url http://159.65.37.166 \
  --api-prefix /api \
  --duration 5m \
  --bet-rate 25 \
  --browse-rate 250 \
  --preauth-users 2000 \
  --setup-timeout 10m \
  --monitor-env loadtest-basic-amd \
  --monitor-host root@159.65.37.166 \
  --monitor-key ~/.keys/socialpredict/loadtest/id_ed25519 \
  --monitor-interval 5
```

Observed result:

| Metric | Value |
| --- | ---: |
| Result | Interrupted failure signal |
| Checks passed | `98.46%` |
| HTTP failure rate | `1.84%` |
| Bets attempted | `1530` |
| Bets succeeded | `856` |
| Site reads attempted | `8606` |
| Site reads succeeded | `7417` |
| Site reads failed | `192` |
| HTTP p95 | `10.67s` |
| Iteration p95 | `13.71s` |
| Dropped iterations | `13034` |
| Max active VUs | `1675` |

Host telemetry:

| Metric | Value |
| --- | ---: |
| Max CPU user | `75.84%` |
| Max CPU system | `26.39%` |
| Min CPU idle | `1.34%` |
| Max Docker CPU sum | `775.15%` |
| Max backend CPU | `362.8%` |
| Max Postgres CPU | `400.02%` |
| Min RAM available | `29952 MiB` |

Interpretation:

- This was not RAM-bound.
- The host was effectively CPU-bound across backend/Postgres.
- The reporting slice distorted the default browsing mix because global reporting snapshots can refresh synchronously when stale.
- Global reporting should be tested as a separate control until it is changed to stale-while-revalidate.
- This run should not be used as the final mixed-workload capacity ceiling; rerun a lower mix without reporting endpoints first.

Raw artifacts:

```text
loadtest/results/site-mix-20260611T031346Z-summary.json
loadtest/hostops/site-mix-loadtest-basic-amd-20260611T031346Z-host.csv
loadtest/hostops/site-mix-loadtest-basic-amd-20260611T031346Z-host-summary.json
loadtest/hostops/site-mix-loadtest-basic-amd-20260611T031346Z-host-profile.json
```

Harness adjustment: global reporting endpoints were removed from the default `site-mix` distribution and the next mixed run was reduced to `10` bets/sec plus `50` reads/sec.

Next command:

```bash
./loadtest/cli/loadtest run site-mix \
  --base-url http://159.65.37.166 \
  --api-prefix /api \
  --duration 5m \
  --bet-rate 10 \
  --browse-rate 50 \
  --preauth-users 500 \
  --setup-timeout 5m \
  --monitor-env loadtest-basic-amd \
  --monitor-host root@159.65.37.166 \
  --monitor-key ~/.keys/socialpredict/loadtest/id_ed25519 \
  --monitor-interval 5
```

## Open Questions

- Does `10` bets/sec plus `50` cached reads/sec stay clean for `5m` after reporting endpoints are removed?
- Does cached-read traffic materially affect write-path p95 compared with hot-market-only runs?
- At what mixed workload do Postgres or backend CPU become dominant?
- Should separate control tests be added for raw full market detail and live bets table routes after cached-read runs complete?

### Mixed Workload Pass: 10 Bets/sec Plus 50 Reads/sec

After the higher `25` bets/sec plus `250` reads/sec mixed workload saturated the host, the mixed scenario was reduced to a smaller representative read/write baseline. This run completed cleanly.

Command:

```bash
./loadtest/cli/loadtest run site-mix \
  --base-url http://159.65.37.166 \
  --api-prefix /api \
  --duration 5m \
  --bet-rate 10 \
  --browse-rate 50 \
  --preauth-users 500 \
  --setup-timeout 5m \
  --monitor-env loadtest-basic-amd \
  --monitor-host root@159.65.37.166 \
  --monitor-key ~/.keys/socialpredict/loadtest/id_ed25519 \
  --monitor-interval 5
```

Scenario mix:

| Scenario | Target | Duration | Notes |
| --- | ---: | --- | --- |
| Hot-market bets | `10/sec` | `5m` | Pre-authenticated users, hot-market-only writes. |
| Site reads | `50/sec` | `5m` | Cached/read-model browsing paths only. |

Read distribution for this run:

| Share | Approx/sec at `50` reads/sec | Endpoint family |
| ---: | ---: | --- |
| `25%` | `12.5/sec` | `/v0/read/market-discovery/markets?...` |
| `15%` | `7.5/sec` | `/v0/read/market-discovery/{tagSlug}?...` |
| `43%` | `21.5/sec` | `/v0/read/markets/{id}/summary` |
| `10%` | `5.0/sec` | `/v0/markets/positions/{id}?limit=21&offset=0` |
| `7%` | `3.5/sec` | `/v0/markets/{id}/leaderboard?limit=21&offset=0` |

Result:

| Metric | Value |
| --- | ---: |
| Result | Pass |
| Wall-clock runtime | `5m27.5s` including setup/graceful stop |
| Measured scenario duration | `5m` |
| Bets attempted | `3001` |
| Bets succeeded | `3001` |
| Failed bets | `0` |
| Site reads attempted | `15001` |
| Site reads succeeded | `15001` |
| HTTP requests | `18506` |
| HTTP failures | `0` |
| HTTP request rate | `56.51/sec` |
| HTTP p95 | `183.83ms` |
| HTTP max | `364.47ms` |
| Iteration p95 | `196.95ms` |
| Dropped/interrupted iterations | `0` |
| Max VUs observed | `13` |

Endpoint check counts:

| Check | Passes | Fails |
| --- | ---: | ---: |
| `bet returned 201` | `3001` | `0` |
| `market discovery read model returned expected status` | `6050` | `0` |
| `market summary read model returned expected status` | `6417` | `0` |
| `market positions read model returned expected status` | `1503` | `0` |
| `market leaderboard read model returned expected status` | `1031` | `0` |

Host telemetry:

| Metric | Value |
| --- | ---: |
| Samples | `41` |
| Window | `2026-06-11T03:37:26Z` to `2026-06-11T03:42:53Z` |
| Max CPU user | `25.35%` |
| Max CPU system | `9.89%` |
| Min CPU idle | `65.65%` |
| Min RAM available | `30867 MiB` |
| Max RAM used | `1228 MiB` |
| Max disk used | `2%` |
| Max disk write | `38616 KiB/s` |
| Max network RX | `66.34 KiB/s` |
| Max network TX | `2291.69 KiB/s` |
| Max Docker CPU sum | `298.01%` |
| Max Docker RAM sum | `200.94 MiB` |
| Max backend CPU | `200.21%` |
| Max Postgres CPU | `89.66%` |
| Max Traefik CPU | `11.34%` |

Raw artifacts:

```text
loadtest/results/site-mix-20260611T033720Z-summary.json
loadtest/hostops/site-mix-loadtest-basic-amd-20260611T033720Z-host.csv
loadtest/hostops/site-mix-loadtest-basic-amd-20260611T033720Z-host-summary.json
loadtest/hostops/site-mix-loadtest-basic-amd-20260611T033720Z-host-profile.json
```

Interpretation:

- This is the first clean mixed read/write datapoint in the June 10/11 rerun sequence.
- The host had substantial CPU and RAM headroom at this reduced mix.
- Backend CPU was the largest app-side slice, while Postgres stayed under `90%` container CPU in the observed maximum sample.
- This run is not a maximum-capacity claim. It is a stable mixed baseline after the higher `25` bets/sec plus `250` reads/sec attempt exposed read-path pressure.
- The run was executed before the pure market-summary read-model fix that removes synchronous accounting refresh from `/v0/read/markets/{id}/summary`. After that fix is deployed, rerun this same command first, then ladder upward.

## Current Interpretation After Pre-v3.4.0 June 10/11 Mixed Runs

- Pure hot-market betting still has a stronger clean datapoint from the earlier large-host experiment than the mixed workload has so far.
- The mixed workload is more representative of real site behavior because it combines hot-market writes with market discovery, market summary cards, positions pages, and leaderboards.
- The failed `25` bets/sec plus `250` reads/sec run should not be treated as final capacity evidence until the read-model fixes are deployed, because the market summary endpoint was still capable of synchronous accounting refresh.
- The clean `10` bets/sec plus `50` reads/sec run is the current pre-`v3.4.0` mixed baseline for this host.

Recommended next sequence after the `v3.4.0` load-test deploy:

1. Use a larger bracket, not tiny increments. The pre-`v3.4.0` `10/50` pass had substantial headroom and is too low to estimate the edge.
2. Rerun a meaningful `v3.4.0` baseline at `25` bets/sec plus `250` reads/sec for `5m`.
3. If clean, jump to `50` bets/sec plus `500` reads/sec for `5m`.
4. If that is clean, test `75` bets/sec plus `750` reads/sec for `5m`; if it fails, bisect between the last clean and failed rate.
5. Keep global reporting and user financial summaries out of the default site mix until those paths have explicit warmup/default-snapshot behavior and separate tests.

## v3.4.0 Mixed Workload Edge-Finding Plan

The goal of the `v3.4.0` series is to bracket the mixed read/write edge quickly.

The prior clean mixed run was `10` bets/sec plus `50` reads/sec. Its host profile showed enough headroom that smaller steps are unlikely to be informative. The next series should use larger jumps and then bisect.

| Step | Command target | Purpose | Decision rule |
| --- | ---: | --- | --- |
| A | `25` bets/sec + `250` reads/sec for `5m` | Re-test the prior failed mixed target after `v3.4.0` removes synchronous market-summary refresh. | If clean, move up. If it fails, inspect failure family before assuming host limit. |
| B | `50` bets/sec + `500` reads/sec for `5m` | Find whether the fix creates substantial new mixed-read headroom. | If p95 stays below `1s`, no failed bets, and no dropped iterations, move up. |
| C | `75` bets/sec + `750` reads/sec for `5m` | Stress the host near a likely mixed-workload edge. | If it fails, bisect between B and C. |
| D | `100` bets/sec + `1000` reads/sec for `5m` | Optional only if C is clean and host telemetry still has headroom. | Stop if Postgres/backend CPU pins or request failures appear. |

Pass/fail criteria:

- Pass: zero failed bets, effectively zero HTTP failures, no dropped iterations, p95 below `1s`, and host CPU not pinned near `0%` idle for sustained samples.
- Degraded: no failed bets but dropped iterations, p95 above `1s`, or CPU pinned for sustained samples.
- Fail: any meaningful failed bets, sustained HTTP failures, site unreachability, or k6 abort.

The `v3.4.0` series should be documented separately from the pre-`v3.4.0` runs even though it uses the same Droplet and fixtures.

## v3.4.0 Mixed Workload Results

### Invalid Setup Run: Fixture Mismatch After v3.4.0 Deploy

A first post-deploy attempt started at `2026-06-11T04:07:22Z` using `25` bets/sec plus `250` reads/sec, but it never reached the measured scenarios. The load-test users failed authentication during setup.

| Metric | Value |
| --- | ---: |
| Target | `25` bets/sec + `250` reads/sec |
| Result | Invalid setup run |
| Failed login requests | `8` |
| Login failure reason | `AUTHORIZATION_DENIED` |
| HTTP failure rate | `89.28%` during setup only |
| Host CPU idle | `99.75%` |

Interpretation:

- This is not capacity evidence.
- The server was idle and ready; the failure was fixture/auth state mismatch after redeploy.
- Corrective action was to reseed remote load-test fixtures and pull fresh local `users.csv` / `markets.csv` before rerunning.

Raw artifacts:

```text
loadtest/results/site-mix-20260611T040722Z-summary.json
loadtest/hostops/site-mix-loadtest-basic-amd-20260611T040722Z-host.csv
loadtest/hostops/site-mix-loadtest-basic-amd-20260611T040722Z-host-summary.json
loadtest/hostops/site-mix-loadtest-basic-amd-20260611T040722Z-host-profile.json
```

### Partial Probe: 25 Bets/sec Plus 250 Reads/sec

A later `25` bets/sec plus `250` reads/sec attempt started at `2026-06-11T04:08:48Z`. It looked clean, but it did not run to the full five-minute confirmation target, so it is recorded as a partial probe rather than proof.

| Metric | Value |
| --- | ---: |
| Target | `25` bets/sec + `250` reads/sec |
| Result | Partial clean probe, not full proof |
| Bets attempted | `2597` |
| Bets succeeded | `2595` |
| Failed bets | `0` recorded in check failures |
| Site reads attempted | `25967` |
| Site reads succeeded | `25948` |
| HTTP failures | `0` |
| HTTP p95 | `108.23ms` |
| Max VUs | `200` |
| Host min CPU idle | `58.24%` |
| Max Docker CPU sum | `276.41%` |
| Max backend CPU | `83.18%` |
| Max Postgres CPU | `42.14%` |
| Min RAM available | `31091 MiB` |

Interpretation:

- This run suggests `25/250` is plausible on `v3.4.0`, but it must be rerun to completion before being treated as a passing evidence point.
- The host had meaningful headroom during the partial window.
- Because the follow-up `50/500` failed badly, the full `25/250` confirmation is still necessary before testing the midpoint.

Raw artifacts:

```text
loadtest/results/site-mix-20260611T040848Z-summary.json
loadtest/hostops/site-mix-loadtest-basic-amd-20260611T040848Z-host.csv
loadtest/hostops/site-mix-loadtest-basic-amd-20260611T040848Z-host-summary.json
loadtest/hostops/site-mix-loadtest-basic-amd-20260611T040848Z-host-profile.json
```

### Failed Bracket: 50 Bets/sec Plus 500 Reads/sec

A `50` bets/sec plus `500` reads/sec attempt started at `2026-06-11T04:12:07Z` and was manually aborted after the run degraded heavily.

| Metric | Value |
| --- | ---: |
| Target | `50` bets/sec + `500` reads/sec |
| Result | Fail / upper-bound bracket |
| Scenario time before abort | About `1m20s` |
| Bets attempted | `1911` |
| Bets succeeded | `196` |
| Bets failed | `715` |
| Site reads attempted | `8236` |
| Site reads succeeded | `7229` |
| Site reads failed | `7` |
| HTTP failure rate | `6.77%` |
| HTTP p95 | `33.05s` |
| HTTP max | `60.26s` |
| Dropped iterations | `34104` |
| VUs maxed | `2000/2000` |
| Host min CPU idle | `88.18%` |
| Max Docker CPU sum | `117.48%` |
| Max backend CPU | `38.37%` |
| Max Postgres CPU | `51.78%` |
| Min RAM available | `30540 MiB` |

Interpretation:

- `50/500` is above the current mixed workload envelope.
- The host was mostly idle while k6 reached `2000` VUs and requests timed out, so this does not look like a simple server CPU saturation limit.
- The failure pattern points toward request queueing, connection pool pressure, client/load-generator saturation, network behavior, or another concurrency bottleneck.
- This is a useful upper-bound bracket, but it should not be interpreted as the raw CPU capacity of the host.

Raw artifacts:

```text
loadtest/results/site-mix-20260611T041207Z-summary.json
loadtest/hostops/site-mix-loadtest-basic-amd-20260611T041207Z-host.csv
loadtest/hostops/site-mix-loadtest-basic-amd-20260611T041207Z-host-summary.json
loadtest/hostops/site-mix-loadtest-basic-amd-20260611T041207Z-host-profile.json
```

### Interim v3.4.0 Bracket After Failed 50/500

| Bound | Status | Evidence |
| --- | --- | --- |
| Lower candidate | `25/250` looked clean but partial | Needs full five-minute confirmation. |
| Upper bound | `50/500` failed | Too aggressive under current harness/system behavior. |
| Next midpoint after full lower confirmation | `35/350` | Run only after full `25/250` completes cleanly. |

This was the interim bracket immediately after the failed `50/500` attempt. Later sections record the completed `25/250` and `35/350` passes.

### Confirmed Pass: 25 Bets/sec Plus 250 Reads/sec

A full `25` bets/sec plus `250` reads/sec confirmation run started at `2026-06-11T04:18:13Z` and completed the five-minute scenario cleanly.

| Metric | Value |
| --- | ---: |
| Target | `25` bets/sec + `250` reads/sec |
| Result | Pass |
| Scenario duration | `5m` |
| Wall-clock runtime | `6m24.6s` including setup/graceful stop |
| Bets attempted | `7500` |
| Bets succeeded | `7500` |
| Failed bets | `0` |
| Site reads attempted | `75000` |
| Site reads succeeded | `75000` |
| Site reads failed | `0` |
| HTTP requests | `84004` |
| HTTP failures | `0` |
| HTTP p95 | `109.08ms` |
| HTTP max | `402.17ms` |
| Iteration p95 | `121.65ms` |
| Dropped/interrupted iterations | `0` |
| Max VUs observed | `34` |

Important interpretation note: k6 custom metric rates include setup and teardown wall time, so `sp_bets_succeeded` reports `19.50/sec` and `sp_site_reads_succeeded` reports `195.03/sec`. The scenario counts confirm the configured rate over the measured five-minute phase: `7500 / 300s = 25` bets/sec and `75000 / 300s = 250` reads/sec.

Host telemetry:

| Metric | Value |
| --- | ---: |
| Samples | `47` |
| Window | `2026-06-11T04:18:19Z` to `2026-06-11T04:24:35Z` |
| Max CPU user | `11.52%` |
| Max CPU system | `31.5%` |
| Min CPU idle | `56.98%` |
| Min RAM available | `30824 MiB` |
| Max RAM used | `1271 MiB` |
| Max disk used | `2%` |
| Max disk write | `48864 KiB/s` |
| Max network RX | `132.35 KiB/s` |
| Max network TX | `2156.83 KiB/s` |
| Max Docker CPU sum | `291.16%` |
| Max Docker RAM sum | `314.44 MiB` |
| Max backend CPU | `75.44%` |
| Max Postgres CPU | `57.61%` |
| Max Traefik CPU | `28.35%` |

Interpretation:

- `25/250` is now a confirmed clean `v3.4.0` mixed workload datapoint on this host.
- The host still had substantial CPU and RAM headroom.
- This establishes a solid lower bound below the failed `50/500` upper bracket.
- The next edge-finding run should be `35/350`, not another `25/250` and not another `50/500`.

Raw artifacts:

```text
loadtest/results/site-mix-20260611T041813Z-summary.json
loadtest/hostops/site-mix-loadtest-basic-amd-20260611T041813Z-host.csv
loadtest/hostops/site-mix-loadtest-basic-amd-20260611T041813Z-host-summary.json
loadtest/hostops/site-mix-loadtest-basic-amd-20260611T041813Z-host-profile.json
```

### Confirmed Pass: 35 Bets/sec Plus 350 Reads/sec

A full `35` bets/sec plus `350` reads/sec midpoint run started at `2026-06-11T04:26:37Z` and completed the five-minute scenario cleanly.

| Metric | Value |
| --- | ---: |
| Target | `35` bets/sec + `350` reads/sec |
| Result | Pass |
| Scenario duration | `5m` |
| Wall-clock runtime | `6m53.0s` including setup/graceful stop |
| Bets attempted | `10501` |
| Bets succeeded | `10501` |
| Failed bets | `0` |
| Site reads attempted | `105000` |
| Site reads succeeded | `105000` |
| Site reads failed | `0` |
| HTTP requests | `117505` |
| HTTP failures | `0` |
| HTTP p95 | `112.84ms` |
| HTTP max | `249.85ms` |
| Iteration p95 | `125.38ms` |
| Dropped/interrupted iterations | `0` |
| Max VUs observed | `51` |

Important interpretation note: k6 custom metric rates include setup and teardown wall time, so `sp_bets_succeeded` reports `25.43/sec` and `sp_site_reads_succeeded` reports `254.25/sec`. The scenario counts confirm the configured rate over the measured five-minute phase: `10501 / 300s ~= 35` bets/sec and `105000 / 300s = 350` reads/sec.

Host telemetry:

| Metric | Value |
| --- | ---: |
| Samples | `51` |
| Window | `2026-06-11T04:26:43Z` to `2026-06-11T04:33:32Z` |
| Max CPU user | `17.83%` |
| Max CPU system | `43.88%` |
| Min CPU idle | `40.94%` |
| Min RAM available | `31019 MiB` |
| Max RAM used | `1076 MiB` |
| Max disk used | `2%` |
| Max disk write | `40760 KiB/s` |
| Max network RX | `185.54 KiB/s` |
| Max network TX | `3002.83 KiB/s` |
| Max Docker CPU sum | `414.72%` |
| Max Docker RAM sum | `212.31 MiB` |
| Max backend CPU | `100.81%` |
| Max Postgres CPU | `58.1%` |
| Max Traefik CPU | `40.25%` |

Interpretation:

- `35/350` is now a confirmed clean `v3.4.0` mixed workload datapoint on this host.
- Latency stayed tight: HTTP p95 remained near `113ms` with zero HTTP failures.
- Host CPU and RAM still had room, so `35/350` is not the observed edge.
- The next edge-finding run should test between clean `35/350` and failed `50/500`. A direct midpoint is `42/420`; a more conservative next step is `40/400`.

Raw artifacts:

```text
loadtest/results/site-mix-20260611T042637Z-summary.json
loadtest/hostops/site-mix-loadtest-basic-amd-20260611T042637Z-host.csv
loadtest/hostops/site-mix-loadtest-basic-amd-20260611T042637Z-host-summary.json
loadtest/hostops/site-mix-loadtest-basic-amd-20260611T042637Z-host-profile.json
```

### Updated v3.4.0 Bracket After Full 35/350 Pass

| Bound | Status | Evidence |
| --- | --- | --- |
| Lower bound | `35/350` passed full `5m` | Clean: `10501` bets, `105000` reads, `0` failures, p95 `112.84ms`. |
| Upper bound | `50/500` failed | Heavy bet failures, p95 `33.05s`, `34104` dropped iterations. |
| Next midpoint | `42/420` | Direct midpoint between clean `35/350` and failed `50/500`. |
| Conservative next step | `40/400` | Smaller step if the operator wants less risk of another hard failure. |

## User-Equivalent Interpretation For v3.4.0 Mixed Tests

The mixed `site-mix` scenario has two different interpretations:

- **Measured server load:** configured API action rate from k6.
- **Human user equivalent:** estimated active users required to naturally generate that action rate under normal rate limits.

Normal/model-office rate limits are intentionally conservative:

```env
RATE_LIMIT_GENERAL_RATE_PER_SECOND=1
RATE_LIMIT_GENERAL_BURST=10
```

The load-test profile is intentionally permissive so one k6 source can generate aggregate traffic. Therefore, the user counts below are estimates, not authentication counts. They are best read as **minimum client identities** and **active-user equivalents**.

### Formulas

```text
total_actions_per_second = bets_per_second + reads_per_second
minimum_client_identities = ceil(total_actions_per_second / 1 general action per second)
active_users = total_actions_per_second / actions_per_second_per_active_user
active_bettors = bets_per_second * seconds_between_bets_per_bettor
hot_window_bettors = bets_per_second * hot_window_seconds
```

Important caveats:

- A client identity is effectively a rate-limit identity, commonly an IP/client identity depending on deployment headers and proxy trust settings.
- Many human users behind one NAT/proxy can share a limiter identity, so this is not always equal to unique users.
- A real browser page can issue multiple API actions. The `site-mix` scenario intentionally models API action load, not browser tab count.
- Public cached reads and authenticated betting both still consume backend/proxy capacity, even if they differ in business risk.
- k6 `--preauth-users` is a fixture credential pool. It is not the same as the number of concurrently active human users. The clean full `25/250` run used a larger authenticated fixture pool but only needed `34` max observed VUs to deliver the target arrival rate.

### Actual Measured Throughput To User Equivalents

| Run | Status | Fixture auth pool | Max observed VUs | Measured/target bets/sec | Measured/target reads/sec | Modeled API actions/sec | Minimum normal-limit client identities | Active users at `1` action every `5s` | Active users at `1` action every `10s` |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| `25/250`, full `5m`, `v3.4.0` | Pass | `1500` | `34` | `25` actual | `250` actual | `275` | `275` | `1375` | `2750` |
| `35/350`, full `5m`, `v3.4.0` | Pass | `2000` | `51` | `35` actual | `350` actual | `385` | `385` | `1925` | `3850` |
| `50/500`, aborted `v3.4.0` | Failed upper bracket | `2000` | `2000` ceiling hit | `50` target, not sustained | `500` target, not sustained | `550` target | `550` target only | `2750` target only | `5500` target only |
| `42/420`, next midpoint | Pending | `2000` planned | TBD | `42` target | `420` target | `462` | `462` | `2310` | `4620` |

Interpretation:

- The clean evidence-backed claim is currently `385` modeled API actions/sec for five minutes on this host and release.
- Under the normal sustained `1` general API action/sec/client-identity policy, that is equivalent to at least `385` independent client identities if every identity is fully active.
- Under a more human browsing model of one API action every `5-10s`, the same clean run corresponds to roughly `1925-3850` simultaneously active users.
- The failed `50/500` row is intentionally labeled as a target-only upper bracket. It should not be used as capacity until a clean five-minute pass exists.

### Hot-Window Platform User Interpretation

This table translates betting rate into the number of humans who place one bet during a one-minute hot market window. It ignores read traffic, so it should be read alongside the mixed API-action table above.

| Betting rate | Bettors in `1m` | Share of `10,000` users | Share of `30,000` users | Share of `50,000` users | Evidence status |
| ---: | ---: | ---: | ---: | ---: | --- |
| `25` bets/sec | `1500` bettors | `15%` | `5%` | `3%` | Clean full `25/250` mixed pass. |
| `35` bets/sec | `2100` bettors | `21%` | `7%` | `4.2%` | Clean full `35/350` mixed pass. |
| `42` bets/sec | `2520` bettors | `25.2%` | `8.4%` | `5.04%` | Next midpoint target. |
| `50` bets/sec | `3000` bettors | `30%` | `10%` | `6%` | Failed as part of `50/500`; not supported yet. |
| `100` bets/sec | `6000` bettors | `60%` | `20%` | `12%` | Clean pure hot-market `1m` baseline only; not a mixed five-minute proof. |

The practical read is that the current clean mixed evidence supports a very active `10,000` user event if about `21%` of users place one bet in the same minute, or a broader `30,000-50,000` user event if `4.2-7%` of the platform is actively betting in that hot minute while other users are browsing cached read paths.

### Clean 25/250 Result: User Equivalents

The confirmed clean `v3.4.0` mixed result was:

```text
25 bets/sec + 250 reads/sec = 275 API actions/sec
```

Minimum normal-limit client identities:

| Assumption | Required identities/users |
| --- | ---: |
| Minimum client identities at `1` action/sec each | `275` |
| Active users at `1` action/sec each | `275` |
| Active users at `1` action every `2s` | `550` |
| Active users at `1` action every `5s` | `1375` |
| Active users at `1` action every `10s` | `2750` |
| Active users at `1` action every `30s` | `8250` |

Betting-only equivalents for the same clean run:

| Bettor behavior | Active bettor equivalent |
| --- | ---: |
| `1` bet/sec/bettor | `25` bettors |
| `1` bet every `10s` | `250` bettors |
| `1` bet every `30s` | `750` bettors |
| `1` bet every `60s` | `1500` bettors |
| `1` bet every `5m` | `7500` bettors |

Read-only equivalents for the same clean run:

| Reader behavior | Active reader equivalent |
| --- | ---: |
| `1` read/sec/reader | `250` readers |
| `1` read every `2s` | `500` readers |
| `1` read every `5s` | `1250` readers |
| `1` read every `10s` | `2500` readers |
| `1` read every `30s` | `7500` readers |

Interpretation:

- The clean `25/250` run is not only `25 users`. It is `275` API actions/sec.
- Under normal per-client rate limits, it requires at least about `275` independent client identities if each identity consumes the full sustained allowance.
- Under a more realistic active-user model of one API action every `5-10s`, the clean run corresponds to roughly `1375-2750` simultaneously active users generating the modeled site mix.
- If the hot-market event is mostly users placing one bet per minute, the `25` bets/sec component corresponds to about `1500` active bettors in that minute, plus concurrent readers.

### Failed 50/500 Bracket: User Equivalents

The failed upper bracket attempted:

```text
50 bets/sec + 500 reads/sec = 550 API actions/sec
```

Minimum normal-limit client identities and user equivalents:

| Assumption | Required identities/users |
| --- | ---: |
| Minimum client identities at `1` action/sec each | `550` |
| Active users at `1` action/sec each | `550` |
| Active users at `1` action every `2s` | `1100` |
| Active users at `1` action every `5s` | `2750` |
| Active users at `1` action every `10s` | `5500` |
| Active users at `1` action every `30s` | `16500` |

Because `50/500` failed, these are not supported capacity claims. They are the current failed upper-bound target for this single-host topology and current harness behavior.

### Clean 35/350 Result: User Equivalents

The confirmed clean `v3.4.0` mixed result was:

```text
35 bets/sec + 350 reads/sec = 385 API actions/sec
```

Minimum normal-limit client identities and user equivalents:

| Assumption | Required identities/users |
| --- | ---: |
| Minimum client identities at `1` action/sec each | `385` |
| Active users at `1` action/sec each | `385` |
| Active users at `1` action every `2s` | `770` |
| Active users at `1` action every `5s` | `1925` |
| Active users at `1` action every `10s` | `3850` |
| Active users at `1` action every `30s` | `11550` |

Betting-only equivalents for `35` bets/sec:

| Bettor behavior | Active bettor equivalent |
| --- | ---: |
| `1` bet every `10s` | `350` bettors |
| `1` bet every `30s` | `1050` bettors |
| `1` bet every `60s` | `2100` bettors |
| `1` bet every `5m` | `10500` bettors |

### Next Midpoint Target: 42/420 User Equivalents

The next direct midpoint target is:

```text
42 bets/sec + 420 reads/sec = 462 API actions/sec
```

If it passes cleanly, the user-equivalent table will be:

| Assumption | Required identities/users |
| --- | ---: |
| Minimum client identities at `1` action/sec each | `462` |
| Active users at `1` action/sec each | `462` |
| Active users at `1` action every `2s` | `924` |
| Active users at `1` action every `5s` | `2310` |
| Active users at `1` action every `10s` | `4620` |
| Active users at `1` action every `30s` | `13860` |

Betting-only equivalents for `42` bets/sec:

| Bettor behavior | Active bettor equivalent |
| --- | ---: |
| `1` bet every `10s` | `420` bettors |
| `1` bet every `30s` | `1260` bettors |
| `1` bet every `60s` | `2520` bettors |
| `1` bet every `5m` | `12600` bettors |
