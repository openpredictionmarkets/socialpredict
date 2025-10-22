# Performance Optimization Implementation Plan

## Overview
Optimize the backend system for production-level performance including database query optimization, caching strategies, connection pooling, and system-level performance improvements.

## Current State Analysis
- Basic GORM queries without optimization
- No caching layer implemented
- Simple connection pooling configuration
- No performance monitoring in place
- No query analysis or optimization
- Limited concurrency handling

## Implementation Steps

### Step 1: Database Query Optimization
**Timeline: 3-4 days**

Analyze and optimize database queries for better performance:

```go
// performance/query_optimizer.go
type QueryOptimizer struct {
    db      *gorm.DB
    monitor *QueryMonitor
    cache   *QueryCache
}

func (qo *QueryOptimizer) OptimizeMarketQueries() {
    // Add indexes for frequently queried columns
    qo.db.Exec(`
        CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_markets_status_end_date
        ON markets(status, end_date) WHERE status = 'active';

        CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_markets_creator_status
        ON markets(creator_id, status);

        CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_bets_market_user
        ON bets(market_id, user_id);

        CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_bets_created_at
        ON bets(created_at DESC);
    `)
}

// Optimized queries with proper joins and indexes
func (r *marketRepository) GetActiveMarketsWithStats(ctx context.Context, limit, offset int) ([]*models.MarketWithStats, error) {
    var results []*models.MarketWithStats

    err := r.db.WithContext(ctx).
        Select(`
            markets.*,
            COUNT(DISTINCT bets.id) as bet_count,
            COALESCE(SUM(bets.amount), 0) as total_volume,
            COUNT(DISTINCT bets.user_id) as participant_count
        `).
        Table("markets").
        Joins("LEFT JOIN bets ON markets.id = bets.market_id").
        Where("markets.status = ? AND markets.end_date > ?", "active", time.Now()).
        Group("markets.id").
        Order("markets.created_at DESC").
        Limit(limit).
        Offset(offset).
        Find(&results).Error

    return results, err
}
```

**Query optimization techniques:**
- Strategic index creation
- Query plan analysis
- N+1 query elimination
- Proper JOIN usage
- Subquery optimization
- Bulk operations

### Step 2: Caching Layer Implementation
**Timeline: 4-5 days**

Implement multi-level caching strategy:

```go
// cache/manager.go
type CacheManager struct {
    redis       *redis.Client
    localCache  *ristretto.Cache
    config      CacheConfig
    metrics     *CacheMetrics
}

type CacheConfig struct {
    Redis       RedisConfig       `yaml:"redis"`
    Local       LocalCacheConfig  `yaml:"local"`
    Strategies  []CacheStrategy   `yaml:"strategies"`
}

func (cm *CacheManager) Get(ctx context.Context, key string) (interface{}, error) {
    // Try local cache first (L1)
    if value, found := cm.localCache.Get(key); found {
        cm.metrics.LocalHits.Inc()
        return value, nil
    }

    // Try Redis cache (L2)
    result := cm.redis.Get(ctx, key)
    if result.Err() == nil {
        cm.metrics.RedisHits.Inc()
        value := result.Val()

        // Store in local cache
        cm.localCache.Set(key, value, 1)
        return value, nil
    }

    cm.metrics.Misses.Inc()
    return nil, ErrCacheMiss
}

// Cache-aside pattern implementation
func (s *marketService) GetMarketWithCache(ctx context.Context, marketID uint) (*models.Market, error) {
    cacheKey := fmt.Sprintf("market:%d", marketID)

    // Try cache first
    if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
        var market models.Market
        if err := json.Unmarshal([]byte(cached.(string)), &market); err == nil {
            return &market, nil
        }
    }

    // Cache miss - fetch from database
    market, err := s.repo.GetByID(ctx, marketID)
    if err != nil {
        return nil, err
    }

    // Store in cache
    if data, err := json.Marshal(market); err == nil {
        s.cache.Set(ctx, cacheKey, string(data), 5*time.Minute)
    }

    return market, nil
}
```

**Caching strategies:**
- Application-level caching (local/Redis)
- Database query result caching
- API response caching
- Static content caching
- Session caching
- Cache invalidation patterns

### Step 3: Connection Pool Optimization
**Timeline: 2 days**

Optimize database and external service connection pools:

