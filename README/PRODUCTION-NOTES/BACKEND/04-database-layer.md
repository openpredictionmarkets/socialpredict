# Database Layer Implementation Plan

## Overview
Enhance the current database layer with proper connection management, transaction handling, query optimization, migration management, and production-ready database patterns.

## Current State Analysis
- Basic GORM setup in `util/postgres.go`
- Simple database initialization without connection pooling configuration
- Basic migration system in `migration/migrate.go`
- No transaction management patterns
- No query optimization or monitoring
- No database health checks

## Implementation Steps

### Step 1: Connection Pool Management
**Timeline: 2 days**

Enhance database connection configuration for production:

```go
// database/connection.go
type DatabaseConfig struct {
    MaxOpenConns    int           `yaml:"max_open_conns"`
    MaxIdleConns    int           `yaml:"max_idle_conns"`
    ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
    ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time"`
}

func NewDatabase(cfg DatabaseConfig) (*gorm.DB, error) {
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    })

    sqlDB, err := db.DB()
    sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
    sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
    sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
    sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

    return db, nil
}
```

**Configuration features:**
- Connection pool sizing
- Connection lifetime management
- Idle connection timeout
- Connection retry logic
- Health check queries

### Step 2: Repository Pattern Implementation
**Timeline: 3-4 days**

Implement repository pattern for better testability and separation of concerns:

```go
// repository/interfaces.go
type UserRepository interface {
    Create(ctx context.Context, user *models.User) error
    GetByID(ctx context.Context, id uint) (*models.User, error)
    GetByUsername(ctx context.Context, username string) (*models.User, error)
    Update(ctx context.Context, user *models.User) error
    Delete(ctx context.Context, id uint) error
    List(ctx context.Context, filters UserFilters) ([]*models.User, error)
}

// repository/user_repository.go
type userRepository struct {
    db *gorm.DB
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
    return r.db.WithContext(ctx).Create(user).Error
}
```

**Repository features:**
- Interface-based design for testability
- Context-aware operations
- Standardized CRUD operations
- Query filtering and pagination
- Bulk operations support

### Step 3: Transaction Management
**Timeline: 2-3 days**

Implement comprehensive transaction management:

```go
// database/transaction.go
type TransactionManager struct {
    db *gorm.DB
}

func (tm *TransactionManager) WithTransaction(ctx context.Context, fn func(*gorm.DB) error) error {
    tx := tm.db.WithContext(ctx).Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
            panic(r)
        }
    }()

    if err := fn(tx); err != nil {
        tx.Rollback()
        return err
    }

    return tx.Commit().Error
}

// Service layer usage
func (s *BetService) PlaceBet(ctx context.Context, bet *models.Bet) error {
    return s.txManager.WithTransaction(ctx, func(tx *gorm.DB) error {
        // Multiple database operations in transaction
        if err := s.betRepo.CreateWithTx(tx, bet); err != nil {
            return err
        }

        return s.userRepo.UpdateCreditsWithTx(tx, bet.UserID, -bet.Amount)
    })
}
```

**Transaction features:**
- Automatic rollback on errors
- Panic recovery with rollback
- Nested transaction support
- Transaction context propagation
- Deadlock detection and retry

### Step 4: Query Optimization and Monitoring
**Timeline: 2-3 days**

Implement query performance monitoring and optimization:

```go
// database/query_monitor.go
type QueryMonitor struct {
    slowQueryThreshold time.Duration
    logger            *logging.Logger
    metrics           *prometheus.HistogramVec
}

func (qm *QueryMonitor) LogSlowQuery(sql string, duration time.Duration, args []interface{}) {
    if duration > qm.slowQueryThreshold {
        qm.logger.WithFields(map[string]interface{}{
            "sql":      sql,
            "duration": duration,
            "args":     args,
        }).Warn("Slow query detected")

        qm.metrics.WithLabelValues("slow").Observe(duration.Seconds())
    }
}
```

**Query optimization features:**
- Slow query logging
- Query performance metrics
- Index usage analysis
- Query plan explanation
- N+1 query detection

### Step 5: Migration Management
**Timeline: 2 days**

Enhance the migration system for production use:

```go
// migration/manager.go
type MigrationManager struct {
    db            *gorm.DB
    migrationPath string
    logger        *logging.Logger
}

type Migration struct {
    Version     string
    Description string
    Up          func(*gorm.DB) error
    Down        func(*gorm.DB) error
}

