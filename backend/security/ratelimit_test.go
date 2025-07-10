package security

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func TestRateLimiter_GetLimiter(t *testing.T) {
	rl := NewRateLimiter(rate.Every(time.Second), 5, time.Minute)

	// Test getting a limiter for a new IP
	limiter1 := rl.GetLimiter("192.168.1.1")
	if limiter1 == nil {
		t.Error("Expected limiter to be created for new IP")
	}

	// Test getting the same limiter for the same IP
	limiter2 := rl.GetLimiter("192.168.1.1")
	if limiter1 != limiter2 {
		t.Error("Expected same limiter instance for same IP")
	}

	// Test getting a different limiter for a different IP
	limiter3 := rl.GetLimiter("192.168.1.2")
	if limiter1 == limiter3 {
		t.Error("Expected different limiter instance for different IP")
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	// Create a rate limiter that allows 2 requests with burst of 1
	rl := NewRateLimiter(rate.Every(time.Second), 1, time.Minute)
	middleware := RateLimitMiddleware(rl)

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Wrap handler with middleware
	wrappedHandler := middleware(testHandler)

	t.Run("allows request within rate limit", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		rr := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %d", rr.Code)
		}
	})

	t.Run("blocks request exceeding rate limit", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.2:12345"
		rr := httptest.NewRecorder()

		// First request should succeed
		wrappedHandler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("First request should succeed, got %d", rr.Code)
		}

		// Second request should be rate limited
		rr = httptest.NewRecorder()
		wrappedHandler.ServeHTTP(rr, req)
		if rr.Code != http.StatusTooManyRequests {
			t.Errorf("Expected status TooManyRequests, got %d", rr.Code)
		}
	})
}

func TestLoginRateLimitMiddleware(t *testing.T) {
	// Create a stricter rate limiter for login
	rl := NewRateLimiter(rate.Every(10*time.Second), 1, time.Minute)
	middleware := LoginRateLimitMiddleware(rl)

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("login success"))
	})

	// Wrap handler with middleware
	wrappedHandler := middleware(testHandler)

	t.Run("allows login within rate limit", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/login", nil)
		req.RemoteAddr = "192.168.1.3:12345"
		rr := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %d", rr.Code)
		}
	})

	t.Run("blocks login exceeding rate limit", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/login", nil)
		req.RemoteAddr = "192.168.1.4:12345"
		rr := httptest.NewRecorder()

		// First request should succeed
		wrappedHandler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("First login should succeed, got %d", rr.Code)
		}

		// Second request should be rate limited
		rr = httptest.NewRecorder()
		wrappedHandler.ServeHTTP(rr, req)
		if rr.Code != http.StatusTooManyRequests {
			t.Errorf("Expected status TooManyRequests, got %d", rr.Code)
		}

		// Check error message is specific to login
		if !contains(rr.Body.String(), "login attempts") {
			t.Error("Expected login-specific error message")
		}
	})
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name           string
		remoteAddr     string
		forwardedFor   string
		realIP         string
		expectedPrefix string
	}{
		{
			name:           "direct connection",
			remoteAddr:     "192.168.1.1:12345",
			expectedPrefix: "192.168.1.1:12345",
		},
		{
			name:           "with X-Forwarded-For",
			remoteAddr:     "10.0.0.1:12345",
			forwardedFor:   "203.0.113.1, 10.0.0.1",
			expectedPrefix: "203.0.113.1",
		},
		{
			name:           "with X-Real-IP",
			remoteAddr:     "10.0.0.1:12345",
			realIP:         "203.0.113.2",
			expectedPrefix: "203.0.113.2",
		},
		{
			name:           "X-Forwarded-For takes precedence over X-Real-IP",
			remoteAddr:     "10.0.0.1:12345",
			forwardedFor:   "203.0.113.1",
			realIP:         "203.0.113.2",
			expectedPrefix: "203.0.113.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = tt.remoteAddr

			if tt.forwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.forwardedFor)
			}
			if tt.realIP != "" {
				req.Header.Set("X-Real-IP", tt.realIP)
			}

			result := getClientIP(req)
			if result != tt.expectedPrefix {
				t.Errorf("getClientIP() = %v, want %v", result, tt.expectedPrefix)
			}
		})
	}
}

func TestRateLimitManager(t *testing.T) {
	manager := NewRateLimitManager()

	if manager.loginLimiter == nil {
		t.Error("Expected login limiter to be initialized")
	}

	if manager.generalLimiter == nil {
		t.Error("Expected general limiter to be initialized")
	}

	// Test middleware creation
	loginMiddleware := manager.GetLoginMiddleware()
	if loginMiddleware == nil {
		t.Error("Expected login middleware to be created")
	}

	generalMiddleware := manager.GetGeneralMiddleware()
	if generalMiddleware == nil {
		t.Error("Expected general middleware to be created")
	}
}

func TestCustomRateLimitManager(t *testing.T) {
	config := RateLimitConfig{
		LoginRate:       rate.Every(5 * time.Second),
		LoginBurst:      2,
		GeneralRate:     rate.Every(500 * time.Millisecond),
		GeneralBurst:    5,
		CleanupInterval: time.Minute,
	}

	manager := NewCustomRateLimitManager(config)

	if manager.loginLimiter == nil {
		t.Error("Expected login limiter to be initialized")
	}

	if manager.generalLimiter == nil {
		t.Error("Expected general limiter to be initialized")
	}
}

func TestDefaultRateLimitConfig(t *testing.T) {
	config := DefaultRateLimitConfig()

	if config.LoginRate <= 0 {
		t.Error("Expected positive login rate")
	}

	if config.LoginBurst <= 0 {
		t.Error("Expected positive login burst")
	}

	if config.GeneralRate <= 0 {
		t.Error("Expected positive general rate")
	}

	if config.GeneralBurst <= 0 {
		t.Error("Expected positive general burst")
	}

	if config.CleanupInterval <= 0 {
		t.Error("Expected positive cleanup interval")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr ||
		(len(s) > len(substr) && containsHelper(s[1:], substr))
}

func containsHelper(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	if s[:len(substr)] == substr {
		return true
	}
	return containsHelper(s[1:], substr)
}
