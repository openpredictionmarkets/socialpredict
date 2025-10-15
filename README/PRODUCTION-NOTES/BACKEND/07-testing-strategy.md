# Testing Strategy Implementation Plan

## Overview
Implement comprehensive testing strategy including unit tests, integration tests, API tests, and performance tests to ensure code quality and system reliability.

## Current State Analysis
- Limited testing exists in the codebase
- Some test files in `models/bets_test.go`, `middleware/middleware_test.go`
- Basic testing structure but not comprehensive
- No integration testing framework
- No API testing suite
- No performance testing

## Implementation Steps

### Step 1: Testing Infrastructure Setup
**Timeline: 2-3 days**

Set up comprehensive testing infrastructure:

```go
// testing/infrastructure.go
type TestSuite struct {
    DB          *gorm.DB
    TestDB      *gorm.DB
    Server      *httptest.Server
    Container   *testcontainers.Container
    Logger      *logging.Logger
    Config      *config.Config
}

func NewTestSuite() *TestSuite {
    // Start test database container
    container, db := setupTestDatabase()

    // Create test server
    server := setupTestServer(db)

    return &TestSuite{
        DB:        db,
        TestDB:    db,
        Server:    server,
        Container: container,
        Logger:    setupTestLogger(),
        Config:    setupTestConfig(),
    }
}

func (ts *TestSuite) SetupTest() {
    // Clear database before each test
    ts.ClearDatabase()

    // Seed test data if needed
    ts.SeedTestData()
}

func (ts *TestSuite) TearDownTest() {
    // Clean up after test
    ts.ClearDatabase()
}
```

**Infrastructure components:**
- Test database management
- Test server setup
- Test containers integration
- Mock services
- Test data factories
- Test utilities

### Step 2: Unit Testing Framework
**Timeline: 3-4 days**

Implement comprehensive unit tests for all packages:

```go
// Example: handlers/markets/markets_test.go
func TestMarketService_CreateMarket(t *testing.T) {
    tests := []struct {
        name    string
        input   models.CreateMarketRequest
        setup   func(*testing.T, *TestSuite)
        want    *models.Market
        wantErr bool
        errType error
    }{
        {
            name: "successful market creation",
            input: models.CreateMarketRequest{
                Title:       "Test Market",
                Description: "Test Description",
                EndDate:     time.Now().Add(24 * time.Hour),
            },
            setup: func(t *testing.T, ts *TestSuite) {
                // Setup test user
                ts.CreateTestUser("testuser")
            },
            want: &models.Market{
                Title:       "Test Market",
                Description: "Test Description",
                Status:      "active",
            },
            wantErr: false,
        },
        {
            name: "market creation with invalid end date",
            input: models.CreateMarketRequest{
                Title:       "Test Market",
                Description: "Test Description",
                EndDate:     time.Now().Add(-24 * time.Hour), // Past date
            },
            wantErr: true,
            errType: &ValidationError{},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ts := NewTestSuite()
            defer ts.Cleanup()

            if tt.setup != nil {
                tt.setup(t, ts)
            }

            service := NewMarketService(ts.DB, ts.Logger)
            got, err := service.CreateMarket(context.Background(), tt.input)

            if tt.wantErr {
                assert.Error(t, err)
                if tt.errType != nil {
                    assert.IsType(t, tt.errType, err)
                }
                return
            }

            assert.NoError(t, err)
            assert.Equal(t, tt.want.Title, got.Title)
            assert.Equal(t, tt.want.Status, got.Status)
        })
    }
}
```

**Unit test coverage:**
- All service layer functions
- Repository methods
- Utility functions
- Validation logic
- Error handling
- Business logic

### Step 3: Integration Testing
**Timeline: 4-5 days**

Implement integration tests for database operations and service interactions:

```go
// testing/integration/database_test.go
func TestDatabaseIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration tests in short mode")
    }

    ts := NewTestSuite()
    defer ts.Cleanup()

    t.Run("user operations", func(t *testing.T) {
        // Test user creation, update, deletion
        user := &models.User{
            Username: "testuser",
            Email:    "test@example.com",
        }

        err := ts.UserRepo.Create(context.Background(), user)
        assert.NoError(t, err)
        assert.NotZero(t, user.ID)

        // Test retrieval
        retrieved, err := ts.UserRepo.GetByUsername(context.Background(), "testuser")
        assert.NoError(t, err)
        assert.Equal(t, user.Email, retrieved.Email)
    })

    t.Run("transaction rollback", func(t *testing.T) {
        err := ts.TxManager.WithTransaction(context.Background(), func(tx *gorm.DB) error {
            // Create user
            user := &models.User{Username: "txuser", Email: "tx@example.com"}
            if err := tx.Create(user).Error; err != nil {
                return err
            }

            // Simulate error to trigger rollback
            return errors.New("simulated error")
        })

        assert.Error(t, err)

        // Verify user was not created due to rollback
        _, err = ts.UserRepo.GetByUsername(context.Background(), "txuser")
        assert.Error(t, err)
        assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
    })
}
```