func (mm *MigrationManager) Migrate() error {
    migrations, err := mm.loadMigrations()
    if err != nil {
        return err
    }

    for _, migration := range migrations {
        if err := mm.runMigration(migration); err != nil {
            return fmt.Errorf("migration %s failed: %w", migration.Version, err)
        }
    }

    return nil
}
```

**Migration features:**
- Version tracking
- Forward and backward migrations
- Migration history logging
- Rollback capabilities
- Migration validation

### Step 6: Database Health Monitoring
**Timeline: 1-2 days**

Implement comprehensive database health checks:

```go
// database/health.go
type DatabaseHealthChecker struct {
    db              *gorm.DB
    maxConnections  int
    queryTimeout    time.Duration
}

func (dhc *DatabaseHealthChecker) Check(ctx context.Context) error {
    // Check basic connectivity
    sqlDB, err := dhc.db.DB()
    if err != nil {
        return fmt.Errorf("failed to get underlying sql.DB: %w", err)
    }

    // Check connection pool status
    stats := sqlDB.Stats()
    if stats.OpenConnections >= dhc.maxConnections {
        return fmt.Errorf("connection pool exhausted: %d/%d", stats.OpenConnections, dhc.maxConnections)
    }

    // Test query execution
    ctx, cancel := context.WithTimeout(ctx, dhc.queryTimeout)
    defer cancel()

    var result int
    err = dhc.db.WithContext(ctx).Raw("SELECT 1").Scan(&result).Error
    if err != nil {
        return fmt.Errorf("test query failed: %w", err)
    }

    return nil
}
```

**Health check features:**
- Connection pool monitoring
- Query execution testing
- Response time monitoring
- Connection leak detection
- Database locks monitoring

### Step 7: Data Access Layer (DAL)
**Timeline: 2-3 days**

Create a unified data access layer:

```go
// dal/dal.go
type DataAccessLayer struct {
    userRepo   repository.UserRepository
    marketRepo repository.MarketRepository
    betRepo    repository.BetRepository
    txManager  *database.TransactionManager
}

func (dal *DataAccessLayer) Users() repository.UserRepository {
    return dal.userRepo
}

func (dal *DataAccessLayer) WithTransaction(ctx context.Context, fn func(*DataAccessLayer) error) error {
    return dal.txManager.WithTransaction(ctx, func(tx *gorm.DB) error {
        txDAL := &DataAccessLayer{
            userRepo:   repository.NewUserRepository(tx),
            marketRepo: repository.NewMarketRepository(tx),
            betRepo:    repository.NewBetRepository(tx),
        }
        return fn(txDAL)
    })
}
```

## Directory Structure
```
database/
├── connection.go          # Database connection management
├── transaction.go         # Transaction management
├── health.go             # Database health checks
├── query_monitor.go      # Query performance monitoring
└── config.go             # Database configuration

repository/
├── interfaces.go         # Repository interfaces
├── user_repository.go    # User data access
├── market_repository.go  # Market data access
├── bet_repository.go     # Bet data access
└── base_repository.go    # Common repository functionality

migration/
├── manager.go            # Migration management
├── migrations/           # Individual migration files
│   ├── 001_initial.go
│   ├── 002_add_indexes.go
│   └── 003_add_constraints.go
└── schema.sql           # Current schema definition

dal/
├── dal.go               # Data access layer
├── factory.go           # DAL factory
└── testing.go           # Test utilities
```

## Database Configuration
```yaml
database:
  host: "localhost"
  port: 5432
  database: "socialpredict"
  username: "postgres"
  password: "${POSTGRES_PASSWORD}"
  ssl_mode: "require"

  connection_pool:
    max_open_conns: 25
    max_idle_conns: 10
    conn_max_lifetime: "1h"
    conn_max_idle_time: "30m"

  monitoring:
    slow_query_threshold: "1s"
    enable_query_logging: true
    log_level: "warn"

  health_check:
    interval: "30s"
    timeout: "5s"
    retries: 3
```

## Performance Optimizations
- **Indexes**: Proper indexing strategy for all queries
- **Query optimization**: Avoid N+1 queries, use joins efficiently
- **Connection pooling**: Optimal pool sizing for workload
- **Prepared statements**: Use prepared statements for repeated queries
- **Bulk operations**: Batch inserts/updates for better performance

## Testing Strategy
- Unit tests for all repository methods
- Integration tests with test database
- Transaction rollback testing
- Connection pool stress testing
- Migration testing (up/down)
- Performance benchmarking

## Migration Strategy
1. Implement new database layer alongside existing code
2. Create repository interfaces and implementations
3. Update services to use repositories gradually
4. Add transaction management to critical operations
5. Enable query monitoring and health checks
6. Remove direct GORM usage from handlers

## Benefits
- Better separation of concerns
- Improved testability with mock repositories
- Proper transaction management
- Performance monitoring and optimization
- Production-ready connection management
- Simplified database operations

## Monitoring and Metrics
- Connection pool utilization
- Query execution times
- Slow query frequency
- Transaction success/failure rates
- Database health check status
- Migration execution history