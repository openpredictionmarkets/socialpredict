# CI/CD Testing Recommendations

This document outlines recommended CI/CD enhancements for the testing infrastructure.

## Coverage Reporting

Add a coverage job to `.github/workflows/backend.yml`:

```yaml
coverage:
  name: coverage
  needs: [smoke]
  if: |
    startsWith(github.head_ref, 'feature/')
    || startsWith(github.head_ref, 'fix/')
    || startsWith(github.head_ref, 'refactor/')
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - name: Setup Go
      uses: actions/setup-go@v5
      with:
        go-version-file: backend/go.mod
        check-latest: true
        cache: true
    - name: Get Dependencies
      working-directory: ./backend
      run: go mod download
    - name: Run tests with coverage
      working-directory: ./backend
      env:
        JWT_SIGNING_KEY: "test-secret-key-for-testing"
      run: |
        go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
        go tool cover -func=coverage.out | tail -1
    - name: Upload coverage to Codecov
      uses: codecov/codecov-action@v4
      with:
        file: ./backend/coverage.out
        flags: backend
        fail_ci_if_error: false
```

## Integration Testing

For integration tests that require a database:

```yaml
integration:
  name: integration
  needs: [unit]
  runs-on: ubuntu-latest
  services:
    postgres:
      image: postgres:15
      env:
        POSTGRES_PASSWORD: postgres
        POSTGRES_DB: socialpredict_test
      options: >-
        --health-cmd pg_isready
        --health-interval 10s
        --health-timeout 5s
        --health-retries 5
      ports:
        - 5432:5432
  steps:
    - uses: actions/checkout@v4
    - name: Setup Go
      uses: actions/setup-go@v5
      with:
        go-version-file: backend/go.mod
    - name: Run integration tests
      working-directory: ./backend
      env:
        DATABASE_URL: postgres://postgres:postgres@localhost:5432/socialpredict_test?sslmode=disable
        JWT_SIGNING_KEY: "test-secret-key-for-testing"
      run: go test -v -tags=integration ./testing/integration/...
```

## Test Commands

### Run all tests
```bash
go test ./...
```

### Run tests with coverage
```bash
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Run specific test packages
```bash
# Unit tests only
go test ./handlers/... ./internal/...

# Integration tests (when tagged)
go test -tags=integration ./testing/integration/...
```

## Coverage Thresholds

Target: >85% code coverage on new code

Exclude from coverage:
- `main.go`
- `*/mocks/*`
- `testing/*` (test utilities themselves)
