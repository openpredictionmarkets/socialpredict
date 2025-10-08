# Monitoring & Alerting Implementation Plan

## Overview
Implement comprehensive monitoring, alerting, and observability systems to ensure system reliability, performance tracking, and proactive incident response.

## Current State Analysis
- No centralized monitoring system
- Basic logging without aggregation
- No alerting mechanisms
- Limited metrics collection
- No dashboards or visualization
- No incident response procedures

## Implementation Steps

### Step 1: Metrics Collection and Exposition
**Timeline: 2-3 days**

Implement comprehensive metrics collection using Prometheus:

```go
// monitoring/metrics.go
package monitoring

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // HTTP metrics
    HTTPRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )

    HTTPRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "endpoint"},
    )

    // Database metrics
    DatabaseConnections = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "database_connections",
            Help: "Number of database connections",
        },
        []string{"state"}, // open, idle, in_use
    )

    DatabaseQueryDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "database_query_duration_seconds",
            Help:    "Database query duration in seconds",
            Buckets: []float64{.001, .005, .01, .05, .1, .5, 1, 5},
        },
        []string{"operation", "table"},
    )

    // Business metrics
    MarketsCreated = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "markets_created_total",
            Help: "Total number of markets created",
        },
    )

    BetsPlaced = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "bets_placed_total",
            Help: "Total number of bets placed",
        },
        []string{"market_id", "outcome"},
    )

    UserRegistrations = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "user_registrations_total",
            Help: "Total number of user registrations",
        },
    )

    // System metrics
    GoRoutines = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "goroutines_count",
            Help: "Number of goroutines",
        },
    )

    MemoryUsage = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "memory_usage_bytes",
            Help: "Memory usage in bytes",
        },
    )
)

// Middleware to collect HTTP metrics
func MetricsMiddleware() mux.MiddlewareFunc {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()

            // Wrap response writer to capture status code
            ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

            next.ServeHTTP(ww, r)

            duration := time.Since(start).Seconds()
            endpoint := getEndpointName(r.URL.Path)

            HTTPRequestsTotal.WithLabelValues(r.Method, endpoint, strconv.Itoa(ww.statusCode)).Inc()
            HTTPRequestDuration.WithLabelValues(r.Method, endpoint).Observe(duration)
        })
    }
}
```

### Step 2: Prometheus Setup and Configuration
**Timeline: 1-2 days**

Configure Prometheus for metrics collection:

```yaml
# monitoring/prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "rules/*.yml"

alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - alertmanager:9093

scrape_configs:
  - job_name: 'socialpredict-backend'
    static_configs:
      - targets: ['socialpredict:8080']
    metrics_path: '/metrics'
    scrape_interval: 30s
    scrape_timeout: 10s

  - job_name: 'postgres'
    static_configs:
      - targets: ['postgres-exporter:9187']
    scrape_interval: 30s

  - job_name: 'redis'
    static_configs:
      - targets: ['redis-exporter:9121']
    scrape_interval: 30s

  - job_name: 'node-exporter'
    static_configs:
      - targets: ['node-exporter:9100']
    scrape_interval: 30s

  - job_name: 'kubernetes-pods'
    kubernetes_sd_configs:
      - role: pod
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
        action: keep
        regex: true
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
        action: replace
        target_label: __metrics_path__
        regex: (.+)
```

### Step 3: Alerting Rules and Alert Manager
**Timeline: 2-3 days**

Configure comprehensive alerting rules:

```yaml
# monitoring/rules/socialpredict.yml
groups:
- name: socialpredict.rules
  rules:
  # High-level service alerts
  - alert: ServiceDown
    expr: up{job="socialpredict-backend"} == 0
    for: 1m
    labels:
      severity: critical
      service: socialpredict
    annotations:
      summary: "SocialPredict service is down"
      description: "SocialPredict service has been down for more than 1 minute"
      runbook_url: "https://wiki.example.com/runbooks/service-down"

  - alert: HighErrorRate
    expr: rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m]) > 0.05
    for: 5m
    labels:
      severity: critical
      service: socialpredict
    annotations:
      summary: "High error rate detected"
      description: "Error rate is {{ $value | humanizePercentage }} for the last 5 minutes"

  - alert: HighLatency
    expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 0.5
    for: 5m
    labels:
      severity: warning
      service: socialpredict
    annotations:
      summary: "High latency detected"
      description: "95th percentile latency is {{ $value }}s"

  # Database alerts
  - alert: DatabaseConnectionPoolExhausted
    expr: database_connections{state="in_use"} / database_connections{state="max"} > 0.9
    for: 2m
    labels:
      severity: warning
      service: database
    annotations:
      summary: "Database connection pool nearly exhausted"
      description: "{{ $value | humanizePercentage }} of database connections are in use"

  - alert: SlowDatabaseQueries
    expr: histogram_quantile(0.95, rate(database_query_duration_seconds_bucket[5m])) > 1
    for: 5m
    labels:
      severity: warning
      service: database
    annotations:
      summary: "Slow database queries detected"
      description: "95th percentile query time is {{ $value }}s"

  # Business logic alerts
  - alert: NoMarketsCreated
    expr: increase(markets_created_total[1h]) == 0
    for: 1h
    labels:
      severity: warning
      service: business
    annotations:
      summary: "No markets created in the last hour"
      description: "This might indicate an issue with market creation functionality"

  - alert: UnusualBettingActivity
    expr: rate(bets_placed_total[5m]) > 100
    for: 5m
    labels:
      severity: warning
      service: business
    annotations:
      summary: "Unusual betting activity detected"
      description: "Betting rate is {{ $value }} bets per second"

  # System alerts
  - alert: HighMemoryUsage
    expr: memory_usage_bytes / (1024*1024*1024) > 1
    for: 5m
    labels:
      severity: warning
      service: system
    annotations:
      summary: "High memory usage"
      description: "Memory usage is {{ $value }}GB"

  - alert: TooManyGoroutines
    expr: goroutines_count > 1000
    for: 5m
    labels:
      severity: warning
      service: system
    annotations:
      summary: "Too many goroutines"
      description: "Goroutine count is {{ $value }}"

# Alert Manager configuration
# monitoring/alertmanager.yml
global:
  smtp_smarthost: 'localhost:587'
  smtp_from: 'alerts@socialpredict.com'

route:
  group_by: ['alertname', 'service']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'web.hook'
  routes:
  - match:
      severity: critical
    receiver: 'critical-alerts'
  - match:
      service: database
    receiver: 'database-team'

receivers:
- name: 'web.hook'
  webhook_configs:
  - url: 'http://webhook:5001/'

- name: 'critical-alerts'
  email_configs:
  - to: 'oncall@socialpredict.com'
    subject: 'CRITICAL: {{ .GroupLabels.alertname }}'
    body: |
      {{ range .Alerts }}
      Alert: {{ .Annotations.summary }}
      Description: {{ .Annotations.description }}
      {{ end }}
  slack_configs:
  - api_url: 'YOUR_SLACK_WEBHOOK_URL'
    channel: '#alerts'
    title: 'CRITICAL Alert'
    text: '{{ range .Alerts }}{{ .Annotations.summary }}{{ end }}'

- name: 'database-team'
  email_configs:
  - to: 'db-team@socialpredict.com'
    subject: 'Database Alert: {{ .GroupLabels.alertname }}'
```

### Step 4: Grafana Dashboards
**Timeline: 3-4 days**

Create comprehensive monitoring dashboards:

```json
// monitoring/dashboards/socialpredict-overview.json
{
  "dashboard": {
    "id": null,
    "title": "SocialPredict Overview",
    "tags": ["socialpredict"],
    "timezone": "browser",
    "panels": [
      {
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total[5m])",
            "legendFormat": "{{method}} {{endpoint}}"
          }
        ],
        "yAxes": [
          {
            "label": "Requests/sec"
          }
        ]
      },
      {
        "title": "Error Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total{status=~\"4..|5..\"}[5m]) / rate(http_requests_total[5m])",
            "legendFormat": "Error Rate"
          }
        ],
        "yAxes": [
          {
            "label": "Error Rate",
            "max": 1,
            "min": 0
          }
        ]
      },
      {
        "title": "Response Time Percentiles",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.50, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "50th percentile"
          },
          {
            "expr": "histogram_quantile(0.90, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "90th percentile"
          },
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "95th percentile"
          }
        ]
      },
      {
        "title": "Database Connections",
        "type": "graph",
        "targets": [
          {
            "expr": "database_connections",
            "legendFormat": "{{state}}"
          }
        ]
      },
      {
        "title": "Business Metrics",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(markets_created_total[1h])",
            "legendFormat": "Markets Created/hour"
          },
          {
            "expr": "rate(bets_placed_total[1h])",
            "legendFormat": "Bets Placed/hour"
          },
          {
            "expr": "rate(user_registrations_total[1h])",
            "legendFormat": "User Registrations/hour"
          }
        ]
      }
    ],
    "time": {
      "from": "now-1h",
      "to": "now"
    },
    "refresh": "30s"
  }
}
```

