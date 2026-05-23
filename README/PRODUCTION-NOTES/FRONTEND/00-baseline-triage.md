# Frontend Baseline Triage

## Purpose

This document triages the large frontend production-note set into a smaller decision framework for the next implementation work.

The existing frontend notes are useful as a backlog, but many of them describe broad greenfield platform programs. This triage keeps the next steps grounded in the current SocialPredict frontend codebase and separates baseline production-readiness work from larger product/platform ideas.

## Source Documents

This triage synthesizes:

| Source | Main theme |
| --- | --- |
| [plan.md](./plan.md) | Overall frontend production-readiness roadmap |
| [01-state-management.md](./01-state-management.md) | State architecture and data ownership |
| [02-performance-optimization.md](./02-performance-optimization.md) | Bundle, rendering, caching, and performance measurement |
| [03-testing-strategy.md](./03-testing-strategy.md) | Unit, component, integration, E2E, and performance testing |
| [04-security.md](./04-security.md) | Frontend auth, validation, XSS, CSP, and API safety |
| [05-accessibility.md](./05-accessibility.md) | Semantic HTML, keyboard navigation, forms, and visual accessibility |
| [06-error-handling.md](./06-error-handling.md) | Error boundaries, API error handling, reporting, and recovery |
| [07-internationalization.md](./07-internationalization.md) | i18n setup, formatting, and RTL support |
| [08-pwa-features.md](./08-pwa-features.md) | Service worker, offline, push, and installability |
| [09-analytics-tracking.md](./09-analytics-tracking.md) | Product analytics, funnels, A/B tests, and feature flags |
| [10-deployment-cicd.md](./10-deployment-cicd.md) | Frontend deployment, environment handling, Docker, and CI/CD |
| [11-monitoring-observability.md](./11-monitoring-observability.md) | APM, dashboards, browser logging, and alerting |
| [12-maintenance-updates.md](./12-maintenance-updates.md) | Dependency maintenance, regression checks, and operational upkeep |

## Current Code Snapshot

As of this triage pass, the frontend is a Vite React app under `frontend/` with:

- React 18, Vite, Tailwind, React Router v5, charting libraries, and `react-error-boundary` in [package.json](../../../frontend/package.json).
- App-level routing and auth context in [App.jsx](../../../frontend/src/App.jsx), [AppRoutes.jsx](../../../frontend/src/helpers/AppRoutes.jsx), and [AuthContent.jsx](../../../frontend/src/helpers/AuthContent.jsx).
- A global error boundary in [App.jsx](../../../frontend/src/App.jsx), but the fallback currently renders `error.message` directly to the user.
- Persistent auth state in `localStorage` from [AuthContent.jsx](../../../frontend/src/helpers/AuthContent.jsx), plus several direct `localStorage.getItem('token')` call sites across pages/components.
- One visible frontend test file, [useDocumentMeta.test.jsx](../../../frontend/src/test/useDocumentMeta.test.jsx), and no `test` script in [package.json](../../../frontend/package.json).
- Backend and Docker workflows in `.github/workflows`, but no dedicated frontend CI workflow that runs install, build, lint, and tests for frontend-only changes.
- No current service worker, i18n runtime, product analytics runtime, browser APM client, or PWA installation surface in `frontend/src`.

This means the immediate work should not start with Redux, PWA, analytics, Grafana, or a full design-system rewrite. The immediate work should make the existing app safer to change and easier to validate.

## Triage Principles

Use these rules to decide what comes next:

1. Prefer work that reduces deployment or security risk before work that adds product surface area.
2. Prefer a small verified seam over a broad architecture migration.
3. Do not introduce a large platform dependency without a concrete current failure mode.
4. Make the frontend build/test path visible in GitHub Actions before relying on manual checks.
5. Keep backend-owned contracts, such as auth/session semantics and API envelopes, aligned with frontend behavior rather than inventing parallel client-only contracts.

## Priority 1: Frontend CI and Test Baseline

Primary sources:

- [03-testing-strategy.md](./03-testing-strategy.md)
- [10-deployment-cicd.md](./10-deployment-cicd.md)
- [12-maintenance-updates.md](./12-maintenance-updates.md)

Why this is first:

The current frontend can be built locally, but the repo does not yet have a dedicated frontend CI signal comparable to the backend workflow. Before making larger UI, auth, accessibility, or performance changes, the branch should have a reliable GitHub check that proves the frontend still installs and builds.

Recommended first slice:

- Add a `frontend` workflow or extend an existing workflow with a frontend job.
- Run dependency install from `frontend/`.
- Run `npm run build`.
- Add a minimal `npm test` or equivalent script only if the current tooling supports it cleanly.
- Keep the first workflow intentionally small; do not introduce Playwright, Lighthouse, or visual regression in the same slice.

Acceptance criteria:

- Frontend PRs produce a visible GitHub Actions status.
- A broken Vite build fails before merge.
- The workflow is documented in the frontend triage or deployment note.

## Priority 2: Auth Storage and API Boundary Cleanup

Primary sources:

- [04-security.md](./04-security.md)
- [06-error-handling.md](./06-error-handling.md)
- [01-state-management.md](./01-state-management.md)

Why this is second:

The frontend still treats the browser as the auth state source of truth and stores bearer tokens in `localStorage`. A previous CodeQL finding already pushed one smaller fix around `mustChangePassword`; the broader risk remains that security-relevant auth state is spread across helper, page, and component call sites.

Recommended first slice:

