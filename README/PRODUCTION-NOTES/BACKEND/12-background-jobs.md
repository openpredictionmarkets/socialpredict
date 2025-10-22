# Background Jobs & Task Processing Implementation Plan

## Overview
Implement a robust background job processing system to handle asynchronous tasks, scheduled operations, and heavy computations without blocking the main application flow.

## Current State Analysis
- No background job processing system
- All operations are synchronous
- Heavy operations block HTTP requests
- No scheduled task capabilities
- No async processing for emails, notifications, etc.
- Limited scalability for compute-intensive tasks

## Implementation Steps

### Step 1: Job Queue System Design
**Timeline: 3-4 days**

Design and implement a flexible job queue system:

```go
// jobs/queue.go
package jobs

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
)

type Job interface {
    GetID() string
    GetType() string
    GetPriority() Priority
    GetPayload() []byte
    GetRetryPolicy() RetryPolicy
    Execute(ctx context.Context) error
}

type Priority int

const (
    PriorityLow Priority = iota
    PriorityNormal
    PriorityHigh
    PriorityUrgent
)

type RetryPolicy struct {
    MaxRetries    int           `json:"max_retries"`
    InitialDelay  time.Duration `json:"initial_delay"`
    MaxDelay      time.Duration `json:"max_delay"`
    BackoffFactor float64       `json:"backoff_factor"`
}

type JobStatus string

const (
    StatusPending   JobStatus = "pending"
    StatusRunning   JobStatus = "running"
    StatusCompleted JobStatus = "completed"
    StatusFailed    JobStatus = "failed"
    StatusRetrying  JobStatus = "retrying"
)

type JobRecord struct {
    ID          string                 `json:"id" gorm:"primaryKey"`
    Type        string                 `json:"type" gorm:"index"`
    Priority    Priority               `json:"priority" gorm:"index"`
    Status      JobStatus              `json:"status" gorm:"index"`
    Payload     []byte                 `json:"payload"`
    Result      []byte                 `json:"result,omitempty"`
    Error       string                 `json:"error,omitempty"`
    RetryCount  int                    `json:"retry_count"`
    RetryPolicy RetryPolicy            `json:"retry_policy" gorm:"embedded"`
    Metadata    map[string]interface{} `json:"metadata" gorm:"serializer:json"`
    CreatedAt   time.Time              `json:"created_at"`
    StartedAt   *time.Time             `json:"started_at,omitempty"`
    CompletedAt *time.Time             `json:"completed_at,omitempty"`
    ScheduledAt time.Time              `json:"scheduled_at" gorm:"index"`
}

type Queue interface {
    Enqueue(ctx context.Context, job Job) error
    EnqueueDelayed(ctx context.Context, job Job, delay time.Duration) error
    EnqueueAt(ctx context.Context, job Job, at time.Time) error
    Dequeue(ctx context.Context, jobTypes []string) (*JobRecord, error)
    UpdateJobStatus(ctx context.Context, jobID string, status JobStatus, result []byte, err error) error
    GetJob(ctx context.Context, jobID string) (*JobRecord, error)
    ListJobs(ctx context.Context, filters JobFilters) ([]*JobRecord, error)
}

// Redis-based queue implementation
type RedisQueue struct {
    redis       *redis.Client
    db          *gorm.DB
    keyPrefix   string
    metrics     *JobMetrics
    logger      *logging.Logger
}

func NewRedisQueue(redis *redis.Client, db *gorm.DB) *RedisQueue {
    return &RedisQueue{
        redis:     redis,
        db:        db,
        keyPrefix: "socialpredict:jobs:",
        metrics:   NewJobMetrics(),
        logger:    logging.NewLogger(),
    }
}

func (rq *RedisQueue) Enqueue(ctx context.Context, job Job) error {
    // Save job to database
    record := &JobRecord{
        ID:          job.GetID(),
        Type:        job.GetType(),
        Priority:    job.GetPriority(),
        Status:      StatusPending,
        Payload:     job.GetPayload(),
        RetryPolicy: job.GetRetryPolicy(),
        CreatedAt:   time.Now(),
        ScheduledAt: time.Now(),
    }

    if err := rq.db.WithContext(ctx).Create(record).Error; err != nil {
        return fmt.Errorf("failed to save job record: %w", err)
    }

    // Add to Redis queue with priority
    queueKey := fmt.Sprintf("%squeue:%s", rq.keyPrefix, job.GetType())
    score := float64(time.Now().Unix()) - float64(job.GetPriority()*1000000)

    if err := rq.redis.ZAdd(ctx, queueKey, &redis.Z{
        Score:  score,
        Member: job.GetID(),
    }).Err(); err != nil {
        return fmt.Errorf("failed to enqueue job: %w", err)
    }

    rq.metrics.JobsEnqueued.WithLabelValues(job.GetType()).Inc()
    return nil
}

func (rq *RedisQueue) Dequeue(ctx context.Context, jobTypes []string) (*JobRecord, error) {
    // Try to dequeue from multiple job type queues in priority order
    for _, jobType := range jobTypes {
        queueKey := fmt.Sprintf("%squeue:%s", rq.keyPrefix, jobType)

        // Get highest priority job (lowest score)
        result := rq.redis.ZPopMin(ctx, queueKey, 1)
        if result.Err() != nil {
            if result.Err() == redis.Nil {
                continue // No jobs in this queue
            }
            return nil, result.Err()
        }

        if len(result.Val()) == 0 {
            continue
        }

        jobID := result.Val()[0].Member.(string)

        // Get job record from database
        var record JobRecord
        if err := rq.db.WithContext(ctx).Where("id = ?", jobID).First(&record).Error; err != nil {
            rq.logger.WithFields(map[string]interface{}{
                "job_id": jobID,
                "error":  err.Error(),
            }).Error("Failed to get job record")
            continue
        }

        // Update status to running
        now := time.Now()
        record.Status = StatusRunning
        record.StartedAt = &now

        if err := rq.db.WithContext(ctx).Save(&record).Error; err != nil {
            rq.logger.WithFields(map[string]interface{}{
                "job_id": jobID,
                "error":  err.Error(),
            }).Error("Failed to update job status")
            continue
        }

        rq.metrics.JobsDequeued.WithLabelValues(jobType).Inc()
        return &record, nil
    }

    return nil, nil // No jobs available
}
```

