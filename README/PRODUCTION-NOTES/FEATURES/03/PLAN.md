---
title: Embeddable Pages And Market Sharing Implementation Plan
document_type: feature-plan
domain: features
author: Patrick Delaney
updated_at: 2026-05-26T00:00:00Z
updated_at_display: "Tuesday, May 26, 2026"
update_reason: "Add the implementation sequence for iframe embedding and market-specific Open Graph sharing."
status: draft
---

# Embeddable Pages And Market Sharing Implementation Plan

## Purpose

This plan turns [DESIGN.md](./DESIGN.md) and [03-embeddable-pages-and-market-sharing.md](./03-embeddable-pages-and-market-sharing.md) into an implementation sequence.

The plan is intentionally split into reviewable slices. Iframe embedding touches deployment security headers and frontend route behavior. Market sharing touches public market data, metadata rendering, route/proxy behavior, and external preview verification.

Agents implementing this feature should mark checklist items as they complete them and leave unchecked items in place when intentionally deferred.

## Branching Rule

This feature plan should branch from PR #713 (`chore/moderator-design-alignment`) until that PR lands in `main`.

After #713 merges:

- rebase or recreate this branch from `main`
- retarget the PR to `main`
- keep this feature independent from the resettable-demo docs PR unless the user explicitly asks to stack them

## Planning Principles

- Treat embedding as an explicit operator/deployment policy, not a removed header by accident.
- Use default-deny embedding unless a route and frame ancestor are explicitly allowed.
- Treat public market metadata as backend/domain-owned public market language.
- Do not leak proposed, rejected, admin-only, account, or audit data into public share tags.
- Prefer initial HTML metadata over client-only metadata mutation unless crawler tests prove otherwise.
- Add a small share metadata seam before adding HTML rendering so visibility rules stay in the market/API boundary.
- Keep the first implementation narrow enough to ship without designing a widget SDK.
- Add tests for both header policy and generated metadata.
- Verify against at least one external Open Graph preview tool before calling the feature done.

## Progress Ledger

- [ ] 01. Feature artifact and design alignment
- [ ] 02. Embedding policy decision and environment config
- [ ] 03. Runtime/proxy frame-header implementation
- [ ] 04. Route-scope and clickjacking safety review
- [ ] 05. Market share metadata source and response contract
- [ ] 06. Initial HTML metadata delivery path
- [ ] 07. Open Graph and Twitter card tag coverage
- [ ] 08. Share-card image policy
- [ ] 09. Verification tests and external preview validation
- [ ] 10. Operator and user documentation

## Implementation Checklist

### 01. Feature Artifact And Design Alignment

Status: complete for this documentation PR.

Checklist:

- [x] Create `README/PRODUCTION-NOTES/FEATURES/03/`.
- [x] Add feature overview document.
- [x] Add `DESIGN.md` aligned with canonical frontend, API, market, runtime, and release boundaries.
- [x] Add `PLAN.md` as an agent-usable implementation sequence.
- [x] Update production-note index links.
- [x] Keep the change documentation-only.

Exit criteria:

- [x] Documentation separates iframe embedding policy from market sharing metadata.
- [x] Documentation identifies crawler limitations for client-only metadata.
- [x] Documentation records the temporary branch dependency on PR #713.

Validation:

- [x] `git diff --check`

### 02. Embedding Policy Decision And Environment Config

Service ownership: Runtime Bootstrap and Infrastructure plus Release and Deployment Control.

Checklist:

- [ ] Set production default to deny framing unless an explicit allowlist is configured.
- [ ] Decide staging/demo allowlists separately if they differ from production.
- [ ] Add typed config or deployment variables for allowed frame ancestors.
- [ ] Document first-pass route scope as public, read-only pages.
- [ ] Document that authenticated, admin, trading, and account-changing pages are excluded unless separately approved.

Exit criteria:

- [ ] Embed behavior is a visible operator choice.
- [ ] No environment silently changes frame policy.

Validation:

- [ ] Config tests or workflow checks prove expected frame policy per environment.

### 03. Runtime/Proxy Frame-Header Implementation

Service ownership: Runtime Bootstrap and Infrastructure.

Checklist:

- [ ] Locate current frame-related headers in app, nginx, Traefik, or deployment config.
- [ ] Make runtime/proxy configuration the owner of `Content-Security-Policy: frame-ancestors`.
- [ ] Decide whether `X-Frame-Options` should be removed or adjusted when CSP frame policy is active.
- [ ] Add header behavior for permitted environments.
- [ ] Ensure backend share-shell responses do not override frame policy independently.
- [ ] Add tests or smoke checks for final response headers.

Exit criteria:

- [ ] Embeddable environments emit headers that allow intended iframe use.
- [ ] Restricted environments emit headers that deny unintended iframe use.

Validation:

- [ ] `curl -I <domain>` shows expected frame policy.
- [ ] Browser iframe smoke test succeeds or fails as expected.

### 04. Route-Scope And Clickjacking Safety Review

