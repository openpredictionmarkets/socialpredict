---
title: Market Taxonomy And Hierarchical Navigation
document_type: production-notes
domain: features
author: Patrick Delaney
updated_at: 2026-06-04T00:00:00Z
updated_at_display: "Thursday, June 4, 2026"
update_reason: "Start the feature spec for tags, category pages, featured market pinning, and CMS-driven market discovery."
status: draft
---

# Market Taxonomy And Hierarchical Navigation

## Purpose

SocialPredict needs a richer market discovery model than a single `/markets` page with search and status tabs.

Search should remain the primary entry point because it is direct, familiar, and works even when taxonomy content is sparse. Under that search-first entry point, the platform should support a CMS-managed market taxonomy: tags, top-level and secondary category pages, pinned featured markets, optional and curated page layouts.

This feature creates the design foundation for that hierarchy before implementation.

## Feature Artifact Map

This directory keeps the market taxonomy/navigation feature work together:

- [09-market-taxonomy-navigation.md](./09-market-taxonomy-navigation.md): feature overview, product behavior, and acceptance criteria.
- [DESIGN.md](./DESIGN.md): domain, CMS, persistence, search, and frontend architecture design.
- [PLAN.md](./PLAN.md): backend-first implementation checklist and PR sequencing.

## Product Shape

The market discovery hierarchy should support:

1. A top-level `/markets` page with search as the first control.
2. A small default recommendation panel under search, initially random active markets.
3. CMS-pinned featured markets on the top-level page.
4. CMS-pinned secondary category pages on the top-level page.
5. Secondary category pages organized by tag/category.
6. Search on each secondary page, filtered by that category/tag by default.
7. Featured/pinned markets within a page.
8. Market tags displayed as chips in search results, market cards, admin review, and market detail pages.
9. Moderator-selected tags during market creation, constrained to admin-managed tags.
10. Admin tag review during market approval.
11. Admin-only tag creation, deletion, ordering, and page layout management.

## Current Behavior

Current market discovery is mostly flat:

- `/markets` has a search bar and status tabs.
- Search can query across active, closed, resolved, or all public markets.
- Default tab content shows markets by status.
- There is no durable category/tag vocabulary.
- There is no CMS-managed page hierarchy.
- There is no way to pin markets or category pages.
- Moderator market creation does not attach category metadata.
- Admin approval does not review tags.

## Ubiquitous Language

- Tag: admin-managed label that can be attached to markets and used for search/filtering.
- Category page: CMS-managed market discovery page, usually backed by one primary tag or tag query.
- Secondary page: category page below the top-level `/markets` page.
- Featured market: market manually pinned to a page by an admin.
- Featured category: secondary category page manually pinned to the top-level markets page.
- Recommendation panel: automatic non-CMS fallback list, initially random active markets.
- Tag chip: compact UI label that shows market tags in lists/details/review flows.
- Taxonomy: the admin-managed set of tags, category pages, and layout rules.

Avoid confusing:

- Tag is not a free-form user hashtag.
- Category page is not necessarily a backend route hardcoded in React.
- Featured is manual CMS curation, not the same as algorithmic recommendation.
- Recommendation is automatic fallback content, not an endorsement unless later copy says so.

## Top-Level `/markets` Page

Search remains first.

Default layout when nothing is pinned:

1. Search bar.
2. Up to 20 recommended active markets, using the current loose/random behavior or a backend-owned equivalent.
3. Existing status navigation remains available: Active, Closed, Resolved, All.

Layout when CMS pinning exists:

1. Search bar.
2. Compact recommendation panel, about 5 markets.
3. Featured category/page cards.
4. Featured market cards.

The top-level page should not become blank just because CMS content is empty. Search and recommendation fallback keep the page useful from day one.

## Secondary Category Pages

A secondary category page should feel familiar relative to the top-level page.

Each secondary page should support:

- search bar at the top
- search filtered by the page's category/tag by default
- status tabs: Active, Closed, Resolved, All
- recommendation/fallback markets for that tag if nothing is pinned
- pinned featured markets
- tag chips on market cards/results