### Step 2: Worker Pool Implementation
**Timeline: 2-3 days**

Create a scalable worker pool system:

```go
// jobs/worker.go
package jobs

type WorkerPool struct {
    workerCount    int
    queue          Queue
    jobHandlers    map[string]JobHandler
    workers        []*Worker
    ctx            context.Context
    cancel         context.CancelFunc
    wg             sync.WaitGroup
    metrics        *JobMetrics
    logger         *logging.Logger
    healthChecker  *HealthChecker
}

type JobHandler interface {
    Handle(ctx context.Context, payload []byte) ([]byte, error)
    GetJobType() string
    GetTimeout() time.Duration
}

type Worker struct {
    id            int
    pool          *WorkerPool
    queue         Queue
    handlers      map[string]JobHandler
    currentJob    *JobRecord
    metrics       *JobMetrics
    logger        *logging.Logger
    lastHeartbeat time.Time
    mutex         sync.RWMutex
}

func NewWorkerPool(workerCount int, queue Queue) *WorkerPool {
    ctx, cancel := context.WithCancel(context.Background())

    return &WorkerPool{
        workerCount:   workerCount,
        queue:         queue,
        jobHandlers:   make(map[string]JobHandler),
        ctx:           ctx,
        cancel:        cancel,
        metrics:       NewJobMetrics(),
        logger:        logging.NewLogger(),
        healthChecker: NewHealthChecker(),
    }
}

func (wp *WorkerPool) RegisterHandler(handler JobHandler) {
    wp.jobHandlers[handler.GetJobType()] = handler
}

func (wp *WorkerPool) Start() error {
    wp.logger.WithFields(map[string]interface{}{
        "worker_count": wp.workerCount,
        "job_types":    wp.getJobTypes(),
    }).Info("Starting worker pool")

    // Start workers
    for i := 0; i < wp.workerCount; i++ {
        worker := &Worker{
            id:       i,
            pool:     wp,
            queue:    wp.queue,
            handlers: wp.jobHandlers,
            metrics:  wp.metrics,
            logger:   wp.logger,
        }

        wp.workers = append(wp.workers, worker)
        wp.wg.Add(1)
        go worker.run(wp.ctx)
    }

    // Start health checker
    go wp.healthChecker.start(wp.ctx, wp.workers)

    return nil
}

func (wp *WorkerPool) Stop() error {
    wp.logger.Info("Stopping worker pool")

    wp.cancel()
    wp.wg.Wait()

    wp.logger.Info("Worker pool stopped")
    return nil
}

func (w *Worker) run(ctx context.Context) {
    defer w.pool.wg.Done()

    w.logger.WithFields(map[string]interface{}{
        "worker_id": w.id,
    }).Info("Starting worker")

    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            w.logger.WithFields(map[string]interface{}{
                "worker_id": w.id,
            }).Info("Worker stopping")
            return

        case <-ticker.C:
            // Try to get a job
            job, err := w.queue.Dequeue(ctx, w.pool.getJobTypes())
            if err != nil {
                w.logger.WithFields(map[string]interface{}{
                    "worker_id": w.id,
                    "error":     err.Error(),
                }).Error("Failed to dequeue job")
                continue
            }

            if job == nil {
                // No jobs available, update heartbeat and continue
                w.updateHeartbeat()
                continue
            }

            // Process the job
            w.processJob(ctx, job)
        }
    }
}

func (w *Worker) processJob(ctx context.Context, job *JobRecord) {
    w.mutex.Lock()
    w.currentJob = job
    w.mutex.Unlock()

    defer func() {
        w.mutex.Lock()
        w.currentJob = nil
        w.mutex.Unlock()
    }()

    w.logger.WithFields(map[string]interface{}{
        "worker_id": w.id,
        "job_id":    job.ID,
        "job_type":  job.Type,
    }).Info("Processing job")

    start := time.Now()

    // Get handler for job type
    handler, exists := w.handlers[job.Type]
    if !exists {
        w.logger.WithFields(map[string]interface{}{
            "job_id":   job.ID,
            "job_type": job.Type,
        }).Error("No handler found for job type")

        w.queue.UpdateJobStatus(ctx, job.ID, StatusFailed, nil,
            fmt.Errorf("no handler found for job type: %s", job.Type))
        return
    }

    // Create timeout context
    jobCtx, cancel := context.WithTimeout(ctx, handler.GetTimeout())
    defer cancel()

    // Execute job
    result, err := handler.Handle(jobCtx, job.Payload)
    duration := time.Since(start)

    // Update metrics
    w.metrics.JobDuration.WithLabelValues(job.Type).Observe(duration.Seconds())

    if err != nil {
        w.logger.WithFields(map[string]interface{}{
            "worker_id": w.id,
            "job_id":    job.ID,
            "job_type":  job.Type,
            "error":     err.Error(),
            "duration":  duration,
        }).Error("Job failed")

        w.metrics.JobsFailed.WithLabelValues(job.Type).Inc()

        // Handle retry logic
        if w.shouldRetry(job, err) {
            w.scheduleRetry(ctx, job, err)
        } else {
            w.queue.UpdateJobStatus(ctx, job.ID, StatusFailed, nil, err)
        }
    } else {
        w.logger.WithFields(map[string]interface{}{
            "worker_id": w.id,
            "job_id":    job.ID,
            "job_type":  job.Type,
            "duration":  duration,
        }).Info("Job completed successfully")

        w.metrics.JobsCompleted.WithLabelValues(job.Type).Inc()
        w.queue.UpdateJobStatus(ctx, job.ID, StatusCompleted, result, nil)
    }
}
```

