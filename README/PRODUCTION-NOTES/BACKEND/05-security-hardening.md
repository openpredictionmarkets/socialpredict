# Security Hardening Implementation Plan

## Overview
Enhance the existing security middleware and implement comprehensive security measures for production deployment including authentication improvements, authorization controls, input validation, and security monitoring.

## Current State Analysis
- Basic security middleware in `security/` package
- JWT authentication in `middleware/auth.go`
- Rate limiting implementation in `security/ratelimit.go`
- Input sanitization in `security/sanitizer.go`
- Basic security headers in `security/headers.go`
- CORS configuration in server setup

## Implementation Steps

### Step 1: Enhanced Authentication System
**Timeline: 3-4 days**

Improve the current JWT authentication with additional security features:

```go
// security/auth/jwt.go
type JWTManager struct {
    signingKey     []byte
    refreshKey     []byte
    tokenExpiry    time.Duration
    refreshExpiry  time.Duration
    blacklist      TokenBlacklist
    rateLimiter    *RateLimiter
}

type TokenPair struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresIn    int64  `json:"expires_in"`
    TokenType    string `json:"token_type"`
}

func (jm *JWTManager) GenerateTokenPair(userID string, roles []string) (*TokenPair, error) {
    // Generate access token with short expiry
    accessToken, err := jm.generateAccessToken(userID, roles)
    if err != nil {
        return nil, err
    }

    // Generate refresh token with longer expiry
    refreshToken, err := jm.generateRefreshToken(userID)
    if err != nil {
        return nil, err
    }

    return &TokenPair{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        ExpiresIn:    int64(jm.tokenExpiry.Seconds()),
        TokenType:    "Bearer",
    }, nil
}
```

**Authentication enhancements:**
- Token refresh mechanism
- Token blacklisting for logout
- Multi-factor authentication support
- Session management
- Login attempt tracking
- Account lockout mechanism

### Step 2: Role-Based Access Control (RBAC)
**Timeline: 2-3 days**

Implement comprehensive authorization system:

```go
// security/auth/rbac.go
type Permission struct {
    Resource string `json:"resource"`
    Action   string `json:"action"`
}

type Role struct {
    Name        string       `json:"name"`
    Permissions []Permission `json:"permissions"`
}

type AuthorizationManager struct {
    roles       map[string]Role
    userRoles   map[string][]string
    cache       *Cache
}

func (am *AuthorizationManager) HasPermission(userID, resource, action string) bool {
    userRoles := am.getUserRoles(userID)
    for _, roleName := range userRoles {
        role := am.roles[roleName]
        for _, perm := range role.Permissions {
            if perm.Resource == resource && perm.Action == action {
                return true
            }
        }
    }
    return false
}

// Middleware for authorization
func (am *AuthorizationManager) RequirePermission(resource, action string) mux.MiddlewareFunc {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            userID := getUserIDFromContext(r.Context())
            if !am.HasPermission(userID, resource, action) {
                http.Error(w, "Forbidden", http.StatusForbidden)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

**RBAC features:**
- Fine-grained permissions
- Role hierarchy
- Dynamic permission checking
- Permission caching
- Admin role management
- Resource-based access control

### Step 3: Advanced Rate Limiting
**Timeline: 2 days**

Enhance the existing rate limiting with advanced features:

```go
// security/ratelimit/advanced.go
type AdvancedRateLimiter struct {
    globalLimiter   *TokenBucket
    userLimiters    map[string]*TokenBucket
    endpointLimits  map[string]RateLimit
    ipLimiters      map[string]*TokenBucket
    redis           *redis.Client
    metrics         *prometheus.CounterVec
}

type RateLimit struct {
    Requests     int           `json:"requests"`
    Window       time.Duration `json:"window"`
    BurstAllowed int           `json:"burst_allowed"`
}

