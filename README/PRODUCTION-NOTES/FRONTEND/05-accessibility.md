---
title: Frontend Accessibility Baseline
document_type: production-notes
domain: frontend
author: Patrick Delaney
updated_at: 2026-05-23T00:00:00Z
updated_at_display: "Saturday, May 23, 2026"
update_reason: "Keep active accessibility work focused on current core workflows and move program-level accessibility into FUTURE."
status: draft
---

# Frontend Accessibility Baseline

## Purpose

This active note covers the first accessibility pass for the current UI.

Start with [00-TRIAGE.md](./00-TRIAGE.md). Accessibility is a participant-access contract, not visual polish and not a reason to begin a full component-library rewrite.

Long-term accessibility-program work lives in [FUTURE/05-long-term-accessibility-program.md](./FUTURE/05-long-term-accessibility-program.md).

## Current Baseline

The frontend has domain-critical workflows that must be usable without pointer-only or color-only interaction:

- Browse markets.
- Create market.
- Buy exposure.
- Sell exposure.
- Resolve market.
- View profile/account.
- View positions.

## Active Direction

1. Audit the core workflows above.
2. Fix obvious missing labels, focus behavior, keyboard traps, semantic structure, and error announcements.
3. Ensure outcome/state display does not rely on color alone.
4. Add automated checks only after the frontend CI baseline exists.
5. Keep the first pass local to existing screens and components.

## Design Plan Alignment

The canonical design plan tracks this as:

- `Frontend Accessibility Interaction Seam`
- `Accessibility Interaction Contract`
- `Frontend Visual System Boundary`

## Active Acceptance Criteria

- Core forms have labels and usable focus behavior.
- Primary navigation and domain workflows can be operated by keyboard.
- Screen-reader users are not blocked by obvious missing semantic structure.
- Market/trading state is not conveyed by color alone.
- Any automated check added to CI is stable and scoped.

## Explicitly Deferred

- Full WCAG audit program.
- Cross-browser assistive technology matrix.
- Visual regression or accessibility snapshot platform.
- Component-library rewrite.

See [FUTURE/05-long-term-accessibility-program.md](./FUTURE/05-long-term-accessibility-program.md).
