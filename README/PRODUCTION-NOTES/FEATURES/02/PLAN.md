---
title: Resettable Public Demo Implementation Plan
document_type: feature-plan
domain: features
author: Patrick Delaney
updated_at: 2026-05-26T00:00:00Z
updated_at_display: "Tuesday, May 26, 2026"
update_reason: "Add the implementation sequence for a separate resettable public demo environment."
status: draft
---

# Resettable Public Demo Implementation Plan

## Purpose

This plan turns [DESIGN.md](./DESIGN.md) and [02-resettable-public-demo.md](./02-resettable-public-demo.md) into an implementation sequence.

The plan is intentionally split across repositories and reviewable slices. The demo feature touches app reset commands, Ansible remote operations, GitHub workflows, DigitalOcean host configuration, and public readiness checks.

Agents implementing this feature should mark checklist items as they complete them and leave unchecked items in place when intentionally deferred.

## Planning Principles

- Keep demo isolated from staging and production.
- Preserve production and staging deployment semantics.
- Make destructive reset behavior explicit and demo-only.
- Keep app-specific reset/seed logic inside `./SocialPredict`.
- Keep Ansible focused on remote host orchestration.
- Keep GitHub workflows focused on dispatch, scheduling, and external verification.
- Do not make GitHub workflows depend on local `HostOps` or laptop key paths.
- Add public readiness checks after both deploy and reset.
- Use an explicit demo guard before any destructive reset behavior.
- Start with app-owned truncate-and-reseed reset mechanics; defer volume recreation or snapshot restore until there is evidence they are needed.
- Avoid Terraform until the first manual DigitalOcean demo droplet proves the workflow shape.

## Repository Ownership

SocialPredict repository:

- App-level `./SocialPredict demo-reset` or equivalent command guarded by demo-only configuration.
- Local tests for reset/fixture behavior.
- GitHub workflow that dispatches demo deploy/reset in the Ansible repository.
- Documentation and public readiness expectations.

`openpredictionmarkets/ansible_playbooks` repository:

- Demo inventory/host variables.
- Demo deploy workflow/playbook.
- Demo reset workflow/playbook.
- Demo-scoped GitHub environment secrets.
- Remote invocation of app-owned commands.

DigitalOcean / host setup:

- Separate demo droplet.
- DNS for demo domain.
- Firewall/SSH policy.
- Docker/runtime prerequisites.
- Isolated database volume.

## Progress Ledger

- [ ] 01. Feature artifact and design alignment
- [ ] 02. Demo environment naming, domain, and DigitalOcean droplet decision
- [ ] 03. SocialPredict app-owned demo reset command
- [ ] 04. Demo fixture seed tests
- [ ] 05. Ansible demo deploy target
- [ ] 06. Ansible demo reset target
- [ ] 07. SocialPredict GitHub manual deploy workflow dispatch
- [ ] 08. SocialPredict GitHub scheduled reset workflow dispatch
- [ ] 09. External public readiness verification for demo
- [ ] 10. Demo operator docs and safety review

## Implementation Checklist

### 01. Feature Artifact And Design Alignment

Status: complete for this documentation PR.

Checklist:

- [x] Create `README/PRODUCTION-NOTES/FEATURES/02/`.
- [x] Add feature overview document.
- [x] Add `DESIGN.md` aligned with canonical deployment/runtime boundaries.
- [x] Add `PLAN.md` as an agent-usable implementation sequence.
- [x] Update production-note index links.
- [x] Keep the change documentation-only.

Exit criteria:

- [x] Documentation distinguishes demo, staging, and production.
- [x] Documentation states that demo data is disposable.
- [x] Documentation identifies `./SocialPredict` as owner of app reset/seed semantics.

Validation:

- [x] `git diff --check`

### 02. Demo Environment Naming And Droplet Decision

Service ownership: Release and Deployment Control plus operator infrastructure.

Checklist:

- [ ] Choose demo domain.
- [ ] Create or designate a separate DigitalOcean droplet.
- [ ] Confirm demo host does not share staging or production DB volumes.
- [ ] Confirm demo host does not contain production secrets.
- [ ] Configure remote app environment as demo, for example `APP_ENV=demo`.
- [ ] Configure DNS to point at demo host.
- [ ] Apply firewall/SSH policy consistent with staging/production hardening.
- [ ] Record demo host conventions in infra docs.

Exit criteria:

- [ ] Demo host is reachable by maintainers.
- [ ] Demo host is isolated from staging and production.
- [ ] Demo domain can be externally checked.

Validation:

- [ ] `doctl compute droplet list --format ID,Name,PublicIPv4,PrivateIPv4,Region,Status,Tags`
- [ ] DNS resolves to demo host.
- [ ] SSH access works with demo-scoped credentials.

### 03. SocialPredict App-Owned Demo Reset Command

Service ownership: SocialPredict app runtime and repository/data boundaries.

Checklist:

- [ ] Add `./SocialPredict demo-reset` or equivalent.
- [ ] Require explicit demo environment guard before destructive behavior, for example `APP_ENV=demo`.
- [ ] Require a demo-only reset confirmation value from the remote workflow.
- [ ] Run migrations or verify migrations are already applied through runtime/bootstrap orchestration.
- [ ] Truncate app-owned mutable demo data.
- [ ] Defer Docker volume recreation or snapshot restore unless truncate-and-reseed is insufficient.
- [ ] Seed admin user with password `Password1`.
- [ ] Set admin `must_change_password=false`.
- [ ] Limit the demo admin to isolated demo-only effects; do not wire production secrets or privileged external integrations into demo.
- [ ] Seed optional demo users and markets if needed.
- [ ] Avoid printing secrets or password hashes.

Exit criteria:

- [ ] Command refuses to run outside demo mode.
- [ ] Command leaves the app in a login-ready state.
- [ ] Command can be invoked remotely by Ansible.

Validation:

- [ ] Backend tests for seeded admin user state.
- [ ] Local dev/demo smoke test for login using `admin` / `Password1`.

### 04. Demo Fixture Seed Tests

Service ownership: app domain/repository tests.

Checklist:

- [ ] Test deterministic admin username.
- [ ] Test password hash validates `Password1`.
- [ ] Test `must_change_password=false`.
- [ ] Test demo guard refuses reset outside demo configuration.
- [ ] Test reset confirmation is required.
- [ ] Test reset clears user-created mutable demo data.
- [ ] Test reset is idempotent.
- [ ] Test seeded admin can exercise intended demo-only admin/moderation flows.

Exit criteria:

- [ ] Fixture behavior is proven without relying on Ansible.
- [ ] Reset can be run repeatedly.

Validation:

- [ ] `cd backend && go test ./...`

### 05. Ansible Demo Deploy Target

Service ownership: `openpredictionmarkets/ansible_playbooks`.

Checklist:

- [ ] Add demo environment variables/secrets.
- [ ] Add demo inventory or target mapping.
- [ ] Add demo deploy playbook/workflow.
- [ ] Deploy a workflow-selected image or release artifact built from a specific `main` commit.
- [ ] Log selected commit SHA or image tag.
- [ ] Preserve demo DB volume during deploy.
- [ ] Run app migrations.
- [ ] Restart only necessary app containers where practical.

Exit criteria:

- [ ] Manual Ansible demo deploy succeeds.
- [ ] Existing staging and production deploy workflows are unchanged.

Validation:

- [ ] Ansible workflow completes successfully.
- [ ] External demo `/health` and `/readyz` pass after deploy.

### 06. Ansible Demo Reset Target

Service ownership: `openpredictionmarkets/ansible_playbooks`.

Checklist:

- [ ] Add reset workflow/playbook that targets only demo.
- [ ] Ensure target environment is fixed by workflow identity, not by a free-form environment input.
- [ ] Pass demo-only reset confirmation to the app command.
- [ ] Invoke the app-owned reset command remotely.
- [ ] Log target host/domain clearly.
- [ ] Avoid embedding SQL table knowledge in Ansible.
- [ ] Avoid printing secrets.

Exit criteria:

- [ ] Manual Ansible reset restores deterministic demo fixtures.
- [ ] Reset cannot target staging or production by parameter typo.

Validation:

- [ ] Ansible reset workflow completes successfully.
- [ ] Admin login works after reset.
- [ ] User-created data from before reset is gone.

### 07. SocialPredict Manual Demo Deploy Workflow Dispatch

Service ownership: SocialPredict GitHub workflows.

Checklist:

- [ ] Add manual `workflow_dispatch` workflow for demo deploy.
- [ ] Use the existing Ansible dispatch-token pattern.
- [ ] Dispatch the Ansible demo deploy workflow.
- [ ] Pass only the selected app version and fixed demo workflow target.
- [ ] Wait for Ansible completion.
- [ ] Run external readiness verification.

Exit criteria:

- [ ] Maintainer can deploy demo from SocialPredict GitHub Actions.
- [ ] Failed Ansible deploy or failed readiness makes the SocialPredict workflow red.

Validation:

- [ ] Manual workflow run succeeds on demo.

### 08. SocialPredict Scheduled Demo Reset Workflow Dispatch

Service ownership: SocialPredict GitHub workflows.

Checklist:

- [ ] Add scheduled workflow for daily reset.
- [ ] Use UTC midnight unless the project explicitly chooses another timezone.
- [ ] Allow manual reset trigger for maintainers.
- [ ] Dispatch the Ansible demo reset workflow.
- [ ] Pass the demo-only reset confirmation and no arbitrary target environment.
- [ ] Wait for Ansible completion.
- [ ] Run external readiness verification.

Exit criteria:

- [ ] Scheduled reset runs automatically.
- [ ] Manual reset can be triggered when needed.
- [ ] Reset failures are visible in SocialPredict GitHub Actions.

Validation:

- [ ] Manual reset workflow succeeds.
- [ ] First scheduled run succeeds.

### 09. External Public Readiness Verification

Service ownership: Release and Deployment Control.

Checklist:

- [ ] Add demo URL to readiness verification configuration.
- [ ] Check `/health` for live server response.
- [ ] Check `/readyz` for app readiness.
- [ ] Optionally check a public backend data endpoint after reset.
- [ ] Confirm `admin` / `Password1` login works after reset without logging the password hash.
- [ ] Use bounded retry intervals to avoid hammering the service.

Exit criteria:

- [ ] Demo deploy/reset workflows stay pending until public readiness passes or timeout occurs.
- [ ] Users can diagnose readiness failure from GitHub Actions logs.

Validation:

- [ ] Failed readiness makes workflow red.
- [ ] Successful readiness makes workflow green.

### 10. Demo Operator Docs And Safety Review

Service ownership: documentation and release operations.

Checklist:

- [ ] Document demo purpose and daily reset policy.
- [ ] Document demo GitHub secrets in Ansible repo.
- [ ] Document SocialPredict dispatch-token requirement if reused.
- [ ] Document the demo guard and reset confirmation values.
- [ ] Document that demo deploy uses a selected `main` commit image/release artifact.
- [ ] Document that the first reset implementation truncates and reseeds app-owned mutable data.
- [ ] Document how to disable scheduled reset.
- [ ] Document how to rotate demo admin fixture password if needed.
- [ ] Document how to destroy/rebuild demo host.

Exit criteria:

- [ ] A maintainer can understand demo deploy/reset without reading workflow internals.
- [ ] The destructive reset boundary is clear.

Validation:

- [ ] Docs review confirms demo cannot be mistaken for staging or production.