- Inventory every frontend token read/write call site.
- Centralize token access behind the auth context or a small API client boundary.
- Avoid adding a full Redux migration just to solve token access.
- Preserve current runtime behavior while reducing direct `localStorage` calls.
- Document that server-managed sessions or memory-only token handling would require backend/session design, not just frontend refactoring.

Acceptance criteria:

- Fewer direct token reads outside the auth/API boundary.
- Login/logout behavior remains unchanged from the user's perspective.
- Password-change flow still works in the current session.
- Security notes distinguish immediate frontend cleanup from future server-managed session work.

## Priority 3: Error Handling That Does Not Leak Internals

Primary sources:

- [06-error-handling.md](./06-error-handling.md)
- [04-security.md](./04-security.md)
- [11-monitoring-observability.md](./11-monitoring-observability.md)

Why this is third:

The app has an error boundary, which is a good start, but the fallback renders `error.message` to the user. That is useful during development and risky as a production default. API error parsing also exists, but user-facing error behavior is not yet a clearly owned contract.

Recommended first slice:

- Make the global error fallback user-safe by default.
- Keep detailed errors in development-only console output, not production UI.
- Standardize API error display around backend response envelopes where practical.
- Add at least one focused test for the safe fallback or API error formatter once frontend test tooling is available.

Acceptance criteria:

- Production UI does not render raw exception messages from the global fallback.
- Users receive a stable generic recovery message.
- Developers can still diagnose errors locally.

## Priority 4: Accessibility Baseline on Existing UI

Primary sources:

- [05-accessibility.md](./05-accessibility.md)
- [03-testing-strategy.md](./03-testing-strategy.md)
- [02-performance-optimization.md](./02-performance-optimization.md)

Why this is fourth:

Accessibility improvements should start with the current UI, not with a new component library. The app has forms, navigation, tables, cards, tabs, and charts; those are enough to define a useful first accessibility baseline.

Recommended first slice:

- Check the login, market list, market detail, create market, and profile flows.
- Fix obvious label, focus, keyboard, and semantic structure issues.
- Add lightweight accessibility checks only after the frontend CI baseline exists.
- Avoid broad visual redesign in the first accessibility slice.

Acceptance criteria:

- Core forms have labels and usable focus behavior.
- Navigation can be operated with a keyboard in the primary flows.
- Any automated accessibility check added to CI is narrow enough to be stable.

## Priority 5: Bundle and Runtime Performance Baseline

Primary sources:

- [02-performance-optimization.md](./02-performance-optimization.md)
- [10-deployment-cicd.md](./10-deployment-cicd.md)
- [11-monitoring-observability.md](./11-monitoring-observability.md)

Why this is fifth:

The frontend uses multiple charting libraries and a Vite build, but there is no visible performance budget or CI build artifact check. Performance work should start by measuring the existing bundle before introducing code splitting or replacing libraries.

Recommended first slice:

- Capture current production build output size.
- Decide whether a simple bundle-size budget belongs in CI.
- Identify obvious duplicate/heavy dependencies, especially charting libraries.
- Defer route-based code splitting until after the baseline is measured.

Acceptance criteria:

- Current build size is recorded.
- A future performance PR has a baseline to compare against.
- Any budget is permissive enough not to block unrelated work unexpectedly.

## Deferred Until Baselines Are Stable

These ideas may be valuable, but they should not be first moves:

| Deferred area | Source documents | Reason to defer |
| --- | --- | --- |
| Full Redux or global state rewrite | [01-state-management.md](./01-state-management.md) | Auth/API boundaries can be cleaned up first without a large state migration. |
| Playwright E2E suite | [03-testing-strategy.md](./03-testing-strategy.md) | Needs basic frontend CI and stable selectors first. |
| Full CSP/security-header program | [04-security.md](./04-security.md) | CSP is partly deployment/proxy owned and should be coordinated with backend/infra. |
| i18n and RTL support | [07-internationalization.md](./07-internationalization.md) | Useful product work, but not a current production-risk blocker without localization requirements. |
| PWA/offline/push | [08-pwa-features.md](./08-pwa-features.md) | Adds caching and state complexity before test/CI/auth seams are stable. |
| Product analytics and A/B testing | [09-analytics-tracking.md](./09-analytics-tracking.md) | Needs a privacy/product decision and stable event taxonomy. |
| Browser APM dashboards and log shipping | [11-monitoring-observability.md](./11-monitoring-observability.md) | Should follow backend signal and frontend error-boundary cleanup. |
| Automated backup/recovery from frontend docs | [12-maintenance-updates.md](./12-maintenance-updates.md) | Mostly backend/infra/data-ops owned, not a frontend baseline concern. |

## Suggested First Five PRs

A practical sequence would be:

1. Add frontend CI build check.
2. Centralize auth token access and document remaining session risk.
3. Sanitize global error-boundary fallback.
4. Add a focused accessibility pass for login/navigation/core forms.
5. Record bundle-size baseline and decide whether to enforce a soft budget.

Each PR should update this triage or the relevant source note with what was learned, especially when the implementation proves a source note is too broad or out of date.

## Stop Conditions

Stop and re-triage if any of these are discovered:

- The frontend cannot build reproducibly on a clean install.
- The current auth model cannot be made safer without backend session changes.
- A proposed dependency introduces significant bundle growth before a performance baseline exists.
- Accessibility fixes require a design-system decision rather than local component cleanup.
- Deployment behavior differs materially between local, staging, and production frontend builds.
