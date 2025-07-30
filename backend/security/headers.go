package security

import (
	"net/http"
)

// SecurityHeaders holds configuration for security headers
type SecurityHeaders struct {
	CSP                string
	XSSProtection      string
	ContentTypeOptions string
	FrameOptions       string
	ReferrerPolicy     string
	PermissionsPolicy  string
}

// DefaultSecurityHeaders returns default security header values
func DefaultSecurityHeaders() SecurityHeaders {
	return SecurityHeaders{
		CSP: "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline'; " +
			"style-src 'self' 'unsafe-inline'; " +
			"img-src 'self' data: https:; " +
			"connect-src 'self'; " +
			"font-src 'self'; " +
			"object-src 'none'; " +
			"media-src 'self'; " +
			"frame-src 'none'; " +
			"worker-src 'none'; " +
			"child-src 'none'; " +
			"form-action 'self'; " +
			"upgrade-insecure-requests",
		XSSProtection:      "1; mode=block",
		ContentTypeOptions: "nosniff",
		FrameOptions:       "DENY",
		ReferrerPolicy:     "strict-origin-when-cross-origin",
		PermissionsPolicy:  "camera=(), microphone=(), geolocation=(), gyroscope=(), magnetometer=(), usb=()",
	}
}

// SecurityHeadersMiddleware creates middleware that adds security headers to responses
func SecurityHeadersMiddleware(headers SecurityHeaders) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set Content Security Policy
			if headers.CSP != "" {
				w.Header().Set("Content-Security-Policy", headers.CSP)
			}

			// Set XSS Protection
			if headers.XSSProtection != "" {
				w.Header().Set("X-XSS-Protection", headers.XSSProtection)
			}

			// Set Content Type Options
			if headers.ContentTypeOptions != "" {
				w.Header().Set("X-Content-Type-Options", headers.ContentTypeOptions)
			}

			// Set Frame Options
			if headers.FrameOptions != "" {
				w.Header().Set("X-Frame-Options", headers.FrameOptions)
			}

			// Set Referrer Policy
			if headers.ReferrerPolicy != "" {
				w.Header().Set("Referrer-Policy", headers.ReferrerPolicy)
			}

			// Set Permissions Policy
			if headers.PermissionsPolicy != "" {
				w.Header().Set("Permissions-Policy", headers.PermissionsPolicy)
			}

			// Set additional security headers
			w.Header().Set("X-Permitted-Cross-Domain-Policies", "none")
			w.Header().Set("Cross-Origin-Embedder-Policy", "require-corp")
			w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
			w.Header().Set("Cross-Origin-Resource-Policy", "same-origin")

			// Remove server information
			w.Header().Set("Server", "")

			next.ServeHTTP(w, r)
		})
	}
}

// CreateSecurityHeadersMiddleware creates a security headers middleware with default settings
func CreateSecurityHeadersMiddleware() func(http.Handler) http.Handler {
	return SecurityHeadersMiddleware(DefaultSecurityHeaders())
}
