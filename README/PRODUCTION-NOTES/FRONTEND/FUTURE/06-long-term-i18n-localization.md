---
title: Long-Term Frontend i18n and Localization
document_type: production-notes
domain: frontend
future: true
author: Patrick Delaney
updated_at: 2026-05-23T00:00:00Z
updated_at_display: "Saturday, May 23, 2026"
update_reason: "Move i18n runtime and RTL work behind product requirements and canonical-language decisions."
status: future
---

# Long-Term Frontend i18n and Localization

## Purpose

This note holds internationalization and localization work that is intentionally deferred.

## Deferred Topics

- i18n runtime dependency selection.
- Translation key architecture.
- Locale-specific date, number, and currency formatting.
- RTL support.
- Translation workflow and content ownership.
- Localized routing or SEO metadata.

## Why Deferred

The current frontend baseline needs canonical SocialPredict language first. Localization can improve reach later, but it must not redefine backend-owned canonical codes, public failure reasons, accounting terms, auth/session semantics, or market outcomes.

## Entry Criteria

Reconsider this when:

- Product requirements identify supported locales.
- The frontend published-language inventory is stable.
- The design plan clarifies which terms are translatable display labels versus backend-owned canonical codes.
- Translation ownership and review workflow are known.

## Guardrail

Do not add i18n runtime dependencies as a baseline architecture move. Start with language ownership and canonical term boundaries.
