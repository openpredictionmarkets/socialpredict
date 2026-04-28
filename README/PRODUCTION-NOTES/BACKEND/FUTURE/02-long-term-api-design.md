---
title: Long-Term API Design
document_type: production-notes
domain: backend
author: Patrick Delaney
updated_at: 2026-04-26T16:16:21Z
updated_at_display: "Sunday, April 26, 2026 at 4:16 PM UTC"
update_reason: "Create a non-binding holding note for deferred API-platform ideas so the active API-design note can stay current-state-first."
status: future
---

# Long-Term API Design

## Purpose

This note is a holding area for deferred API-platform ideas that are not part of the active SocialPredict backend design plan, not part of the current production-note wave sequence, and not part of the current runnable task queue.

Its purpose is to preserve broader API ambitions without letting them distort the active near-term architecture in [06-api-design.md](/workspace/socialpredict/README/PRODUCTION-NOTES/BACKEND/06-api-design.md).

## Current Status

As of 2026-04-26:

- the active backend API note should stay focused on the live HTTP contract, route-family migration, auth-visible semantics, and OpenAPI parity
- the active design plan still treats API/auth work as a later bounded alignment wave after runtime, failure, database, and security prerequisites
- the ideas in this document are explicitly deferred until the current route-family and boundary cleanup is substantially more stable

This document is non-binding on the active design plan and on `TASKS.json`.

## Deferred Or Rejected-For-Now Topics

The following ideas may be interesting later, but they are not current architecture commitments and should not re-enter the active note or task queue without an explicit later decision.

### Richer hypermedia contract ideas

- HATEOAS links
- generic `self/next/prev/first/last` response links across the API
- client-discovery-driven resource navigation design

These ideas are deferred because the live backend does not currently have a client-discovery problem that justifies the added transport complexity.

### Broader content negotiation

- XML response support
- multi-format response negotiation
- generic content-negotiation middleware
- non-JSON-first contract expansion

These ideas are deferred because the active backend is still converging JSON/plain-text route behavior and already has a mixed enough surface without adding more transport variants.

### Universal response-wrapper architecture

- a universal `APIResponse` with fields like `data`, `meta`, `links`, `error`, `success`, `timestamp`, and `version`
- flag-day envelope conversion across every route
- success-body rewriting middleware that wraps handler output after the fact

These ideas are deferred because the backend currently has multiple valid route families, infra/documentation endpoints, and middleware-generated transport responses that do not fit a single immediate wrapper design.

### Advanced versioning platform work

- header-based version negotiation
- a general version manager
- `/v1`, `/v2`, and sunset/deprecation orchestration
- automatic version routing or cross-version compatibility middleware

These ideas are deferred because the live backend currently has one application namespace, `/v0`, and does not yet have a stable enough contract to justify broader version infrastructure.

### Generator-first API workflow

- Swagger/OpenAPI generation from annotations as the primary contract source
- generated-client-first workflow
- client SDK trees as active current-wave output
- tooling that starts owning the contract ahead of the route/source-of-truth order used in the repo today

These ideas are deferred because the current backend already has a hand-maintained `openapi.yaml`, embedded Swagger UI, and route/spec parity tests. The active need is contract accuracy, not generator ownership.

### API service layer or platform package

- a dedicated `api/` subsystem
- an API “service” domain
- a `standards.go` catch-all package for transport policy
- generic middleware/platform trees that compete with `server.go`, handlers, and the active contract boundary

These ideas are deferred because they would create a second architecture inside the monolith rather than clarify the existing request boundary.

## Entry Criteria

This note should only become active planning input after the current API/auth boundary is substantially more stable.

Reasonable entry criteria are:

- the active production notes through the current API note are aligned to live code
- route-family migration has reduced the current envelope/plain-text/raw-response drift materially
- `internal/service/auth.HTTPError` and similar transport leakage are no longer the dominant active boundary issue
- the `/v0` contract is stable enough that a real versioning or generation problem exists
- there is a concrete consumer, product, or operational reason for one of the deferred topics

## What Is Explicitly Not In The Current Queue

The following are explicitly not part of the current queue unless reactivated on purpose:

- HATEOAS rollout
- XML or broader content negotiation
- universal `APIResponse` middleware/platform buildout
- header-based versioning
- `/v1` rollout planning
- generated client trees as the main current output
- a top-level `api/` service/platform package tree

## Re-Entry Questions

Before pulling any of these topics into the active plan, SocialPredict should answer:

- What concrete problem are we solving beyond the current route-family inconsistency?
- Is the current `/v0` contract stable enough to justify a versioning or generator platform?
- Would the topic clarify the existing request boundary, or would it create a second architecture?
- Is there an actual client-discovery, interoperability, or partner-integration need, or is the idea still speculative?
- Can the topic be introduced incrementally without obscuring the current source-of-truth order?

## Relationship To The Active Note

The active note at [06-api-design.md](/workspace/socialpredict/README/PRODUCTION-NOTES/BACKEND/06-api-design.md) is the binding current-state-first architecture note.

This `FUTURE` note exists so that:

- the active note can stay grounded in the live contract
- the task queue can stay focused on route-family migration and OpenAPI parity
- long-term API ambitions are not lost
- deferred platform ideas do not get mistaken for near-term design commitments
