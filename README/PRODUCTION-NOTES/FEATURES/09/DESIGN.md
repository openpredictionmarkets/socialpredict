---
title: Market Taxonomy And Hierarchical Navigation Design
document_type: feature-design
domain: features
author: Patrick Delaney
updated_at: 2026-06-04T00:00:00Z
updated_at_display: "Thursday, June 4, 2026"
update_reason: "Define backend-first taxonomy, CMS, search, recommendation, and page-composition design before implementation."
status: draft
---

# Market Taxonomy And Hierarchical Navigation Design

## Design Position

Market taxonomy is a discovery and curation capability. It is not just UI decoration.

Tags affect market creation, admin review, search filtering, page composition, public navigation, and future CMS layout. The backend should own tag definitions, tag assignments, page definitions, pinning, and recommendation/fallback policy. The frontend should render backend-owned read models and provide admin/moderator workflows.

## Design Inputs

Primary inputs:

- [09-market-taxonomy-navigation.md](./09-market-taxonomy-navigation.md)
- Canonical design plan: `/Users/patrick/Projects/spec-socialpredict-tasks/lib/design/design-plan.json`
- Designer-agent postures from `/Users/patrick/Projects/spec-socialpredict-tasks/.codex/agents/`
- Existing `/markets` search/status-tab behavior
- Existing moderator create and admin review flows

## Problem Framing

The flat `/markets` page works as an early discovery surface because search is easy and status tabs are simple. It does not scale well once the platform has many markets across themes, leagues, geographies, communities, or product areas.

The next design step is a taxonomy and page-composition model that preserves search-first discovery while allowing admins to curate market discovery pages.

## Business Outcomes

- Users can quickly search all markets from the top-level page.
- Users can browse meaningful categories without knowing exact search terms.
- Admins can curate featured markets and category pages.
- Moderators can tag proposed markets using a controlled vocabulary.
- Admins can review and correct tags before publication.
- The platform can grow toward CMS-driven market pages without a large rewrite.

## Boundary Alignment

| Boundary | Responsibility |
| --- | --- |
| Prediction Market Context | Owns market-tag assignment policy and public market filtering by tags. |
| CMS / Content Context | Owns discovery pages, page copy, pinning, and layout order. |
| Participant Account Context | Owns admin/moderator authority facts used by taxonomy management and tag assignment workflows. |
| API And Auth Boundary | Owns admin taxonomy endpoints, moderator create/review payloads, public page/search endpoints, and OpenAPI schemas. |
| Frontend Experience Context | Owns top-level and secondary page presentation, tag chips, admin taxonomy UI, and moderator tag selection UI. |
| Repository And Migration Boundary | Owns additive tag/page/pin tables and timestamped Go migrations. |

## Core Domain Model

### Tag

Admin-managed label attached to markets.

Candidate fields:

- `id`
- `slug`
- `display_name`
- `description`
- `color_key` or `style_key`
- `sort_order`
- `is_active`
- `created_by`
- `created_at`
- `updated_at`

Rules:

- slugs are unique and stable enough for URLs
- display names can change
- inactive tags remain attached historically but cannot be newly selected unless admin reactivates them
- hard delete should be restricted when assignments exist

### Market Tag Assignment

Many-to-many relation between market and tag.

Candidate fields:

- `market_id`
- `tag_id`
- `assigned_by`
- `assigned_at`
- `source`: moderator_proposed, admin_adjusted, migration, system

Rules:

- assignments should be unique per market/tag
- admin can adjust during review
- moderator can propose only active tags

### Discovery Page

CMS-managed navigation page.

Candidate fields:

- `id`
- `slug`
- `title`
- `description`
- `page_type`: top, category
- `primary_tag_id` nullable
- `query_mode`: tag, tag_set, custom, all
- `sort_order`

Rules:

- top page can be represented as one special page or as a conventional read model for `/markets`
- secondary pages usually have a primary tag
- page edits publish immediately in the current CMS model

### Pin

Manual CMS curation of a market or category page.

Candidate fields:

- `id`
- `scope_type`: page
- `scope_id`
- `pin_type`: market, discovery_page
- `market_id` nullable
- `target_page_id` nullable
- `sort_order`
- `label` nullable
- `created_by`
- `created_at`

Rules:

- pinned markets must be visible/tradable or explicitly allowed by policy
- pinned category pages must be published
- pins are ordered manually
- pins should degrade gracefully if a target market is cancelled/resolved/hidden

