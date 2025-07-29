package security

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter manages rate limiting for different clients/IPs
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

// GetLimiter returns the rate limiter for a given key (IP address)
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
	LoginRate       rate.Limit // requests per second for login attempts
	LoginBurst      int        // max burst for login attempts
	GeneralRate     rate.Limit // requests per second for general API
	GeneralBurst    int        // max burst for general API
	CleanupInterval time.Duration
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
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get client IP
			ip := getClientIP(r)

			// Check rate limit
			if !limiter.GetLimiter(ip).Allow() {
				http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// LoginRateLimitMiddleware creates a stricter rate limit for login endpoints
func LoginRateLimitMiddleware(limiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get client IP
			ip := getClientIP(r)

			// Check rate limit
			if !limiter.GetLimiter(ip).Allow() {
				http.Error(w, "Too many login attempts. Please try again later.", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check for forwarded IP first (if behind proxy)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// X-Forwarded-For can contain multiple IPs, get the first one
		if idx := len(forwarded); idx > 0 {
			if commaIdx := 0; commaIdx < idx {
				for i, c := range forwarded {
					if c == ',' {
						commaIdx = i
						break
					}
				}
				if commaIdx > 0 {
					return forwarded[:commaIdx]
				}
			}
			return forwarded
		}
	}

	// Check for real IP header
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	// Fall back to remote address
	return r.RemoteAddr
}

// RateLimitManager manages multiple rate limiters for different endpoints
type RateLimitManager struct {
	loginLimiter   *RateLimiter
	generalLimiter *RateLimiter
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
	}
}

// GetLoginMiddleware returns the login rate limiting middleware
func (rlm *RateLimitManager) GetLoginMiddleware() func(http.Handler) http.Handler {
	return LoginRateLimitMiddleware(rlm.loginLimiter)
}

// GetGeneralMiddleware returns the general rate limiting middleware
func (rlm *RateLimitManager) GetGeneralMiddleware() func(http.Handler) http.Handler {
	return RateLimitMiddleware(rlm.generalLimiter)
}
