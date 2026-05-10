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

`openpredictionmarkets/ansible_playbooks` is the separate deployment-control
repo. It owns the GitHub workflow that installs Ansible, loads the production
SSH key from GitHub secrets, and runs the production playbook against the host.
The `socialpredict` repo only needs `ANSIBLE_PLAYBOOK_TOKEN` so it can dispatch
the production deployment event.

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

The `ansible_playbooks` repository may also contain an `ADMIN_PASSWORD` secret,
but the current production workflow does not pass it into the Ansible command.
It is only relevant if the workflow or a manual Ansible run supplies an
`ADMIN_PASSWORD` variable to the playbook.

## Local HostOps Convention

HostOps is optional local operator access. It is not required for the release workflow because the Ansible workflow uses GitHub secrets to connect to the production host.

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
```

## Notes

- The Ansible production deploy uses GitHub secrets in `openpredictionmarkets/ansible_playbooks`.
- HostOps SSH access is separate operator access and is not required for the GitHub release workflow.
- `doctl` is useful for droplet/firewall inspection, but production app deployment is driven by GitHub Actions plus Ansible.
