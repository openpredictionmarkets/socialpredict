---
title: Long-Term Deployment Platform
document_type: production-notes
domain: backend
author: Patrick Delaney
updated_at: 2026-04-30T11:55:00Z
updated_at_display: "Thursday, April 30, 2026 at 11:55 AM UTC"
update_reason: "Record that the serving-path liveness/readiness prerequisite finished on April 30 while keeping platform migration deferred."
status: draft
---

# Long-Term Deployment Platform

## Purpose

This note holds longer-term deployment-platform ideas that should not drive the active production-hardening sequence.

The active deployment work remains in [08-deployment-infrastructure.md](/workspace/socialpredict/README/PRODUCTION-NOTES/BACKEND/08-deployment-infrastructure.md).

## Completed Prerequisite

Real serving-path liveness and readiness behavior finished on April 30, 2026: `/health` now reports liveness, and `/readyz` reports readiness plus database availability. That removes one prerequisite from this future platform note, but it does not activate Kubernetes, Helm, Terraform, autoscaling, or service-mesh work.

## Deferred Topics

Deferred deployment-platform ideas include:

- Kubernetes manifests
- Helm charts
- Terraform or broader infra-as-code programs
- autoscaling policies
- multi-environment cluster standardization
- service-mesh adoption
- advanced secret-management systems
- blue-green or canary deployment orchestration beyond the current stack

## Preconditions

These ideas should stay deferred until the backend has:

- deployment policy that consumes the April 30, 2026 liveness/readiness endpoints intentionally
- clearer startup-writer and migration posture
- safer graceful shutdown behavior
- explicit proxy publishing for docs and infra routes
- enough operational evidence to justify a broader platform migration

## Guardrail

This document is non-binding on the active design plan and task queue until the active runtime and deployment notes are materially landed.
