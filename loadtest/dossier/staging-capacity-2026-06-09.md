# SocialPredict Staging Capacity Dossier Addendum

Date: 2026-06-09

Environment: `staging`

Base URL: `https://kconfs.com`

Topology tested: single DigitalOcean Basic Droplet running Traefik, nginx, frontend, backend, and Postgres in Docker.

## Executive Summary

The June 9 staging rerun confirms that the current `1 vCPU / 1 GiB RAM / 25 GiB Disk` staging Droplet can sustain `35` hot-market bets/second for `5m` with no failed bets or HTTP failures after a current `main` redeploy and fresh load-test seed.

The same host also passed a `50` hot-market bets/second `1m` run with no failed bets, but that run showed CPU saturation and materially worse p95 latency. The practical staging interpretation is:

- `35/sec for 5m` is the current clean sustained staging datapoint.
- `50/sec for 1m` is a functional pass but a warning-zone datapoint on this host shape.
- The tested staging host should remain a functional staging target, not a production capacity target for concentrated hot-market events.

## Environment

Host profile captured during the `35/sec for 5m` run:

| Field | Value |
| --- | ---: |
| Droplet class | DigitalOcean Basic |
| CPU model | `DO-Regular` |
| Host CPU count | `1` |
| Docker CPU count | `1` |
| Host RAM total | `957 MiB` |
| Docker RAM total | `957 MiB` |
| Root disk size | `24,625 MiB` |
| Root disk used at start | `9,251 MiB` |
| Root disk used pct | `38%` |
| Swap | `0 MiB` |
| Explicit container CPU limits | `0/6` containers |
| Explicit container memory limits | `0/6` containers |

Staging rate limits were the high single-source load-test profile:

```env
RATE_LIMIT_LOGIN_RATE_PER_SECOND=50
RATE_LIMIT_LOGIN_BURST=100
RATE_LIMIT_GENERAL_RATE_PER_SECOND=500
RATE_LIMIT_GENERAL_BURST=1000
RATE_LIMIT_CLEANUP_INTERVAL=5m
```

## Fixture State

The staging database was reseeded before the June 9 tests:

```sh
./loadtest/cli/loadtest fixtures seed staging \
  --users 100 \
  --moderators 5 \
  --markets 20 \
  --hot-markets 2 \
  --reset

./loadtest/cli/loadtest fixtures pull staging
```

Seed output confirmed:

| Fixture | Count |
| --- | ---: |
| Regular users | `100` |
| Moderators | `5` |
| Markets | `20` |
| Hot markets | `2` |
| `must_change_password` | `false` |

## Test Results

| Scenario | Duration | Target | Result | Bets | HTTP failures | HTTP p95 | Max HTTP | VU max | Notes |
| --- | ---: | ---: | --- | ---: | ---: | ---: | ---: | ---: | --- |
| Smoke | `~4s` | n/a | Pass | `3/3` | `0%` | `128ms` | `146ms` | `1` | Health, readiness, status, login, market detail, and bet checks passed. |
| Hot market | `1m` | `35/sec` | Pass | `2100/2100` | `0%` | `80ms` | `481ms` | `100` | Clean functional one-minute check before the longer run. |
| Hot market | `1m` | `50/sec` | Pass with warning | `2955/2955` | `0%` | `3.03s` | `5.10s` | `145` | No failures, but `45` dropped iterations and host CPU saturation. |
| Hot market | `5m` | `35/sec` | Pass | `10501/10501` | `0%` | `720ms` | `2.69s` | `100` | Current clean sustained staging datapoint. |

## June 9 Sustained Run Detail

Command:

```sh
./loadtest/cli/loadtest run hot-market-burst \
  --base-url https://kconfs.com \
  --api-prefix /api \
  --duration 5m \
  --target-rate 35 \
  --preauth-users 100 \
  --setup-timeout 3m \
  --monitor-env staging \
  --monitor-interval 5
```

k6 result:

| Metric | Value |
| --- | ---: |
| Checks succeeded | `10704/10704` |
| HTTP requests | `10604` |
| HTTP failures | `0/10604` |
| Bets attempted | `10501` |
| Bets succeeded | `10501` |
| Observed bet rate | `34.56/sec` |
| HTTP duration avg | `130.94ms` |
| HTTP duration median | `49.28ms` |
| HTTP duration p90 | `129.27ms` |
| HTTP duration p95 | `720.21ms` |
| HTTP duration max | `2.69s` |
| Iteration duration p95 | `735.61ms` |
| Interrupted iterations | `0` |

Host telemetry:

| Metric | Value |
| --- | ---: |
| Telemetry samples | `35` |
| Window | `2026-06-09T03:37:25Z` to `2026-06-09T03:42:23Z` |
| Max CPU user | `65.69%` |
| Max CPU system | `38.24%` |
| Min CPU idle | `0%` |
| Avg CPU idle | `34.91%` |
| Min RAM available | `238 MiB` |
| Max RAM used | `535 MiB` |
| Max disk used | `38%` |
| Max disk read | `464 KiB/s` |
| Max disk write | `3368 KiB/s` |
| Max network RX | `22.54 KiB/s` |
| Max network TX | `30.99 KiB/s` |
| Max Docker CPU sum | `97.48%` |
| Max Docker RAM sum | `196.71 MiB` |
| Max backend CPU | `41.49%` |
| Max Postgres CPU | `59.62%` |
| Max Traefik CPU | `7.25%` |

Raw local artifacts:

| Artifact | Path |
| --- | --- |
| k6 summary | `loadtest/results/hot-market-burst-20260609T033719Z-summary.json` |
| host telemetry CSV | `loadtest/hostops/hot-market-burst-staging-20260609T033719Z-host.csv` |
| host summary JSON | `loadtest/hostops/hot-market-burst-staging-20260609T033719Z-host-summary.json` |
| host profile JSON | `loadtest/hostops/hot-market-burst-staging-20260609T033719Z-host-profile.json` |

These raw artifacts are ignored by git and are retained in the local workspace unless copied into a published dossier bundle.

## Interpretation

The `35/sec for 5m` run is a useful sustained staging proof because it ran long enough to expose tail behavior without producing failed bets, HTTP failures, or interrupted iterations.

However, this is not evidence that the host has large production headroom:

- Docker CPU reached `97.48%` on a `1 vCPU` host.
- Host CPU idle reached `0%`.
- Postgres reached `59.62%` CPU while colocated with the backend and proxies.
- Available RAM fell to `238 MiB`, and the host has no swap.
- The `50/sec for 1m` run remained functionally correct but had p95 latency around `3.03s` and dropped iterations.

The practical sustained staging envelope for this machine should therefore be reported as:

> On DigitalOcean Basic `1 vCPU / 1 GiB RAM / 25 GiB Disk` with colocated Docker Postgres, SocialPredict sustained `35` hot-market bets/second for `5m` with `10501/10501` bets succeeded, `0%` HTTP failures, and HTTP p95 around `720ms`. A `50/sec` one-minute run also had `0%` failures but showed CPU saturation and p95 around `3.03s`, so it should be treated as warning-zone evidence rather than comfortable capacity.

## Recommendation

For `kconfs.com` staging:

- Use `35/sec for 5m` as the current clean sustained datapoint.
- Treat `50/sec for 1m` as a short burst datapoint only.
- Do not use this tiny staging host for production sizing claims beyond light/model-office traffic.
- Use larger temporary load-test hosts for any claim above `50/sec`.

For future testing:

1. Repeat `35/sec for 5m` after major transaction-path changes.
2. If clean, test `40/sec for 5m`.
3. Stop escalating on this host when p95 exceeds `1s`, dropped iterations appear, or host CPU idle repeatedly hits `0%`.
4. Keep staging and model-office rate limits conservative outside explicit load-test windows.
