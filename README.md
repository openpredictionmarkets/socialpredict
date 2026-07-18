[![GitHub Release][release-img]][release]
[![14-day clones][clones-14d-img]][traffic]
[![14-day unique cloners][unique-cloners-14d-img]][traffic]
[![Test][test-img]][test]
[![License: MIT][license-img]][license]
[![Trivy Scan][trivy-img]][trivy]
[![Deploy To Staging][staging-img]][staging]
[![Deploy To Production][production-img]][production]

[release-img]: https://img.shields.io/github/v/release/openpredictionmarkets/socialpredict?label=release
[release]: https://github.com/openpredictionmarkets/socialpredict/releases/latest
[clones-14d-img]: https://img.shields.io/badge/14d%20clones-15%2C321-62c3f8
[unique-cloners-14d-img]: https://img.shields.io/badge/14d%20cloners-219-2ea043
[traffic]: https://github.com/openpredictionmarkets/socialpredict/graphs/traffic
[test-img]: https://github.com/openpredictionmarkets/socialpredict/actions/workflows/backend.yml/badge.svg?branch=main
[test]: https://github.com/openpredictionmarkets/socialpredict/actions/workflows/backend.yml
[license-img]: https://img.shields.io/badge/License-MIT-yellow.svg
[license]: https://github.com/openpredictionmarkets/socialpredict/blob/main/LICENSE
[trivy-img]: https://github.com/openpredictionmarkets/socialpredict/actions/workflows/container-security.yml/badge.svg
[trivy]: https://github.com/openpredictionmarkets/socialpredict/actions/workflows/container-security.yml
[staging-img]: https://github.com/openpredictionmarkets/socialpredict/actions/workflows/deploy-to-staging.yml/badge.svg?event=pull_request
[staging]: https://github.com/openpredictionmarkets/socialpredict/actions/workflows/deploy-to-staging.yml
[production-img]: https://github.com/openpredictionmarkets/socialpredict/actions/workflows/deploy-to-production.yml/badge.svg?event=workflow_run
[production]: https://github.com/openpredictionmarkets/socialpredict/actions/workflows/deploy-to-production.yml

# SocialPredict

## The open prediction market engine for everyone

SocialPredict lets **anyone** – individuals, classrooms, companies, and even governments – tap into the power of prediction markets.

### Setting Up SocialPredict Website with HTTPS on a Virtual Private Computer

