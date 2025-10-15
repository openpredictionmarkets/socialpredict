# Logging & Observability Implementation Plan

## Overview
Transform the basic logging system into a comprehensive observability platform with structured logging, metrics, tracing, and health monitoring.

## Current State Analysis
- Basic `log` package usage in `main.go` and `server.go`
- Simple logging package in `logger/simplelogging.go`
- No structured logging
- No metrics collection
- No distributed tracing
- No health check endpoints

## Implementation Steps

### Step 1: Structured Logging Framework
**Timeline: 2-3 days**

Replace basic logging with structured logging using `logrus` or `zap`:

```go
// logging/logger.go
type Logger struct {
    logger *logrus.Logger
    fields logrus.Fields
}

func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
    return &Logger{
        logger: l.logger,
        fields: logrus.Fields(fields),
    }
}
```

**Features:**
- JSON structured logging
- Log levels (DEBUG, INFO, WARN, ERROR, FATAL)
- Contextual fields (request_id, user_id, etc.)
- Performance-optimized logging

### Step 2: Request Logging Middleware
**Timeline: 1 day**

Create comprehensive request logging middleware:

```go
// middleware/logging.go
func LoggingMiddleware(logger *logging.Logger) mux.MiddlewareFunc {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            // Log request details, response codes, timing
        })
    }
}
```

**Logged fields:**
- Request ID (correlation)
- HTTP method and path
- Request/response size
- Response time
- Status code
- User agent and IP
- Error details

### Step 3: Metrics Collection
**Timeline: 2-3 days**

Implement Prometheus metrics collection:

```go
// metrics/metrics.go
var (
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )

    httpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "HTTP request duration in seconds",
        },
        []string{"method", "endpoint"},
    )
)
```

**Metrics to track:**
- HTTP request counts by endpoint/method/status
- Request duration histograms
- Database query metrics
- Business metrics (bets placed, markets created)
- System metrics (memory, CPU, goroutines)

### Step 4: Health Check System
**Timeline: 1-2 days**

Implement comprehensive health checks:

```go
// health/health.go
type HealthChecker struct {
    checks map[string]HealthCheck
}

type HealthCheck interface {
    Name() string
    Check(ctx context.Context) error
}

type DatabaseHealthCheck struct {
    db *gorm.DB
}
```

**Health checks:**
- Database connectivity
- External service dependencies
- Memory usage
- Disk space
- Custom business logic checks

### Step 5: Distributed Tracing
**Timeline: 2-3 days**

Implement OpenTelemetry tracing:

```go
// tracing/tracer.go
func InitTracer(serviceName string) (trace.Tracer, error) {
    exporter, err := jaeger.New(jaeger.WithCollectorEndpoint())
    if err != nil {
        return nil, err
    }

    tp := trace.NewTracerProvider(
        trace.WithBatcher(exporter),
        trace.WithResource(resource.NewWithAttributes(
            semconv.ServiceNameKey.String(serviceName),
        )),
    )

    otel.SetTracerProvider(tp)
    return tp.Tracer(serviceName), nil
}
```

**Tracing features:**
- HTTP request tracing
- Database query tracing
- Cross-service correlation
- Error tracking in spans

### Step 6: Log Aggregation Setup
**Timeline: 1-2 days**

Configure log shipping to centralized logging:

```go
// logging/aggregation.go
type LogShipper struct {
    endpoint string
    buffer   []LogEntry
    interval time.Duration
}

func (ls *LogShipper) Ship() error {
    // Send logs to ELK stack or similar
}
```

**Features:**
- Buffered log shipping
- Retry logic with exponential backoff
- Log filtering and sampling
- Structured log format

## Directory Structure
```
observability/
├── logging/
│   ├── logger.go           # Main logging interface
│   ├── structured.go       # Structured logging implementation
│   ├── middleware.go       # HTTP logging middleware
│   └── shipper.go          # Log aggregation
├── metrics/
│   ├── metrics.go          # Prometheus metrics
│   ├── collectors.go       # Custom metric collectors
│   └── middleware.go       # Metrics middleware
├── tracing/
│   ├── tracer.go           # OpenTelemetry setup
│   ├── middleware.go       # Tracing middleware
│   └── spans.go            # Custom span utilities
├── health/
│   ├── health.go           # Health check framework
│   ├── checks.go           # Individual health checks
│   └── handlers.go         # Health check HTTP endpoints
└── alerts/
    ├── alerting.go         # Alert rule definitions
    └── notifications.go    # Alert notification handlers
```

## Dependencies
- `github.com/sirupsen/logrus` or `go.uber.org/zap` - Structured logging
- `github.com/prometheus/client_golang` - Metrics collection
- `go.opentelemetry.io/otel` - Distributed tracing
- `github.com/google/uuid` - Request ID generation

## New HTTP Endpoints
```
GET /health           # Basic health check
GET /health/live      # Liveness probe
GET /health/ready     # Readiness probe
GET /health/detailed  # Detailed health status
GET /metrics          # Prometheus metrics endpoint
```

## Configuration Integration
```yaml
logging:
  level: "info"
  format: "json"
  output: "stdout"
  aggregation:
    enabled: true
    endpoint: "http://logstash:5000"

metrics:
  enabled: true
  port: 9090
  namespace: "socialpredict"

tracing:
  enabled: true
  jaeger_endpoint: "http://jaeger:14268/api/traces"
  sample_rate: 0.1
```

## Testing Strategy
- Unit tests for all logging components
- Integration tests for metrics collection
- Health check validation tests
- Load testing with observability enabled
- Alert rule testing

## Migration Strategy
1. Implement logging framework alongside existing logging
2. Gradually replace `log` calls with structured logging
3. Deploy metrics collection without breaking changes
4. Add health checks incrementally
5. Enable tracing for critical paths first

## Benefits
- Complete request visibility
- Performance monitoring and optimization
- Proactive issue detection
- Debugging and troubleshooting capabilities
- SLA monitoring and reporting
- Operational insights

## Monitoring Dashboard Requirements
- Request rate and error rate graphs
- Response time percentiles
- Database performance metrics
- System resource utilization
- Business metrics dashboard
- Alert status and history