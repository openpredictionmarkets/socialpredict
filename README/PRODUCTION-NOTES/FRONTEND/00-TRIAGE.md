---
title: Frontend Baseline Triage
document_type: production-notes
domain: frontend
author: Patrick Delaney
updated_at: 2026-05-23T00:00:00Z
updated_at_display: "Saturday, May 23, 2026"
update_reason: "Ground frontend production-note sequencing with Evans/Fowler/Martin design-agent review and current frontend code evidence."
status: draft
---

# Frontend Baseline Triage

## Purpose

This note triages the large frontend production-note set into a smaller decision framework for the next implementation work.

The existing frontend notes are useful as a backlog, but many of them describe broad greenfield platform programs. This triage keeps the next steps grounded in the current SocialPredict frontend codebase and separates baseline production-readiness work from larger product/platform ideas.

The key conclusion from the three design-agent reviews is:

- Evans/domain lens: add a language and API-contract alignment priority before code work, because the frontend is where participants see the prediction-market language.
- Fowler/evolutionary lens: keep frontend CI first for code execution, but keep early PRs small and reversible.
- Martin/clean-architecture lens: treat the immediate implementation problem as a frontend auth/API/error boundary problem, not only a token-storage problem.

## Source Documents

This triage synthesizes active baseline notes and future platform notes. The active notes are the next queue; the `FUTURE/` notes preserve deferred ideas with re-entry criteria.

Active baseline notes:

| Source | Main theme |
| --- | --- |
| [plan.md](./plan.md) | Frontend production-note index |
| [01-state-management.md](./01-state-management.md) | API/auth adapter and state boundary baseline |
| [02-performance-optimization.md](./02-performance-optimization.md) | Build-size and performance measurement baseline |
| [03-testing-strategy.md](./03-testing-strategy.md) | Frontend install/build verification baseline |
| [04-security.md](./04-security.md) | Auth/API/session-adjacent security baseline |
| [05-accessibility.md](./05-accessibility.md) | Core workflow accessibility baseline |
| [06-error-handling.md](./06-error-handling.md) | Safe public failure presentation baseline |
| [10-deployment-cicd.md](./10-deployment-cicd.md) | Frontend PR CI and deployment-feedback baseline |

Future platform notes:

| Source | Main theme |
| --- | --- |
| [FUTURE/00-baseline-triage.md](./FUTURE/00-baseline-triage.md) | Future frontend backlog index and re-entry framework |
| [FUTURE/01-long-term-state-platform.md](./FUTURE/01-long-term-state-platform.md) | Redux, RTK Query, persisted store, offline sync |
| [FUTURE/02-long-term-performance-platform.md](./FUTURE/02-long-term-performance-platform.md) | Code splitting, Web Vitals, strict budgets, caching platform |
| [FUTURE/03-long-term-test-infrastructure.md](./FUTURE/03-long-term-test-infrastructure.md) | Playwright, coverage, visual regression, accessibility automation |
| [FUTURE/04-long-term-security-platform.md](./FUTURE/04-long-term-security-platform.md) | CSP, server sessions, stronger auth, browser security monitoring |
| [FUTURE/05-long-term-accessibility-program.md](./FUTURE/05-long-term-accessibility-program.md) | Full WCAG/accessibility program |
| [FUTURE/06-long-term-i18n-localization.md](./FUTURE/06-long-term-i18n-localization.md) | i18n, localization, RTL |
| [FUTURE/07-long-term-pwa-offline-platform.md](./FUTURE/07-long-term-pwa-offline-platform.md) | Service worker, offline, push, installability |
| [FUTURE/08-long-term-analytics-experimentation.md](./FUTURE/08-long-term-analytics-experimentation.md) | Product analytics, funnels, A/B testing |
| [FUTURE/09-long-term-browser-observability.md](./FUTURE/09-long-term-browser-observability.md) | Browser APM, replay, dashboards, log shipping |
| [FUTURE/10-long-term-maintenance-automation.md](./FUTURE/10-long-term-maintenance-automation.md) | Dependency automation, regression platform, frontend recovery |

## Design-Agent Review Inputs

