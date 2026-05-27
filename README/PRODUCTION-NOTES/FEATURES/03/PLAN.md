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

- [x] 01. Feature artifact and design alignment
- [x] 02. Embedding policy decision and environment config
- [x] 03. Runtime/proxy frame-header implementation
- [x] 04. Route-scope and clickjacking safety review
- [x] 05. Market share metadata source and response contract
- [x] 06. Initial HTML metadata delivery path
- [x] 07. Open Graph and Twitter card tag coverage
- [x] 08. Share-card image policy
- [ ] 09. Verification tests and external preview validation
- [x] 10. Operator and user documentation

## Implementation Checklist

### 01. Feature Artifact And Design Alignment

Status: complete.

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

Status: complete for the first implementation slice.

Checklist:

- [x] Set production default to deny framing unless an explicit allowlist is configured.
- [x] Decide staging/demo allowlists separately if they differ from production.
- [x] Add typed config or deployment variables for allowed frame ancestors.
- [x] Document first-pass route scope as public, read-only pages.
- [x] Document that authenticated, admin, trading, and account-changing pages are excluded unless separately approved.

Exit criteria:

- [x] Embed behavior is a visible operator choice.
- [x] No environment silently changes frame policy.

Validation:

- [x] Runtime config tests prove expected frame policy defaults and allowlist behavior.

### 03. Runtime/Proxy Frame-Header Implementation

Service ownership: Runtime Bootstrap and Infrastructure.

Status: complete for the first implementation slice.

Checklist:

- [x] Locate current frame-related headers in app, nginx, Traefik, or deployment config.
- [x] Make runtime/proxy configuration the owner of `Content-Security-Policy: frame-ancestors`.
- [x] Decide whether `X-Frame-Options` should be removed or adjusted when CSP frame policy is active.
- [x] Add header behavior for permitted environments.
- [x] Ensure backend share-shell responses do not override frame policy independently.
- [x] Add tests or smoke checks for final response headers.

Exit criteria:

- [x] Embeddable environments emit headers that allow intended iframe use.
- [x] Restricted environments emit headers that deny unintended iframe use.

Validation:

- [ ] `curl -I <domain>` shows expected frame policy after deployment.
- [ ] Browser iframe smoke test succeeds or fails as expected after deployment.

### 04. Route-Scope And Clickjacking Safety Review

Service ownership: Frontend Experience Context plus API/Auth Contract Boundary.

Status: complete for the first implementation slice.

Checklist:

- [x] Inventory routes that can be framed.
- [x] Identify account-changing and trade-changing routes.
- [x] Decide whether market detail pages can be framed before trading forms are framed.
- [x] Keep admin routes unembeddable by default.
- [x] Add route-level or header-level exceptions if needed.

Exit criteria:

- [x] The first iframe release does not accidentally embed sensitive/admin routes.
- [x] Any site-wide embedding decision has documented risk acceptance.

Validation:

- [x] Manual route review.
- [x] Header config checks for representative public defaults.

### 05. Market Share Metadata Source And Response Contract

Service ownership: Prediction Market Context and API/Auth Contract Boundary.

Status: complete for the first implementation slice.

Checklist:

- [x] Define a `ShareMetadata` public read model or equivalent helper.
- [x] Define public market metadata fields on that seam.
- [x] Ensure metadata source uses public market visibility rules.
- [x] Define which market statuses are shareable.
- [x] Treat proposed, rejected, cancelled, private, and admin-only markets as not shareable.
- [x] Allow resolved or closed markets only if they remain public market detail pages.
- [x] Add not-found and non-public market behavior.
- [x] Sanitize and length-bound title/description output.
- [x] Source canonical URL and image URL from typed public base URL configuration.

Exit criteria:

- [x] Share metadata cannot expose proposed, rejected, admin-only, or private data.
- [x] Metadata values are deterministic for a given public market.

Validation:

- [x] Backend tests for public and non-public markets.

### 06. Initial HTML Metadata Delivery Path

Service ownership: Runtime Bootstrap and Infrastructure plus Frontend Experience Context.

Status: complete for the first implementation slice.

Checklist:

- [x] Choose backend share shell, proxy injection, or proven client metadata path.
- [x] If using backend share shell, route `/markets/{id}` or an equivalent share route to HTML metadata.
- [x] Ensure the SPA still loads normally after metadata is emitted.
- [x] Ensure canonical URLs remain stable.
- [x] Consume the share metadata seam rather than duplicating market visibility rules in rendering code.

Exit criteria:

- [x] A crawler that does not execute React can read market metadata from initial HTML.
- [x] Normal browser navigation still reaches the market detail page.

Validation:

- [x] Backend route tests prove `<market-url>` includes Open Graph tags.
- [ ] Browser navigation to market details still works after deployment.

### 07. Open Graph And Twitter Card Tag Coverage

Service ownership: Frontend Experience Context and API/Auth Contract Boundary.

Status: complete for the first implementation slice.

Checklist:

- [x] Emit `og:title`.
- [x] Emit `og:type`.
- [x] Emit `og:url`.
- [x] Emit `og:description`.
- [x] Emit `og:image`.
- [x] Consider `og:site_name` and `og:locale`.
- [x] Consider Twitter card tags for `summary_large_image`.
- [x] Ensure all URLs are absolute.

Exit criteria:

- [x] Required Open Graph tags are present for public market URLs.
- [x] Values match public market language.

Validation:

- [x] Unit or contract tests inspect generated HTML.
- [ ] External preview tool shows expected card after deployment.

### 08. Share-Card Image Policy

Service ownership: Frontend Visual System Boundary plus CMS content operations.

Status: complete for the CMS-managed default-image slice.

Checklist:

- [x] Decide static default image vs generated per-market image.
- [x] Ensure image URL is absolute and publicly reachable after share metadata expansion.
- [x] Ensure image dimensions are suitable for common social previews.
- [x] Keep image generation separate from market domain policy.
- [x] Add fallback image for missing/invalid market images.
- [x] Add admin CMS settings for site name, fallback description, default image URL, and image alt text.

Exit criteria:

- [x] Every shareable market has a valid `og:image`.
- [x] Image policy is documented and testable.

Validation:

- [x] Backend tests cover CMS social-share settings and generated metadata.
- [x] Frontend production build.
- [ ] External request to image URL returns expected content type and success status after deployment.

### 09. Verification Tests And External Preview Validation

Service ownership: Release and Deployment Control.

Checklist:

- [x] Add tests for generated Open Graph HTML.
- [x] Add tests for non-public market metadata behavior.
- [ ] Add deployment smoke checks for frame headers.
- [ ] Add an external preview validation step or manual release checklist.
- [ ] Document expected crawler cache/staleness behavior after market edits.
- [ ] Record expected behavior in PR templates or release docs if needed.

Exit criteria:

- [ ] CI or release validation catches missing required tags.
- [ ] Operators can verify iframe behavior and share previews after deploy.

Validation:

- [x] Backend tests as appropriate.
- [x] Frontend production build.
- [ ] `curl -I` header checks.
- [ ] External Open Graph preview tool check.

### 10. Operator And User Documentation

Service ownership: documentation and release operations.

Checklist:

- [x] Document embed policy per environment.
- [x] Document how to allowlist frame ancestors.
- [x] Document market sharing behavior.
- [x] Document limitations for proposed/rejected/private markets.
- [x] Document how to test share cards.
- [x] Document known crawler cache behavior and how to refresh previews where possible.

Exit criteria:

- [x] Maintainers can configure and validate embedding.
- [x] Maintainers can diagnose missing or stale social previews.

Validation:

- [x] Docs review.
