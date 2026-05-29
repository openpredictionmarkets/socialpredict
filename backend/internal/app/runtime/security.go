package runtime

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"socialpredict/security"

	"golang.org/x/time/rate"
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
	Share             ShareConfig
	RateLimit         security.RateLimitConfig
}

// ShareConfig describes public market sharing metadata owned by runtime config.
type ShareConfig struct {
	PublicBaseURL   string
	DefaultImageURL string
	SiteName        string
}

// LoadSecurityConfigFromEnv validates and freezes deployment-sensitive security settings.
func LoadSecurityConfigFromEnv() (SecurityConfig, error) {
	signingKey := []byte(strings.TrimSpace(os.Getenv("JWT_SIGNING_KEY")))
	if len(signingKey) == 0 {
		return SecurityConfig{}, fmt.Errorf("security config: JWT_SIGNING_KEY is required")
	}

	headers := security.DefaultSecurityHeaders()
	headers.StrictTransportSecurity = strictTransportSecurityHeader()
	applyFrameAncestors(&headers, getRuntimeListEnv("SECURITY_FRAME_ANCESTORS", "'none'"))
	trustProxyHeaders := getRuntimeBoolEnv("TRUST_PROXY_HEADERS", false)
	rateLimit, err := rateLimitConfigFromEnv()
	if err != nil {
		return SecurityConfig{}, err
	}
	rateLimit.TrustProxyHeaders = trustProxyHeaders

	return SecurityConfig{
		JWTSigningKey:     signingKey,
		TrustProxyHeaders: trustProxyHeaders,
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
		Share: ShareConfig{
			PublicBaseURL:   publicBaseURL(),
			DefaultImageURL: strings.TrimSpace(os.Getenv("SHARE_DEFAULT_IMAGE_URL")),
			SiteName:        getRuntimeStringEnv("SHARE_SITE_NAME", "SocialPredict"),
		},
		RateLimit: rateLimit,
	}, nil
}

func rateLimitConfigFromEnv() (security.RateLimitConfig, error) {
	config := security.DefaultRateLimitConfig()
	var err error

	if config.LoginRate, err = getRuntimeRateEnv("RATE_LIMIT_LOGIN_RATE_PER_SECOND", config.LoginRate); err != nil {
		return security.RateLimitConfig{}, err
	}
	if config.LoginBurst, err = getRuntimePositiveIntEnv("RATE_LIMIT_LOGIN_BURST", config.LoginBurst); err != nil {
		return security.RateLimitConfig{}, err
	}
	if config.GeneralRate, err = getRuntimeRateEnv("RATE_LIMIT_GENERAL_RATE_PER_SECOND", config.GeneralRate); err != nil {
		return security.RateLimitConfig{}, err
	}
	if config.GeneralBurst, err = getRuntimePositiveIntEnv("RATE_LIMIT_GENERAL_BURST", config.GeneralBurst); err != nil {
		return security.RateLimitConfig{}, err
	}
	if config.CleanupInterval, err = getRuntimePositiveDurationEnv("RATE_LIMIT_CLEANUP_INTERVAL", config.CleanupInterval); err != nil {
		return security.RateLimitConfig{}, err
	}

	return config, nil
}

func applyFrameAncestors(headers *security.SecurityHeaders, ancestors []string) {
	if headers == nil {
		return
	}
	if len(ancestors) == 0 {
		ancestors = []string{"'none'"}
	}

	headers.CSP = appendCSPDirective(headers.CSP, "frame-ancestors", strings.Join(ancestors, " "))
	if len(ancestors) == 1 && ancestors[0] == "'none'" {
		headers.FrameOptions = "DENY"
		return
	}
	headers.FrameOptions = ""
}

func appendCSPDirective(csp string, name string, value string) string {
	csp = strings.TrimSpace(csp)
	if csp == "" {
		return name + " " + value
	}
	csp = strings.TrimRight(csp, "; ")
	directives := strings.Split(csp, ";")
	filtered := make([]string, 0, len(directives)+1)
	prefix := name + " "
	for _, directive := range directives {
		directive = strings.TrimSpace(directive)
		if directive == "" || strings.HasPrefix(directive, prefix) {
			continue
		}
		filtered = append(filtered, directive)
	}
	filtered = append(filtered, prefix+value)
	return strings.Join(filtered, "; ")
}

func publicBaseURL() string {
	if value := strings.TrimSpace(os.Getenv("PUBLIC_BASE_URL")); value != "" {
		return value
	}
	if value := strings.TrimSpace(os.Getenv("DOMAIN_URL")); value != "" {
		return value
	}
	return "http://localhost"
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

func getRuntimeStringEnv(key, def string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return def
	}
	return value
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

func getRuntimePositiveIntEnv(key string, def int) (int, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return def, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("security config: %s must be a positive integer", key)
	}
	return parsed, nil
}

func getRuntimeRateEnv(key string, def rate.Limit) (rate.Limit, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return def, nil
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("security config: %s must be a positive decimal requests-per-second value", key)
	}
	return rate.Limit(parsed), nil
}

func getRuntimePositiveDurationEnv(key string, def time.Duration) (time.Duration, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return def, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("security config: %s must be a positive Go duration, e.g. 5m", key)
	}
	return parsed, nil
}
