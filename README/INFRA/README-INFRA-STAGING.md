# Infrastructure: Staging Deployment Guide

This document explains how the OpenPredictionMarkets project deploys
`openpredictionmarkets/socialpredict` to staging at
[`kconfs.com`](https://kconfs.com), how that differs from a self-hosted
SocialPredict deployment, and how `HostOps` can help an operator connect to a
VPS after deployment.

## OpenPredictionMarkets Deployment Contract

For the OpenPredictionMarkets organization, deployment is intentionally split by
GitHub event:

| Event | Target | Domain | Environment name |
| --- | --- | --- | --- |
| Pull request merged into `main` | Staging | `kconfs.com` | `staging` |
| GitHub release published | Model office / production | `brierfoxforecast.com` | `mo` |

The application repository does not SSH into hosts directly. It dispatches a
deployment event to the Ansible repository, and the Ansible repository performs
the host-level deployment.

Repositories involved:

- Application repo: `openpredictionmarkets/socialpredict`
- Deployment repo: `openpredictionmarkets/ansible_playbooks`
- Staging host: `kconfs.com`
- Production / model-office host: `brierfoxforecast.com`

`openpredictionmarkets/ansible_playbooks` is where OpenPredictionMarkets keeps
the GitHub workflows, inventories, and Ansible playbooks that actually connect
to the VPS hosts. The `socialpredict` repo only dispatches deployment events to
that repo.

This is our organization's pipeline. Someone self-hosting SocialPredict can use
the same pattern, but they will need their own GitHub repositories, secrets,
VPS/Droplet, DNS, firewall, and SSH keys.

## Staging Workflow

Staging is triggered by `.github/workflows/deploy-to-staging.yml` in this repo.

Triggers:

- `pull_request` closed against `main`, only when the PR was merged
- `workflow_dispatch` for a manual staging redeploy

Deployment chain:

1. A PR is merged into `openpredictionmarkets/socialpredict@main`, or the
   staging workflow is run manually.
2. `deploy-to-staging.yml` uses `peter-evans/repository-dispatch`.
3. The workflow sends event type `deploy-to-staging` to
   `openpredictionmarkets/ansible_playbooks`.
4. `ansible_playbooks` runs its staging workflow.
5. The staging Ansible playbook connects to `kconfs.com` over SSH.
6. Ansible pulls `openpredictionmarkets/socialpredict@main` into
   `/opt/socialpredict`.
7. Ansible runs the app-owned install and runtime commands on the host.
8. The `socialpredict` workflow waits for the downstream Ansible workflow to
   finish, then performs an external public check of `https://kconfs.com/health`
   and `https://kconfs.com/readyz`.

The key boundary is:

- GitHub Actions decides when a deployment should happen.
- Ansible connects to the host and orchestrates remote deployment steps.
- `./SocialPredict` owns app runtime behavior on the host.
- `./HostOps` is a local operator convenience wrapper, not the staging deploy
  engine.

## What Ansible Does On Staging

The staging playbook in `openpredictionmarkets/ansible_playbooks` currently
performs this high-level flow:

1. Checks whether `/opt/socialpredict/.env` exists.
2. Stops existing SocialPredict containers when an environment file is present.
3. Removes old staging backend/frontend images.
4. Pulls the latest `main` branch into `/opt/socialpredict`.
5. Runs:

   ```bash
   /opt/socialpredict/SocialPredict install -e production -d "$STAGING_DOMAIN" -m "$STAGING_EMAIL"
   ```

6. Builds staging backend and frontend Docker images on the host.
7. Writes staging image names into `/opt/socialpredict/.env`.
8. Sets `DB_REQUIRE_TLS=false` for the staging local database topology.
9. Runs:

   ```bash
   /opt/socialpredict/SocialPredict up
   ```

10. Retries startup if the startup-writer container needs time to recover.
11. Runs app-owned Docker cleanup when available:

   ```bash
   /opt/socialpredict/SocialPredict cleanup docker
   ```

The generated `.env` on the host contains values such as `ADMIN_PASSWORD`.
That value can change after a deployment because the install flow regenerates
runtime secrets.

## Production / Model Office Difference

Production is documented separately in
[`README-INFRA-PRODUCTION.md`](./README-INFRA-PRODUCTION.md), but the important
contrast is:

- Staging deploys from merges to `main`.
- Production / `mo` deploys from published GitHub releases.
- Production deployment is gated so a manual Docker image workflow run does not
  accidentally deploy `brierfoxforecast.com`.
- Production uses the `PRODUCTION_*` Ansible secrets and the production host.
- Staging uses the `STAGING_*` Ansible secrets and the staging host.

## GitHub Setup For This Pattern

### Application repository secrets

In the application repository (`socialpredict`), the dispatch workflow needs a
secret that can trigger workflows in the Ansible repository:

```text
ANSIBLE_PLAYBOOK_TOKEN
```

For OpenPredictionMarkets this token is used by
`peter-evans/repository-dispatch` to send `deploy-to-staging` and
`deploy-to-production` events to `openpredictionmarkets/ansible_playbooks`.
It is also used by the `socialpredict` workflow to read the matching downstream
Ansible workflow run status, so the token must be able to dispatch repository
events and read Actions runs in `openpredictionmarkets/ansible_playbooks`.

No host SSH keys, hostnames, domains, or sudo/become passwords are required in
the `socialpredict` repo for the current deployment workflows. Older secrets
such as `ACCESS_KEY`, `COMPOSE_ENV`, `DOCKER_TOKEN`, `HOST`, or `USERNAME` are
not referenced by the current `socialpredict/.github/workflows/*` deployment
files.

A self-hosted fork can either:

- keep this two-repository pattern and create a token that can dispatch to its
  own Ansible/deployment repository, or
- replace this with a simpler single-repository workflow if that better fits
  the deployment model.

### Ansible repository secrets

The Ansible deployment repository needs host connection and install settings.
For staging, the current workflow expects these secrets:

```text
STAGING_PRIVATE_KEY
STAGING_USER
STAGING_HOST
STAGING_PORT
STAGING_DOMAIN
STAGING_EMAIL
STAGING_PASSWORD
```

These are CI/CD deployment secrets, not HostOps local variables:

| Secret | Primary purpose |
| --- | --- |
| `STAGING_PRIVATE_KEY` | SSH private key used by GitHub Actions/Ansible to connect to the VPS |
| `STAGING_USER` | SSH user used by Ansible |
| `STAGING_HOST` | Hostname or IP address Ansible connects to |
| `STAGING_PORT` | SSH port Ansible connects to |
| `STAGING_PASSWORD` | Ansible become/sudo password when the host requires one |
| `STAGING_DOMAIN` | Domain passed into `./SocialPredict install -e production -d`; becomes app/domain/proxy config in `.env` |
| `STAGING_EMAIL` | Email passed into `./SocialPredict install -e production -m`; used for Traefik/Let's Encrypt certificate registration |

Typical OpenPredictionMarkets staging values are:

```text
STAGING_HOST=kconfs.com
STAGING_PORT=22
STAGING_USER=root
STAGING_DOMAIN=kconfs.com
```

`STAGING_PRIVATE_KEY` must be the private SSH key whose public key is already in
the staging VPS user's `~/.ssh/authorized_keys`. `STAGING_PASSWORD` is used as
Ansible's become password when required by the target host. If the host uses
passwordless root SSH and does not need sudo escalation, the playbook or secret
model can be simplified.

For production / `mo`, the same pattern uses `PRODUCTION_*` secrets in the
Ansible repository.

The Ansible secrets let GitHub Actions connect to the VPS and invoke
`./SocialPredict install` and `./SocialPredict up`. Docker, nginx, and Traefik
then consume the generated `.env` and config files on the host. By contrast,
HostOps variables live only on an operator laptop and are for human access.

The `ansible_playbooks` repository may also contain an `ADMIN_PASSWORD` secret.
The current staging and production workflows do not pass that secret into the
Ansible command, so it is not required for the active deployment path. The
playbooks can optionally update the admin password only if an `ADMIN_PASSWORD`
variable is supplied by the workflow or a manual Ansible run.

## VPS / DigitalOcean Setup Checklist

A staging VPS or DigitalOcean Droplet needs:

1. DNS pointing the staging domain to the public IPv4 address.
2. Firewall rules that allow SSH, HTTP, and HTTPS.
3. Docker and Docker Compose available on the host.
4. A directory such as `/opt/socialpredict` writable by the deployment user.
5. An SSH public key in the deployment user's `~/.ssh/authorized_keys` matching
   the private key stored as `STAGING_PRIVATE_KEY` in GitHub.
6. Enough disk space for image builds and runtime containers.

For DigitalOcean, `doctl` can inspect droplets and firewalls, but it does not
add SSH keys to an existing droplet's `authorized_keys`. DigitalOcean account
SSH keys are normally injected when a droplet is created. For an existing host,
add the public key through an existing SSH session, the DigitalOcean web console,
or a one-off maintenance workflow that already has host access.

## HostOps Local Operator Setup

`./HostOps` is a local operator convenience wrapper. It is useful after deployment for SSH access and reading generated host `.env` values such as `ADMIN_PASSWORD`, but it is not the GitHub workflow deploy mechanism.

HostOps configuration lives on your laptop under `~/.keys/socialpredict/<env>/`.
It is separate from GitHub repository secrets. It does not trigger a GitHub
workflow and is not read by Ansible.

Detailed setup lives in [`README-INFRA-HOSTOPS.md`](./README-INFRA-HOSTOPS.md). Keep the key-generation, `hostops.env`, DigitalOcean, and command reference there so this staging guide only describes where HostOps fits in the deployment model.

For OpenPredictionMarkets staging, the expected local HostOps environment is:

```bash
~/.keys/socialpredict/staging/hostops.env
~/.keys/socialpredict/staging/id_ed25519
~/.keys/socialpredict/staging/id_ed25519.pub
```

The staging `hostops.env` should resolve to `kconfs.com`, use the deployment user and SSH port for the VPS, and set `HOSTOPS_REPO_PATH=/opt/socialpredict`.

After the HostOps public key has been authorized on the staging host, retrieve the generated admin password with:

```bash
./HostOps env list
./HostOps host env get staging ADMIN_PASSWORD
```

Use the same pattern for production / `mo`, with the `mo` environment directory and `brierfoxforecast.com`.

## How This Differs For Other Deployments

If you are not deploying OpenPredictionMarkets' own staging or production
servers, do not copy the domains or secrets verbatim. Decide your own contract:

- one repo or two repos
- merge-to-main deploys, release-only deploys, or manual-only deploys
- `staging`, `prod`, `demo`, or another environment naming scheme
- DigitalOcean, another VPS provider, or a private server
- GitHub-hosted Ansible, local Ansible, or another deployment runner

The reusable pieces are the boundaries:

- `./SocialPredict` should own app install/runtime operations.
- Ansible or another deploy runner should call `./SocialPredict` remotely.
- GitHub Actions should trigger the runner and hold CI/CD secrets.
- `./HostOps` should help a human operator connect, inspect, and eventually
  orchestrate host/cloud operations without duplicating app runtime logic.

## Post-Deploy Checks

After a staging deployment, verify:

```bash
curl -sS -o /dev/null -w '%{http_code}\n' https://kconfs.com/
curl -sS -o /dev/null -w '%{http_code}\n' https://kconfs.com/api/v0/content/home
curl -sS -o /dev/null -w '%{http_code}\n' https://kconfs.com/health
curl -sS -o /dev/null -w '%{http_code}\n' https://kconfs.com/readyz
```

The GitHub staging workflow now performs the `/health` and `/readyz` checks
externally from GitHub Actions after the Ansible workflow completes. It polls
every 30 seconds for up to 10 minutes and expects `/health` to return `live`
and `/readyz` to return `ready`. The verification job writes a GitHub Actions
job summary with the checked URLs, expected and last observed bodies, attempt
count, timeout, interval, and final result. `/ops/status` remains an operator
status endpoint and is not currently used as a staging deploy gate.

A deploy is not zero-downtime today. Brief 404s or failed probes can occur
while containers are stopped and restarted.
