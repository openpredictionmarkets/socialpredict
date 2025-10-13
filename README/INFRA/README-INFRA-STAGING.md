# Infrastructure: Staging Deployment Guide

This document explains how SocialPredict is deployed automatically to **staging** at [https://kconfs.com](https://kconfs.com), within the Open Prediction Markets Github organization.

---

## Overview

Staging deployment is handled automatically through **GitHub Actions** and **Ansible**.

- **Primary Playbook Repo:**  
  [openpredictionmarkets/ansible_playbooks](https://github.com/openpredictionmarkets/ansible_playbooks/tree/main)

- **Staging Server:**  
  Hosted at `kconfs.com`

- **Trigger:**  
  A **merge into `main`** on [openpredictionmarkets/socialpredict](https://github.com/openpredictionmarkets/socialpredict)  
  automatically redeploys the latest build to staging.

---

## Deployment Flow

1. A **Pull Request** is merged into `socialpredict@main`.
2. The workflow `.github/workflows/deploy-to-staging.yml` runs.
3. That workflow **dispatches a remote event** (`deploy-to-staging`) to the **ansible_playbooks** repository.
4. The `ansible_playbooks` repo has a workflow called `deploy_staging.yml` which:
   - Connects to the staging host at `kconfs.com`.
   - Stops all running SocialPredict containers.
   - Removes old images:
     - `socialpredict-staging-backend`
     - `socialpredict-staging-frontend`
   - Pulls the latest code from `openpredictionmarkets/socialpredict@main`.
   - Builds **new backend and frontend images** using the production build pipeline (with `staging` prefixes).
   - Starts new containers.

> **Note:** The staging workflow (`deploy_staging.yml`) can also be triggered manually via **workflow_dispatch** on the Ansible repo.

---

## Developer Experience

- **You don’t need SSH access** to staging to deploy.
- Just **merge your feature branch into `main`** — staging updates automatically.
- You can manually redeploy anytime by running the workflow in **ansible_playbooks → Actions → deploy_staging**.
continued:  
## Environment & Variables

The following key environment variables are defined in `.env.example` and used across build and deploy workflows:

| Variable | Purpose |
|-----------|----------|
| `BACKEND_IMAGE_NAME` | Image name used when building the backend |
| `FRONTEND_IMAGE_NAME` | Image name used when building the frontend |
| `BACKEND_CONTAINER_NAME` | Name of backend container |
| `FRONTEND_CONTAINER_NAME` | Name of frontend container |
| `DOMAIN_URL` | Domain for the staging environment (`https://kconfs.com`) |
| `ADMIN_PASSWORD` | Admin login password for staging |

> The admin password is persistent across deployments and does **not** reset automatically.

---

## GitHub Workflow Summary

**File:** `.github/workflows/deploy-to-staging.yml`

```
name: Deploy To Staging
on:
  pull_request:
    branches: [main]
    types: [closed]

jobs:
  deploy:
    name: Deploy SocialPredict to Staging
    if: github.event.pull_request.merged == true
    uses: openpredictionmarkets/ansible_playbooks/.github/workflows/deploy_staging.yml@main
This ensures that staging is only rebuilt when a PR is merged to main, maintaining a clean and predictable pipeline.
```


### Manual Trigger Option

You can redeploy staging manually without merging a PR:

* Go to the ansible_playbooks repo.
* Open the Actions tab.
* Select deploy_staging workflow.
* Click Run workflow → main.

## Summary

✅ Merge to main → Auto-deploys to staging
✅ Ansible handles container rebuilds & restarts
✅ Password and environment stay persistent
✅ Manual deploy available via GitHub Actions