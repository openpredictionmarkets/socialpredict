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

For temporary raw-IP load-test hosts, create an arbitrary HostOps environment
directory such as:

```text
~/.keys/socialpredict/loadtest/
```

with `hostops.env` pointing at the temporary Droplet IP. The app install on that
host should use production topology, load-test rate limits, and HTTP-only edge:

```bash
./SocialPredict install \
  -e production \
  -d 45.55.227.1 \
  -r loadtest \
  --tls-mode http

./SocialPredict up
```

Run tests against `http://45.55.227.1`, not `https://...`, unless you attach a
real DNS name and use the default `--tls-mode https`.

## Temporary DigitalOcean Host Smoke Path

This sequence documents the OpenPredictionMarkets raw-IP smoke path used for
short-lived load-test hosts. It intentionally starts with the cheapest useful
Droplet to prove SSH, Docker, Ansible, and HTTP deploy wiring before resizing or
creating a larger host.

1. Create a dedicated local SSH key for the temporary host:

```bash
mkdir -p ~/.keys/socialpredict/loadtest
chmod 700 ~/.keys/socialpredict/loadtest
ssh-keygen -t ed25519 \
  -f ~/.keys/socialpredict/loadtest/id_ed25519 \
  -N '' \
  -C socialpredict-loadtest-YYYYMMDD
```

2. Import the public key into DigitalOcean:

```bash
doctl compute ssh-key import socialpredict-loadtest-YYYYMMDD \
  --public-key-file ~/.keys/socialpredict/loadtest/id_ed25519.pub
```

3. Create a temporary Ubuntu 24.04 Droplet with Docker bootstrap cloud-init:

```bash
doctl compute droplet create socialpredict-loadtest-YYYYMMDD \
  --region nyc3 \
  --image ubuntu-24-04-x64 \
  --size s-1vcpu-1gb \
  --ssh-keys <DIGITALOCEAN_SSH_KEY_ID> \
  --tag-names socialpredict-loadtest,socialpredict \
  --user-data-file loadtest/hostops/cloud-init-docker-ubuntu2404.yml \
  --wait
```

The initial `s-1vcpu-1gb` host is for smoke only. It has a 25GB disk and can be
resized later to DigitalOcean sizes with at least a 25GB disk, such as
`g-2vcpu-8gb`, without permanently increasing disk. Avoid disk resize unless the
test specifically requires it because DigitalOcean disk increases are not
reversible.

4. Attach the normal HTTP/HTTPS/SSH firewall:

```bash
doctl compute firewall add-droplets <FIREWALL_ID> --droplet-ids <DROPLET_ID>
```

For OpenPredictionMarkets, the current shared firewall is `port80-access`.

5. Verify SSH and Docker:

```bash
ssh -i ~/.keys/socialpredict/loadtest/id_ed25519 root@<DROPLET_IP> '
  lsb_release -ds
  docker --version
  docker compose version
  free -h
  df -h /
'
```

6. Point the Ansible load-test workflow at the temporary host:

```bash
gh secret set LOADTEST_HOST --repo openpredictionmarkets/ansible_playbooks --body '<DROPLET_IP>'
gh secret set LOADTEST_PORT --repo openpredictionmarkets/ansible_playbooks --body '22'
gh secret set LOADTEST_USER --repo openpredictionmarkets/ansible_playbooks --body 'root'
gh secret set LOADTEST_DOMAIN_OR_IP --repo openpredictionmarkets/ansible_playbooks --body '<DROPLET_IP>'
gh secret set LOADTEST_RATE_LIMIT_PROFILE --repo openpredictionmarkets/ansible_playbooks --body 'loadtest'
gh secret set LOADTEST_PRIVATE_KEY --repo openpredictionmarkets/ansible_playbooks < ~/.keys/socialpredict/loadtest/id_ed25519
```

7. Run the manual Ansible workflow:

```bash
gh workflow run deploy_loadtest.yml \
  --repo openpredictionmarkets/ansible_playbooks \
  -f socialpredict_ref=main \
  -f tls_mode=http \
  -f domain_or_ip=<DROPLET_IP>
```

8. Verify the raw-IP deployment:

```bash
curl -sS http://<DROPLET_IP>/health
curl -sS http://<DROPLET_IP>/readyz
curl -sS http://<DROPLET_IP>/api/ops/status
```

9. Run a smoke test from the local load generator:

```bash
./loadtest/cli/loadtest fixtures seed loadtest \
  --host root@<DROPLET_IP> \
  --key ~/.keys/socialpredict/loadtest/id_ed25519 \
  --repo-path /opt/socialpredict \
  --users 100 \
  --moderators 5 \
  --markets 20 \
  --hot-markets 2 \
  --reset

./loadtest/cli/loadtest fixtures pull loadtest \
  --host root@<DROPLET_IP> \
  --key ~/.keys/socialpredict/loadtest/id_ed25519 \
  --remote-path /opt/socialpredict/loadtest/fixtures \
  --local-path loadtest/fixtures

./loadtest/cli/loadtest run smoke \
  --base-url http://<DROPLET_IP> \
  --api-prefix /api
```

10. Destroy the temporary host after testing:

```bash
doctl compute droplet delete <DROPLET_ID>
```