This triage was reviewed against the three designer-agent postures from `/Users/patrick/Projects/spec-socialpredict-tasks-auto`:

| Agent posture | Applied concern | Triage result |
| --- | --- | --- |
| Evans/domain design | Ubiquitous language, bounded contexts, stakeholder terms | Add Priority 0 for frontend published language and API-contract alignment. |
| Fowler/evolutionary architecture | Safe sequencing, reversibility, feedback loops | Keep CI/build first for code, move broad auth migration into incremental seams, and defer bundle budgets until after core safety. |
| Martin/clean architecture | Dependency direction, ports/adapters, testability | Reframe token work as auth/API/error boundary cleanup and name current mixed-responsibility modules. |

## Design Plan Relationship

Do not create a second standalone frontend design plan as a competing architecture source of truth.

Durable frontend architecture decisions should update the canonical design artifact at `/Users/patrick/Projects/spec-socialpredict-tasks-auto/lib/design/design-plan.json`. The design-plan library and designer-agent instructions define that file as the canonical artifact. A separate frontend design plan would split the ubiquitous language and make frontend/API contract decisions easier to drift from backend domain language.

Recommended design-plan direction:

- Keep one canonical repo-level design plan.
- Add a frontend slice, workstream, or bounded context to that plan when frontend architecture decisions become durable.
- Narrow the existing design-plan out-of-scope language from broad frontend redesign to broad frontend visual/platform redesign, while keeping frontend published-language and API-contract alignment in scope.
- Treat this `00-TRIAGE.md` as implementation sequencing and code-grounded readiness guidance, not as the canonical design authority.

The canonical design plan now includes frontend extraction tags, a `Frontend Experience Context`, a `Frontend Visual System Boundary`, active frontend workstreams `W08` and `W09`, and ADR guardrails for deferring browser platform capabilities until baseline seams are stable.

## Current Code Snapshot

As of this triage pass, the frontend is a Vite React app under `frontend/` with:

- React 18, Vite, Tailwind, React Router v5, charting libraries, and `react-error-boundary` in [package.json](../../../frontend/package.json).
- App-level routing and auth context in [App.jsx](../../../frontend/src/App.jsx), [AppRoutes.jsx](../../../frontend/src/helpers/AppRoutes.jsx), and [AuthContent.jsx](../../../frontend/src/helpers/AuthContent.jsx).
- A global error boundary in [App.jsx](../../../frontend/src/App.jsx), but the fallback currently renders `error.message` directly to the user.
- Persistent auth state in `localStorage` from [AuthContent.jsx](../../../frontend/src/helpers/AuthContent.jsx), plus several direct `localStorage.getItem('token')` call sites across pages/components.
- API transport and envelope handling spread across pages, hooks, and components; [axios.js](../../../frontend/src/api/axios.js) is a placeholder, while direct `fetch(API_URL...)` call sites and duplicated `unwrapApiResponse` functions remain.
- One visible frontend test file, [useDocumentMeta.test.jsx](../../../frontend/src/test/useDocumentMeta.test.jsx), but [package.json](../../../frontend/package.json) has no `test` script and does not declare `vitest` or `jsdom`.
- Backend and Docker workflows in `.github/workflows`, but no dedicated frontend PR CI workflow that runs install, build, lint, or tests for frontend-only changes. Release/manual Docker builds do build the frontend image through [docker.yml](../../../.github/workflows/docker.yml), but that is not the same as PR feedback.
- No current service worker, i18n runtime, product analytics runtime, browser APM client, or PWA installation surface in `frontend/src`.

This means the immediate work should not start with Redux, PWA, analytics, Grafana, or a full design-system rewrite. The immediate work should make the existing app safer to change, easier to validate, and more consistent with canonical SocialPredict domain language.

## Triage Principles

Use these rules to decide what comes next:

1. Prefer work that reduces deployment or security risk before work that adds product surface area.
2. Prefer a small verified seam over a broad architecture migration.
3. Do not introduce a large platform dependency without a concrete current failure mode.
4. Make the frontend build/test path visible in GitHub Actions before relying on manual checks.
5. Keep backend-owned contracts, such as auth/session semantics and API envelopes, aligned with frontend behavior rather than inventing parallel client-only contracts.
6. Do not let frontend labels invent domain language; conform to canonical design-plan terms unless the design plan is explicitly amended.
7. Split PRs when a change crosses from seam creation into broad call-site migration.

