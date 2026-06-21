---
title: Container Security Scanning Plan
document_type: feature-plan
domain: features
author: Patrick Delaney
updated_at: 2026-06-08T00:00:00Z
updated_at_display: "Monday, June 8, 2026"
update_reason: "Track implementation tasks for image scanning and follow-up hardening."
status: draft
---

# Container Security Scanning Plan

## 01. Baseline CI Scanning

Checklist:

- [x] Add release-image Trivy scans for frontend image.
- [x] Add release-image Trivy scans for backend image.
- [x] Upload frontend image SARIF to GitHub Code Scanning.
- [x] Upload backend image SARIF to GitHub Code Scanning.
- [x] Fail release image workflow on critical frontend vulnerabilities.
- [x] Fail release image workflow on critical backend vulnerabilities.

## 02. Scheduled Drift Scanning

Checklist:

- [x] Add scheduled GHCR image scan workflow.
- [x] Add manual dispatch tag input for targeted rescans.
- [x] Upload scheduled scan SARIF to GitHub Code Scanning.
- [x] Fail scheduled scan workflow on critical vulnerabilities.

## 03. Dependency Update Support

Checklist:

- [x] Add Dependabot GitHub Actions updates.
- [x] Add Dependabot backend Docker base-image updates.
- [x] Add Dependabot frontend Docker base-image updates.

## 04. Future Hardening

Checklist:

- [ ] Pin scanner actions by full commit SHA.
- [ ] Generate and retain SBOMs for frontend/backend images.
- [ ] Sign images with Cosign.
- [ ] Deploy by image digest instead of mutable tags.
- [ ] Define a documented vulnerability exception process with expiry dates.
- [ ] Consider blocking on `HIGH,CRITICAL` after the first cleanup pass.