```go
// performance/connection_pool.go
type ConnectionPoolManager struct {
    db          *gorm.DB
    httpClient  *http.Client
    config      PoolConfig
    monitor     *PoolMonitor
}

type PoolConfig struct {
    Database DatabasePoolConfig `yaml:"database"`
    HTTP     HTTPPoolConfig     `yaml:"http"`
}

func (cpm *ConnectionPoolManager) OptimizeDatabasePool() error {
    sqlDB, err := cpm.db.DB()
    if err != nil {
        return err
    }

    // Calculate optimal pool size based on system resources
    cpuCount := runtime.NumCPU()
    optimalMaxOpen := cpuCount * 4
    optimalMaxIdle := cpuCount * 2

    sqlDB.SetMaxOpenConns(optimalMaxOpen)
    sqlDB.SetMaxIdleConns(optimalMaxIdle)
    sqlDB.SetConnMaxLifetime(time.Hour)
    sqlDB.SetConnMaxIdleTime(30 * time.Minute)

    // Monitor pool statistics
    go cpm.monitorPoolStats(sqlDB)

    return nil
}

func (cpm *ConnectionPoolManager) OptimizeHTTPClient() {
    cpm.httpClient = &http.Client{
        Transport: &http.Transport{
            MaxIdleConns:        100,
            MaxIdleConnsPerHost: 10,
            IdleConnTimeout:     90 * time.Second,
            TLSHandshakeTimeout: 10 * time.Second,
            DialContext: (&net.Dialer{
                Timeout:   30 * time.Second,
                KeepAlive: 30 * time.Second,
            }).DialContext,
        },
        Timeout: 30 * time.Second,
    }
}
```

### Step 4: Response Compression and Optimization
**Timeline: 1-2 days**

Implement response compression and optimization:

```go
// middleware/compression.go
func CompressionMiddleware() mux.MiddlewareFunc {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Check if client accepts compression
            if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
                next.ServeHTTP(w, r)
                return
            }

            // Skip compression for small responses
            cw := &compressResponseWriter{
                ResponseWriter: w,
                minSize:       1024, // Only compress responses > 1KB
            }

            next.ServeHTTP(cw, r)
        })
    }
}

type compressResponseWriter struct {
    http.ResponseWriter
    gzipWriter *gzip.Writer
    minSize    int
    buffer     bytes.Buffer
}

func (cw *compressResponseWriter) Write(data []byte) (int, error) {
    // Buffer response until we know the size
    cw.buffer.Write(data)

    if cw.buffer.Len() > cw.minSize && cw.gzipWriter == nil {
        cw.Header().Set("Content-Encoding", "gzip")
        cw.gzipWriter = gzip.NewWriter(cw.ResponseWriter)
    }

    if cw.gzipWriter != nil {
        return cw.gzipWriter.Write(data)
    }

    return cw.ResponseWriter.Write(data)
}
```

### Step 5: Background Job Processing
**Timeline: 3-4 days**

Implement background job processing for heavy operations:

```go
// jobs/processor.go
type JobProcessor struct {
    workers   int
    queue     chan Job
    wg        sync.WaitGroup
    ctx       context.Context
    cancel    context.CancelFunc
    metrics   *JobMetrics
}

type Job interface {
    Execute(ctx context.Context) error
    GetType() string
    GetPriority() int
    GetRetryCount() int
    ShouldRetry(error) bool
}

type MarketResolutionJob struct {
    MarketID    uint      `json:"market_id"`
    ResolvedBy  uint      `json:"resolved_by"`
    Outcome     string    `json:"outcome"`
    ScheduledAt time.Time `json:"scheduled_at"`
    retryCount  int
}

func (mrj *MarketResolutionJob) Execute(ctx context.Context) error {
    // Heavy market resolution logic
    marketService := GetMarketServiceFromContext(ctx)
    return marketService.ResolveMarket(ctx, mrj.MarketID, mrj.Outcome)
}

func (jp *JobProcessor) Start() {
    jp.ctx, jp.cancel = context.WithCancel(context.Background())

    for i := 0; i < jp.workers; i++ {
        jp.wg.Add(1)
        go jp.worker(i)
    }
}

func (jp *JobProcessor) worker(id int) {
    defer jp.wg.Done()

    for {
        select {
        case job := <-jp.queue:
            start := time.Now()
            err := job.Execute(jp.ctx)
            duration := time.Since(start)

            jp.metrics.JobsProcessed.WithLabelValues(job.GetType()).Inc()
            jp.metrics.JobDuration.WithLabelValues(job.GetType()).Observe(duration.Seconds())

            if err != nil {
                jp.handleJobError(job, err)
            }

        case <-jp.ctx.Done():
            return
        }
    }
}
```

### Step 6: Memory and Resource Optimization
**Timeline: 2-3 days**

Implement memory optimization and resource management:

```go
// performance/memory.go
type MemoryManager struct {
    objectPools map[string]*sync.Pool
    metrics     *MemoryMetrics
}

func (mm *MemoryManager) GetObjectPool(name string, factory func() interface{}) *sync.Pool {
    if pool, exists := mm.objectPools[name]; exists {
        return pool
    }

    pool := &sync.Pool{
        New: factory,
    }
    mm.objectPools[name] = pool
    return pool
}

// Object pooling for frequently allocated objects
var (
    responsePool = sync.Pool{
        New: func() interface{} {
            return &APIResponse{}
        },
    }

    bufferPool = sync.Pool{
        New: func() interface{} {
            return make([]byte, 4096)
        },
    }
)

func GetResponse() *APIResponse {
    resp := responsePool.Get().(*APIResponse)
    // Reset response object
    *resp = APIResponse{}
    return resp
}

func PutResponse(resp *APIResponse) {
    responsePool.Put(resp)
}

// Memory-efficient pagination
func (r *marketRepository) GetMarketsStream(ctx context.Context, filters MarketFilters) (<-chan *models.Market, error) {
    marketChan := make(chan *models.Market, 100) // Buffered channel

    go func() {
        defer close(marketChan)

        offset := 0
        batchSize := 100

        for {
            var markets []*models.Market
            err := r.db.WithContext(ctx).
                Where(filters.ToQuery()).
                Limit(batchSize).
                Offset(offset).
                Find(&markets).Error

            if err != nil {
                return
            }

            if len(markets) == 0 {
                return
            }

            for _, market := range markets {
                select {
                case marketChan <- market:
                case <-ctx.Done():
                    return
                }
            }

            offset += batchSize
        }
    }()

    return marketChan, nil
}
```

