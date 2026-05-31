# 06 Temporary Load-Test Droplets

Status: implemented baseline
Date: 2026-05-31

## Purpose

SocialPredict needs a reusable way to deploy production-like temporary hosts for capacity tests without resizing or disrupting `kconfs.com` staging.

The target use case is an ephemeral DigitalOcean Droplet addressed by raw public IP, tested for a few hours or days, then destroyed. These hosts should run the same packaged production topology, but with two deliberate differences:

- HTTP-only public edge for IP-based testing, because Let's Encrypt requires a domain.
- Permissive load-test rate limits so one load-generator IP can stress the app, database, proxy, and host before the app limiter becomes the bottleneck.

## Non-Goals

This is not a new application environment class. Do not add `-e loadtest` unless production and temporary capacity-test topology diverge materially.

This feature should not add autoscaling, managed Postgres, Terraform, DigitalOcean provisioning, or GitHub workflow orchestration yet. Those can be layered on once the manual temporary host path is proven.

## Operator Shape

Use production environment with explicit load-test knobs:

```bash
./SocialPredict install \
  -e production \
  -d 45.55.227.1 \
  -r loadtest \
  --tls-mode http

./SocialPredict up
```

Then test from a separate load generator:

```bash
./loadtest/cli/loadtest run hot-market-burst \
  --base-url http://45.55.227.1 \
  --api-prefix /api \
  --duration 5m \
  --target-rate 250 \
  --monitor-env loadtest
```

## Ownership Boundary

`./SocialPredict install` owns app runtime configuration:

- app environment remains `production`
- public base URL scheme
- API URL scheme
- Traefik edge mode
- rate-limit profile values

HostOps and future Ansible wrappers may choose the target host and invoke this command, but they should not reimplement app install logic.

## Safety Notes

The `loadtest` rate-limit profile is not a production recommendation. It is for short-lived controlled hosts where the operator is intentionally generating high traffic from one source IP.

HTTP-only mode is not for public model-office or production domains. It is for raw-IP temporary load-test hosts where TLS would add setup friction and Let's Encrypt cannot issue for the IP.

Destroy temporary hosts after testing to avoid cost drift. Prefer creating new temporary Droplets over permanently increasing the disk on staging.
