# SocialPredict Staging Load-Test Dossier

Date: 2026-05-29

Environment: `staging`

Base URL: `https://kconfs.com`

Topology tested: single DigitalOcean Droplet running Traefik, nginx, frontend, backend, and Postgres in Docker.

## Executive Summary

The current staging system has been validated for small-to-moderate external API load from a macOS load generator. The cleanest hot-market write tests show:

- `20` hot-market bets/second for `1m`: clean pass, low latency.
- `35` hot-market bets/second for `1m`: clean functional pass, but tail latency begins to stretch.
- `50` hot-market bets/second for `1m`: not yet a valid pure betting result because that run was contaminated by repeated login traffic and login rate limiting.

The current conservative capacity envelope for this specific staging Droplet is therefore:

- Comfortable hot-market write rate: about `20` bets/second.
- Upper observed clean hot-market write rate: about `35` bets/second.
- Needs retest with pre-authenticated users: `50+` bets/second.

May 31, 2026 update:

- `50/sec` pre-authenticated hot-market run passed cleanly with `3001/3001` bets succeeded and HTTP p95 around `704ms`.
- `75/sec` for `1m` passed functionally with `4412/4412` bets succeeded, but HTTP p95 rose to about `1.67s` and k6 dropped `89` iterations.
- `75/sec` for `5m` was aborted under stress and is a failure/ceiling datapoint: `295` failed bets, HTTP p95 about `11.18s`, and minimum available RAM about `119 MiB`.
- `100/sec` for `1m` was beyond the current tiny Droplet: `42.96%` HTTP failures, HTTP p95 about `21.04s`, and minimum available RAM about `82 MiB`.

This dossier does not prove capacity for `10,000`, `30,000`, or `50,000` simultaneously active users. It translates those user counts into required request/write rates and proposes machine-size ranges to validate next.

## Fixture State

Remote staging was seeded with:

- Regular users: `100`
- Moderators: `5`
- Markets: `20`
- Hot markets: `2`
- Fixture password: redacted in command output
- `must_change_password`: `false`

Seed command pattern:

```sh
./loadtest/cli/loadtest fixtures seed staging \
  --users 100 \
  --moderators 5 \
  --markets 20 \
  --hot-markets 2 \
  --reset
```

If the staging deploy was refreshed after the seed, reseed before continuing load tests.

## Rate-Limit Configuration Observed

Staging was configured for single-source external load testing:

```env
RATE_LIMIT_LOGIN_RATE_PER_SECOND=50
RATE_LIMIT_LOGIN_BURST=100
RATE_LIMIT_GENERAL_RATE_PER_SECOND=500
RATE_LIMIT_GENERAL_BURST=1000
RATE_LIMIT_CLEANUP_INTERVAL=5m
```

These are intentionally permissive staging values. They should not be copied to model-office or production defaults.

## Machine Observed

The user-reported target tier was the smallest DigitalOcean tier:

- `1 vCPU`
- `512 MiB` RAM
- `10 GiB` SSD

The actual runtime observations from `free -h` and Docker showed:

- Host memory visible to Docker: about `957 MiB`
- Swap: `0B`
- App, database, proxy, and frontend all colocated on one host
- Database: local Docker Postgres

This means the actual machine under test behaved closer to a `1 vCPU / 1 GiB` host than a `512 MiB` host.

## Test Results