func (arl *AdvancedRateLimiter) CheckRate(r *http.Request) error {
    // Check global rate limit
    if !arl.globalLimiter.Allow() {
        return ErrGlobalRateLimit
    }

    // Check IP-based rate limit
    clientIP := getClientIP(r)
    if !arl.checkIPLimit(clientIP) {
        return ErrIPRateLimit
    }

    // Check user-based rate limit
    userID := getUserID(r)
    if userID != "" && !arl.checkUserLimit(userID) {
        return ErrUserRateLimit
    }

    // Check endpoint-specific rate limit
    endpoint := getEndpoint(r)
    if !arl.checkEndpointLimit(endpoint, userID) {
        return ErrEndpointRateLimit
    }

    return nil
}
```

**Rate limiting enhancements:**
- Multiple rate limiting strategies
- Distributed rate limiting with Redis
- Endpoint-specific limits
- User and IP-based limits
- Burst allowance
- Rate limit monitoring

### Step 4: Input Validation and Sanitization
**Timeline: 2-3 days**

Expand input validation beyond the current sanitization:

```go
// security/validation/validator.go
type ValidationEngine struct {
    validator   *validator.Validate
    sanitizer   *Sanitizer
    rules       map[string]ValidationRule
}

type ValidationRule struct {
    Required     bool                   `json:"required"`
    Type         string                 `json:"type"`
    MinLength    int                    `json:"min_length,omitempty"`
    MaxLength    int                    `json:"max_length,omitempty"`
    Pattern      string                 `json:"pattern,omitempty"`
    Sanitize     []string               `json:"sanitize,omitempty"`
    CustomRules  []CustomValidationRule `json:"custom_rules,omitempty"`
}

func (ve *ValidationEngine) ValidateRequest(r *http.Request, rules map[string]ValidationRule) (*ValidationResult, error) {
    result := &ValidationResult{
        Valid:  true,
        Errors: make(map[string][]string),
    }

    // Parse request body
    var data map[string]interface{}
    if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
        return nil, err
    }

    // Validate each field
    for field, rule := range rules {
        if err := ve.validateField(field, data[field], rule); err != nil {
            result.Valid = false
            result.Errors[field] = append(result.Errors[field], err.Error())
        }
    }

    return result, nil
}
```

**Validation features:**
- Schema-based validation
- Custom validation rules
- Automatic sanitization
- File upload validation
- SQL injection prevention
- XSS protection

### Step 5: Security Headers and HTTPS
**Timeline: 1-2 days**

Enhance security headers and HTTPS configuration:

```go
// security/headers/security.go
type SecurityHeadersMiddleware struct {
    config SecurityHeadersConfig
}

type SecurityHeadersConfig struct {
    HSTS                HSSTConfig     `yaml:"hsts"`
    CSP                 CSPConfig      `yaml:"csp"`
    ReferrerPolicy      string         `yaml:"referrer_policy"`
    PermissionsPolicy   []string       `yaml:"permissions_policy"`
    CrossOriginPolicy   COOPConfig     `yaml:"cross_origin_policy"`
}

func (shm *SecurityHeadersMiddleware) Apply(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // HSTS
        if shm.config.HSTS.Enabled {
            w.Header().Set("Strict-Transport-Security",
                fmt.Sprintf("max-age=%d; includeSubDomains; preload",
                    shm.config.HSTS.MaxAge))
        }

        // Content Security Policy
        w.Header().Set("Content-Security-Policy", shm.buildCSP())

        // Other security headers
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("X-XSS-Protection", "1; mode=block")
        w.Header().Set("Referrer-Policy", shm.config.ReferrerPolicy)

        next.ServeHTTP(w, r)
    })
}
```

**Security headers:**
- HTTP Strict Transport Security (HSTS)
- Content Security Policy (CSP)
- X-Frame-Options
- X-Content-Type-Options
- Referrer Policy
- Permissions Policy

### Step 6: Security Monitoring and Alerting
**Timeline: 2 days**

Implement security event monitoring:

```go
// security/monitoring/monitor.go
type SecurityMonitor struct {
    logger      *logging.Logger
    alertManager *AlertManager
    metrics     *SecurityMetrics
    events      chan SecurityEvent
}

type SecurityEvent struct {
    Type        string                 `json:"type"`
    Severity    string                 `json:"severity"`
    UserID      string                 `json:"user_id,omitempty"`
    IP          string                 `json:"ip"`
    UserAgent   string                 `json:"user_agent"`
    Details     map[string]interface{} `json:"details"`
    Timestamp   time.Time              `json:"timestamp"`
}