### Step 3: Job Implementations
**Timeline: 3-4 days**

Create specific job implementations for common tasks:

```go
// jobs/handlers/email.go
package handlers

type EmailJob struct {
    ID          string    `json:"id"`
    To          string    `json:"to"`
    Subject     string    `json:"subject"`
    Body        string    `json:"body"`
    Template    string    `json:"template,omitempty"`
    TemplateData map[string]interface{} `json:"template_data,omitempty"`
}

func (ej *EmailJob) GetID() string { return ej.ID }
func (ej *EmailJob) GetType() string { return "email" }
func (ej *EmailJob) GetPriority() Priority { return PriorityNormal }
func (ej *EmailJob) GetPayload() []byte {
    data, _ := json.Marshal(ej)
    return data
}
func (ej *EmailJob) GetRetryPolicy() RetryPolicy {
    return RetryPolicy{
        MaxRetries:    3,
        InitialDelay:  30 * time.Second,
        MaxDelay:      10 * time.Minute,
        BackoffFactor: 2.0,
    }
}

type EmailHandler struct {
    emailService *email.Service
    logger       *logging.Logger
}

func (eh *EmailHandler) Handle(ctx context.Context, payload []byte) ([]byte, error) {
    var job EmailJob
    if err := json.Unmarshal(payload, &job); err != nil {
        return nil, fmt.Errorf("failed to unmarshal email job: %w", err)
    }

    // Send email
    if err := eh.emailService.Send(ctx, job.To, job.Subject, job.Body); err != nil {
        return nil, fmt.Errorf("failed to send email: %w", err)
    }

    return []byte(`{"status": "sent"}`), nil
}

func (eh *EmailHandler) GetJobType() string { return "email" }
func (eh *EmailHandler) GetTimeout() time.Duration { return 30 * time.Second }

// jobs/handlers/market_resolution.go
type MarketResolutionJob struct {
    ID       string `json:"id"`
    MarketID uint   `json:"market_id"`
    Outcome  string `json:"outcome"`
    UserID   uint   `json:"user_id"`
}

func (mrj *MarketResolutionJob) GetID() string { return mrj.ID }
func (mrj *MarketResolutionJob) GetType() string { return "market_resolution" }
func (mrj *MarketResolutionJob) GetPriority() Priority { return PriorityHigh }

type MarketResolutionHandler struct {
    marketService *services.MarketService
    betService    *services.BetService
    userService   *services.UserService
    logger        *logging.Logger
}

func (mrh *MarketResolutionHandler) Handle(ctx context.Context, payload []byte) ([]byte, error) {
    var job MarketResolutionJob
    if err := json.Unmarshal(payload, &job); err != nil {
        return nil, fmt.Errorf("failed to unmarshal market resolution job: %w", err)
    }

    mrh.logger.WithFields(map[string]interface{}{
        "market_id": job.MarketID,
        "outcome":   job.Outcome,
        "user_id":   job.UserID,
    }).Info("Processing market resolution")

    // Get market
    market, err := mrh.marketService.GetByID(ctx, job.MarketID)
    if err != nil {
        return nil, fmt.Errorf("failed to get market: %w", err)
    }

    // Validate user can resolve this market
    if market.CreatorID != job.UserID {
        return nil, fmt.Errorf("user not authorized to resolve this market")
    }

    // Resolve market and distribute payouts
    if err := mrh.marketService.ResolveMarket(ctx, job.MarketID, job.Outcome); err != nil {
        return nil, fmt.Errorf("failed to resolve market: %w", err)
    }

    // Process payouts for all bets
    if err := mrh.processPayouts(ctx, job.MarketID, job.Outcome); err != nil {
        return nil, fmt.Errorf("failed to process payouts: %w", err)
    }

    result := map[string]interface{}{
        "market_id": job.MarketID,
        "outcome":   job.Outcome,
        "status":    "resolved",
    }

    resultData, _ := json.Marshal(result)
    return resultData, nil
}

func (mrh *MarketResolutionHandler) GetJobType() string { return "market_resolution" }
func (mrh *MarketResolutionHandler) GetTimeout() time.Duration { return 5 * time.Minute }

// jobs/handlers/analytics.go
type AnalyticsJob struct {
    ID        string                 `json:"id"`
    EventType string                 `json:"event_type"`
    UserID    uint                   `json:"user_id,omitempty"`
    Data      map[string]interface{} `json:"data"`
    Timestamp time.Time              `json:"timestamp"`
}

type AnalyticsHandler struct {
    analyticsService *analytics.Service
    logger           *logging.Logger
}

func (ah *AnalyticsHandler) Handle(ctx context.Context, payload []byte) ([]byte, error) {
    var job AnalyticsJob
    if err := json.Unmarshal(payload, &job); err != nil {
        return nil, fmt.Errorf("failed to unmarshal analytics job: %w", err)
    }

    // Process analytics event
    if err := ah.analyticsService.TrackEvent(ctx, job.EventType, job.UserID, job.Data); err != nil {
        return nil, fmt.Errorf("failed to track analytics event: %w", err)
    }

    return []byte(`{"status": "tracked"}`), nil
}

func (ah *AnalyticsHandler) GetJobType() string { return "analytics" }
func (ah *AnalyticsHandler) GetTimeout() time.Duration { return 10 * time.Second }
```

