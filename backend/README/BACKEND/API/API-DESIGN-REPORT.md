# SocialPredict API Design Report

## Current State (Post-Refactoring)

### Architecture Overview

The SocialPredict API has been refactored from a handler-centric architecture to a clean layered architecture following Domain-Driven Design principles:

```
handlers/          # HTTP layer - JSON in/out, status codes, error mapping
├── admin/dto/     # HTTP request/response types
├── bets/dto/
├── markets/dto/
└── users/dto/

internal/
├── domain/        # Pure business logic (no HTTP, no GORM)
│   ├── admin/
│   ├── bets/
│   ├── markets/
│   └── users/
├── repository/    # Data access layer (GORM implementations)
│   ├── admin/
│   ├── bets/
│   ├── markets/
│   └── users/
└── app/
    └── container.go  # Dependency injection / composition root

models/           # Database models (GORM)
```

### Clean Architecture Benefits

1. **Separation of Concerns**: HTTP logic, business logic, and data access are cleanly separated
2. **Testability**: Business logic can be tested independently of HTTP and database
3. **Interface-driven**: All dependencies use interfaces, enabling easy mocking and testing
4. **Microservices Ready**: Each domain can be extracted into its own service
5. **OpenAPI First**: API specification drives implementation

### Migration Status

#### ✅ Completed Migrations

**Markets Domain**
- ✅ `internal/domain/markets/` - Business logic and validation
- ✅ `internal/repository/markets/` - GORM repository implementation
- ✅ `handlers/markets/` - HTTP handlers (GORM-free)
- ✅ `handlers/markets/dto/` - Request/response DTOs
- ✅ OpenAPI specification for markets endpoints

**Users Domain**
- ✅ `internal/domain/users/` - Business logic
- ✅ `internal/repository/users/` - GORM repository implementation
- ⚠️ `handlers/users/` - Needs migration to new pattern

**Infrastructure**
- ✅ `internal/app/container.go` - Dependency injection
- ✅ OpenAPI documentation scaffolding

#### 🚧 In Progress

**Remaining Handlers to Migrate:**
- `handlers/admin/` - User administration
- `handlers/bets/` - Betting and positions
- `handlers/cms/` - Content management
- `handlers/metrics/` - System metrics
- `handlers/positions/` - Position management
- `handlers/stats/` - Statistics
- `handlers/tradingdata/` - Trading data
- `handlers/users/` - User profile management

#### 📋 TODO

**Domain Services to Create:**
- `internal/domain/bets/` - Betting business logic
- `internal/domain/positions/` - Position calculation logic
- `internal/domain/stats/` - Statistics calculation
- `internal/domain/admin/` - Administrative operations

**Repository Implementations:**
- `internal/repository/bets/`
- `internal/repository/positions/`
- `internal/repository/stats/`
- `internal/repository/admin/`

## API Endpoints Status

### Markets API ✅

| Method | Endpoint | Status | Description |
|--------|----------|--------|-------------|
| GET    | /markets | ✅ Refactored | List markets with filters |
| POST   | /markets | ✅ Refactored | Create new market |
| GET    | /markets/{id} | ✅ Refactored | Get market details |
| GET    | /markets/search | ✅ Refactored | Search markets |
| POST   | /markets/{id}/resolve | ✅ Refactored | Resolve market |
| PUT    | /markets/{id}/labels | ✅ Refactored | Update custom labels |

### Users API 🚧

| Method | Endpoint | Status | Description |
|--------|----------|--------|-------------|
| GET    | /users | 🚧 Legacy | List users |
| GET    | /users/{username} | 🚧 Legacy | Get user profile |
| PUT    | /users/{username} | 🚧 Legacy | Update user profile |
| GET    | /users/{username}/financial | 🚧 Legacy | Get user financials |
| POST   | /users/{username}/credit | 🚧 Legacy | Add user credit |

### Bets API 🚧