Clean up or rotate `LOADTEST_*` secrets after deleting the host so future manual
workflow runs cannot accidentally target a stale IP.

## Command Sequence For Staging

1. Confirm public readiness:

```bash
curl -s https://kconfs.com/readyz
```

Expected body:

```text
ready
```

Optional: inspect and clean staging disk before capacity runs:

```bash
./HostOps host disk staging
./HostOps host cleanup staging
./HostOps host cleanup staging --all-images
./HostOps host disk staging
```

Do not use `--volumes` during routine cleanup unless you have confirmed the
volumes are unused and safe to delete. The cleanup wrapper calls remote
`./SocialPredict cleanup docker`; HostOps does not own Docker runtime behavior.

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

7. Attach host telemetry when increasing load:

```bash
./loadtest/cli/loadtest run hot-market-burst \
  --base-url https://kconfs.com \
  --api-prefix /api \
  --duration 1m \
  --target-rate 50 \
  --preauth-users 100 \
  --monitor-env staging \
  --monitor-interval 5
```

This writes a CSV under `loadtest/hostops/` with CPU, RAM, disk, and Docker
stats sampled from the remote host over SSH. The CLI also writes sibling host
profile and host summary JSON files and prints the after-run maxima/minima. Use
that summary alongside the k6 summary when producing a release dossier.

The host profile is the test-control record. It captures OS/kernel, CPU count,
memory, root disk, Docker server/storage/cgroup settings, Docker-visible CPU and
RAM, and whether containers have explicit CPU or memory limits.

8. Increase load gradually and record results in the release dossier.

## Hot-Market Burst Sequence

After smoke and a low baseline pass, isolate concentrated betting pressure with
`hot-market-burst`. This scenario pre-authenticates users during k6 setup and
reuses bearer tokens during the measured betting window, so it is intended to
measure betting throughput rather than login churn.

```bash
./loadtest/cli/loadtest run hot-market-burst \
  --base-url https://kconfs.com \
  --api-prefix /api \
  --duration 1m \
  --target-rate 50 \
  --preauth-users 100
```

If `LOGIN_RATE_LIMITED` appears in this scenario, either increase
`--preauth-users` only after confirming the staging login limit, or reseed/pull
fresh fixtures and rerun. Do not interpret login-limit failures as pure betting
capacity failures.

## Normal-Rate-Limit Equivalent

OpenPredictionMarkets staging uses much higher per-IP limits than normal so a
single Mac can generate useful capacity pressure. The normal/model-office
general API limit is currently `1` request/second per client identity with burst
`10`; login is `0.1` request/second with burst `3`.

The dossier converts measured successful betting throughput into normal-limit
client-identity equivalents with:

```text
ceil(successful_bets_per_second / normal_general_rate_per_second)
```

Example: a run that sustains `65.78` successful bets/second corresponds to
`ceil(65.78 / 1) = 66` normal-limit client identities if each identity places
one bet/second. If modeling one bet every ten seconds per identity, use
`ceil(65.78 / 0.1) = 658`.

## Important Interpretation Notes

- The app rate limiter is per client identity/IP, not a global server cap.
- Single-source k6 from a Mac needs higher per-IP staging limits than normal production traffic.
- Model-office/production should keep conservative limits to discourage automation by any one client identity.
- Do not run heavy k6 tests on the app/database droplet itself; use a Mac or separate load-generator host.
- If smoke fails with `AUTHORIZATION_DENIED` or `MARKET_NOT_FOUND`, reseed remote staging and pull fresh fixtures again.
- If tests fail with `RATE_LIMITED` or `LOGIN_RATE_LIMITED`, confirm staging is using the high `.env.staging` rate-limit overlay.
- If host telemetry shows available memory dropping below roughly `150 MiB`, p95 latency above `1s`, dropped iterations, or sustained error rates, stop increasing the target rate and capture that run as a capacity boundary.
- If `docker_cpu_pct_sum` is near `100%` on a `1 vCPU` host while host CPU idle is near `0%`, Docker is not the bottleneck boundary; the whole host is CPU saturated.
- If the host profile shows no explicit container CPU/memory limits, there is no Docker Compose resource cap to raise for more local headroom.

## Small-Droplet Max-Capacity Ladder

For the current Basic `1 vCPU / 1 GiB RAM / 25 GiB Disk` staging Droplet, use a
gradual hot-market ladder and capture host telemetry on every run:

```bash
./loadtest/cli/loadtest run hot-market-burst \
  --base-url https://kconfs.com \
  --api-prefix /api \
  --duration 1m \
  --target-rate 50 \
  --preauth-users 100 \
  --monitor-env staging \
  --monitor-interval 5
```

After the run, create a dossier that includes both the k6 result and the host
summary:

```bash
./loadtest/cli/loadtest dossier \
  --summary loadtest/results/<summary>.json \
  --host-summary loadtest/hostops/<run>-host-summary.json \
  --out loadtest/dossier/runs/<run>.json
```

Suggested sequence after smoke passes:

- `50` bets/sec for `1m`
- `75` bets/sec for `1m`
- `100` bets/sec for `1m`
- `125` bets/sec for `1m`
- repeat the highest passing rate for `5m`

Stop the ladder when any run shows errors, dropped iterations, p95 latency over
the chosen service target, or host telemetry shows sustained memory pressure.
