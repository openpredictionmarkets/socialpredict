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

- Embedding is default-deny unless the operator explicitly configures allowed frame ancestors.
- The first implementation should allow public, read-only pages only.
- Authenticated pages should remain unembeddable unless a route is explicitly approved.
- Admin pages should remain unembeddable by default.
- Trading and account-changing flows must remain unembeddable until they pass a separate clickjacking review.

If the product decision is literally site-wide iframe embedding, the implementation should still make the policy explicit through headers and configuration rather than removing protections silently.

For the initial feature slice, "entire webpage" means the full SocialPredict public app shell can be framed on approved host domains. It does not mean every authenticated, admin, trading, or account-changing route is automatically frameable.

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

The first shareable market states should be the public market states already safe for public lists and public detail pages. Proposed, rejected, private, admin-only, or cancelled markets should return safe fallback or not-found behavior instead of market-specific share metadata. Resolved or closed markets may be shareable only if they are still public market detail pages in the product.

## Metadata Source Of Truth

Market preview metadata should be generated from backend-visible public market data.

Required fields:

- market title
- market description or safe summary
- canonical public market URL
- public market status if included in copy
- market image URL or generated fallback image

These fields should be exposed through an explicit public share metadata read model, such as `ShareMetadata`, owned by the Prediction Market/API boundary. Rendering code may consume that read model, but it must not duplicate market visibility rules.

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
- Runtime/proxy configuration should own the final browser-enforced frame headers.
- Backend share shells or frontend routes must not silently override the runtime/proxy frame policy.
- Headers such as `X-Frame-Options` and `Content-Security-Policy: frame-ancestors` must be reviewed together.
- Allowing all ancestors with `frame-ancestors *` should be treated as a deliberate public-demo-style policy, not a default production posture.
- Market previews must not expose private moderator, admin, account, or audit data.
- Rejected/proposed moderator markets should not produce public share cards unless a separate private/admin sharing policy is designed.
- Preview descriptions should be sanitized and length-bounded.
- Preview images should not allow arbitrary external script or HTML injection.
- Canonical URLs and image URLs must be absolute, public, and derived from environment-owned public base URL configuration.
- Operators should expect social crawlers to cache metadata; stale preview behavior after market edits should be documented.

## Operator Configuration

The first implementation uses runtime-owned configuration for deployment boundaries, plus CMS-managed defaults for share-card content:

```text
PUBLIC_BASE_URL
SECURITY_FRAME_ANCESTORS
SHARE_DEFAULT_IMAGE_URL
SHARE_SITE_NAME

CMS Social Share settings:
siteName
defaultDescription
defaultImageUrl
imageAlt
```

Expected defaults:

- `PUBLIC_BASE_URL` falls back to `DOMAIN_URL`, then `http://localhost`.
- `SECURITY_FRAME_ANCESTORS` defaults to `'none'`, which keeps iframe embedding denied.
- When `SECURITY_FRAME_ANCESTORS` is set to an explicit allowlist, the backend emits `Content-Security-Policy: frame-ancestors ...` and omits `X-Frame-Options` because `X-Frame-Options` cannot express a multi-origin allowlist.
- `SHARE_DEFAULT_IMAGE_URL` may be an absolute URL or a path under `PUBLIC_BASE_URL`; if unset, share cards use `/og/socialpredict-share.png`.
- `SHARE_SITE_NAME` defaults to `SocialPredict`.
- The admin dashboard Social Share CMS tab can upload a PNG, JPEG, or WebP image up to 5 MiB and writes the file under the backend upload directory and points share cards at `/api/v0/content/social-share/image`.
- The same tab can also override `SHARE_DEFAULT_IMAGE_URL`, `SHARE_SITE_NAME`, fallback description, image alt text, and whether local share-image metadata is enabled without a deploy.
- When local share-image metadata is disabled, market pages still emit title, description, and URL metadata, but the uploaded image endpoint returns `404` so operators can stop local crawler image traffic quickly.
- `SOCIAL_SHARE_UPLOAD_DIR` controls where uploaded share images are stored; production compose backs it with the `socialpredict_uploads` Docker volume.
- The default Open Graph image should be public and close to 1200x630px.
- The fallback description should usually be 110-160 characters; the backend allows up to 220 characters.

Example allowlist:

```text
SECURITY_FRAME_ANCESTORS="'self', https://partner.example"
```

After deployment, operators should verify:

```text
curl -I https://example.com/markets/1
curl https://example.com/markets/1 | grep -E 'og:title|og:description|og:image|og:image:alt'
curl -I https://example.com/api/v0/content/social-share/image # expect 200 only when local image sharing is enabled
```

Social preview crawlers may cache Open Graph metadata. If a market title, description, or image changes, an external preview/debugger tool may be needed to refresh the cached card.

## Acceptance Criteria

- Operators can configure whether SocialPredict pages may be embedded in iframes.
- Embedding is default-deny unless an operator-configured allowlist or explicit public-demo policy permits it.
- The embedding policy is visible in deployed response headers.
- Admin, authenticated account-changing, and trading routes remain unembeddable unless explicitly approved.
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
