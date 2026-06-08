---
title: Container Security Scanning Design
document_type: feature-design
domain: features
author: Patrick Delaney
updated_at: 2026-06-08T00:00:00Z
updated_at_display: "Monday, June 8, 2026"
update_reason: "Specify CI boundaries for container image vulnerability scanning and alerting."
status: draft
---

# Container Security Scanning Design

## Boundary

Container vulnerability scanning belongs to the CI/release boundary. It does not change application runtime logic, market math, or deployment configuration.

## Critical Decision

The scan gate is critical for release deployment:

```text
critical image vulnerability found -> Docker workflow fails -> production deploy workflow does not dispatch
```

This prevents a release image with known critical findings from being automatically deployed to the model-office/production-style environment.

## Alerting Model

SARIF uploads make scanner findings visible in GitHub Code Scanning. The scan uploads all severities so maintainers can inspect the full image posture, while the workflow gate initially blocks only `CRITICAL` findings.

## Image Identity

Release scans use immutable image digests from `docker/build-push-action` outputs. This avoids scanning a mutable tag that could drift during the workflow.

Scheduled scans use tags because they are designed to monitor the current published state of `latest` or a manually selected release tag.

## Future Hardening

- Pin scanner actions by full commit SHA.
- Add SBOM generation and retention.
- Sign images with Cosign.
- Deploy images by digest instead of `latest`.
- Add policy exceptions with explicit expiry dates for unavoidable base-image CVEs.
- Raise the blocking threshold from `CRITICAL` to `HIGH,CRITICAL` after baseline cleanup.
