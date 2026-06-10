# Temporary DigitalOcean Load-Test Droplet Runbook

This runbook describes how to create a short-lived DigitalOcean Droplet for larger SocialPredict load tests without permanently resizing `kconfs.com` staging.

Use this for capacity experiments that exceed the small staging ladder. The app under test runs on the temporary Droplet. The k6 load generator should run from your Mac or from a separate load-generator host, not from the same Droplet.

## When To Use This

Use a temporary Droplet when:

- staging has already passed smoke and moderate checks
- you need evidence above the `1 vCPU / 1 GiB` staging envelope
- you want to test CPU/RAM sizes without increasing staging disk permanently
- you want to destroy the host after a few hours or days

Do not run this against `brierfoxforecast.com` or another user-facing model-office/prod domain.

## Prerequisites

Local tools:

```bash
brew install k6 node
```

Required CLIs:

```bash
doctl version
gh --version
./loadtest/cli/loadtest check
```

Required access:

- DigitalOcean account configured in `doctl`
- GitHub CLI authenticated with access to `openpredictionmarkets/ansible_playbooks`
- ability to set `LOADTEST_*` secrets in `openpredictionmarkets/ansible_playbooks`
- the shared DigitalOcean firewall allowing SSH, HTTP, and HTTPS

For OpenPredictionMarkets, the shared firewall currently used by staging/model-office is named `port80-access`.

## Size Strategy

Start with the cheapest useful smoke host:

```text
s-1vcpu-1gb
```

Then resize CPU/RAM only after SSH, Docker, Ansible deploy, raw-IP HTTP, seed, and smoke are proven.

For the larger June 2026 experiment, the target was:

```text
s-8vcpu-32gb-amd
```

Important: do not pass `--resize-disk` unless you intentionally want permanent disk expansion. DigitalOcean CPU/RAM-only resize is reversible; disk increases are not.

## 1. Create Or Reuse A Load-Test SSH Key

```bash
mkdir -p ~/.keys/socialpredict/loadtest
chmod 700 ~/.keys/socialpredict/loadtest

ssh-keygen -t ed25519 \
  -f ~/.keys/socialpredict/loadtest/id_ed25519 \
  -N '' \
  -C socialpredict-loadtest-$(date +%Y%m%d)
```

Import the public key into DigitalOcean:

```bash
doctl compute ssh-key import socialpredict-loadtest-$(date +%Y%m%d) \
  --public-key-file ~/.keys/socialpredict/loadtest/id_ed25519.pub
```

Capture the key ID:

```bash
doctl compute ssh-key list
```

Set a shell variable for the key ID:

```bash
export DO_SSH_KEY_ID=<DIGITALOCEAN_SSH_KEY_ID>
```

## 2. Create The Temporary Droplet

```bash
export LOADTEST_DROPLET_NAME=socialpredict-loadtest-$(date +%Y%m%d-%H%M%S)

doctl compute droplet create "$LOADTEST_DROPLET_NAME" \
  --region nyc3 \
  --image ubuntu-24-04-x64 \
  --size s-1vcpu-1gb \
  --ssh-keys "$DO_SSH_KEY_ID" \
  --tag-names socialpredict-loadtest,socialpredict \
  --user-data-file loadtest/hostops/cloud-init-docker-ubuntu2404.yml \
  --wait
```

Get the Droplet ID and IP:

```bash
doctl compute droplet list \
  --format ID,Name,PublicIPv4,Status,Region,SizeSlug \
  | grep "$LOADTEST_DROPLET_NAME"
```

Set local variables:

```bash
export LOADTEST_DROPLET_ID=<DROPLET_ID>
export LOADTEST_DROPLET_IP=<DROPLET_IP>
```

## 3. Attach The Firewall

Find the firewall:

```bash
doctl compute firewall list --format ID,Name,InboundRules,DropletIDs
```

Attach the Droplet:

```bash
export LOADTEST_FIREWALL_ID=<FIREWALL_ID>

doctl compute firewall add-droplets "$LOADTEST_FIREWALL_ID" \
  --droplet-ids "$LOADTEST_DROPLET_ID"
```

## 4. Verify SSH And Docker

Cloud-init can take a few minutes after Droplet creation. Retry until Docker Compose is available.

```bash
ssh -i ~/.keys/socialpredict/loadtest/id_ed25519 root@"$LOADTEST_DROPLET_IP" '
  lsb_release -ds
  docker --version
  docker compose version
  free -h
  df -h /
  cat /root/docker-bootstrap-complete.txt 2>/dev/null || true
'
```

## 5. Point The Ansible Load-Test Workflow At The Host

Set GitHub secrets in `openpredictionmarkets/ansible_playbooks`:

