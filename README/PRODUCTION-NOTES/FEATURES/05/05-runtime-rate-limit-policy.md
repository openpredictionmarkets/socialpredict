# 05 Runtime Rate Limit Policy

Status: proposed
Date: 2026-05-29

## Purpose

SocialPredict needs runtime-configurable rate limits so staging, production, and load-test environments can be tuned without changing game rules or rebuilding code.

This is not game setup. It should not live in `backend/setup/setup.yaml`, because rate limits are security and operations policy, not market mechanics or economics.

The first measured target is the current staging DigitalOcean baseline:

- Droplet class: regular CPU
- Memory: 512 MiB
- vCPU: 1
- Transfer: 500 GiB
- SSD: 10 GiB
- Cost: about $0.00595/hr, $4/mo
- Current staging domain: `kconfs.com`

This feature should let operators record what this small machine can safely handle before larger instance sizing, database separation, Redis/distributed rate limiting, or proxy-level rate limiting are introduced.

## Current Problem

The first k6 baseline against `kconfs.com` failed before it reached meaningful database or betting capacity pressure. The backend returned stable security failures:

- `LOGIN_RATE_LIMITED`
- `RATE_LIMITED`

That means the current test primarily validated the application rate limiter rather than the market, bet, database, or reverse-proxy capacity.

The current defaults remain appropriate as secure defaults, but they are too low for controlled staging load tests from a single load-generator IP.

## Ownership Boundary

Rate limits are runtime security configuration.

They should be owned by:

- `.env` values generated or updated by `./SocialPredict install`
- backend runtime startup configuration
- infrastructure/deployment documentation
- release dossier evidence for measured environments

They should not be owned by:

- `backend/setup/setup.yaml`
- game-mode setup
- market economics
- frontend-only configuration
- k6 scripts as hard-coded assumptions

## Proposed Runtime Environment Variables

Add environment-driven configuration with safe defaults matching current behavior when unset:

```bash
RATE_LIMIT_LOGIN_RATE_PER_SECOND=0.1
RATE_LIMIT_LOGIN_BURST=3
RATE_LIMIT_GENERAL_RATE_PER_SECOND=1
RATE_LIMIT_GENERAL_BURST=10
RATE_LIMIT_CLEANUP_INTERVAL=5m
```

Existing behavior should remain the fallback:

- login: one request every 10 seconds, burst 3
- general API: one request per second, burst 10
- cleanup interval: 5 minutes

The implementation should parse these values during backend startup and pass them into `security.RateLimitConfig` before constructing runtime middleware.

## Install-Time Policy

`./SocialPredict install` should set rate-limit values for production-like installs.

The install flow should not ask users to understand Go token-bucket internals. It should present a small number of profiles, then write explicit `.env` values.

Candidate profiles:

```text
secure-default
  Best for unknown public deployments and development.
  Keeps current conservative defaults.

small-droplet-staging
  Intended for the current $4/mo 512 MiB / 1 vCPU DigitalOcean droplet.
  Values are provisional and must be verified by load-test evidence.

custom
  Operator supplies explicit login/general rate and burst values.
```

The production install path should default to `secure-default` unless the operator explicitly chooses a staging/load-test profile.

Development installs can keep secure defaults. Developers who need local load testing can override `.env` manually.

## Initial Small-Droplet Staging Candidate

The first candidate values for the current `kconfs.com` staging machine should be conservative enough to avoid masking resource pressure but high enough to stop rate limiting from dominating low-volume k6 tests.

Start with:

```bash
RATE_LIMIT_LOGIN_RATE_PER_SECOND=5
RATE_LIMIT_LOGIN_BURST=20
RATE_LIMIT_GENERAL_RATE_PER_SECOND=25
RATE_LIMIT_GENERAL_BURST=50
RATE_LIMIT_CLEANUP_INTERVAL=5m
```

These are not production capacity claims. They are a starting point for measurement on the $4/mo droplet.

If this profile still rate-limits before CPU, memory, DB, or application latency becomes the bottleneck, raise it incrementally and record each run in the release dossier.

If this profile causes resource pressure before the desired request volume, record the machine limit rather than hiding it with a larger droplet.

## Load-Test Interaction

k6 should continue to run from outside the app/database droplet.

For public staging and production URLs, k6 should use:

```bash
./loadtest/cli/loadtest run smoke --base-url https://kconfs.com --api-prefix /api
```

For lower-than-one-request-per-second scenarios, the loadtest CLI should support k6 `timeUnit` values. k6 `constant-arrival-rate.rate` must be an integer, so this is invalid:

```bash
--bet-rate 0.05
```

The desired expression should be supported as:

```bash
--bet-rate 1 --bet-time-unit 20s
--browse-rate 1 --browse-time-unit 5s
```

This lets operators run tiny tests without abusing fractional rates.

## Release Dossier Evidence

Every capacity-oriented run should record:

- git SHA or release tag
- environment and domain
- DigitalOcean droplet size
- rate-limit profile and exact values
- k6 scenario, rate, time unit, duration, and user/market fixture counts
- p50, p95, p99 latency
- throughput and error rate
- `RATE_LIMITED` and `LOGIN_RATE_LIMITED` counts
- CPU, memory, disk, and network observations
- whether the run hit rate limits, application errors, DB pressure, or host pressure first

For the current $4/mo droplet, the dossier should explicitly say that findings apply to the 512 MiB / 1 vCPU / 10 GiB SSD tier only.

## Non-Goals

This feature should not immediately add:

- Redis-backed distributed rate limiting
- proxy-owned rate limiting
- per-user paid-tier quotas
- abuse-detection policy
- autoscaling
- larger droplet recommendations without evidence

Those may be future work after the first release dossier identifies the actual bottleneck.
