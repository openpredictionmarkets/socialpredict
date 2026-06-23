# Schemathesis Read-Only Contract Report

**Date:** 2026-06-16
**Branch:** `feature/multiple-choice-binary-markets`
**Environment:** local development API at `http://localhost:8080`

## Command

```bash
MAX_EXAMPLES=1 integrationtest/scripts/schemathesis-read.sh
```

## Scope

Default read-only paths:

- `/v0/setup`
- `/v0/setup/frontend`
- `/v0/stats`
- `/v0/market-tags`

Default phase: `coverage`

Default checks:

- `not_a_server_error`
- `status_code_conformance`
- `content_type_conformance`
- `response_schema_conformance`

## Result

| Metric | Value |
| --- | ---: |
| Operations selected | 4 |
| Operations tested | 4 |
| Test cases generated | 4 |
| Passed | 4 |
| Failed | 0 |

Generated artifact, ignored by git:

```text
integrationtest/artifacts/schemathesis-read-20260616T044620Z/junit.xml
```

## Notes

The default runner intentionally avoids mutating routes. Earlier broad runs selected mutation/stateful paths and mostly reported rate-limit/spec-surface findings rather than useful integration failures.