### Step 5: Log Aggregation and Analysis
**Timeline: 2-3 days**

Set up centralized logging with ELK stack:

```yaml
# monitoring/elasticsearch.yml
version: '3.8'
services:
  elasticsearch:
    image: docker.elastic.co/elasticsearch/elasticsearch:8.8.0
    environment:
      - discovery.type=single-node
      - "ES_JAVA_OPTS=-Xms512m -Xmx512m"
      - xpack.security.enabled=false
    ports:
      - "9200:9200"
    volumes:
      - elasticsearch_data:/usr/share/elasticsearch/data

  logstash:
    image: docker.elastic.co/logstash/logstash:8.8.0
    ports:
      - "5044:5044"
    volumes:
      - ./logstash.conf:/usr/share/logstash/pipeline/logstash.conf
    depends_on:
      - elasticsearch

  kibana:
    image: docker.elastic.co/kibana/kibana:8.8.0
    ports:
      - "5601:5601"
    environment:
      - ELASTICSEARCH_HOSTS=http://elasticsearch:9200
    depends_on:
      - elasticsearch

# Logstash configuration
# monitoring/logstash.conf
input {
  beats {
    port => 5044
  }
}

filter {
  if [fields][service] == "socialpredict" {
    json {
      source => "message"
    }

    date {
      match => [ "timestamp", "ISO8601" ]
    }

    if [level] == "ERROR" {
      mutate {
        add_tag => [ "error" ]
      }
    }
  }
}

output {
  elasticsearch {
    hosts => ["http://elasticsearch:9200"]
    index => "socialpredict-%{+YYYY.MM.dd}"
  }
}
```

### Step 6: Health Check System
**Timeline: 1-2 days**

Implement comprehensive health checks:

```go
// health/checker.go
package health

type HealthChecker struct {
    checks   map[string]HealthCheck
    timeout  time.Duration
    logger   *logging.Logger
}

type HealthCheck interface {
    Name() string
    Check(ctx context.Context) error
}

type HealthStatus struct {
    Status    string                 `json:"status"`
    Timestamp time.Time              `json:"timestamp"`
    Checks    map[string]CheckResult `json:"checks"`
    Version   string                 `json:"version"`
    Uptime    time.Duration          `json:"uptime"`
}

type CheckResult struct {
    Status  string        `json:"status"`
    Message string        `json:"message,omitempty"`
    Latency time.Duration `json:"latency"`
}

func (hc *HealthChecker) CheckHealth(ctx context.Context) *HealthStatus {
    status := &HealthStatus{
        Status:    "healthy",
        Timestamp: time.Now(),
        Checks:    make(map[string]CheckResult),
        Version:   version.GetVersion(),
        Uptime:    time.Since(startTime),
    }

    var wg sync.WaitGroup
    resultChan := make(chan struct {
        name   string
        result CheckResult
    }, len(hc.checks))

    // Run all checks concurrently
    for name, check := range hc.checks {
        wg.Add(1)
        go func(name string, check HealthCheck) {
            defer wg.Done()

            start := time.Now()
            ctx, cancel := context.WithTimeout(ctx, hc.timeout)
            defer cancel()

            err := check.Check(ctx)
            latency := time.Since(start)

            result := CheckResult{
                Latency: latency,
            }

            if err != nil {
                result.Status = "unhealthy"
                result.Message = err.Error()
                status.Status = "unhealthy"
            } else {
                result.Status = "healthy"
            }

            resultChan <- struct {
                name   string
                result CheckResult
            }{name, result}
        }(name, check)
    }

    go func() {
        wg.Wait()
        close(resultChan)
    }()

    // Collect results
    for result := range resultChan {
        status.Checks[result.name] = result.result
    }

    return status
}

// HTTP handlers for health checks
func (hc *HealthChecker) LivenessHandler(w http.ResponseWriter, r *http.Request) {
    // Simple liveness check - just return 200 OK
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}

func (hc *HealthChecker) ReadinessHandler(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()

    status := hc.CheckHealth(ctx)

    w.Header().Set("Content-Type", "application/json")

    if status.Status == "healthy" {
        w.WriteHeader(http.StatusOK)
    } else {
        w.WriteHeader(http.StatusServiceUnavailable)
    }

    json.NewEncoder(w).Encode(status)
}
```

