---
title: Container Security Scanning
document_type: feature-overview
domain: features
author: Patrick Delaney
updated_at: 2026-06-08T00:00:00Z
updated_at_display: "Monday, June 8, 2026"
update_reason: "Define image vulnerability scanning, SARIF alerting, scheduled rescans, and release gating for SocialPredict container images."
status: draft
---

# Container Security Scanning

## Purpose

SocialPredict publishes backend and frontend images to GitHub Container Registry. GitHub's default repository security features are useful, but they do not by themselves create a full NIST-style container image security program.

This feature adds a stricter baseline:

- scan release images after they are built
- upload SARIF results into GitHub Code Scanning
- block release image workflows on critical image vulnerabilities
- rescan published GHCR images on a schedule so newly disclosed CVEs are surfaced after release
- use Dependabot to keep Docker base images and GitHub Actions moving forward

## Policy

| Area | Baseline policy |
| --- | --- |
| Release image scan | Scan pushed frontend/backend images by immutable digest. |
| Scheduled scan | Scan `latest` by default, with manual dispatch support for a specific tag such as `3.2.0`. |
| Alert destination | Upload SARIF to GitHub Code Scanning. |
| Blocking threshold | Fail workflow on `CRITICAL` vulnerabilities. |
| Non-blocking visibility | Upload `UNKNOWN`, `LOW`, `MEDIUM`, `HIGH`, and `CRITICAL` findings to SARIF. |
| Initial high-severity posture | `HIGH` findings are visible but non-blocking until the image baseline is cleaned up. |
| Future stricter posture | After baseline cleanup, consider blocking on both `HIGH` and `CRITICAL`. |

## Why This Is Not The Whole NIST Program

NIST SP 800-190 covers more than vulnerability scanning. A fuller program would also cover trusted base images, image immutability, provenance, signing, secret hygiene, runtime hardening, least privilege, network policy, host hardening, and continuous monitoring.

This feature is the image-scanning slice only. It complements, but does not replace, runtime and infrastructure hardening.

## Workflow Shape

The release Docker workflow now performs this order:

1. Build and push frontend image.
2. Attest frontend build provenance.
3. Scan the pushed frontend image by digest.
4. Upload frontend SARIF.
5. Fail if critical frontend vulnerabilities exist.
6. Build and push backend image.
7. Attest backend build provenance.
8. Scan the pushed backend image by digest.
9. Upload backend SARIF.
10. Fail if critical backend vulnerabilities exist.

Production deployment depends on the release Docker workflow succeeding, so a critical image finding blocks production deployment.

## Scheduled Rescans

The scheduled container scan runs weekly and scans GHCR images by tag. This catches vulnerabilities that are disclosed after an image was originally published.

Default scheduled target:

```text
ghcr.io/openpredictionmarkets/socialpredict-frontend:latest
ghcr.io/openpredictionmarkets/socialpredict-backend:latest
```

Manual dispatch can scan a release tag, for example:

```text
3.2.0
```

## Dependabot Coverage

Dependabot now checks:

- GitHub Actions in `.github/workflows/`
- backend Docker base images under `docker/backend/`
- frontend Docker base images under `docker/frontend/`

This should create reviewable PRs when base images or CI actions can move forward.

## Supply Chain Note

The first implementation uses versioned GitHub Actions for Trivy and SARIF upload. For a stricter supply-chain posture, future work should pin security-scanner actions to full commit SHAs or install the scanner binary from a verified release artifact.

## Acceptance Criteria

- Release image workflow scans frontend and backend images.
- Release image workflow uploads SARIF scan results.
- Release image workflow fails on critical image vulnerabilities.
- Production deploy remains blocked unless image publishing and critical scan gates pass.
- Scheduled scan workflow exists for recurring GHCR image checks.
- Manual scan dispatch can target a specific image tag.
- Dependabot can propose Docker base image and GitHub Actions updates.