```bash
gh secret set LOADTEST_HOST --repo openpredictionmarkets/ansible_playbooks --body "$LOADTEST_DROPLET_IP"
gh secret set LOADTEST_PORT --repo openpredictionmarkets/ansible_playbooks --body '22'
gh secret set LOADTEST_USER --repo openpredictionmarkets/ansible_playbooks --body 'root'
gh secret set LOADTEST_DOMAIN_OR_IP --repo openpredictionmarkets/ansible_playbooks --body "$LOADTEST_DROPLET_IP"
gh secret set LOADTEST_RATE_LIMIT_PROFILE --repo openpredictionmarkets/ansible_playbooks --body 'loadtest'
gh secret set LOADTEST_PRIVATE_KEY --repo openpredictionmarkets/ansible_playbooks < ~/.keys/socialpredict/loadtest/id_ed25519
```

Optional image/ref controls:

```bash
gh secret set LOADTEST_IMAGE_TAG --repo openpredictionmarkets/ansible_playbooks --body 'latest'
gh secret set LOADTEST_BACKEND_IMAGE_NAME --repo openpredictionmarkets/ansible_playbooks --body 'ghcr.io/openpredictionmarkets/socialpredict-backend'
gh secret set LOADTEST_FRONTEND_IMAGE_NAME --repo openpredictionmarkets/ansible_playbooks --body 'ghcr.io/openpredictionmarkets/socialpredict-frontend'
```

## 6. Deploy With The Load-Test Workflow

Use HTTP mode for raw IP hosts:

```bash
gh workflow run deploy_loadtest.yml \
  --repo openpredictionmarkets/ansible_playbooks \
  -f socialpredict_ref=main \
  -f tls_mode=http \
  -f domain_or_ip="$LOADTEST_DROPLET_IP"
```

Watch the workflow in GitHub Actions or from the CLI:

```bash
gh run list --repo openpredictionmarkets/ansible_playbooks --workflow deploy_loadtest.yml --limit 5
```

## 7. Verify The Raw-IP Deployment

```bash
curl -fsS "http://$LOADTEST_DROPLET_IP/health"
curl -fsS "http://$LOADTEST_DROPLET_IP/readyz"
curl -fsS "http://$LOADTEST_DROPLET_IP/api/ops/status"
```

Expected readiness body:

```text
ready
```

If this fails, inspect remote containers:

```bash
ssh -i ~/.keys/socialpredict/loadtest/id_ed25519 root@"$LOADTEST_DROPLET_IP" '
  cd /opt/socialpredict || exit 1
  docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
  docker compose ps
'
```

## 8. Run A Smoke Seed And Smoke Test

```bash
./loadtest/cli/loadtest fixtures seed loadtest \
  --host root@"$LOADTEST_DROPLET_IP" \
  --key ~/.keys/socialpredict/loadtest/id_ed25519 \
  --repo-path /opt/socialpredict \
  --users 100 \
  --moderators 5 \
  --markets 20 \
  --hot-markets 2 \
  --reset

./loadtest/cli/loadtest fixtures pull loadtest \
  --host root@"$LOADTEST_DROPLET_IP" \
  --key ~/.keys/socialpredict/loadtest/id_ed25519 \
  --remote-path /opt/socialpredict/loadtest/fixtures \
  --local-path loadtest/fixtures

./loadtest/cli/loadtest run smoke \
  --base-url "http://$LOADTEST_DROPLET_IP" \
  --api-prefix /api
```

Smoke must pass before resizing or running capacity tests.

## 9. Resize For Larger Tests

Resize CPU/RAM only:

```bash
doctl compute droplet-action resize "$LOADTEST_DROPLET_ID" \
  --size s-8vcpu-32gb-amd \
  --wait
```

Verify the resized host:

```bash
doctl compute droplet get "$LOADTEST_DROPLET_ID" \
  --format ID,Name,PublicIPv4,Status,Memory,VCPUs,Disk,Region,SizeSlug

ssh -i ~/.keys/socialpredict/loadtest/id_ed25519 root@"$LOADTEST_DROPLET_IP" '
  nproc
  free -h
  df -h /
  docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
'

curl -fsS "http://$LOADTEST_DROPLET_IP/readyz"
```

If containers are not healthy after resize, rerun the Ansible load-test deploy from step 6.

## 10. Seed Larger Fixtures

For the June 2026 larger-host experiment:

```bash
./loadtest/cli/loadtest fixtures seed loadtest \
  --host root@"$LOADTEST_DROPLET_IP" \
  --key ~/.keys/socialpredict/loadtest/id_ed25519 \
  --repo-path /opt/socialpredict \
  --users 10000 \
  --moderators 100 \
  --markets 1000 \
  --hot-markets 10 \
  --reset

./loadtest/cli/loadtest fixtures pull loadtest \
  --host root@"$LOADTEST_DROPLET_IP" \
  --key ~/.keys/socialpredict/loadtest/id_ed25519 \
  --remote-path /opt/socialpredict/loadtest/fixtures \
  --local-path loadtest/fixtures
```

Run smoke again:

```bash
./loadtest/cli/loadtest run smoke \
  --base-url "http://$LOADTEST_DROPLET_IP" \
  --api-prefix /api
```

## 11. Verify Fixture Integrity Before Capacity Runs

Confirm server-side market IDs match local `markets.csv`:

```bash
ssh -i ~/.keys/socialpredict/loadtest/id_ed25519 root@"$LOADTEST_DROPLET_IP" <<'REMOTE'
cd /opt/socialpredict || exit 1
set -a && . ./.env && set +a
docker exec socialpredict-postgres-container psql \
  -U "$POSTGRES_USER" \
  -d "$POSTGRES_DATABASE" \
  -v ON_ERROR_STOP=1 <<SQL
SELECT min(id) AS min_market_id, max(id) AS max_market_id, count(*) AS markets FROM markets;
SELECT count(*) AS bets FROM bets;
SQL
REMOTE

head -20 loadtest/fixtures/markets.csv
tail -20 loadtest/fixtures/markets.csv
```

If local fixture IDs are not within the server market range, reseed with `--reset`, pull fixtures again, and rerun smoke.

## 12. Run A Discovery Ladder

One-minute discovery ladder:

```bash
for rate in 100 150 200 250 300 350 400; do
  ./loadtest/cli/loadtest run hot-market-burst \
    --base-url "http://$LOADTEST_DROPLET_IP" \
    --api-prefix /api \
    --duration 1m \
    --target-rate "$rate" \
    --preauth-users 2000 \
    --setup-timeout 5m \
    --monitor-env loadtest-basic-amd \
    --monitor-host root@"$LOADTEST_DROPLET_IP" \
    --monitor-key ~/.keys/socialpredict/loadtest/id_ed25519 \
    --monitor-interval 5
done
```

Stop the ladder when the first degraded/failure result appears:

- failed bets
- HTTP failures
- dropped iterations
- p95 above the chosen service target
- host CPU idle near `0%`
- Postgres CPU dominating the host

## 13. Run A Five-Minute Confirmation

Use the highest clean discovery rate, or step down if the discovery run showed warning signs:

```bash
./loadtest/cli/loadtest run hot-market-burst \
  --base-url "http://$LOADTEST_DROPLET_IP" \
  --api-prefix /api \
  --duration 5m \
  --target-rate <RATE> \
  --preauth-users 5000 \
  --setup-timeout 10m \
  --monitor-env loadtest-basic-amd \
  --monitor-host root@"$LOADTEST_DROPLET_IP" \
  --monitor-key ~/.keys/socialpredict/loadtest/id_ed25519 \
  --monitor-interval 5
```

For the June 2026 Basic AMD experiment, the best clean sustained datapoint was:

```text
250 bets/sec for 5m
```

The `300/sec for 5m` repeat did not satisfy the strict clean-run standard.

## 14. Generate Dossier Artifacts

The CLI writes:

- k6 summary JSON under `loadtest/results/`
- host telemetry CSV under `loadtest/hostops/`
- host summary JSON under `loadtest/hostops/`
- host profile JSON under `loadtest/hostops/`

List recent artifacts:

```bash
ls -t loadtest/results/*summary.json | head
ls -t loadtest/hostops/*host-summary.json | head
ls -t loadtest/hostops/*host-profile.json | head
```

Generate compact dossier JSON:

```bash
./loadtest/cli/loadtest dossier \
  --summary loadtest/results/<summary>.json \
  --host-summary loadtest/hostops/<host-summary>.json \
  --out loadtest/dossier/runs/<run>.json
```

Update or create the markdown dossier under `loadtest/dossier/`.

## 15. Destroy The Temporary Host

Destroy the host when the experiment is done:

```bash
doctl compute droplet delete "$LOADTEST_DROPLET_ID"
```

Clean up or rotate `LOADTEST_*` secrets after deleting the host so future manual workflow runs cannot accidentally target a stale IP:

```bash
gh secret delete LOADTEST_HOST --repo openpredictionmarkets/ansible_playbooks
gh secret delete LOADTEST_DOMAIN_OR_IP --repo openpredictionmarkets/ansible_playbooks
gh secret delete LOADTEST_PRIVATE_KEY --repo openpredictionmarkets/ansible_playbooks
```

Keep `LOADTEST_PORT`, `LOADTEST_USER`, and rate-limit/profile secrets only if you intentionally reuse them across temporary hosts.

## Known Caveats

- A raw-IP host should use `tls_mode=http`; use `https` only with real DNS and `LOADTEST_EMAIL`.
- Do not run heavy k6 tests on the app/database Droplet itself.
- A Basic/shared-CPU success proves only that observed host during that test window. It is not dedicated-CPU evidence.
- CPU/RAM-only resize is reversible; disk resize is not.
- Always reseed and pull fixtures after redeploying or resetting the database.
- Smoke is necessary but not sufficient. Verify fixture IDs before long capacity runs.
