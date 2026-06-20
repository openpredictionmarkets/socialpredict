---
title: Grouped Answer Limit Contract Alignment Design
document_type: feature-design
domain: features
author: Patrick Delaney
updated_at: 2026-06-17T00:00:00Z
updated_at_display: "Wednesday, June 17, 2026"
update_reason: "Define one answer-count contract across API, setup, domain, and frontend."
status: draft
---

# Grouped Answer Limit Contract Alignment Design

## Design Posture

The hard answer cap is an operational abuse/performance guardrail, not a statement about how many real-world outcomes can exist.

## Contract Rule

The maximum representable contract should be the domain/runtime maximum. Deployment setup may lower the effective cap, but OpenAPI and DTO validation should not claim a lower fixed cap than the configured product can expose.

## Recommended Baseline

- Domain maximum: `50` unless changed by explicit design.
- Setup may configure a lower hard safety cap.
- OpenAPI documents the maximum supported cap and notes deployment configuration may lower it.
- Frontend uses setup-provided cap for user-facing controls.