| Scenario | Command shape | Result | Bets | HTTP req/s | HTTP p95 | Max HTTP | Notes |
| --- | --- | --- | ---: | ---: | ---: | ---: | --- |
| Smoke | `run smoke` | Pass | `3/3` | `3.69` | `110ms` | `127ms` | End-to-end health, login, market, bet path worked. |
| Baseline low | `browse=5/s`, `bet=2/s`, `1m` | Pass | `121/121` | `8.95` | `100ms` | `140ms` | Clean baseline. |
| Baseline moderate | `browse=20/s`, `bet=10/s`, `2m` | Pass with warning | `1200/1200` | `36.79` | `1.81s` | `7.46s` | No failures, but `37` dropped iterations and stretched tail latency. |
| Hot market | `target=20/s`, `1m` | Pass | `1200/1200` | `38.77` | `63ms` | `122ms` | Clean hot-market write result. |
| Hot market | `target=35/s`, `1m` | Pass with warning | `2101/2101` | `65.43` | `787ms` | `2.68s` | No failures, but tail latency is rising. |
| Hot market | `target=50/s`, `1m` | Invalid for write ceiling | `2037/2081` succeeded | `71.16` | `11.04s` | `14.00s` | Login churn caused `204` login failures and `191` login rate-limit events. Retest with pre-auth is required. |
| Hot market pre-auth | `target=50/s`, `1m` | Pass | `3001/3001` | `47.04` | `704ms` | `2.49s` | Cleanest current capacity-pass datapoint. |
| Hot market pre-auth | `target=75/s`, `1m` | Pass with warning | `4412/4412` | `67.31` | `1.67s` | `3.86s` | No failed bets, but p95 crossed `1s` and k6 dropped `89` iterations. |
| Hot market pre-auth | `target=75/s`, `5m` | Fail / aborted | `7619/8785` attempted | `59.62` | `11.18s` | `17.06s` | Aborted under pressure; `295` failed bets and `982` dropped iterations. |
| Hot market pre-auth | `target=100/s`, `1m` | Fail / aborted | `1502/3297` attempted | `42.63` | `21.04s` | `30.47s` | Beyond current host capacity; `42.96%` HTTP failures. |

## Host Stats During 50/sec Contaminated Run

Samples collected during the `50/sec` hot-market run:

| Time UTC | Used RAM | Available RAM | Backend CPU | Postgres CPU | Traefik CPU | Notes |
| --- | ---: | ---: | ---: | ---: | ---: | --- |
| `12:00:13` | `457 MiB` | `328 MiB` | `19.89%` | `38.33%` | `48.98%` | Proxy and DB both active. |
| `12:00:29` | `511 MiB` | `273 MiB` | `24.36%` | `17.33%` | `39.67%` | Memory headroom narrowing. |
| `12:00:52` | `607 MiB` | `176 MiB` | `34.00%` | `30.14%` | `14.56%` | No swap; low available memory becomes a material risk. |

Interpretation:

- CPU was not pinned in the observed samples.
- Memory pressure is the more immediate risk on this tiny single-node setup.
- The `50/sec` result is not a clean database/write ceiling because login rate limiting and repeated authentication distorted the run.
- Traefik CPU was non-trivial during external HTTPS tests, so an end-to-end external test is more useful than an internal-only container test.

## Active-User Translation

Raw active-user counts do not directly size the system. The useful sizing unit is request rate, especially hot-market write rate.

For a one-minute hot window, assuming each participating user places one bet:

| Active users | 5% bet in 1 minute | 10% bet in 1 minute | 25% bet in 1 minute | 50% bet in 1 minute |
| ---: | ---: | ---: | ---: | ---: |
| `10,000` | `8.3` bets/s | `16.7` bets/s | `41.7` bets/s | `83.3` bets/s |
| `30,000` | `25.0` bets/s | `50.0` bets/s | `125.0` bets/s | `250.0` bets/s |
| `50,000` | `41.7` bets/s | `83.3` bets/s | `208.3` bets/s | `416.7` bets/s |

For a one-hour spread, the same volume is much easier:

| Active users | 25% bet over 1 hour | 50% bet over 1 hour |
| ---: | ---: | ---: |
| `10,000` | `0.7` bets/s | `1.4` bets/s |
| `30,000` | `2.1` bets/s | `4.2` bets/s |
| `50,000` | `3.5` bets/s | `6.9` bets/s |

The risky scenario is not total daily or hourly volume. The risky scenario is a hot market where many users act inside a one-minute or sub-minute window.

## Normal Rate-Limit Identity Matrix

OpenPredictionMarkets staging uses intentionally high per-IP limits so a single
Mac/k6 source can produce load:

```env
RATE_LIMIT_GENERAL_RATE_PER_SECOND=500
RATE_LIMIT_GENERAL_BURST=1000
```

The normal/model-office profile is conservative:

```env
RATE_LIMIT_LOGIN_RATE_PER_SECOND=0.1
RATE_LIMIT_LOGIN_BURST=3
RATE_LIMIT_GENERAL_RATE_PER_SECOND=1
RATE_LIMIT_GENERAL_BURST=10
```