Service ownership: Frontend Experience Context plus API/Auth Contract Boundary.

Checklist:

- [ ] Inventory routes that can be framed.
- [ ] Identify account-changing and trade-changing routes.
- [ ] Decide whether market detail pages can be framed before trading forms are framed.
- [ ] Keep admin routes unembeddable by default.
- [ ] Add route-level or header-level exceptions if needed.

Exit criteria:

- [ ] The first iframe release does not accidentally embed sensitive/admin routes.
- [ ] Any site-wide embedding decision has documented risk acceptance.

Validation:

- [ ] Manual route review.
- [ ] Header checks for representative public, authenticated, and admin routes.

### 05. Market Share Metadata Source And Response Contract

Service ownership: Prediction Market Context and API/Auth Contract Boundary.

Checklist:

- [ ] Define a `ShareMetadata` public read model or equivalent helper.
- [ ] Define public market metadata fields on that seam.
- [ ] Ensure metadata source uses public market visibility rules.
- [ ] Define which market statuses are shareable.
- [ ] Treat proposed, rejected, cancelled, private, and admin-only markets as not shareable.
- [ ] Allow resolved or closed markets only if they remain public market detail pages.
- [ ] Add not-found and non-public market behavior.
- [ ] Sanitize and length-bound title/description output.
- [ ] Source canonical URL and image URL from typed public base URL configuration.

Exit criteria:

- [ ] Share metadata cannot expose proposed, rejected, admin-only, or private data.
- [ ] Metadata values are deterministic for a given public market.

Validation:

- [ ] Backend tests for public and non-public markets.

### 06. Initial HTML Metadata Delivery Path

Service ownership: Runtime Bootstrap and Infrastructure plus Frontend Experience Context.

Checklist:

- [ ] Choose backend share shell, proxy injection, or proven client metadata path.
- [ ] If using backend share shell, route `/markets/{id}` or an equivalent share route to HTML metadata.
- [ ] Ensure the SPA still loads normally after metadata is emitted.
- [ ] Ensure canonical URLs remain stable.
- [ ] Consume the share metadata seam rather than duplicating market visibility rules in rendering code.

Exit criteria:

- [ ] A crawler that does not execute React can read market metadata from initial HTML.
- [ ] Normal browser navigation still reaches the market detail page.

Validation:

- [ ] `curl <market-url>` includes Open Graph tags.
- [ ] Browser navigation to market details still works.

### 07. Open Graph And Twitter Card Tag Coverage

Service ownership: Frontend Experience Context and API/Auth Contract Boundary.

Checklist:

- [ ] Emit `og:title`.
- [ ] Emit `og:type`.
- [ ] Emit `og:url`.
- [ ] Emit `og:description`.
- [ ] Emit `og:image`.
- [ ] Consider `og:site_name` and `og:locale`.
- [ ] Consider Twitter card tags for `summary_large_image`.
- [ ] Ensure all URLs are absolute.

Exit criteria:

- [ ] Required Open Graph tags are present for public market URLs.
- [ ] Values match public market language.

Validation:

- [ ] Unit or contract tests inspect generated HTML.
- [ ] External preview tool shows expected card.

### 08. Share-Card Image Policy

Service ownership: Frontend Visual System Boundary.

Checklist:

- [ ] Decide static default image vs generated per-market image.
- [ ] Ensure image URL is absolute and publicly reachable.
- [ ] Ensure image dimensions are suitable for common social previews.
- [ ] Keep image generation separate from market domain policy.
- [ ] Add fallback image for missing/invalid market images.

Exit criteria:

- [ ] Every shareable market has a valid `og:image`.
- [ ] Image policy is documented and testable.

Validation:

- [ ] External request to image URL returns expected content type and success status.

### 09. Verification Tests And External Preview Validation

Service ownership: Release and Deployment Control.

Checklist:

- [ ] Add tests for generated Open Graph HTML.
- [ ] Add tests for non-public market metadata behavior.
- [ ] Add deployment smoke checks for frame headers.
- [ ] Add an external preview validation step or manual release checklist.
- [ ] Document expected crawler cache/staleness behavior after market edits.
- [ ] Record expected behavior in PR templates or release docs if needed.

Exit criteria:

- [ ] CI or release validation catches missing required tags.
- [ ] Operators can verify iframe behavior and share previews after deploy.

Validation:

- [ ] Backend/frontend tests as appropriate.
- [ ] `curl -I` header checks.
- [ ] External Open Graph preview tool check.

### 10. Operator And User Documentation

Service ownership: documentation and release operations.

Checklist:

- [ ] Document embed policy per environment.
- [ ] Document how to allowlist frame ancestors.
- [ ] Document market sharing behavior.
- [ ] Document limitations for proposed/rejected/private markets.
- [ ] Document how to test share cards.
- [ ] Document known crawler cache behavior and how to refresh previews where possible.

Exit criteria:

- [ ] Maintainers can configure and validate embedding.
- [ ] Maintainers can diagnose missing or stale social previews.

Validation:

- [ ] Docs review.
