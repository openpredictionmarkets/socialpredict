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

The production deploy workflow is intentionally gated to release-triggered image builds:

```yaml
github.event.workflow_run.conclusion == 'success' &&
github.event.workflow_run.event == 'release'
```

This prevents a manual run of the Docker image workflow from accidentally deploying production.

## Local HostOps Convention

HostOps local environment files for production should live under:

```bash
~/.keys/socialpredict/mo/
```

Expected local files:

```bash
~/.keys/socialpredict/mo/hostops.env
~/.keys/socialpredict/mo/digitalocean.env
~/.keys/socialpredict/mo/id_ed25519
~/.keys/socialpredict/mo/id_ed25519.pub
```

Example `hostops.env`:

```bash
HOSTOPS_HOST=brierfoxforecast.com
HOSTOPS_HOST_IP=143.198.177.112
HOSTOPS_USER=root
HOSTOPS_PORT=22
HOSTOPS_KEY=~/.keys/socialpredict/mo/id_ed25519
HOSTOPS_REPO_PATH=/opt/socialpredict
```

Example `digitalocean.env`:

```bash
DIGITALOCEAN_CONTEXT=socialpredict
DIGITALOCEAN_DROPLET_ID=422726269
DIGITALOCEAN_DROPLET_NAME=breirfoxforecast-alpha
DIGITALOCEAN_FIREWALL_ID=1aa45531-83bc-4371-821c-8fa2abcda411
DIGITALOCEAN_FIREWALL_NAME=port80-access
```

## Operator Checks

After a production release deploy, verify:

```bash
curl -sS -o /dev/null -w '%{http_code}\n' https://brierfoxforecast.com/
curl -sS -o /dev/null -w '%{http_code}\n' https://brierfoxforecast.com/api/v0/content/home
```

If HostOps SSH is authorized for `mo`, retrieve the generated admin password:

```bash
./HostOps host env get mo ADMIN_PASSWORD
```

## Notes

- The Ansible production deploy uses GitHub secrets in `openpredictionmarkets/ansible_playbooks`.
- HostOps SSH access is separate operator access and is not required for the GitHub release workflow.
- `doctl` is useful for droplet/firewall inspection, but production app deployment is driven by GitHub Actions plus Ansible.
