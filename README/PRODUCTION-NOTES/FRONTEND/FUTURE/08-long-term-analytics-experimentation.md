---
title: Long-Term Frontend Analytics and Experimentation
document_type: production-notes
domain: frontend
future: true
author: Patrick Delaney
updated_at: 2026-05-23T00:00:00Z
updated_at_display: "Saturday, May 23, 2026"
update_reason: "Move product analytics, funnels, and A/B testing behind event vocabulary and privacy decisions."
status: future
---

# Long-Term Frontend Analytics and Experimentation

## Purpose

This note holds analytics and experimentation ideas that should follow event taxonomy and privacy decisions.

## Deferred Topics

- Google Analytics, Mixpanel, or similar SDKs.
- Product funnels and conversion tracking.
- Financial or market interaction event tracking.
- A/B testing runtime.
- Feature flag experimentation platform.
- User journey dashboards.

## Why Deferred

Analytics can shape product decisions and collect sensitive behavioral data. It needs SocialPredict-owned event names, payload boundaries, privacy/redaction rules, and consent posture before any vendor runtime is installed.

## Entry Criteria

Reconsider this when:

- The browser telemetry event vocabulary is defined.
- Forbidden PII/accounting fields are explicit.
- Product owners know which decisions analytics should support.
- Privacy and consent posture is reviewed.
- Bundle-size and runtime impact are acceptable.

## Guardrail

Do not let vendor analytics schemas define SocialPredict's product or domain language.
