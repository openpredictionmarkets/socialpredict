---
title: Market Taxonomy And Hierarchical Navigation Plan
document_type: feature-plan
domain: features
author: Patrick Delaney
updated_at: 2026-06-04T00:00:00Z
updated_at_display: "Thursday, June 4, 2026"
update_reason: "Track first implementation slice for market tag persistence, admin tag management, moderator selection, and tag display."
status: in-progress
---

# Market Taxonomy And Hierarchical Navigation Plan

## Purpose

This plan turns [09-market-taxonomy-navigation.md](./09-market-taxonomy-navigation.md) and [DESIGN.md](./DESIGN.md) into a backend-first implementation sequence.

Agents implementing this feature should mark checklist items as they complete them and leave deferred work unchecked.

## Planning Principles

- Search stays first on top-level and secondary market pages.
- Tags and CMS page composition are backend-owned policy/read models.
- Admins manage tag vocabulary; moderators select from active tags.
- Additive timestamped migrations come before UI dependencies.
- Start with simple random active recommendations before a recommendation platform.
- Keep pinned content manual and auditable.
- Keep every PR independently reviewable.

## 01. Feature Artifact And Design Alignment

Checklist:

- [x] Create `README/PRODUCTION-NOTES/FEATURES/09/`.
- [x] Add feature overview.
- [x] Add design artifact.
- [x] Add implementation plan.
- [x] Cross-reference canonical design plan and designer-agent postures.
- [ ] Review final design with product/user-facing terminology before implementation.

## 02. Tag Persistence And Domain Model

Service ownership: prediction market context and repository/migration boundary.

Checklist:

- [x] Add timestamped migration for `market_tags`.
- [x] Add timestamped migration for `market_tag_assignments`.
- [x] Add indexes for tag slug and market/tag assignment lookup.
- [x] Add domain models for tag and assignment.
- [x] Add repository mapping tests.
- [x] Add service policy for active tag validation.
- [x] Prefer archive/disable over destructive delete when assignments exist.

## 03. Market Create And Admin Review Tag Flow

Service ownership: market creation/review use cases and API boundary.

Checklist:

- [x] Add tag IDs/slugs to create-market request.
- [x] Validate moderator-selected tags exist and are active.
- [x] Persist moderator-proposed tag assignments.
- [x] Include tag chips/data in admin market review payloads.
- [ ] Allow admin adjustment of tags during proposal review if policy approves.
- [x] Include tags in published market payloads.
- [ ] Update `backend/docs/openapi.yaml`.
- [x] Add handler/domain tests.
- [ ] Run schemathesis/go-kin validation once OpenAPI is updated.

## 04. Admin Tag Management

Service ownership: CMS/content context plus API/auth boundary.

Checklist:

- [x] Add admin list/create/update/archive tag APIs.
- [x] Add guarded delete or archive-only policy.
- [ ] Show market count before destructive tag action.
- [x] Require confirmation for delete/archive.
- [x] Add admin dashboard tag management UI.
- [x] Add tests for admin-only access and validation.

## 05. Search And Tag Filtering

Service ownership: market search/read model.

Checklist:

- [ ] Add optional tag filter to public market search.
- [ ] Add optional tag filter to status-based market listing.
- [x] Include tags in search/list result DTOs.
- [ ] Keep Active/Closed/Resolved/All behavior compatible.
- [ ] Add query/index tests for tag-filtered search.
- [ ] Verify performance posture for many tags/markets.

## 06. Discovery Pages And Sections

Service ownership: CMS/content context and composed read model.

Checklist:

- [x] Add admin CMS scaffold for TOP and SECONDARY market discovery layout options.
- [x] Add migration for `market_discovery_pages`.
- [x] Add migration for `market_discovery_sections`.
- [x] Add domain/read models for top-level and secondary category pages.
- [ ] Support implicit `All` section when no sections exist.
- [x] Add public page composition endpoint.
- [x] Add admin page layout management API.
- [ ] Add admin section management APIs.
- [ ] Add tests for published/unpublished pages and section ordering.

## 07. Pins And Featured Content

Service ownership: CMS/content context.

Checklist:

- [x] Add migration for `market_discovery_pins`.
- [ ] Support pinned markets by page and section.
- [ ] Support pinned secondary category pages on top-level page.
- [ ] Add ordering controls.
- [ ] Define behavior for cancelled/resolved/hidden pinned targets.
- [ ] Add admin pin/unpin APIs.
- [ ] Add tests for ordering and target validation.

## 08. Recommendation Fallback

Service ownership: market read model and page composition use case.

Checklist:

- [ ] Implement random active market fallback for top-level page.
- [ ] Implement tag-scoped random active fallback for secondary pages.
- [ ] Use limit 20 when no pinned content exists.
- [ ] Use limit 5 when pinned content exists.
- [ ] Decide stable daily/session randomization versus per-request randomization.
- [ ] Add tests for empty CMS fallback behavior.

## 09. Frontend Market Discovery UX

Service ownership: frontend experience context.

Checklist:

- [x] Add admin CMS panel that distinguishes Home Page, Market Discovery Layout, and Social Share settings.
- [x] Update `/markets` to consume composed top-level page model.
- [ ] Preserve search-first layout.
- [ ] Render compact recommendations when pinned content exists.
- [ ] Render featured category cards.
- [ ] Render featured market cards.
- [x] Render tag chips in search/list cards.
- [ ] Add secondary category page route/layout.
- [ ] Scope secondary-page search to page tag/category by default.
- [ ] Keep status tabs familiar across top-level and secondary pages.

## 10. Moderator And Admin Frontend UX

Service ownership: frontend moderator/admin workflows.

Checklist:

- [x] Add tag selector to moderator create market form.
- [ ] Use typeahead/search when tags exceed a small number.
- [x] Show selected tags before submit.
- [x] Show proposed tags in admin review table/details.
- [ ] Allow admin correction if backend policy supports it.
- [x] Add tag chips on market detail page.
- [x] Add tag chips on profile/admin lifecycle tables where useful.

## Exit Criteria

- Tags are durable and admin-managed.
- Moderators can only select active existing tags.
- Admin review surfaces proposed tags.
- Search can be filtered by tag/category.
- Top-level `/markets` remains useful with or without CMS pinning.
- Secondary category pages reuse familiar search/status behavior.
- Pinned markets/pages and sections are CMS-managed and ordered.
- Empty pages fall back to recommendation content.
- OpenAPI and tests are updated before frontend depends on new endpoints.