Default layout when nothing is pinned:

1. Category title/description.
2. Search bar filtered to the category.
3. Up to 20 random/recommended markets for that category.

Layout when page is curated:

1. Category title/description.
2. Search bar filtered to the category.
3. Compact recommendation panel, about 5 markets.
4. Pinned featured markets.

## Tag Assignment Flow

### Moderator Create Flow

Moderators should select tags from admin-managed options while creating a market.

Rules:

- moderators cannot create new tags
- moderators can select one or more available tags
- UI should support search/typeahead because tag lists can grow
- default can be no tag only if policy allows it
- future policy may require at least one tag before submission

### Admin Review Flow

Admin market review should show proposed tags.

Admins should be able to:

- approve tags as proposed
- add/remove tags before approval if policy allows
- see tag chips next to title/description/custom labels
- treat inappropriate tags as a reason for rejection or adjustment

### Published Market Display

Published markets should show tag chips:

- market detail page
- `/markets` list/search results
- category page results
- admin review/governance tables
- moderator profile market lists where useful

## CMS And Admin Management

Admins should own taxonomy management.

Admin capabilities:

- create tag
- edit tag display name/slug/description/color/order
- disable/archive tag
- delete tag only when safe or after confirmation
- create category page from a tag or tag query
- pin/unpin featured markets
- pin/unpin featured category pages

Deletion requires safety guardrails:

- require confirmation
- show count of markets using the tag
- prefer archive/disable over destructive delete
- prevent deletion if market references exist unless a migration/removal path is explicit

## Persistence Shape Candidate

Candidate tables:

- `market_tags`: admin-owned tag definitions.
- `market_tag_assignments`: many-to-many relation between markets and tags.
- `market_discovery_pages`: top/secondary page definitions.
- `market_discovery_pins`: pinned markets and pinned category pages, ordered by admin.

Important design choice: use additive timestamped Go migrations under `backend/migration/migrations`, following the existing migration convention.

## Search And Recommendation

Search should remain familiar.

Baseline behavior:

- existing market search continues to support text query and lifecycle/public status
- add optional tag/category filter
- tag-filtered search should use the same text matching semantics as unfiltered search
- category pages pass their tag filter automatically

Recommendation fallback can start simple:

- random active markets on top-level page
- random active markets matching the category tag on secondary pages
- limit 20 when no pinned content exists
- limit 5 when pinned content exists

This should be backend-owned enough that frontend does not invent recommendation rules independently.

## Design-Agent Cross-Reference

Evans/domain posture:

- Treat taxonomy terms as ubiquitous language before schema work.
- Separate tags, category pages, featured markets, and recommendations because each has a different business meaning.
- Do not let React route structure define the domain taxonomy.

Fowler/evolutionary posture:

- Start with tags and page read models before building a large CMS platform.
- Prefer reversible admin curation tables and simple random recommendations before a real recommender system.
- Keep a path from flat `/markets` to hierarchical pages without breaking the existing search/status tabs.

Martin/clean-architecture posture:

- Domain/use-case services own tag policy, page composition, and market filtering.
- Handlers adapt HTTP to use cases; frontend consumes read models.
- Database tables and React components are details around the taxonomy/navigation policy.

## Acceptance Criteria

Planning-level acceptance criteria:

- Feature docs define tags, category pages, featured markets, and recommendations distinctly.
- Database migration plan is additive and timestamped.
- Admin ownership of tag creation/deletion is explicit.
- Moderator tag selection is constrained to existing admin-managed tags.
- Admin review includes tag visibility and correction path.
- Top-level and secondary pages preserve search-first behavior.
- Empty CMS pages remain useful through recommendation fallback.
- Pinned content changes page layout but does not remove search.
- Tag chips appear in relevant market display surfaces.

## Out Of Scope For First Pass

- Machine-learning recommendation system.
- User-personalized recommendations.
- Public user-created tags.
- Tag moderation by non-admins.
- Full drag-and-drop CMS builder.
- SEO/URL slug migration strategy beyond basic category page slugs.
- Analytics-driven auto pinning.