## Priority 0: Frontend Published Language and API Contract Alignment

Primary sources:

- [01-state-management.md](./01-state-management.md)
- [04-security.md](./04-security.md)
- [06-error-handling.md](./06-error-handling.md)
- Canonical design plan at `/Users/patrick/Projects/spec-socialpredict-tasks-auto/lib/design/design-plan.json`

Why this is priority zero:

The frontend is where participants encounter the prediction-market domain. If UI labels, API reason handling, and route access language drift from canonical backend/domain language, the system can appear inconsistent even when backend behavior is correct.

The first design/doc move should be to inventory and protect frontend published language. The first code move can still be CI.

Language areas to inventory:

| Area | Terms to align |
| --- | --- |
| Market browsing | market, market status, market trading state, close time, resolved, pending, closed |
| Create market | question, description, close time, outcome, outcome label, creator |
| Buy exposure | buy, bet, purchase shares, amount, trade fee, projected probability, account balance |
| Sell exposure | sell, shares owned, sale credit, position value, remaining exposure |
| Resolution | resolution, resolved outcome, payout, profit/loss, accounting result |
| Profile/account | participant, regular participant, admin, creator, authenticated participant, public profile |
| Auth/session | session, token, must change password, logout, authorization denied |
| Failures | public failure reason, recovery message, raw error, backend envelope |

Acceptance criteria:

- The triage names frontend language as a production-readiness concern, not only a copywriting concern.
- Frontend docs state that canonical outcome codes must remain distinct from outcome display labels.
- Public failure messages are treated as user-facing domain language, not raw runtime errors.
- Future implementation PRs have a place to record language discoveries before changing labels broadly.

## Priority 1: Frontend CI and Test Baseline

Primary sources:

- [03-testing-strategy.md](./03-testing-strategy.md)
- [10-deployment-cicd.md](./10-deployment-cicd.md)
- [FUTURE/10-long-term-maintenance-automation.md](./FUTURE/10-long-term-maintenance-automation.md)

Why this is first for code:

The current frontend can be built locally, but the repo does not yet have a dedicated frontend CI signal comparable to the backend workflow. Before making larger UI, auth, accessibility, or performance changes, the branch should have a reliable GitHub check that proves the frontend still installs and builds.

Recommended first slice:

- Add a frontend PR workflow or extend an existing workflow with a frontend job.
- Use Node 22 unless the project deliberately chooses another supported runtime.
- Run `npm ci` from `frontend/`.
- Run `npm run build` from `frontend/`.
- Do not claim test coverage until test tooling is declared. The current test imports Vitest, but `vitest`, `jsdom`, and a `test` script are not declared in [package.json](../../../frontend/package.json).
- If tests are added in this slice, add an explicit `vitest --environment jsdom` script and the required dev dependencies.
- Keep the first workflow intentionally small; do not introduce Playwright, Lighthouse, visual regression, or bundle budgets in the same slice.

Acceptance criteria:

- Frontend PRs produce a visible GitHub Actions status.
- A broken Vite build fails before merge.
- Any test command that CI runs is actually declared and reproducible from a clean install.
- The workflow is documented in this triage or the deployment note.

## Priority 2: Auth/API Boundary Discovery and Incremental Cleanup

Primary sources:

- [04-security.md](./04-security.md)
- [06-error-handling.md](./06-error-handling.md)
- [01-state-management.md](./01-state-management.md)

Why this is second:

The frontend still treats the browser as the auth state source of truth and stores bearer tokens in `localStorage`. A previous CodeQL finding already pushed one smaller fix around `mustChangePassword`; the broader risk remains that security-relevant auth state, transport, headers, backend envelope parsing, and user-facing error display are spread across helper, page, hook, and component call sites.

This is not only a storage issue. It is a frontend boundary issue.

Recommended first slice:

- Inventory every frontend token read/write call site.
- Inventory direct `fetch(API_URL...)` call sites and duplicated `unwrapApiResponse` helpers.
- Define a small API client port/adapter for authenticated requests.
- Centralize auth header injection.
- Centralize backend envelope parsing.
- Keep React components dependent on use-case functions/hooks, not directly on `API_URL`, `fetch`, or `localStorage`.
- Preserve current runtime behavior while reducing direct storage and transport coupling.
- Document that server-managed sessions or memory-only token handling would require backend/session design, not just frontend refactoring.
- Split implementation PRs if a branch changes more than the seam plus one or two representative call sites.

Acceptance criteria:

- No direct token reads outside the auth storage adapter/API client boundary, except temporary documented compatibility call sites.
- `Authenticated Participant`, `Regular Participant`, `Admin`, and `Must Change Password` have explicit frontend meanings aligned with backend/domain language.
- Login/logout behavior remains unchanged from the user's perspective.
- Password-change flow still works in the current session.
- Security notes distinguish immediate frontend cleanup from future server-managed session work.

## Priority 3: Public Failure Language and Safe Error Handling

Primary sources:

- [06-error-handling.md](./06-error-handling.md)
- [04-security.md](./04-security.md)
- [FUTURE/09-long-term-browser-observability.md](./FUTURE/09-long-term-browser-observability.md)

Why this is third:

The app has an error boundary, which is a good start, but the fallback renders `error.message` to the user. That is useful during development and risky as a production default. Many components also render or alert raw `err.message` values. API error parsing exists, but user-facing error behavior is not yet a clearly owned contract.

Fowler sequencing note: the tiny global fallback leak can be fixed before broad auth/API migration because it is small, reversible, and low-risk.

Recommended first slice:

- Make the global error fallback user-safe by default.
- Split diagnostic errors from safe user-facing recovery messages.
- Keep detailed errors in development-only console output, not production UI.
- Map backend envelopes/reasons into user-safe copy without exposing raw runtime details.
- Standardize API error display around backend response envelopes where practical.
- Add at least one focused test for the safe fallback or API error formatter once frontend test tooling is available.

Acceptance criteria:

- Production UI does not render raw exception messages from the global fallback.
- Expected business rejections do not read like runtime crashes.
- Users receive a stable generic recovery message when the app fails unexpectedly.
- Developers can still diagnose errors locally.

## Priority 4: Accessibility Baseline on Existing Core Workflows

Primary sources:

- [05-accessibility.md](./05-accessibility.md)
- [03-testing-strategy.md](./03-testing-strategy.md)
- [02-performance-optimization.md](./02-performance-optimization.md)

Why this is fourth:

Accessibility improvements should start with the current UI, not with a new component library. The app has forms, navigation, tables, cards, tabs, charts, and domain-critical trading flows; those are enough to define a useful first accessibility baseline.

Recommended first slice:

- Check browse markets, create market, buy exposure, sell exposure, resolve market, view profile/account, and view position flows.
- Fix obvious label, focus, keyboard, and semantic structure issues.
- Add lightweight accessibility checks only after the frontend CI baseline exists.
- Avoid broad visual redesign in the first accessibility slice.

Acceptance criteria:

- Core forms have labels and usable focus behavior.
- Navigation can be operated with a keyboard in the primary flows.
- Market/trading workflows are not blocked for keyboard or screen-reader users by obvious structural issues.
- Any automated accessibility check added to CI is narrow enough to be stable.

## Priority 5: Bundle and Runtime Performance Baseline

Primary sources:

- [02-performance-optimization.md](./02-performance-optimization.md)
- [10-deployment-cicd.md](./10-deployment-cicd.md)
- [FUTURE/09-long-term-browser-observability.md](./FUTURE/09-long-term-browser-observability.md)

Why this is fifth:

The frontend uses multiple charting libraries and a Vite build, but there is no visible performance baseline or PR build artifact check. Performance work should start by measuring the existing bundle before introducing code splitting or replacing libraries.

Recommended first slice:

- Capture current production build output size.
- Decide whether a simple bundle-size budget belongs in CI.
- Identify obvious duplicate/heavy dependencies, especially charting libraries.
- Defer route-based code splitting until after the baseline is measured.
- Keep any budget permissive at first. Do not let a new budget block unrelated work unexpectedly.

