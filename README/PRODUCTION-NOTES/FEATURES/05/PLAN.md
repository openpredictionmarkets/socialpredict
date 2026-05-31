# Runtime Rate Limit Policy Plan

Status: implemented baseline; staging evidence remains pending
Date: 2026-05-29

## 01. Backend Env Parsing

Service ownership: backend runtime and security middleware.

Checklist:

- [x] Add env parsing for login/general rate and burst values.
- [x] Preserve current defaults when env values are unset.
- [x] Reject invalid values during security config startup.
- [x] Unit test default behavior and env override behavior.
- [x] Confirm `TRUST_PROXY_HEADERS` behavior remains unchanged.

## 02. Installer Profiles

Service ownership: `./SocialPredict install` and deployment documentation.

Checklist:

- [x] Add install profile selection for production-like installs.
- [x] Keep `secure-default` as the safe default.
- [x] Add `small-droplet-staging` profile for the current Basic `1 vCPU / 1 GiB / 25 GiB` DigitalOcean staging target.
- [x] Add `env-file` profile support for non-secret deploy env overlays.
- [x] Add `custom` prompts for explicit values.
- [x] Write resolved values into `.env`.
- [x] Document how to change values after install.

## 03. Loadtest CLI Low-Rate Support

Service ownership: `loadtest/cli` and k6 scenario configuration.

Checklist:

- [x] Add `--browse-time-unit` for baseline browse scenario.
- [x] Add `--bet-time-unit` for baseline bet scenario.
- [x] Add equivalent time-unit options where useful for hot-market and soak scenarios.
- [x] Document that k6 rates are integers and sub-1/sec traffic should use larger time units.
- [ ] Validate a tiny public staging run with fresh fixtures and no threshold failures.

Implementation note: a local 2026-05-29 k6 baseline run using
`--browse-rate 1 --browse-time-unit 5s --bet-rate 1 --bet-time-unit 20s`
initialized successfully and no longer hit the k6 fractional-rate parser error.
It still failed thresholds against public staging because the local fixture CSVs
were stale relative to the deployed staging database, so final staging evidence
remains pending.

## 04. Staging Evidence Pass

Service ownership: release dossier and operations.

Checklist:

- [ ] Deploy env-configurable rate limits to `kconfs.com`.
- [x] Land Ansible deploy support for sourcing environment-specific rate-limit overlays.
- [ ] Run smoke after deployment.
- [ ] Run a low baseline test.
- [ ] Increment rates until the first clear bottleneck appears.
- [ ] Record rate-limit settings and DigitalOcean host observations in the release dossier.

Ansible note: `openpredictionmarkets/ansible_playbooks` sources
`deploy/env/.env.staging` for staging and `deploy/env/.env.mo` for
production/model-office. Optional `STAGING_RATE_LIMIT_ENV_FILE` and
`PRODUCTION_RATE_LIMIT_ENV_FILE` secrets can point to different overlay paths.

## 05. Documentation

Checklist:

- [x] Update infra docs with the rate-limit env vars.
- [x] Update loadtest docs with the staging profile and time-unit examples.
- [x] Update release dossier schema/metadata guidance to include rate-limit settings.
- [x] Cross-link from FEATURE/04 load testing notes.

## Exit Criteria

- Operators can choose conservative defaults or staging/load-test limits during install.
- Backend runtime applies env-configured rate limits without code changes.
- k6 low-rate tests can run without fractional-rate errors.
- Release dossier evidence can distinguish rate-limit bottlenecks from app/database/host bottlenecks.
