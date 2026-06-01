# Temporary Load-Test Droplets Design

Status: implemented baseline
Date: 2026-05-31

## Design Position

Temporary capacity-test hosts are production-like runtime deployments, not a new app environment. Keep `APP_ENV=production` so the same Docker Compose, startup-writer, Postgres, nginx, Traefik, and backend readiness paths are exercised.

Add narrow install-time knobs instead of a broad `-e loadtest` mode:

- `-r loadtest` for permissive single-source load-test rate limits.
- `--tls-mode http` for raw-IP HTTP-only public edge.

## Rate-Limit Profile

The `loadtest` profile should be high enough that a single Mac/k6 source does not trip the app limiter before host capacity becomes visible:

```env
RATE_LIMIT_LOGIN_RATE_PER_SECOND=100
RATE_LIMIT_LOGIN_BURST=200
RATE_LIMIT_GENERAL_RATE_PER_SECOND=1000
RATE_LIMIT_GENERAL_BURST=2000
RATE_LIMIT_CLEANUP_INTERVAL=5m
```

These values are intentionally higher than staging and much higher than model-office/production defaults.

## TLS Mode

Supported install TLS modes:

- `https`: default production/domain mode. Writes `https://DOMAIN`, renders Traefik with HTTP-to-HTTPS redirect and Let's Encrypt.
- `http`: temporary IP mode. Writes `http://DOMAIN_OR_IP`, renders Traefik with plain HTTP only, and skips Let's Encrypt certificate registration.

The CLI name is `--tls-mode` even though `http` disables TLS, because the choice describes the public edge mode.

## Compose and Proxy Boundary

Do not fork the production compose file for load testing. The same `scripts/docker-compose-prod.yaml` should run in both modes.

Traefik config can be generated from different templates while the nginx production template continues to route `/api`, `/health`, `/readyz`, `/ops/status`, Swagger/OpenAPI, market share shell pages, and frontend traffic.

## Future Re-Entry Points

Revisit this design when:

- Ansible needs a `loadtest` dispatch target.
- HostOps needs first-class temporary host provisioning.
- Capacity tests need managed Postgres instead of colocated Docker Postgres.
- We need DNS-backed temporary subdomains and real HTTPS for public demos.