### Step 7: Incident Response and Runbooks
**Timeline: 2 days**

Create incident response procedures and runbooks:

```yaml
# monitoring/runbooks/service-down.md
# Service Down Runbook

## Alert: ServiceDown

### Description
The SocialPredict backend service is not responding to health checks.

### Immediate Actions
1. Check service status: `kubectl get pods -n socialpredict`
2. Check recent logs: `kubectl logs -n socialpredict -l app=socialpredict-backend --tail=100`
3. Check resource usage: `kubectl top pods -n socialpredict`

### Investigation Steps
1. **Check if pods are running**
   ```bash
   kubectl get pods -n socialpredict
   ```

2. **Check pod events**
   ```bash
   kubectl describe pod <pod-name> -n socialpredict
   ```

3. **Check application logs**
   ```bash
   kubectl logs <pod-name> -n socialpredict --previous
   ```

4. **Check database connectivity**
   ```bash
   kubectl exec -it <pod-name> -n socialpredict -- /bin/sh
   # Test database connection
   ```

### Resolution Steps
1. **If pod is CrashLooping**: Check logs for error messages and fix the underlying issue
2. **If pod is OOMKilled**: Increase memory limits in deployment
3. **If database is unreachable**: Check database pod status and network policies
4. **If configuration issue**: Update ConfigMap and restart pods

### Escalation
- If issue persists for >15 minutes, escalate to @oncall-team
- For database issues, escalate to @database-team
- For infrastructure issues, escalate to @platform-team

### Post-Incident
1. Update runbook with lessons learned
2. Create post-mortem if outage >5 minutes
3. Review monitoring and alerting effectiveness
```

## Directory Structure
```
monitoring/
├── metrics/
│   ├── collectors.go           # Custom metric collectors
│   ├── middleware.go          # Metrics middleware
│   └── registry.go            # Metrics registry
├── health/
│   ├── checker.go             # Health check system
│   ├── checks/                # Individual health checks
│   └── handlers.go            # HTTP health endpoints
├── config/
│   ├── prometheus.yml         # Prometheus configuration
│   ├── alertmanager.yml       # Alert Manager configuration
│   └── grafana/               # Grafana dashboards
├── rules/
│   ├── socialpredict.yml      # Application alerts
│   ├── infrastructure.yml     # Infrastructure alerts
│   └── business.yml           # Business logic alerts
├── dashboards/
│   ├── overview.json          # Main overview dashboard
│   ├── performance.json       # Performance dashboard
│   ├── business.json          # Business metrics dashboard
│   └── infrastructure.json    # Infrastructure dashboard
├── logging/
│   ├── logstash.conf          # Logstash configuration
│   ├── filebeat.yml           # Filebeat configuration
│   └── kibana/                # Kibana configurations
└── runbooks/
    ├── service-down.md        # Service outage runbook
    ├── high-latency.md        # High latency runbook
    ├── database-issues.md     # Database issue runbook
    └── scaling.md             # Scaling procedures
```

## Key Metrics to Monitor

### Application Metrics
- Request rate and latency
- Error rates by endpoint
- Business metrics (markets, bets, users)
- Authentication success/failure rates

### Infrastructure Metrics
- CPU and memory usage
- Database connection pool status
- Cache hit/miss ratios
- Network I/O

### Business Metrics
- Market creation rate
- Betting volume and frequency
- User engagement metrics
- Revenue metrics

## Alert Severity Levels

### Critical (Page immediately)
- Service completely down
- High error rates (>5%)
- Database unavailable
- Security incidents

### Warning (Notify during business hours)
- High latency (>500ms p95)
- Resource utilization >80%
- Slow database queries
- Unusual business activity

### Info (Log for review)
- Performance degradation
- Non-critical feature failures
- Capacity planning triggers

## Benefits
- Proactive issue detection
- Faster incident response
- Better system visibility
- Data-driven optimization decisions
- Improved reliability and uptime
- Comprehensive audit trail