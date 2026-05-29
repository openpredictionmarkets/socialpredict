# Runtime Rate Limit Policy Design

Status: implemented baseline
Date: 2026-05-29

## Design Position

Rate limits are runtime security controls. They should be configured beside deployment/runtime settings and should be visible in operational evidence.

They are not game rules. Moving them into `setup.yaml` would couple infrastructure security posture to market mechanics and would make production tuning look like a product/game-mode change.

## Configuration Shape

The backend constructs `security.RateLimitConfig` during runtime startup. The baseline implementation now parses explicit environment values before server wiring and preserves existing defaults when values are unset. Invalid configured values fail startup rather than silently weakening or misrepresenting the runtime posture.

Environment names:

```bash
RATE_LIMIT_LOGIN_RATE_PER_SECOND
RATE_LIMIT_LOGIN_BURST
RATE_LIMIT_GENERAL_RATE_PER_SECOND
RATE_LIMIT_GENERAL_BURST
RATE_LIMIT_CLEANUP_INTERVAL
```

Use decimal parsing for rates so values such as `0.1` remain possible. Use integer parsing for burst. Use Go duration parsing for cleanup interval.

## Install Integration

`./SocialPredict install` writes explicit `.env` values for development, localhost, and production-like installs.

The install UI should prefer profiles over raw fields:

- `secure-default`
- `small-droplet-staging`
- `custom`

For `custom`, the installer prompts for explicit values. For profiles, the installer writes the resolved values so the operator can audit and change `.env` later.

Non-interactive production installs can pass the profile with:

```bash
./SocialPredict install -e production -d example.com -m ops@example.com -r small-droplet-staging
```

or by setting `RATE_LIMIT_PROFILE` before invoking the installer.

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

The current compose files do not intentionally constrain backend CPU with Docker `cpus`, `mem_limit`, or `deploy.resources` settings. On the current one-vCPU droplet, host CPU observations should therefore be treated as real capacity signals unless a future host-level or compose-level limit is added.

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
