# Runtime Rate Limit Policy Design

Status: proposed
Date: 2026-05-29

## Design Position

Rate limits are runtime security controls. They should be configured beside deployment/runtime settings and should be visible in operational evidence.

They are not game rules. Moving them into `setup.yaml` would couple infrastructure security posture to market mechanics and would make production tuning look like a product/game-mode change.

## Configuration Shape

The backend already constructs `security.RateLimitConfig` during server startup. The proposed implementation should add a small env parser before that construction and preserve existing defaults when env values are unset or invalid.

Recommended names:

```bash
RATE_LIMIT_LOGIN_RATE_PER_SECOND
RATE_LIMIT_LOGIN_BURST
RATE_LIMIT_GENERAL_RATE_PER_SECOND
RATE_LIMIT_GENERAL_BURST
RATE_LIMIT_CLEANUP_INTERVAL
```

Use decimal parsing for rates so values such as `0.1` remain possible. Use integer parsing for burst. Use Go duration parsing for cleanup interval.

## Install Integration

`./SocialPredict install` should write explicit `.env` values for production-like installs.

The install UI should prefer profiles over raw fields:

- `secure-default`
- `small-droplet-staging`
- `custom`

For `custom`, the installer can prompt for explicit values. For profiles, the installer should write the resolved values so the operator can audit and change `.env` later.

## Current Measurement Target

The first staging target is intentionally small:

```text
DigitalOcean regular CPU
512 MiB memory
1 vCPU
500 GiB transfer
10 GiB SSD
$4/mo
```

The design goal is not to promise high traffic on this instance. The goal is to make the limit measurable and to avoid confusing rate-limit failures with actual application/database capacity failures.

## Safety

Defaults must stay conservative.

A production install should not silently choose permissive load-test settings. Operators should explicitly choose the staging/load-test profile or custom values.

Load-test-specific values should be documented as staging evidence inputs, not as universal production recommendations.

## Future Re-Entry Points

Revisit this design if evidence shows:

- one app replica is not enough and rate limits need shared/distributed state
- Traefik or another ingress should enforce coarse rate limits before the app
- auth/login rate limits need separate trusted automation behavior
- customer-facing plans require per-user or per-organization quotas