## Page Composition Read Model

The backend should expose a composed read model for discovery pages instead of making React assemble the taxonomy from many unrelated calls.

Candidate response shape:

```json
{
  "page": {
    "slug": "markets",
    "title": "Markets",
    "description": "Browse and search markets"
  },
  "search": {
    "defaultTagSlug": null,
    "statusFilters": ["active", "closed", "resolved", "all"]
  },
  "recommendations": {
    "mode": "random_active",
    "limit": 5,
    "markets": []
  },
  "featuredPages": [],
  "featuredMarkets": []
}
```

Frontend pages should render this read model and call search endpoints with `tag` or `page` filters.

## Search Design

Search remains the primary entry point.

Rules:

- top-level search defaults to all public markets unless status tab narrows it
- secondary page search defaults to that page's primary tag/category
- users can still switch Active/Closed/Resolved/All within the scoped search
- search results include tags as chips
- empty query falls back to recommendation/page composition, not a failed search state

## Recommendation Design

Initial recommendation is intentionally simple.

Recommended baseline:

- random active markets
- filtered by page tag on secondary pages
- limit 20 when no pinned content exists
- limit 5 when pinned content exists

This is a fallback/recommendation seam, not a personalized recommendation platform. Keep it small and replaceable.

## Moderator And Admin Workflows

### Moderator Create

The create-market use case should accept tag IDs/slugs as proposed market metadata.

Backend rules:

- validate tags exist and are active
- persist assignments with source `moderator_proposed`
- reject invalid tag IDs/slugs with a public validation reason

### Admin Review

Admin review should display proposed tags and allow correction.

Backend rules:

- admin can add/remove tags while proposal is pending
- approved market carries final tag assignments into publication
- tag changes should be auditable or reconstructable

### Admin CMS

Admin taxonomy management should be separate from ordinary market review.

Admin can manage:

- tags
- discovery pages
- pins

## Migration Design

Use additive timestamped Go migrations under `backend/migration/migrations`.

Candidate migration slices:

1. Tag definitions and assignments.
2. Market create/review/read payload support for tags.
3. Discovery pages .
4. Pins and page composition read model.

Avoid combining all tables and all UI behavior in one PR unless the migration is purely schema-only and low-risk.

## Designer Lens Review

Evans/domain lens:

- The domain language must distinguish tag, category page, pin, featured market, and recommendation.
- Taxonomy is shared language between moderators, admins, and market browsers.
- CMS curation and market lifecycle are related but not the same bounded context.

Fowler/evolutionary lens:

- Keep the first implementation narrow: tags plus chips, then page composition, then CMS pinning.
- Avoid committing to a recommendation platform; random active markets are a reversible seam.
- Composed read models reduce frontend churn while preserving an evolutionary backend path.

Martin/clean-architecture lens:

- Tag validation and page composition are use cases, not React conditionals.
- Database tables are details; use cases should expose simple request/response models.
- Admin/moderator authority is an input to use cases, not a UI-only permission flag.

## Risks

| Risk | Mitigation |
| --- | --- |
| CMS scope creep | Start with tags and simple page composition; defer full page builder. |
| Tag deletion breaks history | Prefer archive/disable; restrict delete with references. |
| Frontend invents taxonomy rules | Backend composed read model and API contracts. |
| Search performance degrades | Add indexes on tag assignments, slug, lifecycle/status, and consider query plans before large datasets. |
| Confusing page/category/tag terms | Maintain glossary and display language in docs/API. |
| Migration ordering conflicts | Stack from #734 and use timestamped additive migrations later. |

## Open Questions

- Should tags be single-level only, or should parent/child tag hierarchy exist?
- Should secondary category pages always map to one tag, or support tag queries?
- Should moderators be required to select at least one tag?
- Can admins add tags during approval that moderators did not propose?
- Should tags have colors from a fixed style guide or arbitrary admin-selected colors?
- Should public users ever see inactive tags on historical markets?
- Should cancelled/yanked markets be excluded from pins automatically?
- Should recommendation randomness be stable per day/session or fully random per request?

## Design Exit Criteria

- Tag, page, pin, and recommendation language is stable enough for API names.
- Database migration slices are chosen.
- Backend-first API contracts are drafted before frontend UI implementation.
- Admin tag deletion policy is decided.
- Empty-page fallback behavior is decided.
- Category search semantics are decided.