**Integration test scenarios:**
- Database CRUD operations
- Transaction management
- Service layer interactions
- External service integrations
- Authentication flows
- Business workflow testing

### Step 4: API Testing Suite
**Timeline: 3-4 days**

Create comprehensive API tests:

```go
// testing/api/api_test.go
func TestMarketsAPI(t *testing.T) {
    ts := NewTestSuite()
    defer ts.Cleanup()

    // Create test user and get auth token
    user := ts.CreateTestUser("apiuser")
    token := ts.GenerateAuthToken(user)

    t.Run("GET /v1/markets", func(t *testing.T) {
        // Seed test markets
        ts.SeedTestMarkets(5)

        req := httptest.NewRequest("GET", "/v1/markets", nil)
        req.Header.Set("Authorization", "Bearer "+token)

        w := httptest.NewRecorder()
        ts.Server.Handler.ServeHTTP(w, req)

        assert.Equal(t, http.StatusOK, w.Code)

        var response APIResponse
        err := json.Unmarshal(w.Body.Bytes(), &response)
        assert.NoError(t, err)
        assert.True(t, response.Success)
        assert.NotNil(t, response.Data)
    })

    t.Run("POST /v1/markets", func(t *testing.T) {
        marketData := map[string]interface{}{
            "title":       "API Test Market",
            "description": "Created via API test",
            "end_date":    time.Now().Add(24 * time.Hour).Format(time.RFC3339),
        }

        body, _ := json.Marshal(marketData)
        req := httptest.NewRequest("POST", "/v1/markets", bytes.NewBuffer(body))
        req.Header.Set("Authorization", "Bearer "+token)
        req.Header.Set("Content-Type", "application/json")

        w := httptest.NewRecorder()
        ts.Server.Handler.ServeHTTP(w, req)

        assert.Equal(t, http.StatusCreated, w.Code)

        // Verify market was created
        var response APIResponse
        err := json.Unmarshal(w.Body.Bytes(), &response)
        assert.NoError(t, err)
        assert.True(t, response.Success)
    })

    t.Run("authentication required", func(t *testing.T) {
        req := httptest.NewRequest("GET", "/v1/markets", nil)
        // No auth token

        w := httptest.NewRecorder()
        ts.Server.Handler.ServeHTTP(w, req)

        assert.Equal(t, http.StatusUnauthorized, w.Code)
    })
}
```

**API test coverage:**
- All endpoint functionality
- Authentication and authorization
- Input validation
- Error responses
- Rate limiting
- Content negotiation

### Step 5: Performance Testing
**Timeline: 2-3 days**

Implement performance and load testing:

```go
// testing/performance/load_test.go
func TestAPIPerformance(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping performance tests in short mode")
    }

    ts := NewTestSuite()
    defer ts.Cleanup()

    // Setup test data
    ts.SeedTestData()

    t.Run("concurrent user requests", func(t *testing.T) {
        const (
            numUsers    = 100
            reqPerUser  = 10
            maxDuration = 5 * time.Second
        )

        var wg sync.WaitGroup
        results := make(chan time.Duration, numUsers*reqPerUser)

        start := time.Now()

        for i := 0; i < numUsers; i++ {
            wg.Add(1)
            go func(userID int) {
                defer wg.Done()

                token := ts.GenerateAuthToken(ts.GetTestUser(userID))

                for j := 0; j < reqPerUser; j++ {
                    reqStart := time.Now()

                    req := httptest.NewRequest("GET", "/v1/markets", nil)
                    req.Header.Set("Authorization", "Bearer "+token)

                    w := httptest.NewRecorder()
                    ts.Server.Handler.ServeHTTP(w, req)

                    duration := time.Since(reqStart)
                    results <- duration

                    assert.Equal(t, http.StatusOK, w.Code)
                }
            }(i)
        }

        wg.Wait()
        close(results)

        totalDuration := time.Since(start)
        assert.Less(t, totalDuration, maxDuration)

        // Analyze response times
        var durations []time.Duration
        for d := range results {
            durations = append(durations, d)
        }

        sort.Slice(durations, func(i, j int) bool {
            return durations[i] < durations[j]
        })

        p95 := durations[int(0.95*float64(len(durations)))]
        assert.Less(t, p95, 200*time.Millisecond, "95th percentile response time too high")

        t.Logf("Performance results: P95: %v, Total time: %v", p95, totalDuration)
    })
}
```

### Step 6: Test Data Management
**Timeline: 2 days**

Create comprehensive test data factories and fixtures:

