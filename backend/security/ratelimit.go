package security

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const (
	// RateLimitScopeInProcess records that this limiter is local to one Go process.
	// It is not a distributed, replica-wide, or ingress-wide rate limit.
	RateLimitScopeInProcess = "in_process"
)

// RateLimiter manages process-local rate limiting for different client identities.
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
	cleanup  time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(r rate.Limit, b int, cleanup time.Duration) *RateLimiter {
	rl := &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    b,
		cleanup:  cleanup,
	}

	// Start cleanup goroutine
	go rl.cleanupRoutine()

	return rl
}

// Scope reports the enforcement scope for this limiter.
func (rl *RateLimiter) Scope() string {
	return RateLimitScopeInProcess
}

// GetLimiter returns the in-process rate limiter for a given client identity.
func (rl *RateLimiter) GetLimiter(key string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[key]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.limiters[key] = limiter
	}

	return limiter
}

// cleanupRoutine periodically removes unused limiters
func (rl *RateLimiter) cleanupRoutine() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		for key, limiter := range rl.limiters {
			// Remove limiters that haven't been used recently
			if limiter.TokensAt(time.Now()) == float64(rl.burst) {
				delete(rl.limiters, key)
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimitConfig holds configuration for different rate limits
type RateLimitConfig struct {
	LoginRate         rate.Limit // requests per second for login attempts
	LoginBurst        int        // max burst for login attempts
	GeneralRate       rate.Limit // requests per second for general API
	GeneralBurst      int        // max burst for general API
	CleanupInterval   time.Duration
	TrustProxyHeaders bool
}

// DefaultRateLimitConfig returns sensible default rate limits
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		LoginRate:       rate.Every(10 * time.Second), // 1 login attempt per 10 seconds
		LoginBurst:      3,                            // Allow burst of 3 attempts
		GeneralRate:     rate.Every(time.Second),      // 1 request per second
		GeneralBurst:    10,                           // Allow burst of 10 requests
		CleanupInterval: 5 * time.Minute,              // Cleanup every 5 minutes
	}
}

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(limiter *RateLimiter) func(http.Handler) http.Handler {
	return RateLimitMiddlewareWithProxyTrust(limiter, false)
}

// RateLimitMiddlewareWithProxyTrust creates rate limiting middleware with explicit proxy-header posture.
func RateLimitMiddlewareWithProxyTrust(limiter *RateLimiter, trustProxy bool) func(http.Handler) http.Handler {
	return RateLimitMiddlewareWithClientIdentity(limiter, NewClientIdentityExtractor(trustProxy))
}