### Step 4: Scheduled Jobs System
**Timeline: 2-3 days**

Implement cron-like scheduled job functionality:

```go
// jobs/scheduler.go
package jobs

type Scheduler struct {
    queue       Queue
    schedules   map[string]*Schedule
    ticker      *time.Ticker
    ctx         context.Context
    cancel      context.CancelFunc
    logger      *logging.Logger
    mutex       sync.RWMutex
}

type Schedule struct {
    ID          string        `json:"id"`
    CronExpr    string        `json:"cron_expr"`
    JobType     string        `json:"job_type"`
    Payload     []byte        `json:"payload"`
    Enabled     bool          `json:"enabled"`
    LastRun     *time.Time    `json:"last_run,omitempty"`
    NextRun     time.Time     `json:"next_run"`
    Timezone    string        `json:"timezone"`
    MaxRuns     int           `json:"max_runs,omitempty"` // 0 = unlimited
    RunCount    int           `json:"run_count"`
}

func NewScheduler(queue Queue) *Scheduler {
    ctx, cancel := context.WithCancel(context.Background())

    return &Scheduler{
        queue:     queue,
        schedules: make(map[string]*Schedule),
        ticker:    time.NewTicker(1 * time.Minute),
        ctx:       ctx,
        cancel:    cancel,
        logger:    logging.NewLogger(),
    }
}

func (s *Scheduler) AddSchedule(schedule *Schedule) error {
    s.mutex.Lock()
    defer s.mutex.Unlock()

    // Parse cron expression and calculate next run
    nextRun, err := s.calculateNextRun(schedule.CronExpr, schedule.Timezone)
    if err != nil {
        return fmt.Errorf("invalid cron expression: %w", err)
    }

    schedule.NextRun = nextRun
    s.schedules[schedule.ID] = schedule

    s.logger.WithFields(map[string]interface{}{
        "schedule_id": schedule.ID,
        "job_type":    schedule.JobType,
        "next_run":    nextRun,
    }).Info("Added scheduled job")

    return nil
}

func (s *Scheduler) Start() {
    s.logger.Info("Starting job scheduler")

    go func() {
        for {
            select {
            case <-s.ctx.Done():
                return
            case <-s.ticker.C:
                s.processSchedules()
            }
        }
    }()
}

func (s *Scheduler) processSchedules() {
    s.mutex.RLock()
    schedules := make([]*Schedule, 0, len(s.schedules))
    for _, schedule := range s.schedules {
        schedules = append(schedules, schedule)
    }
    s.mutex.RUnlock()

    now := time.Now()

    for _, schedule := range schedules {
        if !schedule.Enabled {
            continue
        }

        if schedule.MaxRuns > 0 && schedule.RunCount >= schedule.MaxRuns {
            continue
        }

        if now.After(schedule.NextRun) || now.Equal(schedule.NextRun) {
            s.runScheduledJob(schedule)
        }
    }
}

func (s *Scheduler) runScheduledJob(schedule *Schedule) {
    s.logger.WithFields(map[string]interface{}{
        "schedule_id": schedule.ID,
        "job_type":    schedule.JobType,
    }).Info("Running scheduled job")

    // Create job
    job := &ScheduledJob{
        ID:          fmt.Sprintf("scheduled_%s_%d", schedule.ID, time.Now().Unix()),
        ScheduleID:  schedule.ID,
        Type:        schedule.JobType,
        Payload:     schedule.Payload,
        Priority:    PriorityNormal,
    }

    // Enqueue job
    if err := s.queue.Enqueue(s.ctx, job); err != nil {
        s.logger.WithFields(map[string]interface{}{
            "schedule_id": schedule.ID,
            "error":       err.Error(),
        }).Error("Failed to enqueue scheduled job")
        return
    }

    // Update schedule
    s.mutex.Lock()
    now := time.Now()
    schedule.LastRun = &now
    schedule.RunCount++

    nextRun, err := s.calculateNextRun(schedule.CronExpr, schedule.Timezone)
    if err != nil {
        s.logger.WithFields(map[string]interface{}{
            "schedule_id": schedule.ID,
            "error":       err.Error(),
        }).Error("Failed to calculate next run time")
    } else {
        schedule.NextRun = nextRun
    }
    s.mutex.Unlock()
}

// Scheduled job implementations
type ScheduledJob struct {
    ID         string    `json:"id"`
    ScheduleID string    `json:"schedule_id"`
    Type       string    `json:"type"`
    Payload    []byte    `json:"payload"`
    Priority   Priority  `json:"priority"`
}

func (sj *ScheduledJob) GetID() string { return sj.ID }
func (sj *ScheduledJob) GetType() string { return sj.Type }
func (sj *ScheduledJob) GetPriority() Priority { return sj.Priority }
func (sj *ScheduledJob) GetPayload() []byte { return sj.Payload }
```

