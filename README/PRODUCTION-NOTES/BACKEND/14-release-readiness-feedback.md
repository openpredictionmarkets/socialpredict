---
title: Release-To-Readiness Feedback
document_type: production-notes
domain: backend
author: Patrick Delaney
updated_at: 2026-05-11T21:45:00Z
updated_at_display: "Monday, May 11, 2026 at 09:45 PM UTC"
update_reason: "Document release-to-readiness feedback policy for GitHub deploy verification and backend probe semantics."
status: active
---

# Release-To-Readiness Feedback

## TL;DR

SocialPredict deployment feedback is split across two ownership boundaries:

- Ansible owns host-level deployment and reports whether the deployment job
  completed.
- The `socialpredict` GitHub Actions workflows own external public verification
  and report whether the deployed domain became live and ready from outside the
  host.

For the current OpenPredictionMarkets deployment, the public gate is intentionally
small: `GET /health` must return `live`, and `GET /readyz` must return `ready`.
The richer `GET /ops/status` endpoint is operator status, not a deploy gate yet.

## Current Release Contract

OpenPredictionMarkets uses this release-to-environment mapping:

| Event | Environment | Domain |
| --- | --- | --- |
| Pull request merged into `main` | Staging | `https://kconfs.com` |
| GitHub release published | Model office / production | `https://brierfoxforecast.com` |

The `openpredictionmarkets/socialpredict` repository dispatches deployment work
to `openpredictionmarkets/ansible_playbooks`. The Ansible repository connects to
the VPS, updates the checkout, runs `./SocialPredict install`, starts the app,
and performs host-level deployment steps. After that downstream Ansible workflow
finishes successfully, the `socialpredict` workflow polls the public domain from
GitHub Actions.

This gives reviewers one visible path in the application repository:

1. image build or deployment trigger starts in `socialpredict`
2. Ansible workflow is dispatched in `ansible_playbooks`
3. `socialpredict` waits for the matching Ansible workflow result
4. `socialpredict` performs external public readiness verification
5. GitHub Actions shows a final success or failure in the application repo

## Probe Semantics

| Endpoint | Expected success body | What it proves | What it does not prove |
| --- | --- | --- | --- |
| `GET /health` | `live` | The public route reaches a serving backend HTTP process. | Database readiness, business correctness, latency, or request success rate. |
| `GET /readyz` | `ready` | The public route reaches a backend instance whose readiness gate is open and whose primary database ping succeeds. | Background job health, third-party dependency health, or full user journey success. |
| `GET /ops/status` | JSON status object | Operator-visible process status including liveness, readiness, request failure count, and DB pool stats. | A stable deploy gate, fleet-wide monitoring signal, or early startup progress before the backend is listening. |

`/health` and `/readyz` are deliberately plain-text and exact-body probes so
Docker, compose, GitHub Actions, and black-box checks can use them without
parsing a larger operator payload.

`/ops/status` is intentionally not part of the deployment gate in this wave. It
is richer and process-local. It is useful for operator investigation after a
deploy, but making it a release gate should be a separate design decision that
defines which JSON fields are stable enough to enforce from CI.

## GitHub Actions Feedback

The shared `.github/actions/verify-public-readiness` action polls the configured
public origin every 30 seconds for up to 10 minutes by default. It now writes a
GitHub Actions job summary containing:

- the verified base URL
- the number of completed attempts
- elapsed time, timeout, and interval
- the checked `/health` and `/readyz` URLs
- the expected and last observed probe bodies
- the final pass or timeout result

This summary is meant to answer the operator question: "Did the public site
become live and ready after the deployment job finished?"

## Design Review Triggers

Review the active design plan before changing any of these release-facing
contracts:

- changing `/health` or `/readyz` status code or response-body semantics
- adding or removing an endpoint from the GitHub deployment gate
- changing whether `/ops/status` is operator-only or deploy-gating
- changing staging or production workflow trigger conditions
- changing the Ansible dispatch/wait boundary between `socialpredict` and
  `ansible_playbooks`
- changing startup writer, readiness gate, database TLS, or production compose
  behavior that affects when `/readyz` should become true

Timeout and polling interval changes are smaller operational changes, but they
should still be reflected in the infra docs because they affect what deployers
see in GitHub Actions.

## Non-Goals

This note does not introduce:

- Kubernetes, blue-green deployment, or a new orchestration platform
- Prometheus, Grafana, alert rules, or a monitoring stack
- host-internal Ansible health checks as a replacement for external public
  verification
- a full browser/user journey smoke test
- `/ops/status` JSON conformance as a required release gate

Those may be valid later, but the current baseline is intentionally a narrow
release-to-readiness feedback loop around the deployed public domain.

