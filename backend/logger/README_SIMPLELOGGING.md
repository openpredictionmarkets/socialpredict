### SocialPredict Runtime Logger Contract

`backend/logger` is the sole backend runtime log-emission adapter for startup, server, and middleware code.

Use the concrete package helpers:

```go
logger.Info("server", "HTTP server listening", logger.Operation("Start"), logger.Address(":8080"))
logger.Warn("startup", "development environment override not loaded", logger.Operation("LoadDevFile"), logger.Err(err))
logger.Error("auth", "request authentication failed", err, logger.Operation("Authenticate"))
logger.Fatal("startup", "database initialization failed", err, logger.Operation("InitDB"))
```

Stable runtime field vocabulary:

- `component`: runtime area such as `startup`, `server`, or `middleware`
- `operation`: the concrete function or boundary action
- `trace_id`
- `span_id`
- `trace_flags`
- `address`
- `error`
- `source`: caller file and line injected by the logger

Trace correlation:

- Use `logger.TraceContext(traceID, spanID, traceFlags)` when identifiers already exist.
- Use `logger.TraceContextFromTraceparent(r.Header.Get("Traceparent"))` in middleware or transport code that receives W3C trace headers.
- If correlation data is unavailable, omit those fields rather than inventing placeholders.
- These helpers preserve correlation fields only; future tracing or metric export rollout stays in runtime wiring and is not defined by the logger's stdout format.

Secret-safety rules:

- Never log passwords, raw tokens, API keys, cookies, authorization headers, or full request bodies.
- The logger redacts obvious secret-bearing field keys and common `key=value` or `Bearer ...` patterns before emission.
- Redaction is a guardrail, not a license to pass secrets into log messages.

Compatibility:

- `LogInfo`, `LogWarn`, and `LogError` remain as shims for older call sites.
- New runtime callers should prefer `Info`, `Warn`, `Error`, `Fatal`, and the field helpers above so the package vocabulary stays uniform.