### Step 5: Job Management API
**Timeline: 2 days**

Create HTTP endpoints for job management:

```go
// handlers/jobs.go
package handlers

type JobsHandler struct {
    queue     jobs.Queue
    scheduler *jobs.Scheduler
    pool      *jobs.WorkerPool
    logger    *logging.Logger
}

func (jh *JobsHandler) ListJobs(w http.ResponseWriter, r *http.Request) {
    filters := jobs.JobFilters{
        Status:   r.URL.Query().Get("status"),
        JobType:  r.URL.Query().Get("type"),
        Page:     parseInt(r.URL.Query().Get("page"), 1),
        PageSize: parseInt(r.URL.Query().Get("page_size"), 20),
    }

    jobList, err := jh.queue.ListJobs(r.Context(), filters)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    WriteJSONResponse(w, map[string]interface{}{
        "jobs":       jobList,
        "pagination": filters.GetPagination(),
    })
}

func (jh *JobsHandler) GetJob(w http.ResponseWriter, r *http.Request) {
    jobID := mux.Vars(r)["id"]

    job, err := jh.queue.GetJob(r.Context(), jobID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }

    WriteJSONResponse(w, job)
}

func (jh *JobsHandler) RetryJob(w http.ResponseWriter, r *http.Request) {
    jobID := mux.Vars(r)["id"]

    job, err := jh.queue.GetJob(r.Context(), jobID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }

    if job.Status != jobs.StatusFailed {
        http.Error(w, "Job is not in failed state", http.StatusBadRequest)
        return
    }

    // Reset job status and requeue
    if err := jh.queue.UpdateJobStatus(r.Context(), jobID, jobs.StatusPending, nil, nil); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    WriteJSONResponse(w, map[string]string{"status": "queued for retry"})
}

func (jh *JobsHandler) GetWorkerStats(w http.ResponseWriter, r *http.Request) {
    stats := jh.pool.GetStats()
    WriteJSONResponse(w, stats)
}
```