```go
// testing/factories/factories.go
type UserFactory struct {
    db *gorm.DB
}

func (uf *UserFactory) Create(overrides ...func(*models.User)) *models.User {
    user := &models.User{
        Username:    faker.Username(),
        Email:       faker.Email(),
        DisplayName: faker.Name(),
        Credits:     1000,
        CreatedAt:   time.Now(),
    }

    // Apply overrides
    for _, override := range overrides {
        override(user)
    }

    if err := uf.db.Create(user).Error; err != nil {
        panic(fmt.Sprintf("Failed to create test user: %v", err))
    }

    return user
}

func (uf *UserFactory) CreateMany(count int, overrides ...func(*models.User)) []*models.User {
    users := make([]*models.User, count)
    for i := 0; i < count; i++ {
        users[i] = uf.Create(overrides...)
    }
    return users
}

type MarketFactory struct {
    db *gorm.DB
}

func (mf *MarketFactory) Create(overrides ...func(*models.Market)) *models.Market {
    market := &models.Market{
        Title:       faker.Sentence(),
        Description: faker.Paragraph(),
        CreatorID:   1, // Default creator
        Status:      "active",
        EndDate:     time.Now().Add(24 * time.Hour),
        CreatedAt:   time.Now(),
    }

    // Apply overrides
    for _, override := range overrides {
        override(market)
    }

    if err := mf.db.Create(market).Error; err != nil {
        panic(fmt.Sprintf("Failed to create test market: %v", err))
    }

    return market
}
```

### Step 7: Test Automation and CI Integration
**Timeline: 2 days**

Set up automated testing in CI/CD pipeline:

```yaml
# .github/workflows/test.yml
name: Test Suite

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:13
        env:
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: socialpredict_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
    - uses: actions/checkout@v3

    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.23

    - name: Install dependencies
      run: go mod download

    - name: Run unit tests
      run: go test -v -race -coverprofile=coverage.txt ./...

    - name: Run integration tests
      run: go test -v -tags=integration ./testing/integration/...
      env:
        DATABASE_URL: postgres://postgres:postgres@localhost:5432/socialpredict_test?sslmode=disable

    - name: Run API tests
      run: go test -v -tags=api ./testing/api/...

    - name: Upload coverage to Codecov
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.txt
```

## Directory Structure
```
testing/
├── infrastructure.go      # Test infrastructure setup
├── factories/
│   ├── user_factory.go    # User test data factory
│   ├── market_factory.go  # Market test data factory
│   └── bet_factory.go     # Bet test data factory
├── fixtures/
│   ├── users.json         # Static test data
│   ├── markets.json       # Static test data
│   └── scenarios/         # Complex test scenarios
├── unit/
│   ├── services/          # Service layer unit tests
│   ├── repositories/      # Repository layer unit tests
│   └── utils/             # Utility function tests
├── integration/
│   ├── database_test.go   # Database integration tests
│   ├── services_test.go   # Service integration tests
│   └── workflows_test.go  # Business workflow tests
├── api/
│   ├── auth_test.go       # Authentication API tests
│   ├── markets_test.go    # Markets API tests
│   ├── users_test.go      # Users API tests
│   └── bets_test.go       # Betting API tests
├── performance/
│   ├── load_test.go       # Load testing
│   ├── stress_test.go     # Stress testing
│   └── benchmark_test.go  # Benchmark tests
└── mocks/
    ├── repositories.go    # Mock repository interfaces
    ├── services.go        # Mock service interfaces
    └── external.go        # Mock external services
```

## Test Configuration
```yaml
# testing/config.yaml
database:
  driver: "postgres"
  host: "localhost"
  port: 5432
  database: "socialpredict_test"
  username: "postgres"
  password: "postgres"

test_data:
  users: 100
  markets: 50
  bets: 500

performance:
  max_response_time: "200ms"
  concurrent_users: 100
  requests_per_user: 10

coverage:
  minimum: 85
  exclude:
    - "main.go"
    - "*/mocks/*"
    - "testing/*"
```

## Testing Commands
```makefile
# Makefile
.PHONY: test test-unit test-integration test-api test-performance test-coverage

test: test-unit test-integration test-api

test-unit:
	go test -v -race ./...

test-integration:
	go test -v -tags=integration ./testing/integration/...

test-api:
	go test -v -tags=api ./testing/api/...

test-performance:
	go test -v -tags=performance ./testing/performance/...

test-coverage:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

test-clean:
	docker-compose -f docker-compose.test.yml down -v
```

## Benefits
- High code quality assurance
- Regression prevention
- Faster development cycles
- Better refactoring confidence
- Performance monitoring
- Documentation through tests

## Success Metrics
- >85% code coverage
- <200ms API response times (95th percentile)
- Zero failing tests in CI/CD
- All critical paths covered by integration tests
- Performance benchmarks within acceptable limits