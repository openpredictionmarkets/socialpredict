package runtime

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"socialpredict/security"
)

// CORSConfig describes the request-boundary CORS posture owned by runtime bootstrap.
type CORSConfig struct {
	Enabled          bool
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// SecurityConfig is the runtime-owned security posture for the process.
type SecurityConfig struct {
	JWTSigningKey     []byte
	TrustProxyHeaders bool
	CORS              CORSConfig
	Headers           security.SecurityHeaders
}

// LoadSecurityConfigFromEnv validates and freezes deployment-sensitive security settings.
func LoadSecurityConfigFromEnv() (SecurityConfig, error) {
	signingKey := []byte(strings.TrimSpace(os.Getenv("JWT_SIGNING_KEY")))
	if len(signingKey) == 0 {
		return SecurityConfig{}, fmt.Errorf("security config: JWT_SIGNING_KEY is required")
	}

	headers := security.DefaultSecurityHeaders()
	headers.StrictTransportSecurity = strictTransportSecurityHeader()

	return SecurityConfig{
		JWTSigningKey:     signingKey,
		TrustProxyHeaders: getRuntimeBoolEnv("TRUST_PROXY_HEADERS", false),
		CORS: CORSConfig{
			Enabled:          getRuntimeBoolEnv("CORS_ENABLED", true),
			AllowedOrigins:   getRuntimeListEnv("CORS_ALLOW_ORIGINS", "*"),
			AllowedMethods:   getRuntimeListEnv("CORS_ALLOW_METHODS", "GET,POST,PUT,PATCH,DELETE,OPTIONS"),
			AllowedHeaders:   getRuntimeListEnv("CORS_ALLOW_HEADERS", "Content-Type,Authorization"),
			ExposedHeaders:   getRuntimeListEnv("CORS_EXPOSE_HEADERS", ""),
			AllowCredentials: getRuntimeBoolEnv("CORS_ALLOW_CREDENTIALS", false),
			MaxAge:           getRuntimeIntEnv("CORS_MAX_AGE", 600),
		},
		Headers: headers,
	}, nil
}

func strictTransportSecurityHeader() string {
	if !getRuntimeBoolEnv("SECURITY_HSTS_ENABLED", false) {
		return ""
	}

	maxAge := getRuntimeIntEnv("SECURITY_HSTS_MAX_AGE", 31536000)
	parts := []string{fmt.Sprintf("max-age=%d", maxAge)}
	if getRuntimeBoolEnv("SECURITY_HSTS_INCLUDE_SUBDOMAINS", false) {
		parts = append(parts, "includeSubDomains")
	}
	if getRuntimeBoolEnv("SECURITY_HSTS_PRELOAD", false) {
		parts = append(parts, "preload")
	}
	return strings.Join(parts, "; ")
}

func getRuntimeListEnv(key, def string) []string {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		val = def
	}
	if val == "" {
		return nil
	}

	parts := strings.Split(val, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func getRuntimeBoolEnv(key string, def bool) bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if value == "" {
		return def
	}
	return value == "1" || value == "true" || value == "yes" || value == "on"
}

func getRuntimeIntEnv(key string, def int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return def
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return def
	}
	return parsed
}
