# Runtime Rate Limit Policy Plan

Status: proposed
Date: 2026-05-29

## 01. Backend Env Parsing

Service ownership: backend runtime and security middleware.

Checklist:

- [ ] Add env parsing for login/general rate and burst values.
- [ ] Preserve current defaults when env values are unset.
- [ ] Reject or safely fall back on invalid values with clear startup logging.
- [ ] Unit test default behavior and env override behavior.
- [ ] Confirm `TRUST_PROXY_HEADERS` behavior remains unchanged.

## 02. Installer Profiles

Service ownership: `./SocialPredict install` and deployment documentation.

Checklist:

- [ ] Add install profile selection for production-like installs.
- [ ] Keep `secure-default` as the safe default.
- [ ] Add `small-droplet-staging` profile for the current $4/mo DigitalOcean target.
- [ ] Add `custom` prompts for explicit values.
- [ ] Write resolved values into `.env`.
- [ ] Document how to change values after install.

## 03. Loadtest CLI Low-Rate Support

Service ownership: `loadtest/cli` and k6 scenario configuration.

Checklist:

- [ ] Add `--browse-time-unit` for baseline browse scenario.
- [ ] Add `--bet-time-unit` for baseline bet scenario.
- [ ] Add equivalent time-unit options where useful for hot-market and soak scenarios.
- [ ] Document that k6 rates are integers and sub-1/sec traffic should use larger time units.
- [ ] Validate a tiny public staging run without fractional-rate parse errors.

## 04. Staging Evidence Pass

Service ownership: release dossier and operations.

Checklist:

- [ ] Deploy env-configurable rate limits to `kconfs.com`.
- [ ] Set the `small-droplet-staging` profile on the current 512 MiB / 1 vCPU droplet.
- [ ] Run smoke after deployment.
- [ ] Run a low baseline test.
- [ ] Increment rates until the first clear bottleneck appears.
- [ ] Record rate-limit settings and DigitalOcean host observations in the release dossier.

## 05. Documentation

Checklist:

- [ ] Update infra docs with the rate-limit env vars.
- [ ] Update loadtest docs with the staging profile and time-unit examples.
- [ ] Update release dossier schema/metadata guidance to include rate-limit settings.
- [ ] Cross-link from FEATURE/04 load testing notes.

## Exit Criteria

- Operators can choose conservative defaults or staging/load-test limits during install.
- Backend runtime applies env-configured rate limits without code changes.
- k6 low-rate tests can run without fractional-rate errors.
- Release dossier evidence can distinguish rate-limit bottlenecks from app/database/host bottlenecks.
