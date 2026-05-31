---
title: Frontend Accessibility Baseline
document_type: production-notes
domain: frontend
author: Patrick Delaney
updated_at: 2026-05-24T00:00:00Z
updated_at_display: "Sunday, May 24, 2026"
update_reason: "Record the first accessibility pass and identify the remaining market, trade, profile, and admin workflow audit."
status: draft
---

# Frontend Accessibility Baseline

## Purpose

This active note covers accessibility work for the current UI after the first baseline pass.

Start with [00-TRIAGE.md](./00-TRIAGE.md). Accessibility is a participant-access contract, not visual polish and not a reason to begin a full component-library rewrite.

Long-term accessibility-program work lives in [FUTURE/05-long-term-accessibility-program.md](./FUTURE/05-long-term-accessibility-program.md).

## Current Baseline

Create-market, change-password, global fallback, and navigation controls now have a first pass of explicit labels, status/error announcements, menu button names, and datetime field labeling. Manual review confirmed create-market submission, route guard behavior, change-password layout, and datetime label behavior.

The frontend still needs a broader workflow audit.

The frontend has domain-critical workflows that must be usable without pointer-only or color-only interaction:

- Browse markets.
- Create market.
- Buy exposure.
- Sell exposure.
- Resolve market.
- View profile/account.
- View positions.

## Active Direction

1. Audit the remaining core workflows above, especially buy, sell, resolve, profile/account, market details, and admin content editing.
2. Fix obvious missing labels, focus behavior, keyboard traps, semantic structure, and error announcements.
3. Ensure outcome/state display does not rely on color alone.
4. Add automated checks only after frontend test tooling is declared.
5. Keep the next pass local to existing screens and components.

## Design Plan Alignment

The canonical design plan tracks this as:

- `Frontend Accessibility Interaction Seam`
- `Accessibility Interaction Contract`
- `Frontend Visual System Boundary`

## Active Acceptance Criteria

- Create-market and change-password forms have labels and usable focus behavior.
- Primary navigation and domain workflows can be operated by keyboard.
- Screen-reader users are not blocked by obvious missing semantic structure.
- Error and success messages use status or alert semantics where they are introduced.
- Market/trading state is not conveyed by color alone.
- Any automated check added to CI is stable and scoped.

## Explicitly Deferred

- Full WCAG audit program.
- Cross-browser assistive technology matrix.
- Visual regression or accessibility snapshot platform.
- Component-library rewrite.

See [FUTURE/05-long-term-accessibility-program.md](./FUTURE/05-long-term-accessibility-program.md).
