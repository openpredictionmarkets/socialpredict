---
title: Deployment Infrastructure Reading
document_type: reading-notes
domain: backend
author: Patrick Delaney
updated_at: 2026-05-03T12:13:40Z
updated_at_display: "Sunday, May 03, 2026 at 12:13 PM UTC"
update_reason: "Add focused reading references for the WAVE08 deployment health, readiness, startup ownership, proxy publishing, and graceful shutdown contract."
status: active
---

# Deployment Infrastructure Reading

This reading note supports
[08-deployment-infrastructure.md](../08-deployment-infrastructure.md). It is
not a general software bookshelf. These references are selected because they
directly explain the production patterns behind WAVE08: health probes, readiness
gates, graceful lifecycle handling, deployment pipelines, production-ready
service standards, and cloud-native routing behavior.

The selection bias is toward direct topical fit, established publishers or
primary authors, operator credibility, and community reception. Marketplace
ratings drift over time, so this note treats ratings as a secondary signal
rather than the deciding factor.

## Best First Reads

1. [**Kubernetes Patterns, 2nd Edition**](https://www.oreilly.com/library/view/kubernetes-patterns-2nd/9781098131678/)
   by Bilgin Ibryam and Roland Huss, O'Reilly, 2023.
   Start here for the direct health-probe pattern. The book has an explicit
   Health Probe pattern and separates liveness from readiness as application
   lifecycle signals. Even though WAVE08 uses Docker Compose and nginx rather
   than Kubernetes, the conceptual model is the same: one check answers whether
   the process is alive, and another answers whether traffic should be routed to
   it.

2. [**Release It!, 2nd Edition**](https://www.oreilly.com/library/view/release-it-2nd/9781680504552/)
   by Michael T. Nygard, Pragmatic Bookshelf, 2018.
   Read this for the production-systems mindset behind WAVE08. The relevant
   idea is that reliability is designed into runtime behavior, deployment
   shape, failure handling, and operational signals. The SocialPredict split
   between `/health`, `/readyz`, startup writer ownership, and graceful shutdown
   is a small version of that production-readiness discipline.

3. [**Production-Ready Microservices**](https://www.oreilly.com/library/view/production-ready-microservices/9781491965962/)
   by Susan J. Fowler, O'Reilly, 2016.
   Read this for the service-standardization frame. WAVE08 is turning deployment
   behavior into an explicit service contract: which process may mutate startup
   state, what health means, what readiness means, what the proxy publishes, and
   what a reviewer can verify.

4. [**Cloud Native Patterns**](https://www.manning.com/books/cloud-native-patterns)
   by Cornelia Davis, Manning, 2019.
   Read this for the application-lifecycle and routing frame. The most relevant
   chapters are the ones on running cloud-native applications in production,
   application lifecycle, services, routing, and service discovery. WAVE08 is
   applying those ideas to the current compose/nginx topology instead of waiting
   for a larger platform migration.

5. [**The Site Reliability Workbook**](https://sre.google/workbook/table-of-contents/)
   edited by Betsy Beyer, Niall Richard Murphy, David K. Rensin, Kent Kawahara,
   and Stephen Thorne, O'Reilly, 2018.
   Read this after the first four if the goal is to connect runtime contracts to
   operational review. It is useful for thinking about reliable services in
   cloud environments, concrete verification, service-level thinking, and
   rollout procedures.

6. [**Continuous Delivery**](https://www.pearson.com/en-us/subject-catalog/p/Humble-Continuous-Delivery-Reliable-Software-Releases-through-Build-Test-and-Deployment-Automation/P200000009113)
   by Jez Humble and David Farley, Addison-Wesley Professional, 2010.
   Read this for the deployment-pipeline and release-verification frame. WAVE08
   intentionally leaves one remaining check for the downstream deploy runner:
   after dispatch, the real host should prove that compose service roles and
   nginx root-path routes were applied together.

## Primary References

- Kubernetes documentation:
  [**Liveness, Readiness, and Startup Probes**](https://kubernetes.io/docs/concepts/configuration/liveness-readiness-startup-probes/).
  This is the clearest primary reference for the exact vocabulary behind
  `/health` and `/readyz`.
- Google SRE books:
  [**Site Reliability Engineering**](https://sre.google/sre-book/table-of-contents/)
  and **The Site Reliability Workbook** are available from Google and are useful
  when the deployment note needs to grow from local probes into service-level
  objectives, rollout checks, alerting, and production review.

## How This Maps To WAVE08

- `/health` maps to liveness: the backend process is serving HTTP.
- `/readyz` maps to readiness: the backend is safe to receive traffic because
  startup is complete and database availability checks pass.
- `backend-startup-writer` maps to explicit startup ownership: one service may
  run migrations and startup-owned seeds.
- request-serving `backend` maps to non-writer serving replicas: it verifies
  migrations before opening readiness and should be the only backend target for
  nginx and frontend traffic.
- nginx root-path publishing maps to explicit routing: reviewers should check
  `/health`, `/readyz`, `/openapi.yaml`, `/swagger`, and `/swagger/` at the
  public host root.
- graceful shutdown maps to lifecycle signaling: close readiness first, wait for
  the readiness-drain window, then let HTTP shutdown drain in-flight requests.