// RateLimitMiddlewareWithClientIdentity creates rate limiting middleware using an explicit client identity contract.
func RateLimitMiddlewareWithClientIdentity(limiter *RateLimiter, clientIdentity ClientIdentityExtractor) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientID := clientIdentity.Extract(r)

			if !limiter.GetLimiter(clientID).Allow() {
				WriteRateLimited(w, RuntimeReasonRateLimited)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// LoginRateLimitMiddleware creates a stricter rate limit for login endpoints
func LoginRateLimitMiddleware(limiter *RateLimiter) func(http.Handler) http.Handler {
	return LoginRateLimitMiddlewareWithProxyTrust(limiter, false)
}

// LoginRateLimitMiddlewareWithProxyTrust creates login rate limiting middleware with explicit proxy-header posture.
func LoginRateLimitMiddlewareWithProxyTrust(limiter *RateLimiter, trustProxy bool) func(http.Handler) http.Handler {
	return LoginRateLimitMiddlewareWithClientIdentity(limiter, NewClientIdentityExtractor(trustProxy))
}

// LoginRateLimitMiddlewareWithClientIdentity creates login rate limiting middleware using an explicit client identity contract.
func LoginRateLimitMiddlewareWithClientIdentity(limiter *RateLimiter, clientIdentity ClientIdentityExtractor) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientID := clientIdentity.Extract(r)

			if !limiter.GetLimiter(clientID).Allow() {
				WriteRateLimited(w, RuntimeReasonLoginRateLimited)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ClientIdentityExtractor defines which request attributes may identify a client at the HTTP boundary.
type ClientIdentityExtractor struct {
	trustProxyHeaders bool
}

// NewClientIdentityExtractor returns a client identity extractor bound to the runtime proxy-trust contract.
func NewClientIdentityExtractor(trustProxyHeaders bool) ClientIdentityExtractor {
	return ClientIdentityExtractor{trustProxyHeaders: trustProxyHeaders}
}

// Extract returns the client identity used by request-boundary security controls.
func (e ClientIdentityExtractor) Extract(r *http.Request) string {
	if r == nil {
		return ""
	}

	if e.trustProxyHeaders {
		if forwarded := firstForwardedFor(r.Header.Get("X-Forwarded-For")); forwarded != "" {
			return forwarded
		}

		if realIP := headerIP(r.Header.Get("X-Real-IP")); realIP != "" {
			return realIP
		}
	}

	return remoteAddressHost(r.RemoteAddr)
}

// getClientIP extracts the client IP address from the request.
// Forwarded headers are only trusted when the deployment explicitly opts in.
func getClientIP(r *http.Request) string {
	return getClientIPWithProxyTrust(r, false)
}

func getClientIPWithProxyTrust(r *http.Request, trustProxy bool) string {
	return NewClientIdentityExtractor(trustProxy).Extract(r)
}

func firstForwardedFor(value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	parts := strings.Split(value, ",")
	return headerIP(parts[0])
}

func headerIP(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || net.ParseIP(value) == nil {
		return ""
	}
	return value
}

func remoteAddressHost(remoteAddr string) string {
	remoteAddr = strings.TrimSpace(remoteAddr)
	if remoteAddr == "" {
		return ""
	}
	host, _, err := net.SplitHostPort(remoteAddr)
	if err == nil {
		return host
	}
	return remoteAddr
}

// RateLimitManager manages multiple rate limiters for different endpoints
type RateLimitManager struct {
	loginLimiter   *RateLimiter
	generalLimiter *RateLimiter
	clientIdentity ClientIdentityExtractor
}

// NewRateLimitManager creates a new rate limit manager with default configuration
func NewRateLimitManager() *RateLimitManager {
	config := DefaultRateLimitConfig()

	return &RateLimitManager{
		loginLimiter: NewRateLimiter(
			config.LoginRate,
			config.LoginBurst,
			config.CleanupInterval,
		),
		generalLimiter: NewRateLimiter(
			config.GeneralRate,
			config.GeneralBurst,
			config.CleanupInterval,
		),
		clientIdentity: NewClientIdentityExtractor(config.TrustProxyHeaders),
	}
}

// NewCustomRateLimitManager creates a rate limit manager with custom configuration
func NewCustomRateLimitManager(config RateLimitConfig) *RateLimitManager {
	return &RateLimitManager{
		loginLimiter: NewRateLimiter(
			config.LoginRate,
			config.LoginBurst,
			config.CleanupInterval,
		),
		generalLimiter: NewRateLimiter(
			config.GeneralRate,
			config.GeneralBurst,
			config.CleanupInterval,
		),
		clientIdentity: NewClientIdentityExtractor(config.TrustProxyHeaders),
	}
}

// GetLoginMiddleware returns the login rate limiting middleware
func (rlm *RateLimitManager) GetLoginMiddleware() func(http.Handler) http.Handler {
	return LoginRateLimitMiddlewareWithClientIdentity(rlm.loginLimiter, rlm.clientIdentity)
}

// GetGeneralMiddleware returns the general rate limiting middleware
func (rlm *RateLimitManager) GetGeneralMiddleware() func(http.Handler) http.Handler {
	return RateLimitMiddlewareWithClientIdentity(rlm.generalLimiter, rlm.clientIdentity)
}
