# SocialPredict Large-Host Capacity Rerun Notebook

Date opened: 2026-06-11

Status: active experiment; temporary large host is deployed, seeded, smoke-tested, and has one clean hot-market baseline datapoint.

Base URL: `http://159.65.37.166`

## Research Question

Can the current `main` build, including cached reporting/read-model improvements, reproduce or improve the previous larger-host SocialPredict capacity evidence on a temporary DigitalOcean Basic AMD host?

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

## Mixed Cached-Read Workload Plan

The next run should use `site-mix`, which combines hot-market betting with cached/read-model browsing paths.

Default read distribution at `250` reads/sec:

| Share | Approx/sec | Endpoint family |
| ---: | ---: | --- |
| `25%` | `62.5/sec` | `/v0/read/market-discovery/markets?...` |
| `15%` | `37.5/sec` | `/v0/read/market-discovery/{tagSlug}?...` |
| `38%` | `95/sec` | `/v0/read/markets/{id}/summary` |
| `10%` | `25/sec` | `/v0/markets/positions/{id}?limit=21&offset=0` |
| `7%` | `17.5/sec` | `/v0/markets/{id}/leaderboard?limit=21&offset=0` |
| `5%` | `12.5/sec` | `/v0/system/metrics`, `/v0/global/leaderboard?limit=21&offset=0` |

The default mixed run intentionally excludes the raw full market detail route, the live bets table route, and user financial summaries. Those should be measured separately as control or worst-case paths after the harness can warm the relevant snapshots.

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

Next command:

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

## Open Questions

- Does `25` bets/sec plus `250` cached reads/sec stay clean for `5m`?
- Does cached-read traffic materially affect write-path p95 compared with hot-market-only runs?
- At what mixed workload do Postgres or backend CPU become dominant?
- Should separate control tests be added for raw full market detail and live bets table routes after cached-read runs complete?