### Step 6: Monitoring and Metrics
**Timeline: 1-2 days**

Add comprehensive monitoring for the job system:

```go
// jobs/metrics.go
package jobs

type JobMetrics struct {
    JobsEnqueued   *prometheus.CounterVec
    JobsDequeued   *prometheus.CounterVec
    JobsCompleted  *prometheus.CounterVec
    JobsFailed     *prometheus.CounterVec
    JobDuration    *prometheus.HistogramVec
    QueueSize      *prometheus.GaugeVec
    WorkersBusy    prometheus.Gauge
    WorkersIdle    prometheus.Gauge
}

func NewJobMetrics() *JobMetrics {
    return &JobMetrics{
        JobsEnqueued: promauto.NewCounterVec(
            prometheus.CounterOpts{
                Name: "jobs_enqueued_total",
                Help: "Total number of jobs enqueued",
            },
            []string{"job_type"},
        ),
        JobsDequeued: promauto.NewCounterVec(
            prometheus.CounterOpts{
                Name: "jobs_dequeued_total",
                Help: "Total number of jobs dequeued",
            },
            []string{"job_type"},
        ),
        JobsCompleted: promauto.NewCounterVec(
            prometheus.CounterOpts{
                Name: "jobs_completed_total",
                Help: "Total number of jobs completed successfully",
            },
            []string{"job_type"},
        ),
        JobsFailed: promauto.NewCounterVec(
            prometheus.CounterOpts{
                Name: "jobs_failed_total",
                Help: "Total number of jobs that failed",
            },
            []string{"job_type"},
        ),
        JobDuration: promauto.NewHistogramVec(
            prometheus.HistogramOpts{
                Name:    "job_duration_seconds",
                Help:    "Job execution duration in seconds",
                Buckets: []float64{.1, .5, 1, 5, 10, 30, 60, 300, 600},
            },
            []string{"job_type"},
        ),
        QueueSize: promauto.NewGaugeVec(
            prometheus.GaugeOpts{
                Name: "job_queue_size",
                Help: "Number of jobs in queue",
            },
            []string{"job_type", "status"},
        ),
        WorkersBusy: promauto.NewGauge(
            prometheus.GaugeOpts{
                Name: "workers_busy",
                Help: "Number of busy workers",
            },
        ),
        WorkersIdle: promauto.NewGauge(
            prometheus.GaugeOpts{
                Name: "workers_idle",
                Help: "Number of idle workers",
            },
        ),
    }
}
```