The current runtime limiter is keyed by client identity/IP and is in-process.
Therefore, the matrix below estimates normal-limit **client identities**, not
guaranteed unique human users if many users share one NAT/proxy identity.

Formula:

```text
required_client_identities =
  ceil(successful_bets_per_second / allowed_bets_per_second_per_client_identity)
```

The `1.0` bet/sec row is the normal sustained general API limit midpoint/anchor:
one client identity using the full normal sustained allowance for one betting
request per second. Slower rows model less aggressive users.

| Assumed betting rate per client identity | `50/s` pass, `45.48` successful bets/s | `75/s` 1m warning, `65.78` successful bets/s | `75/s` 5m aborted, `56.66` successful bets/s | `100/s` aborted, `22.76` successful bets/s |
| ---: | ---: | ---: | ---: | ---: |
| `0.10` bet/s, 1 bet every 10s | `455` identities | `658` identities | `567` identities | `228` identities |
| `0.25` bet/s, 1 bet every 4s | `182` identities | `264` identities | `227` identities | `92` identities |
| `0.50` bet/s, 1 bet every 2s | `91` identities | `132` identities | `114` identities | `46` identities |
| `1.00` bet/s, normal-limit anchor | `46` identities | `66` identities | `57` identities | `23` identities |

Interpretation:

- The clean `50/s` run corresponds to about `46` normal-limit client identities
  if each identity places one bet per second, or about `455` identities if each
  identity places one bet every ten seconds.
- The `75/s` one-minute run corresponds to about `66` normal-limit client
  identities at one bet per second, but the host already showed warning signs:
  p95 latency above `1s` and dropped iterations.
- The aborted `75/s` five-minute and `100/s` one-minute runs should be treated
  as ceiling/failure evidence, not usable capacity targets.

## Machine-Size Estimate

These are planning estimates, not validated production limits.

DigitalOcean's current Basic Droplet table lists relevant reference sizes including `512 MiB / 1 vCPU`, `4 GiB / 2 vCPUs`, `8 GiB / 4 vCPUs`, and `16 GiB / 8 vCPUs`. The table below uses those resource bands as sizing shorthand; it does not assert that Basic shared CPU is the right final production class for every tier.

| Target | Traffic assumption | Estimated starting point | Why | Estimated DigitalOcean Droplet price |
| --- | --- | --- | --- | --- |
| `10,000` active users | Up to `10%` betting in a one-minute hot window, about `17` bets/s | `2 vCPU / 4 GiB` single node, or current node for staging only | Current staging passed `20` bets/s cleanly, but production needs memory headroom, deploy headroom, migrations, logs, and spikes. | Basic `4 GiB / 2 vCPU`: `$0.03571/hr`, `$24/mo`[^digitalocean-basic-pricing-2026-05-29] |
| `10,000` active users | `25%` betting in one minute, about `42` bets/s | `4 vCPU / 8 GiB`, preferably with managed Postgres | This exceeds the conservative `20-35` bets/s envelope and enters the unvalidated `50/sec` range. | Basic `8 GiB / 4 vCPU`: `$0.07143/hr`, `$48/mo`[^digitalocean-basic-pricing-2026-05-29] |
| `30,000` active users | `5-10%` betting in one minute, about `25-50` bets/s | `4 vCPU / 8 GiB` app host plus managed Postgres | The app may handle the low end, but colocated Postgres on a tiny host is the wrong production shape. | Basic app host `8 GiB / 4 vCPU`: `$0.07143/hr`, `$48/mo`; managed Postgres not included[^digitalocean-basic-pricing-2026-05-29] |
| `30,000` active users | `25%` betting in one minute, about `125` bets/s | Multiple app nodes or `8 vCPU / 16 GiB` app tier plus managed Postgres | This is several times the validated envelope and needs horizontal app capacity and database tuning. | Basic `16 GiB / 8 vCPU`: `$0.14286/hr`, `$96/mo`; managed Postgres/load balancing not included[^digitalocean-basic-pricing-2026-05-29] |
| `50,000` active users | `5-10%` betting in one minute, about `42-83` bets/s | `8 vCPU / 16 GiB` app tier plus managed Postgres | The low end is just beyond current clean validation; the high end requires dedicated database capacity. | Basic `16 GiB / 8 vCPU`: `$0.14286/hr`, `$96/mo`; managed Postgres not included[^digitalocean-basic-pricing-2026-05-29] |
| `50,000` active users | `25-50%` betting in one minute, about `208-417` bets/s | Load-balanced app nodes, managed Postgres sized separately, connection pooling, observability, and possibly async/caching changes | This is not a single tiny-Droplet workload. It should be treated as a dedicated scale project. | At least multiple Basic `16 GiB / 8 vCPU` app nodes at `$96/mo` each, plus managed Postgres/load balancer costs not included[^digitalocean-basic-pricing-2026-05-29] |

