package security

import (
	"net/http"
)

// SecurityService provides a comprehensive security layer for the application
type SecurityService struct {
	Sanitizer   *Sanitizer
	Validator   *Validator
	RateManager *RateLimitManager
	Headers     SecurityHeaders
}

// NewSecurityService creates a new security service with default configurations
func NewSecurityService() *SecurityService {
	return &SecurityService{
		Sanitizer:   NewSanitizer(),
		Validator:   NewValidator(),
		RateManager: NewRateLimitManager(),
		Headers:     DefaultSecurityHeaders(),
	}
}

// NewCustomSecurityService creates a security service with custom rate limiting configuration
func NewCustomSecurityService(rateLimitConfig RateLimitConfig) *SecurityService {
	return &SecurityService{
		Sanitizer:   NewSanitizer(),
		Validator:   NewValidator(),
		RateManager: NewCustomRateLimitManager(rateLimitConfig),
		Headers:     DefaultSecurityHeaders(),
	}
}

// SecurityMiddleware combines all security middleware into a single middleware stack
func (s *SecurityService) SecurityMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		// Apply middleware in order: headers first, then rate limiting, then the handler
		return SecurityHeadersMiddleware(s.Headers)(
			s.RateManager.GetGeneralMiddleware()(next),
		)
	}
}

// LoginSecurityMiddleware provides enhanced security for login endpoints
func (s *SecurityService) LoginSecurityMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		// Apply security headers and stricter rate limiting for login
		return SecurityHeadersMiddleware(s.Headers)(
			s.RateManager.GetLoginMiddleware()(next),
		)
	}
}

// ValidateAndSanitizeUserInput validates and sanitizes user registration/update data
func (s *SecurityService) ValidateAndSanitizeUserInput(input UserInput) (*SanitizedUserInput, error) {
	// First validate the input structure
	if err := s.Validator.ValidateStruct(input); err != nil {
		return nil, err
	}

	// Then sanitize each field
	sanitizedUsername, err := s.Sanitizer.SanitizeUsername(input.Username)
	if err != nil {
		return nil, err
	}

	sanitizedDisplayName, err := s.Sanitizer.SanitizeDisplayName(input.DisplayName)
	if err != nil {
		return nil, err
	}

	sanitizedDescription, err := s.Sanitizer.SanitizeDescription(input.Description)
	if err != nil {
		return nil, err
	}

	sanitizedEmoji, err := s.Sanitizer.SanitizeEmoji(input.PersonalEmoji)
	if err != nil {
		return nil, err
	}

	// Sanitize personal links
	var sanitizedLinks [4]string
	links := [4]string{input.PersonalLink1, input.PersonalLink2, input.PersonalLink3, input.PersonalLink4}
	for i, link := range links {
		sanitized, err := s.Sanitizer.SanitizePersonalLink(link)
		if err != nil {
			return nil, err
		}
		sanitizedLinks[i] = sanitized
	}

	// Validate password if provided
	if input.Password != "" {
		if err := s.Sanitizer.SanitizePassword(input.Password); err != nil {
			return nil, err
		}
	}

	return &SanitizedUserInput{
		Username:      sanitizedUsername,
		DisplayName:   sanitizedDisplayName,
		Description:   sanitizedDescription,
		PersonalEmoji: sanitizedEmoji,
		PersonalLink1: sanitizedLinks[0],
		PersonalLink2: sanitizedLinks[1],
		PersonalLink3: sanitizedLinks[2],
		PersonalLink4: sanitizedLinks[3],
		Password:      input.Password, // Keep original for hashing
	}, nil
}

// ValidateAndSanitizeMarketInput validates and sanitizes market creation data
func (s *SecurityService) ValidateAndSanitizeMarketInput(input MarketInput) (*SanitizedMarketInput, error) {
	// First validate the input structure
	if err := s.Validator.ValidateStruct(input); err != nil {
		return nil, err
	}

	// Sanitize the market title
	sanitizedTitle, err := s.Sanitizer.SanitizeMarketTitle(input.Title)
	if err != nil {
		return nil, err
	}

	// Sanitize the description
	sanitizedDescription, err := s.Sanitizer.SanitizeDescription(input.Description)
	if err != nil {
		return nil, err
	}

	return &SanitizedMarketInput{
		Title:       sanitizedTitle,
		Description: sanitizedDescription,
		EndTime:     input.EndTime,
	}, nil
}

// Input structures for validation
type UserInput struct {
	Username      string `validate:"required,min=3,max=30,username"`
	DisplayName   string `validate:"required,min=1,max=50,safe_string"`
	Description   string `validate:"max=2000,safe_string"`
	PersonalEmoji string `validate:"max=20"`
	PersonalLink1 string `validate:"max=200,url"`
	PersonalLink2 string `validate:"max=200,url"`
	PersonalLink3 string `validate:"max=200,url"`
	PersonalLink4 string `validate:"max=200,url"`
	Password      string `validate:"required,strong_password"`
}

type MarketInput struct {
	Title       string `validate:"required,min=1,max=160,safe_string"`
	Description string `validate:"max=2000,safe_string"`
	EndTime     string `validate:"required,future_date"`
}

type BetInput struct {
	MarketID string  `validate:"required,market_id"`
	Amount   float64 `validate:"required,positive_amount"`
	Outcome  string  `validate:"required,market_outcome"`
}

// Sanitized output structures
type SanitizedUserInput struct {
	Username      string
	DisplayName   string
	Description   string
	PersonalEmoji string
	PersonalLink1 string
	PersonalLink2 string
	PersonalLink3 string
	PersonalLink4 string
	Password      string
}

type SanitizedMarketInput struct {
	Title       string
	Description string
	EndTime     string
}

type SanitizedBetInput struct {
	MarketID string
	Amount   float64
	Outcome  string
}

// ValidateAndSanitizeBetInput validates and sanitizes betting data
func (s *SecurityService) ValidateAndSanitizeBetInput(input BetInput) (*SanitizedBetInput, error) {
	// Validate the input structure
	if err := s.Validator.ValidateStruct(input); err != nil {
		return nil, err
	}

	// No additional sanitization needed for betting data since validation covers it
	return &SanitizedBetInput{
		MarketID: input.MarketID,
		Amount:   input.Amount,
		Outcome:  input.Outcome,
	}, nil
}

// GetDefaultConfig returns the default security configuration
func GetDefaultConfig() SecurityConfig {
	return SecurityConfig{
		RateLimit: DefaultRateLimitConfig(),
		Headers:   DefaultSecurityHeaders(),
	}
}

// SecurityConfig holds all security configuration
type SecurityConfig struct {
	RateLimit RateLimitConfig
	Headers   SecurityHeaders
}