### Step 7: Performance Monitoring and Profiling
**Timeline: 2 days**

Implement comprehensive performance monitoring:

```go
// performance/monitoring.go
type PerformanceMonitor struct {
    metrics     *PerformanceMetrics
    profiler    *Profiler
    alerts      *AlertManager
}

func (pm *PerformanceMonitor) StartProfiling() {
    // Enable pprof endpoints
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()

    // Memory usage monitoring
    go pm.monitorMemory()

    // Goroutine monitoring
    go pm.monitorGoroutines()

    // Response time monitoring
    go pm.monitorResponseTimes()
}

func (pm *PerformanceMonitor) monitorMemory() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        var m runtime.MemStats
        runtime.ReadMemStats(&m)

        pm.metrics.MemoryUsage.Set(float64(m.Alloc))
        pm.metrics.MemorySystem.Set(float64(m.Sys))
        pm.metrics.GCCount.Set(float64(m.NumGC))

        // Alert if memory usage is too high
        if m.Alloc > 500*1024*1024 { // 500MB
            pm.alerts.SendAlert("High memory usage detected", map[string]interface{}{
                "alloc": m.Alloc,
                "sys":   m.Sys,
            })
        }
    }
}
```

## Directory Structure
```
performance/
├── query_optimizer.go     # Database query optimization
├── cache/
│   ├── manager.go         # Cache management
│   ├── redis.go           # Redis cache implementation
│   ├── local.go           # Local cache implementation
│   └── strategies.go      # Caching strategies
├── connection_pool.go     # Connection pool optimization
├── compression.go         # Response compression
├── memory.go              # Memory optimization
├── monitoring.go          # Performance monitoring
├── profiling.go           # Performance profiling
└── jobs/
    ├── processor.go       # Background job processing
    ├── queue.go           # Job queue management
    └── workers.go         # Worker pool management

middleware/
├── compression.go         # Compression middleware
├── caching.go            # Caching middleware
└── performance.go        # Performance tracking middleware
```

## Performance Configuration
```yaml
performance:
  database:
    max_open_conns: 25
    max_idle_conns: 10
    conn_max_lifetime: "1h"
    slow_query_threshold: "100ms"

  cache:
    redis:
      enabled: true
      address: "redis:6379"
      db: 0
      max_retries: 3
    local:
      max_entries: 10000
      ttl: "5m"

  compression:
    enabled: true
    min_size: 1024
    level: 6

  jobs:
    workers: 4
    queue_size: 1000
    retry_attempts: 3

  monitoring:
    enable_profiling: true
    profiling_port: 6060
    memory_threshold: "500MB"
    response_time_threshold: "200ms"
```

## Performance Benchmarks
```go
// benchmarks/api_benchmark_test.go
func BenchmarkMarketsAPI(b *testing.B) {
    ts := NewTestSuite()
    defer ts.Cleanup()

    // Seed test data
    ts.SeedTestMarkets(1000)

    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            req := httptest.NewRequest("GET", "/v1/markets", nil)
            req.Header.Set("Authorization", "Bearer "+ts.token)

            w := httptest.NewRecorder()
            ts.Server.Handler.ServeHTTP(w, req)

            if w.Code != http.StatusOK {
                b.Errorf("Expected 200, got %d", w.Code)
            }
        }
    })
}

func BenchmarkDatabaseQuery(b *testing.B) {
    ts := NewTestSuite()
    defer ts.Cleanup()

    ts.SeedTestMarkets(10000)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        markets, err := ts.MarketRepo.GetActiveMarkets(context.Background(), 20, 0)
        if err != nil {
            b.Fatal(err)
        }
        if len(markets) == 0 {
            b.Fatal("No markets returned")
        }
    }
}
```

## Performance Targets
- API response time: <200ms (95th percentile)
- Database query time: <50ms (average)
- Memory usage: <512MB under normal load
- Cache hit ratio: >80%
- Concurrent users: 1000+
- Throughput: 1000 requests/second

## Monitoring Dashboards
- Response time percentiles
- Database query performance
- Cache hit/miss ratios
- Memory and CPU usage
- Goroutine count
- Connection pool utilization
- Background job processing rates

## Benefits
- Improved user experience with faster response times
- Better resource utilization
- Higher system capacity and scalability
- Reduced infrastructure costs
- Better system reliability under load
- Proactive performance issue detection