| Method | Endpoint | Status | Description |
|--------|----------|--------|-------------|
| GET    | /bets | 🚧 Legacy | List bets |
| POST   | /bets | 🚧 Legacy | Place bet |
| POST   | /positions/sell | 🚧 Legacy | Sell position |
| GET    | /positions/{username} | 🚧 Legacy | Get user positions |

### Admin API 🚧

| Method | Endpoint | Status | Description |
|--------|----------|--------|-------------|
| POST   | /admin/users | 🚧 Legacy | Create user |
| DELETE | /admin/users/{username} | 🚧 Legacy | Delete user |
| POST   | /admin/markets/{id}/resolve | 🚧 Legacy | Admin resolve market |

## Microservices Readiness

### Current Monolith Benefits
- Simple deployment and development
- ACID transactions across domains
- No network latency between services

### Microservices Migration Path

When ready to split into microservices, each domain can be extracted:

#### Markets Service
```
internal/domain/markets/     → markets-service/domain/
internal/repository/markets/ → markets-service/repository/
handlers/markets/           → markets-service/handlers/
```

#### Users Service
```
internal/domain/users/      → users-service/domain/
internal/repository/users/  → users-service/repository/
handlers/users/            → users-service/handlers/
```

#### Bets Service
```
internal/domain/bets/       → bets-service/domain/
internal/repository/bets/   → bets-service/repository/
handlers/bets/             → bets-service/handlers/
```

### Service Communication Strategy

**Option 1: HTTP APIs**
- Generate HTTP clients from OpenAPI specs
- Replace repository interfaces with HTTP client implementations
- Use circuit breakers and retries for resilience

**Option 2: gRPC**
- Generate gRPC clients from protobuf definitions
- High performance, type-safe communication
- Built-in load balancing and health checking

**Option 3: Event-Driven**
- Use message queues (Redis Streams, Kafka)
- Eventually consistent, highly scalable
- Complex error handling and ordering

## Performance Considerations

### Database Access Patterns
- Each repository encapsulates database access patterns
- Query optimization can be done at repository level
- Connection pooling and caching strategies isolated

### Caching Strategy
- Domain services can implement caching logic
- Repository layer can cache frequent queries
- HTTP layer can implement response caching

### Monitoring and Observability
- Domain services emit business metrics
- Repository layer tracks database performance
- HTTP layer monitors request/response patterns

## Security Architecture

### Authentication Flow
```
HTTP Request → Handler → Middleware → Domain Service
                ↑
            Validates JWT
```

### Authorization Patterns
- Domain services implement business-level authorization
- Repository layer handles data-level permissions
- HTTP layer manages session and token validation

## Testing Strategy

### Unit Testing
- ✅ Domain services with mocked repositories
- ✅ Repository layer with test database
- ✅ HTTP handlers with mocked services

### Integration Testing
- ✅ Full request/response cycle testing
- ✅ Database transaction testing
- ✅ OpenAPI contract testing

### End-to-End Testing
- API client generation from OpenAPI spec
- Automated testing of complete user journeys
- Performance testing with realistic load

## Next Steps

### Phase 1: Complete Core Migrations (Current)
1. Migrate remaining handlers to clean architecture
2. Create missing domain services and repositories
3. Update OpenAPI specification for all endpoints

### Phase 2: Enhanced Testing and Documentation
1. Add comprehensive unit tests for all domain services
2. Implement integration tests for all API endpoints
3. Generate API client libraries for frontend

### Phase 3: Performance and Monitoring
1. Implement caching layer
2. Add metrics and observability
3. Performance testing and optimization

### Phase 4: Microservices Preparation (Future)
1. Define service boundaries based on business domains
2. Implement service communication patterns
3. Set up infrastructure for distributed systems

## Conclusion

The refactoring to clean architecture provides a solid foundation for:
- Maintainable and testable code
- OpenAPI-first development workflow
- Future microservices migration
- Improved developer productivity

The current implementation demonstrates the pattern with the Markets domain, and the remaining domains will follow the same architectural principles.
