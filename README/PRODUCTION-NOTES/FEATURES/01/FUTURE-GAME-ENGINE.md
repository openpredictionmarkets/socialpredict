---
title: Moderator Mode Future Game Engine
document_type: feature-future-note
domain: features
author: Patrick Delaney
updated_at: 2026-05-24T00:00:00Z
updated_at_display: "Sunday, May 24, 2026"
update_reason: "Capture the later game-engine concern separately from the immediate moderator-mode smoke-test UI."
status: draft
---

# Future Game Engine

## Purpose

Moderator mode starts from a game setup change in `backend/setup/setup.yaml`, but the first implementation should not grow into a generic rules engine.

For now, the correct boundary is:

- `setup.yaml` declares the game mode and moderation policy.
- The configuration service parses that policy into typed application config.
- Domain services consume typed policy and enforce concrete rules.
- Handlers and frontend views adapt to backend-owned state instead of inventing game rules.

## Later Trigger

Create a dedicated game-engine layer only when one of these becomes true:

- A third game mode needs materially different creation, trading, resolution, or cancellation rules.
- Moderator mode adds enough branching that policy checks are scattered across handlers, repositories, or frontend code.
- Admin-configurable game behavior needs safe runtime validation rather than deploy-time setup changes.
- Rule combinations need explicit compatibility checks before a game can start.

## Candidate Shape

A later game engine should remain inside the application/domain boundary, not the frontend.

Possible seam:

- `GamePolicy`: typed immutable policy loaded from setup.
- `GameRules`: interface for creation, trading, resolution, amendment, and cancellation eligibility.
- `GameRulesRegistry`: maps `open`, `moderator`, and future modes to concrete rule sets.
- `GameDecision`: shared result type carrying allow/deny, public reason, audit metadata, and optional operator diagnostics.

The engine should be introduced only after the current direct-policy implementation has clear duplication or incompatible rules.

## Current Moderator Work

The current moderator feature should continue with direct, explicit domain rules:

- users domain owns role, moderator status, and suspension facts
- markets domain owns proposal, approval, rejection, publication, amendment, resolution, and cancellation lifecycle
- bets domain owns only trade guard integration and accounting behavior
- frontend owns smoke-test and workflow presentation, not rule authority

This keeps the feature reviewable while preserving a clear path to a future game-engine abstraction.
