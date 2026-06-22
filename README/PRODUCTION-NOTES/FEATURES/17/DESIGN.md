---
title: Grouped Activity Freshness Metadata Design
document_type: feature-design
domain: features
author: Patrick Delaney
updated_at: 2026-06-17T00:00:00Z
updated_at_display: "Wednesday, June 17, 2026"
update_reason: "Record implemented freshness semantics for grouped activity display endpoints."
status: implemented
---

# Grouped Activity Freshness Metadata Design

## Design Posture

Freshness metadata is part of the display/read-model boundary. It must not imply transaction safety. Transaction confirmations remain authoritative.

## Options

| Option | Benefit | Cost |
| --- | --- | --- |
| Populate freshness from child snapshots | Consistent with cached display posture. | Need aggregation semantics across children. |
| Populate live freshness | Honest if endpoints compute live every request. | Does not reduce load. |
| Remove freshness until grouped snapshots exist | Avoids misleading UI. | Less user visibility into stale display state. |

## Preferred Direction

Short term: if grouped activity is computed live, return freshness with `source = live`, `transactionSafeRead = false`, and generated-at time.

Long term: move grouped activity into read models and aggregate freshness from child snapshots.

## Implemented Contract

Grouped activity endpoints currently compute live. They therefore return freshness metadata with:

- `source = live`
- `targetFreshnessSeconds = 0`
- `transactionSafeRead = false`
- `generatedAt` set at response generation time

This is intentionally honest rather than optimized: users get a clear freshness signal, while the UI does not imply the response is a transaction-safe source of truth.
