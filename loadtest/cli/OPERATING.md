# Load Test CLI Operating Runbook

This runbook is written for a human operator or an LLM agent that needs to reproduce the OpenPredictionMarkets staging load-test sequence safely.

## Local Prerequisites

Install local tools on the load generator machine:

```bash
brew install k6 node
```

Verify:

```bash
./loadtest/cli/loadtest check
```

## SSH Setup

The CLI uses SSH for remote host operations. For OpenPredictionMarkets staging, the expected local key is:

```text
~/.keys/socialpredict/staging/id_ed25519
```

The corresponding public key must already be present in the staging host user's `~/.ssh/authorized_keys`.

Default staging SSH target:

```text
root@kconfs.com
```

Default staging repo path:

```text
/opt/socialpredict
```

Override those when needed with:

```bash
--host root@45.55.227.1
--key ~/.keys/socialpredict/staging/id_ed25519
--port 22
--repo-path /opt/socialpredict
```

## Command Sequence For Staging

1. Confirm public readiness:

```bash
curl -s https://kconfs.com/readyz
```

Expected body:

```text
ready
```

2. Confirm staging rate limits:

```bash
./loadtest/cli/loadtest host rate-limits staging
```

Expected current staging values for single-source load testing:

```text
RATE_LIMIT_LOGIN_RATE_PER_SECOND=50
RATE_LIMIT_LOGIN_BURST=100
RATE_LIMIT_GENERAL_RATE_PER_SECOND=500
RATE_LIMIT_GENERAL_BURST=1000
RATE_LIMIT_CLEANUP_INTERVAL=5m
```

3. Seed remote staging fixtures:

```bash
./loadtest/cli/loadtest fixtures seed staging \
  --users 100 \
  --moderators 5 \
  --markets 20 \
  --hot-markets 2 \
  --reset
```

This runs `./SocialPredict load seed` on the remote host with:

```text
LOAD_TEST_ENABLED=true
LOAD_TEST_ALLOW_PRODUCTION=true
```

`--reset` removes only load-test-prefixed fixture data before recreating fixtures.

4. Pull fresh fixture CSVs to the local load generator:

```bash
./loadtest/cli/loadtest fixtures pull staging
```

5. Run smoke:

```bash
./loadtest/cli/loadtest run smoke --base-url https://kconfs.com --api-prefix /api
```

Smoke should pass before any baseline or burst test.

6. Run a cautious baseline:

```bash
./loadtest/cli/loadtest run baseline \
  --base-url https://kconfs.com \
  --api-prefix /api \
  --duration 1m \
  --browse-rate 5 \
  --bet-rate 2
```

7. Increase load gradually and record results in the release dossier.

## Important Interpretation Notes

- The app rate limiter is per client identity/IP, not a global server cap.
- Single-source k6 from a Mac needs higher per-IP staging limits than normal production traffic.
- Model-office/production should keep conservative limits to discourage automation by any one client identity.
- Do not run heavy k6 tests on the app/database droplet itself; use a Mac or separate load-generator host.
- If smoke fails with `AUTHORIZATION_DENIED` or `MARKET_NOT_FOUND`, reseed remote staging and pull fresh fixtures again.
- If tests fail with `RATE_LIMITED` or `LOGIN_RATE_LIMITED`, confirm staging is using the high `.env.staging` rate-limit overlay.