## Directory Structure
```
jobs/
├── queue.go              # Job queue interface and implementation
├── worker.go             # Worker pool implementation
├── scheduler.go          # Scheduled jobs system
├── metrics.go            # Job system metrics
├── health.go             # Health checking for workers
├── handlers/
│   ├── email.go          # Email job handler
│   ├── analytics.go      # Analytics job handler
│   ├── market_resolution.go # Market resolution handler
│   ├── notifications.go  # Notification handler
│   └── cleanup.go        # Cleanup job handler
├── middleware/
│   └── job_context.go    # Job execution context middleware
└── config/
    ├── jobs.yaml         # Job system configuration
    └── schedules.yaml    # Scheduled jobs configuration
```

## Job Types and Use Cases

### High Priority Jobs
- Market resolution and payout processing
- Security incident response
- Critical system maintenance

### Normal Priority Jobs
- User notifications
- Email sending
- Data synchronization
- Report generation

### Low Priority Jobs
- Analytics processing
- Log archival
- Cache warming
- Database cleanup

### Scheduled Jobs
- Daily user engagement reports
- Weekly market statistics
- Monthly cleanup tasks
- System health checks

## Configuration
```yaml
jobs:
  redis:
    url: "redis://localhost:6379"
    db: 1

  workers:
    count: 4
    job_types: ["email", "analytics", "market_resolution", "notifications"]

  retry_policy:
    max_retries: 3
    initial_delay: "30s"
    max_delay: "10m"
    backoff_factor: 2.0

  scheduler:
    enabled: true
    check_interval: "1m"

  monitoring:
    metrics_enabled: true
    health_check_interval: "30s"
```

## Benefits
- Improved application responsiveness
- Better resource utilization
- Scalable background processing
- Reliable job execution with retry logic
- Scheduled task automation
- Comprehensive monitoring and observability
- Fault tolerance and recovery