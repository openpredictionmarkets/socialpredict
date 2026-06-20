---
title: Analytics Repository Boundary Design
document_type: feature-design
domain: features
author: Patrick Delaney
updated_at: 2026-06-17T00:00:00Z
updated_at_display: "Wednesday, June 17, 2026"
update_reason: "Define clean architecture boundary for analytics persistence."
status: draft
---

# Analytics Repository Boundary Design

## Design Posture

Analytics calculations and reporting concepts are domain/application policy. GORM, SQL, and database table structure are mechanisms at the persistence edge.

## Target Boundary

| Package | Owns | Must not own |
| --- | --- | --- |
| `internal/domain/analytics` | interfaces, domain records, reporting formulas, service policy | GORM imports, SQL clauses, database row scanning |
| `internal/repository/analytics` | GORM adapter, SQL aggregation, table joins | reporting policy decisions |
| handlers | HTTP request/response mapping | analytics SQL or formulas |

## Migration Strategy

Move one query group at a time. Keep the existing analytics service interface stable while introducing a repository adapter.

## Risks

- Large mechanical moves can obscure behavior changes.
- Query performance can change if tests only assert shape and not representative data.
