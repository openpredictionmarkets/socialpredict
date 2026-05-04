---
title: Long-Term API Automation Authentication
document_type: production-notes
domain: backend
author: Patrick Delaney
updated_at: 2026-05-04T02:15:00Z
updated_at_display: "Monday, May 4, 2026 at 2:15 AM UTC"
update_reason: "Record deferred machine-to-machine and automation authentication options without changing the current JWT login contract."
status: future
---

# Long-Term API Automation Authentication

## Purpose

This note preserves a future authentication direction for API automation clients.
It does not change the current SocialPredict API contract.

Today, API clients authenticate by calling `/v0/login` with a username and password over HTTPS and then sending the returned JWT as a bearer token. That is acceptable for browser login and local curl testing, but it is not the ideal long-term interface for scripts, integrations, bots, classroom tooling, or other automated clients.

## Future Direction

SocialPredict should consider a dedicated automation-authentication path before treating the public API as integration-friendly.

Reasonable future options include:

- user-managed API keys with scoped permissions, expiration, revocation, and last-used audit metadata
- OAuth 2.0 authorization flows for third-party integrations
- OAuth device flow for CLI or device-style clients where pasting a password into a shell is a poor fit
- service-account or bot credentials if SocialPredict later needs first-class automated market makers, classroom imports, or partner integrations

The goal is to avoid requiring automation users to place account passwords in shell history, scripts, CI logs, or integration configuration.

## Current Non-Goal

This is not part of the current WAVE10 validation work.

The current API should continue to require authenticated bearer tokens for private actions such as:

- creating markets with `POST /v0/markets`
- placing bets with `POST /v0/bet`
- selling positions with `POST /v0/sell`
- accessing other private account actions

The near-term contract remains: users log in, receive a JWT, and present that token in the `Authorization: Bearer <token>` header. The frontend handles that flow automatically after login.

## Entry Criteria

This topic should only move into the active plan when at least one concrete automation use case exists, such as:

- documented API use outside the browser frontend
- a CLI, scheduled job, bot, classroom import, or partner integration
- repeated operational need to script authenticated API actions
- a security review that rejects password-based login for non-browser API clients

Before implementation, SocialPredict should decide:

- whether API keys, OAuth, device flow, or service accounts best match the product need
- whether tokens are user-scoped, app-scoped, organization-scoped, or service-account-scoped
- how scopes map to market creation, trading, admin, read-only, and account actions
- how users create, view, revoke, rotate, and audit automation credentials
- how automation credentials interact with `mustChangePassword`, rate limits, and future abuse controls

## Explicitly Deferred

The following are deferred until this note is reactivated on purpose:

- API key tables or secret storage
- OAuth provider implementation
- device-flow implementation
- service-account or bot-account model changes
- API-key scopes and permission UI
- automation-token audit trails

## Relationship To Existing Notes

This note complements:

- [01-long-term-security-hardening.md](/workspace/socialpredict/README/PRODUCTION-NOTES/BACKEND/FUTURE/01-long-term-security-hardening.md), which defers broader identity, session, request-signing, and machine-to-machine security posture.
- [02-long-term-api-design.md](/workspace/socialpredict/README/PRODUCTION-NOTES/BACKEND/FUTURE/02-long-term-api-design.md), which defers broader API-platform work until the current `/v0` contract is more stable.
- [05-security-hardening.md](/workspace/socialpredict/README/PRODUCTION-NOTES/BACKEND/05-security-hardening.md), which remains the active current-state security note.
- [06-api-design.md](/workspace/socialpredict/README/PRODUCTION-NOTES/BACKEND/06-api-design.md), which remains the active current-state API contract note.