func (sm *SecurityMonitor) LogSecurityEvent(event SecurityEvent) {
    // Log the event
    sm.logger.WithFields(map[string]interface{}{
        "security_event": true,
        "event_type":     event.Type,
        "severity":       event.Severity,
        "user_id":        event.UserID,
        "ip":            event.IP,
    }).Warn("Security event detected")

    // Update metrics
    sm.metrics.SecurityEvents.WithLabelValues(event.Type, event.Severity).Inc()

    // Send alert for high-severity events
    if event.Severity == "HIGH" || event.Severity == "CRITICAL" {
        sm.alertManager.SendAlert(event)
    }

    // Queue for further processing
    select {
    case sm.events <- event:
    default:
        sm.logger.Error("Security event queue full, dropping event")
    }
}
```

**Security monitoring:**
- Failed authentication attempts
- Suspicious IP activity
- Rate limit violations
- Permission violations
- Malformed requests
- Security policy violations

### Step 7: API Security Best Practices
**Timeline: 2 days**

Implement API-specific security measures:

```go
// security/api/protection.go
type APIProtection struct {
    requestSigning   *RequestSigning
    antiReplay      *AntiReplayProtection
    dataEncryption  *DataEncryption
}

func (ap *APIProtection) ValidateRequest(r *http.Request) error {
    // Validate request signature if required
    if ap.requestSigning.IsRequired(r.URL.Path) {
        if err := ap.requestSigning.Validate(r); err != nil {
            return err
        }
    }

    // Check for replay attacks
    if err := ap.antiReplay.CheckRequest(r); err != nil {
        return err
    }

    return nil
}
```

## Directory Structure
```
security/
├── auth/
│   ├── jwt.go              # JWT token management
│   ├── rbac.go             # Role-based access control
│   ├── mfa.go              # Multi-factor authentication
│   └── session.go          # Session management
├── ratelimit/
│   ├── advanced.go         # Advanced rate limiting
│   ├── distributed.go      # Redis-based rate limiting
│   └── metrics.go          # Rate limiting metrics
├── validation/
│   ├── validator.go        # Input validation engine
│   ├── sanitizer.go        # Input sanitization (enhanced)
│   └── rules.go            # Validation rule definitions
├── headers/
│   ├── security.go         # Security headers middleware
│   └── csp.go              # Content Security Policy
├── monitoring/
│   ├── monitor.go          # Security event monitoring
│   ├── alerts.go           # Security alerting
│   └── metrics.go          # Security metrics
├── api/
│   ├── protection.go       # API-specific protections
│   ├── signing.go          # Request signing
│   └── encryption.go       # Data encryption
└── middleware/
    ├── auth.go             # Authentication middleware (enhanced)
    ├── authz.go            # Authorization middleware
    └── security.go         # Combined security middleware
```

## Security Configuration
```yaml
security:
  authentication:
    jwt:
      access_token_expiry: "15m"
      refresh_token_expiry: "7d"
      signing_method: "RS256"
      key_rotation_interval: "30d"

    mfa:
      enabled: true
      methods: ["totp", "sms"]

  authorization:
    rbac_enabled: true
    cache_ttl: "5m"

  rate_limiting:
    global:
      requests: 1000
      window: "1m"
    per_user:
      requests: 100
      window: "1m"
    per_ip:
      requests: 50
      window: "1m"

  headers:
    hsts:
      enabled: true
      max_age: 31536000
    csp:
      default_src: ["'self'"]
      script_src: ["'self'", "'unsafe-inline'"]

  monitoring:
    failed_login_threshold: 5
    alert_on_suspicious_activity: true
```

## Security Testing
- Authentication and authorization tests
- Rate limiting effectiveness tests
- Input validation and sanitization tests
- Security header verification tests
- Penetration testing integration
- Vulnerability scanning automation

## Compliance and Standards
- OWASP Top 10 compliance
- OAuth 2.0 and OpenID Connect standards
- GDPR compliance for data handling
- SOC 2 Type II controls
- PCI DSS for payment data (if applicable)

## Security Metrics and Monitoring
- Authentication success/failure rates
- Authorization denials
- Rate limiting triggers
- Security event frequencies
- Vulnerability scan results
- Compliance score tracking

## Benefits
- Comprehensive security coverage
- Automated threat detection
- Compliance with security standards
- Improved user trust and data protection
- Reduced security incident response time
- Scalable security architecture