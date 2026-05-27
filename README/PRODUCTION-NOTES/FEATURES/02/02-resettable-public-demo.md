---
title: Resettable Public Demo Environment
document_type: production-notes
domain: features
author: Patrick Delaney
updated_at: 2026-05-26T00:00:00Z
updated_at_display: "Tuesday, May 26, 2026"
update_reason: "Start the feature spec for a disposable public demo environment backed by a separate DigitalOcean droplet."
status: draft
---

# Resettable Public Demo Environment

## Purpose

SocialPredict needs a public demo environment where anyone can try the current application without affecting staging or production data.

The demo environment should run on its own DigitalOcean droplet, deploy from the current `main` branch on demand, and reset its database to deterministic demo fixtures once per day. User-entered data in this environment is intentionally disposable.

This note is a feature-level spec. It cuts across SocialPredict app commands, GitHub workflows, the `openpredictionmarkets/ansible_playbooks` repository, DigitalOcean infrastructure, and operational documentation.

## Feature Artifact Map

This directory keeps the resettable-demo feature work together:

- [02-resettable-public-demo.md](./02-resettable-public-demo.md): feature overview, product behavior, and acceptance criteria.
- [DESIGN.md](./DESIGN.md): architecture and boundary design aligned with the canonical design plan.
- [PLAN.md](./PLAN.md): implementation sequencing and PR slicing plan derived from the design.

## Environment Intent

The resettable public demo is neither staging nor production.

- Staging validates deployment and release mechanics for maintainers.
- Production or `mo` serves durable real users and must preserve data.
- Demo serves public experimentation and must be safe to wipe.

The demo droplet must not share database volumes, secrets, host keys, or production credentials with staging or production.

## Demo Login Contract

The baseline public fixture should include a deterministic admin account:

```text
username: admin
password: Password1
must_change_password: false
```

The password is intentionally public only because the demo environment is disposable and isolated. This password must not be used for staging, production, or any environment that contains real data.

Additional demo users, markets, and moderator fixtures can be added later, but they should be produced by an app-owned seed/reset command rather than by ad hoc SQL in Ansible.

## Deployment Behavior

The demo should support manual deployment of the newest `main` branch.

Desired behavior:

1. A maintainer manually triggers a GitHub workflow in `openpredictionmarkets/socialpredict`.
2. The SocialPredict workflow dispatches a demo deploy workflow in `openpredictionmarkets/ansible_playbooks`.
3. Ansible connects to the demo droplet using demo-scoped secrets.
4. Ansible deploys the newest SocialPredict release or image to the demo host.
5. The app runs migrations.
6. The demo stays up or restarts only the necessary app containers.
7. The workflow verifies public readiness from outside the server.

This is an application upgrade path, not a destructive reset path.

## Reset Behavior

The demo should also support a scheduled reset every day at midnight.

Desired behavior:

1. A scheduled GitHub workflow triggers the reset path.
2. Ansible connects to the demo droplet using demo-scoped secrets.
3. Ansible invokes an app-owned reset command, such as `./SocialPredict demo-reset`.
4. The reset command wipes demo-mutated data or restores a clean snapshot.
5. The reset command runs migrations and deterministic seed fixtures.
6. The workflow verifies the public demo is healthy and ready.

The reset workflow must be explicit about its destructive nature and must only target the demo environment.

## Data Policy

Public demo data is temporary.

Rules:

- Demo data may be deleted at any time.
- User-entered demo data is not backed up as durable product data.
- The demo environment must not contain real production user data.
- Reset behavior must be scoped to the demo database or demo Docker volume.
- Operators must be able to verify which host and domain the reset is targeting before enabling the schedule.

## Operations Boundary

The desired boundary is:

- `./SocialPredict` owns app-specific reset, migration, and seed behavior.
- `openpredictionmarkets/ansible_playbooks` owns remote host execution and DigitalOcean-targeted deploy/reset playbooks.
- GitHub workflows own manual and scheduled orchestration.
- `./HostOps` may remain a local convenience wrapper for maintainers, but GitHub workflows should not depend on a maintainer laptop or local `~/.keys` material.

## Required External Configuration

The Ansible repository should use demo-scoped GitHub secrets or environment variables, similar to staging and production.

Likely demo values:

```text
DEMO_DOMAIN
DEMO_EMAIL
DEMO_HOST
DEMO_PASSWORD       # only if password auth is intentionally supported; prefer private key
DEMO_PORT
DEMO_PRIVATE_KEY
DEMO_USER
```

The SocialPredict repository should only need the dispatch token or repository token required to trigger the Ansible workflow, following the existing deployment pattern.

## Acceptance Criteria

- Demo has its own DigitalOcean droplet or equivalent isolated host.
- Demo has its own database volume and does not share staging or production data.
- A maintainer can manually deploy the newest `main` to demo from GitHub Actions.
- The demo reset runs on a daily schedule and can also be manually triggered.
- After reset, `admin` can log in with `Password1` and is not forced to change the password.
- User-created data disappears after reset.
- The reset path cannot target staging or production by accident.
- The public readiness check confirms the demo is externally reachable after deploy/reset.

## Non-Goals

This feature does not initially require:

- Terraform provisioning.
- Blue/green demo deploys.
- Production-grade backup policy for demo data.
- A browser-visible demo reset button.
- Demo user signup controls beyond normal app behavior.
- A generic environment management platform.
