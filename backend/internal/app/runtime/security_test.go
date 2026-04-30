package runtime

import "testing"

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
}