[^digitalocean-basic-pricing-2026-05-29]: DigitalOcean Basic Droplet pricing shown on 2026-05-29 from <https://www.digitalocean.com/pricing/droplets>: `512 MiB / 1 vCPU / 500 GiB transfer / 10 GiB SSD` at `$0.00595/hr`, `$4/mo`; `1 GiB / 1 vCPU / 1,000 GiB transfer / 25 GiB SSD` at `$0.00893/hr`, `$6/mo`; `2 GiB / 1 vCPU / 2,000 GiB transfer / 50 GiB SSD` at `$0.01786/hr`, `$12/mo`; `2 GiB / 2 vCPU / 3,000 GiB transfer / 60 GiB SSD` at `$0.02679/hr`, `$18/mo`; `4 GiB / 2 vCPU / 4,000 GiB transfer / 80 GiB SSD` at `$0.03571/hr`, `$24/mo`; `8 GiB / 4 vCPU / 5,000 GiB transfer / 160 GiB SSD` at `$0.07143/hr`, `$48/mo`; `16 GiB / 8 vCPU / 6,000 GiB transfer / 320 GiB SSD` at `$0.14286/hr`, `$96/mo`.

## Current Recommendation

For the next release dossier, report the current staging result as:

> On a single small DigitalOcean staging Droplet with colocated Docker Postgres, SocialPredict passed a pre-authenticated `50/sec` hot-market burst with `3001/3001` bets succeeded, `0%` HTTP failures, and HTTP p95 around `704ms`. A `75/sec` one-minute burst also had no failed bets, but p95 rose to about `1.67s` and k6 dropped iterations. Sustained `75/sec` and `100/sec` runs crossed into failure/ceiling territory.

For production-style planning:

- Use the current smallest staging Droplet only as a functional staging target.
- Treat `2 vCPU / 4 GiB` as the minimum serious single-node model-office target.
- Treat `4 vCPU / 8 GiB` plus managed Postgres as the first credible target for hot-market load beyond the current `50-65` successful bets/second stressed envelope.
- Treat `50,000` active users with one-minute hot windows as requiring a split app/database architecture, not a single small Droplet.

## Next Engineering Steps

1. Reseed staging after the merged deployment if the database was reset.
2. Pull fixtures locally:

```sh
./loadtest/cli/loadtest fixtures pull staging
```

3. Rerun hot-market burst with pre-authenticated users:

```sh
./loadtest/cli/loadtest run hot-market-burst \
  --base-url https://kconfs.com \
  --api-prefix /api \
  --duration 1m \
  --target-rate 50 \
  --preauth-users 100
```

4. Collect host stats during the run:

```sh
ssh -i ~/.keys/socialpredict/staging/id_ed25519 root@kconfs.com '
  date
  free -h
  docker stats --no-stream
'
```

5. Use `50/sec` as the current clean reference run for this tiny Droplet.
6. Treat `75/sec` one-minute as a warning-zone run, not a comfortable target.
7. Treat sustained `75/sec` and `100/sec` as current ceiling/failure evidence unless host resources are increased.

## Known Gaps

- No validated run has used `10,000`, `30,000`, or `50,000` unique users yet.
- The current fixtures only include `100` regular users and `20` markets.
- The `50/sec` run must be repeated after the pre-authenticated hot-market runner is deployed.
- No Postgres-level metrics were captured yet, such as active connections, locks, slow queries, buffer hit ratio, or transaction latency.
- No swap is configured on staging; memory exhaustion would likely fail abruptly.
- Current tests are from one external client IP, which is useful for staging limits but not a realistic global traffic distribution.
