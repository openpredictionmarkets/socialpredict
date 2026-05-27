package runtime

import (
	"strings"
	"testing"
)

func TestLoadSecurityConfigFromEnvRequiresJWTSigningKey(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", "   ")

	config, err := LoadSecurityConfigFromEnv()
	if err == nil {
		t.Fatalf("expected missing JWT signing key error")
	}
	if len(config.JWTSigningKey) != 0 {
		t.Fatalf("expected no signing key on error")
	}
}

func TestLoadSecurityConfigFromEnvOwnsDefaults(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key")

	config, err := LoadSecurityConfigFromEnv()
	if err != nil {
		t.Fatalf("LoadSecurityConfigFromEnv returned error: %v", err)
	}
	if string(config.JWTSigningKey) != "test-secret-key" {
		t.Fatalf("unexpected signing key")
	}
	if config.TrustProxyHeaders {
		t.Fatalf("trusted proxy headers should default to false")
	}
	if !config.CORS.Enabled {
		t.Fatalf("CORS should default to enabled")
	}
	if got := config.CORS.AllowedOrigins; len(got) != 1 || got[0] != "*" {
		t.Fatalf("CORS allowed origins = %v, want [*]", got)
	}
	if config.Headers.StrictTransportSecurity != "" {
		t.Fatalf("HSTS should default to disabled, got %q", config.Headers.StrictTransportSecurity)
	}
	if got := config.Headers.FrameOptions; got != "DENY" {
		t.Fatalf("X-Frame-Options = %q, want DENY", got)
	}
	if got := config.Headers.CSP; !strings.Contains(got, "frame-ancestors 'none'") {
		t.Fatalf("CSP missing default frame-ancestors: %q", got)
	}
	if config.Share.PublicBaseURL != "http://localhost" {
		t.Fatalf("PublicBaseURL = %q", config.Share.PublicBaseURL)
	}
}

func TestLoadSecurityConfigFromEnvOwnsProxyCORSAndHSTSOverrides(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key")
	t.Setenv("TRUST_PROXY_HEADERS", "true")
	t.Setenv("CORS_ALLOW_ORIGINS", "https://app.example, https://admin.example")
	t.Setenv("CORS_ALLOW_METHODS", "GET,POST")
	t.Setenv("CORS_ALLOW_CREDENTIALS", "true")
	t.Setenv("SECURITY_HSTS_ENABLED", "true")
	t.Setenv("SECURITY_HSTS_MAX_AGE", "123")
	t.Setenv("SECURITY_HSTS_INCLUDE_SUBDOMAINS", "true")
	t.Setenv("SECURITY_HSTS_PRELOAD", "true")
	t.Setenv("SECURITY_FRAME_ANCESTORS", "'self', https://partner.example")
	t.Setenv("PUBLIC_BASE_URL", "https://kconfs.com")
	t.Setenv("SHARE_DEFAULT_IMAGE_URL", "https://cdn.example/share.png")
	t.Setenv("SHARE_SITE_NAME", "KConfs")

	config, err := LoadSecurityConfigFromEnv()
	if err != nil {
		t.Fatalf("LoadSecurityConfigFromEnv returned error: %v", err)
	}
	if !config.TrustProxyHeaders {
		t.Fatalf("trusted proxy headers should be enabled")
	}
	if got := config.CORS.AllowedOrigins; len(got) != 2 || got[0] != "https://app.example" || got[1] != "https://admin.example" {
		t.Fatalf("CORS allowed origins = %v", got)
	}
	if got := config.CORS.AllowedMethods; len(got) != 2 || got[0] != "GET" || got[1] != "POST" {
		t.Fatalf("CORS allowed methods = %v", got)
	}
	if !config.CORS.AllowCredentials {
		t.Fatalf("CORS allow credentials should be enabled")
	}
	if got := config.Headers.StrictTransportSecurity; got != "max-age=123; includeSubDomains; preload" {
		t.Fatalf("Strict-Transport-Security = %q", got)
	}
	if got := config.Headers.CSP; !strings.Contains(got, "frame-ancestors 'self' https://partner.example") {
		t.Fatalf("CSP missing frame allowlist: %q", got)
	}
	if got := config.Headers.FrameOptions; got != "" {
		t.Fatalf("X-Frame-Options should be omitted when CSP frame-ancestors allowlist is configured, got %q", got)
	}
	if config.Share.PublicBaseURL != "https://kconfs.com" || config.Share.DefaultImageURL != "https://cdn.example/share.png" || config.Share.SiteName != "KConfs" {
		t.Fatalf("unexpected share config: %+v", config.Share)
	}
}
