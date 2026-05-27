---
title: Embeddable Pages And Market Sharing
document_type: production-notes
domain: features
author: Patrick Delaney
updated_at: 2026-05-26T00:00:00Z
updated_at_display: "Tuesday, May 26, 2026"
update_reason: "Start the feature spec for iframe embedding and market-specific Open Graph sharing."
status: draft
---

# Embeddable Pages And Market Sharing

## Purpose

SocialPredict should be easier to distribute outside the main website.

This feature has two related goals:

1. Allow SocialPredict pages to be embedded in an iframe where the operator intentionally permits it.
2. Allow individual market pages to produce high-quality social/link previews using Open Graph metadata.

The Open Graph goal follows the public Open Graph sharing model described by OpenGraph.io: page-level metadata controls how links are previewed by social platforms and messaging clients. Reference: <https://www.opengraph.io/open-graph-meta-tags>.

This note is a feature-level spec. It cuts across frontend routing, backend market data, deployment security headers, CSP policy, server-side or edge-rendered metadata, and operator configuration.

## Feature Artifact Map

This directory keeps the embedding and sharing feature work together:

- [03-embeddable-pages-and-market-sharing.md](./03-embeddable-pages-and-market-sharing.md): feature overview, product behavior, and acceptance criteria.
- [DESIGN.md](./DESIGN.md): architecture and boundary design aligned with the canonical design plan.
- [PLAN.md](./PLAN.md): implementation sequencing and PR slicing plan derived from the design.

## Product Outcomes

Embeddable pages should enable:

- A partner, publication, classroom, or community site to embed SocialPredict pages.
- Operators to intentionally allow or deny iframe embedding by environment/domain.
- Future embeddable market widgets without immediately designing a widget platform.

Market sharing should enable:

- A user to share a market URL and get a useful preview card.
- Preview cards to include market title, description, canonical URL, and preview image.
- Market previews to use backend-owned market language instead of frontend-only approximations.

## Embed Scope

The user-facing goal is to allow the entire webpage to be embedded as an iframe.

The safety interpretation is:

- Public pages may be embeddable when operator policy permits it.
- Authenticated pages should require explicit review before being embeddable.
- Admin pages should not become embeddable by accident.
- Trading and account-changing flows must be reviewed for clickjacking risk before broad iframe enablement.

If the product decision is literally site-wide iframe embedding, the implementation should still make the policy explicit through headers and configuration rather than removing protections silently.

## Sharing Scope

The first market-sharing target is individual market detail pages, such as:

```text
/markets/{marketId}
```

Each shareable market page should provide at minimum:

```html
<meta property="og:title" content="..." />
<meta property="og:type" content="website" />
<meta property="og:url" content="..." />
<meta property="og:description" content="..." />
<meta property="og:image" content="..." />
```

Likely optional tags:

```html
<meta property="og:site_name" content="SocialPredict" />
<meta property="og:locale" content="en_US" />
<meta name="twitter:card" content="summary_large_image" />
<meta name="twitter:title" content="..." />
<meta name="twitter:description" content="..." />
<meta name="twitter:image" content="..." />
```

Open Graph tags must reflect the canonical public market state, not private moderator/admin state.

## Metadata Source Of Truth

Market preview metadata should be generated from backend-visible public market data.

Required fields:

- market title
- market description or safe summary
- canonical public market URL
- public market status if included in copy
- market image URL or generated fallback image

The frontend may render the market page, but social crawlers often do not execute the full React app reliably. The implementation must decide how metadata reaches crawlers before relying on client-side mutation of `<meta>` tags.

Candidate metadata delivery options:

1. Backend-served market share page that emits HTML metadata before the SPA loads.
2. Edge/proxy route that injects metadata for `/markets/{marketId}`.
3. Server-side rendered share shell for market routes.
4. Static default tags plus a later unfurling endpoint, if rich market cards are deferred.

Client-only React Helmet-style updates are likely insufficient as the primary sharing mechanism unless tested against target crawlers.

## Security And Privacy

Iframe embedding and link previews are security-sensitive surfaces.

Rules:

- Embedding policy must be intentional and environment-configurable.
- Headers such as `X-Frame-Options` and `Content-Security-Policy: frame-ancestors` must be reviewed together.
- Allowing all ancestors with `frame-ancestors *` should be treated as a deliberate public-demo-style policy, not a default production posture.
- Market previews must not expose private moderator, admin, account, or audit data.
- Rejected/proposed moderator markets should not produce public share cards unless a separate private/admin sharing policy is designed.
- Preview descriptions should be sanitized and length-bounded.
- Preview images should not allow arbitrary external script or HTML injection.

## Acceptance Criteria

- Operators can configure whether SocialPredict pages may be embedded in iframes.
- The embedding policy is visible in deployed response headers.
- Public market detail URLs emit Open Graph metadata with title, description, URL, type, and image.
- Market metadata is based on public market data only.
- Proposed, rejected, cancelled, private, or admin-only states do not leak through public Open Graph tags.
- Sharing previews work in at least one external Open Graph preview tester before the feature is called complete.
- Documentation explains which environments allow embedding and why.

## Non-Goals

This feature does not initially require:

- A full widget SDK.
- Per-market embed customization UI.
- Partner-specific OAuth or signed embeds.
- Server-side rendering for the entire frontend unless metadata delivery proves it necessary.
- Personalized Open Graph previews.
- Embedding admin dashboards without separate security review.