Acceptance criteria:

- Current build size is recorded.
- A future performance PR has a baseline to compare against.
- Any budget is permissive enough not to block unrelated work unexpectedly.

## Boundary Violations To Track

These are not all defects to fix in one PR. They are seams that should guide small refactors:

| Seam | Current issue | Direction |
| --- | --- | --- |
| `AuthContent` | Mixes React context, browser persistence, JWT decoding, login HTTP transport, backend envelope parsing, and auth state policy | Split storage, transport, and auth-state policy behind explicit helpers/adapters. |
| `AppRoutes` | Embeds access-policy decisions directly in route JSX | Keep short-term, then move route guards toward auth/access decision helpers. |
| Page/component fetches | Many components import `API_URL` and call `fetch` directly | Move authenticated and envelope-aware requests through an API client boundary. |
| Duplicated `unwrapApiResponse` | Central helper exists, but local copies remain | Use one frontend API response helper unless a route has a documented exception. |
| Route-coupled hooks | Hooks such as market details mix router detail, auth detail, transport, DTO normalization, and view calculation | Extract transport and normalization seams before broad state-library migration. |
| Raw error display | Global boundary and component alerts can show raw exception/backend text | Split diagnostic details from public recovery messages. |

## Deferred Until Baselines Are Stable

These ideas may be valuable, but they should not be first moves:

| Deferred area | Source documents | Reason to defer |
| --- | --- | --- |
| Full Redux or global state rewrite | [FUTURE/01-long-term-state-platform.md](./FUTURE/01-long-term-state-platform.md) | Auth/API boundaries can be cleaned up first without a large state migration. |
| Playwright E2E suite | [FUTURE/03-long-term-test-infrastructure.md](./FUTURE/03-long-term-test-infrastructure.md) | Needs basic frontend CI and stable selectors first. |
| Full CSP/security-header program | [FUTURE/04-long-term-security-platform.md](./FUTURE/04-long-term-security-platform.md) | CSP is partly deployment/proxy owned and should be coordinated with backend/infra. |
| i18n and RTL support | [FUTURE/06-long-term-i18n-localization.md](./FUTURE/06-long-term-i18n-localization.md) | Useful product work, but not a current production-risk blocker without localization requirements. |
| PWA/offline/push | [FUTURE/07-long-term-pwa-offline-platform.md](./FUTURE/07-long-term-pwa-offline-platform.md) | Adds caching and state complexity before test/CI/auth seams are stable. |
| Product analytics and A/B testing | [FUTURE/08-long-term-analytics-experimentation.md](./FUTURE/08-long-term-analytics-experimentation.md) | Needs a privacy/product decision and stable event taxonomy. |
| Browser APM dashboards and log shipping | [FUTURE/09-long-term-browser-observability.md](./FUTURE/09-long-term-browser-observability.md) | Should follow backend signal and frontend error-boundary cleanup. |
| Automated backup/recovery from frontend docs | [FUTURE/10-long-term-maintenance-automation.md](./FUTURE/10-long-term-maintenance-automation.md) | Mostly backend/infra/data-ops owned, not a frontend baseline concern. |

## Suggested First Five PRs

A practical sequence would be:

1. Document frontend published-language and API-contract alignment.
2. Add frontend CI build check with Node 22, `npm ci`, and `npm run build`.
3. Sanitize the global error-boundary fallback and define safe public failure copy.
4. Add auth/API boundary discovery plus a small API client/auth-header seam.
5. Migrate token/API call sites incrementally, starting with one or two representative flows, then run an accessibility pass on core workflows.

If frontend tests are added before or during this sequence, make that an explicit tooling slice: add `vitest`, `jsdom`, and a reproducible `npm test` script before requiring test execution in CI.

## Stop Conditions

Stop and re-triage if any of these are discovered:

- The frontend cannot build reproducibly on a clean install.
- The current auth model cannot be made safer without backend session changes.
- A proposed dependency introduces significant bundle growth before a performance baseline exists.
- Accessibility fixes require a design-system decision rather than local component cleanup.
- Deployment behavior differs materially between local, staging, and production frontend builds.
