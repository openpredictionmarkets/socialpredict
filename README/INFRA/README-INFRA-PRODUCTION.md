# Infrastructure: Production Deployment Guide

This document explains how SocialPredict is deployed to **mo/production** at [https://brierfoxforecast.com](https://brierfoxforecast.com).

## Deployment Contract

- Merge to `main` deploys staging at `kconfs.com`.
- Publishing a GitHub release deploys production at `brierfoxforecast.com`.
- Manual Docker image workflow runs should not deploy production.

## Production Host

- Environment name: `mo`
- Domain: `brierfoxforecast.com`
- DigitalOcean droplet: `breirfoxforecast-alpha`
- Public IPv4: `143.198.177.112`
- Firewall: `port80-access`

## Workflow Chain

1. A GitHub release is published in `openpredictionmarkets/socialpredict`.
2. `.github/workflows/docker.yml` builds and publishes frontend/backend Docker images to GHCR.
3. `.github/workflows/deploy-to-production.yml` runs after that image workflow completes.
4. `deploy-to-production.yml` dispatches `deploy-to-production` to `openpredictionmarkets/ansible_playbooks`.
5. The Ansible production playbook connects to the production host and runs the deployment.
6. The `socialpredict` workflow waits for the downstream Ansible workflow to
   finish, then performs an external public check of
   `https://brierfoxforecast.com/health` and
   `https://brierfoxforecast.com/readyz`.

`openpredictionmarkets/ansible_playbooks` is the separate deployment-control
repo. It owns the GitHub workflow that installs Ansible, loads the production
SSH key from GitHub secrets, and runs the production playbook against the host.
The `socialpredict` repo only needs `ANSIBLE_PLAYBOOK_TOKEN` so it can dispatch
the production deployment event and read the matching downstream Ansible
workflow run status. For a fine-grained token, make sure it can dispatch
repository events and read Actions runs in `openpredictionmarkets/ansible_playbooks`.

The production deploy workflow is intentionally gated to release-triggered image builds:

```yaml
github.event.workflow_run.conclusion == 'success' &&
github.event.workflow_run.event == 'release'
```

This prevents a manual run of the Docker image workflow from accidentally deploying production.

## Required GitHub Secrets

In `openpredictionmarkets/socialpredict`, production deployment requires only:

```text
ANSIBLE_PLAYBOOK_TOKEN
```

In `openpredictionmarkets/ansible_playbooks`, the production workflow expects:

```text
PRODUCTION_PRIVATE_KEY
PRODUCTION_USER
PRODUCTION_HOST
PRODUCTION_PORT
PRODUCTION_DOMAIN
PRODUCTION_EMAIL
PRODUCTION_PASSWORD
```

These are CI/CD deployment secrets, not HostOps local variables:

| Secret | Primary purpose |
| --- | --- |
| `PRODUCTION_PRIVATE_KEY` | SSH private key used by GitHub Actions/Ansible to connect to the VPS |
| `PRODUCTION_USER` | SSH user used by Ansible |
| `PRODUCTION_HOST` | Hostname or IP address Ansible connects to |
| `PRODUCTION_PORT` | SSH port Ansible connects to |
| `PRODUCTION_PASSWORD` | Ansible become/sudo password when the host requires one |
| `PRODUCTION_DOMAIN` | Domain passed into `./SocialPredict install -e production -d`; becomes app/domain/proxy config in `.env` |
| `PRODUCTION_EMAIL` | Email passed into `./SocialPredict install -e production -m`; used for Traefik/Let's Encrypt certificate registration |

The Ansible secrets let GitHub Actions connect to the VPS and invoke
`./SocialPredict install` and `./SocialPredict up`. Docker, nginx, and Traefik
then consume the generated `.env` and config files on the host. By contrast,
HostOps variables live only on an operator laptop and are for human access.

The packaged production compose topology uses a local Docker Postgres service.
For that topology, `./SocialPredict install -e production` writes
`DB_REQUIRE_TLS=false` so the backend does not reject the in-container
`sslmode=disable` connection. Operators who replace the local Docker database
with an external production database should review `DB_REQUIRE_TLS` and
`DB_SSLMODE` explicitly. In practice, keep `DB_REQUIRE_TLS=false` only for the
packaged local compose database. For an external production database, set a
provider-appropriate TLS mode, for example `DB_REQUIRE_TLS=true` with
`DB_SSLMODE=verify-full` when certificate validation is configured.

The `ansible_playbooks` repository may also contain an `ADMIN_PASSWORD` secret,
but the current production workflow does not pass it into the Ansible command.
It is only relevant if the workflow or a manual Ansible run supplies an
`ADMIN_PASSWORD` variable to the playbook.

## Local HostOps Convention

HostOps is optional local operator access. It is not required for the release workflow because the Ansible workflow uses GitHub secrets to connect to the production host.

HostOps configuration lives on your laptop under `~/.keys/socialpredict/<env>/`.
It is separate from GitHub repository secrets. It does not trigger a GitHub
workflow and is not read by Ansible.

Detailed setup lives in [`README-INFRA-HOSTOPS.md`](./README-INFRA-HOSTOPS.md). For OpenPredictionMarkets production / `mo`, use the local environment directory:

```bash
~/.keys/socialpredict/mo/
```

Expected production settings are `HOSTOPS_HOST=brierfoxforecast.com`, `HOSTOPS_HOST_IP=143.198.177.112`, `HOSTOPS_USER=root`, `HOSTOPS_PORT=22`, and `HOSTOPS_REPO_PATH=/opt/socialpredict`. If HostOps SSH is authorized for `mo`, retrieve the generated admin password with:

```bash
./HostOps host env get mo ADMIN_PASSWORD
```

## Operator Checks

After a production release deploy, verify:

```bash
curl -sS -o /dev/null -w '%{http_code}\n' https://brierfoxforecast.com/
curl -sS -o /dev/null -w '%{http_code}\n' https://brierfoxforecast.com/api/v0/content/home
curl -sS https://brierfoxforecast.com/health
curl -sS https://brierfoxforecast.com/readyz
```

The GitHub production deploy workflow now performs the `/health` and `/readyz`
checks externally from GitHub Actions after the Ansible workflow completes. It
polls every 30 seconds for up to 10 minutes and expects `/health` to return
`live` and `/readyz` to return `ready`.

## Notes

- The Ansible production deploy uses GitHub secrets in `openpredictionmarkets/ansible_playbooks`.
- HostOps SSH access is separate operator access and is not required for the GitHub release workflow.
- `doctl` is useful for droplet/firewall inspection, but production app deployment is driven by GitHub Actions plus Ansible.