![socialpredict_setup_demo](https://github.com/user-attachments/assets/1062a3b2-46d0-4a40-a648-39e71f0e7cec)

### SocialPredict Logging In and Creating Market

![socialpredict_demo_3x](https://github.com/user-attachments/assets/438bb424-670b-4cd3-a69f-5581c4d9fcbf)

---

### Join us in shaping the future of prediction markets by building connections and expertise within your community!

## Used By

<img src="https://github.com/openpredictionmarkets/socialpredict/raw/main/README/IMG/logotype_kenyon-purple_rgb.png" alt = "Kenyon College Logo" width=40% height=40%>

* Kenyon College (Political Science course PSCI 303, Campaigns & Elections; syllabus [here](https://www.zacharymcgee.net/syllabi/PSCI_303_public.pdf))
                        
## Stargazers Over Time
[![Stargazers over time](README/METRICS/stargazers-over-time.svg)](https://github.com/openpredictionmarkets/socialpredict/stargazers)

## Licensing

SocialPredict is available under the [MIT License](https://github.com/openpredictionmarkets/socialpredict/blob/main/LICENSE).

---

## Roadmap at a Glance

We’re building SocialPredict as the **best free prediction-market infrastructure** you can run yourself — today and into the future.

```mermaid
flowchart LR

    classDef roadmap fill:#dbeafe,stroke:#1d4ed8,color:#111827,stroke-width:2px;

    Y2025["🧩 Service Architecture<br/>2025"]
    Y2026["🧱 Microservices & Math<br/>2026"]
    Y2027["☁️ Cloud & UX<br/>2027"]
    Y2030["🚀 HPC & Analytics<br/>2028–2030"]

    Y2025 --> Y2026 --> Y2027 --> Y2030

    class Y2025,Y2026,Y2027,Y2030 roadmap
```

---

## Staging

Check out our staging instance at [kconfs.com](https://kconfs.com/) to see the newest `main` branch deployment in action. This environment may be reset, load tested, or temporarily unavailable while we test deployment and capacity changes.

Our model office / production-style release target is [brierfoxforecast.com](https://brierfoxforecast.com/). This environment is intended to track published releases rather than every merge to `main`.

## Feature Highlights

### Moderator Mode

SocialPredict now supports a moderator-driven market workflow. Approved moderators can propose markets, while admins retain final publishing control. Proposed markets are not tradable until an admin approves them.

Highlights:

- Moderator approval/status management
- Proposed, published, and rejected market lifecycle states
- Proposal cost accounting and refund behavior for rejected proposals
- Moderator profile tabs for proposed, published, and rejected markets
- Admin review queues for proposed, approved, and rejected markets
- Full market descriptions and custom YES/NO labels shown during review

![Moderator profile showing proposed, published, and rejected market tabs](README/IMG/screenshots/moderator-profile-market-tabs.png)

![Admin published market review queue with full-description controls and tag adjustments](README/IMG/screenshots/admin-market-review-published.png)

### Admin and CMS Tools

Admins have expanded operational controls for reviewing users, reviewing proposed markets, stewarding market operations, organizing market discovery, and managing public sharing metadata.

Highlights:

- Admin user queue for moderator eligibility/status review
- Market review tools with approval, rejection, review trail, and refund behavior
- Market stewardship reassignment so admins can move operational responsibility from a suspended or unavailable moderator to another active moderator
- Admin-managed market tags, topic navigation, market pins, and market discovery layout controls
- CMS-managed social share title, description, image, and image-enable setting
- Default OpenGraph/social share image support

![Admin user governance queue](README/IMG/screenshots/admin-user-governance.png)

![Admin market discovery layout editor](README/IMG/screenshots/admin-market-discovery-layout.png)

![Admin social share CMS settings with share preview](README/IMG/screenshots/admin-social-share-settings.png)

### Market Discovery and Governance

Market discovery has moved beyond a flat market list. Admin-managed tags and market discovery layout controls now support topic navigation, pinned markets, and tag-filtered topic pages while keeping search as the primary entry point.

Highlights:

- Admin-managed tags shown on market creation, admin review, market lists, and market detail pages
- `/markets` topic navigation with persistent tag/category pins
- Tag-filtered secondary pages under `/markets/topic/:slug`
- Pinned market cards for curated discovery
- Infinite-scroll market lists for `/markets` and topic pages
- Search and pagination improvements for admin market review and user/moderator queues

![Markets discovery page with topic navigation and pinned market cards](README/IMG/screenshots/markets-discovery.png)

![Topic page filtered by Category A](README/IMG/screenshots/markets-topic-category-a.png)

### Sharing and Embeddable Markets

Market pages now have a stronger sharing foundation, including OpenGraph and Twitter card metadata for link previews. The embeddable-market feature plan is documented so individual markets can be shared and embedded more cleanly as that work continues.

Highlights:

- Market-aware OpenGraph metadata
- Social share image endpoint and default image
- Security-header updates needed for controlled embedding
- Feature planning for iframe-friendly market embeds

![Market detail page with chart and public sharing-ready metadata](README/IMG/screenshots/market-detail-share-ready.png)

### Market Amendments, Stewardship, and Work Profits

Market governance now includes clearer controls around who can operate a market after it is published and how contract-like text can be clarified.

Highlights:

- Immutable market titles with additive-only description amendments
- Moderator/steward amendment proposal queues with admin pending/approved/rejected review
- Optional admin auto-approval for market proposals and amendments
- Suspended moderators lose market-creation/governance capabilities
- Current stewards, not necessarily original creators, can govern and resolve assigned markets
- Thresholded moderator work-profit payouts after successful market resolution
- User financial reporting for moderator work profits

![Market page showing expanded contract text and approved amendments](README/IMG/screenshots/market-description-amendments.png)

![Admin amendment review queue](README/IMG/screenshots/admin-amendment-review-queue.png)

### Read Models and Performance Foundations

The backend now has a stronger performance boundary between transaction-time math and display/read-model data.

Highlights:

- Display-oriented market accounting snapshots for volume, dust, probability, participants, and chart data
- Cached/read-model paths for system metrics, user financial metrics, market cards, leaderboards, and market widgets
- Freshness messaging for cached display widgets
- Pagination for market bets, positions, market leaderboards, global leaderboard, and market lists
- Boundary tests to keep snapshots out of buy/sell/resolution transaction paths

### Load Testing and Release Dossiers

The repository now includes an external load-testing harness under [`loadtest/`](./loadtest/README.md). It uses k6 from a separate load-generator machine and can seed remote staging fixtures, pull fixture CSVs, run smoke/baseline/hot-market tests, and capture host telemetry through HostOps.

Current capacity evidence is summarized in the [staging load-test dossier](./loadtest/dossier/staging-capacity-2026-05-29.md), the [large Basic AMD experiment notebook](./loadtest/dossier/general-purpose-capacity-experiment-2026-06-02.md), and the [capacity forecast dossier](./loadtest/dossier/capacity-forecast-2026-06-02.md).

Highlights:

- k6 smoke, baseline, hot-market burst, and soak scenarios
- Remote staging fixture seeding and fixture pull commands
- Host telemetry capture for CPU, RAM, disk, network, Docker, backend, Postgres, and Traefik
- Host profile capture to record the exact tested Droplet shape and Docker/container limits
- Release dossier tooling for turning load-test results into reviewable evidence

Example dossier: [Staging Capacity Dossier, 2026-05-29](./loadtest/dossier/staging-capacity-2026-05-29.md). Look for the most recent dossier before making deployment or sizing decisions, because capacity results depend on the exact release, host size, database topology, and rate-limit configuration under test.

Latest forecast note: [Capacity Forecast Dossier, 2026-06-02](./loadtest/dossier/capacity-forecast-2026-06-02.md) records the strongest current sustained hot-market result as `250` bets/sec for `5m` on a single-node DigitalOcean Basic AMD `8 vCPU / 32 GiB RAM` host, while `300` bets/sec for `5m` is not yet a clean supported target.

## Getting Started

### Setting up a Local Instance

- [Info on Local Setup](/README/LOCAL_SETUP.md)
- [Info on How Economics Can Be Customized](/README/README-CONFIG.md)
- Development fixture helpers are available through `./SocialPredict dev-bootstrap-users` for local smoke testing.

### Deploying to the Web

- [How to Set Up Your Own Website](/README/STAGE_SETUP.md)
- [HostOps Scaffold and Infra Boundary Notes](/README/INFRA/README-INFRA-HOSTOPS.md)
- [Staging Deployment Guide](/README/INFRA/README-INFRA-STAGING.md)
- [Production Deployment Guide](/README/INFRA/README-INFRA-PRODUCTION.md)

OpenPredictionMarkets deployment conventions:

- Merges to `main` deploy staging at [kconfs.com](https://kconfs.com/).
- Published GitHub releases deploy the model office / production-style target at [brierfoxforecast.com](https://brierfoxforecast.com/).
- GitHub Actions dispatch Ansible-based deploys and then verify public `/health` and `/readyz`.
- HostOps is a local convenience wrapper for host SSH, environment discovery, disk checks, cleanup, and load-test telemetry support.

### API, Operations, and Load Testing

- [Canonical Backend API documentation](/backend/docs/README.md)
- [Load Testing Guide](/loadtest/README.md)
- [Load Test CLI Runbook](/loadtest/cli/OPERATING.md)
- [Current Staging Capacity Dossier](/loadtest/dossier/staging-capacity-2026-05-29.md)
- [Capacity Forecast Dossier](/loadtest/dossier/capacity-forecast-2026-06-02.md)

### How Do Prediction Markets Work?

Here's a quick primer about how (and why) [prediction markets work](/README/MATH/README-MATH.md). Want more info? We maintain a list of resources where you can see research on [prediction markets in action](https://github.com/openpredictionmarkets/resources).

## Contributing

We welcome and appreciate every contribution. Get started by reading our [guide](https://github.com/openpredictionmarkets/socialpredict/blob/main/CONTRIBUTING.md) and make sure to follow our [Code of Conduct](https://github.com/openpredictionmarkets/socialpredict/blob/main/CODE_OF_CONDUCT.md).

### Where to Next?

- Brush up on our [Development Conventions](/README/README-CONVENTIONS.md)
- Review the canonical [Backend API documentation](/backend/docs/README.md)
- Check out our [ongoing Projects](https://github.com/openpredictionmarkets/socialpredict/projects?query=is%3Aopen)
- Look at our [Issues](https://github.com/openpredictionmarkets/socialpredict/issues)
- Have your say on [GitHub Discussions](https://github.com/orgs/openpredictionmarkets/discussions)
