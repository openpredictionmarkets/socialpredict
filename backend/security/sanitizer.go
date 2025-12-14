package security

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode"

	"github.com/microcosm-cc/bluemonday"
)

// Sanitizer holds the bluemonday policies for different content types
type Sanitizer struct {
	strictPolicy *bluemonday.Policy
	basicPolicy  *bluemonday.Policy
}

// NewSanitizer creates a new sanitizer with predefined policies
func NewSanitizer() *Sanitizer {
	return &Sanitizer{
		strictPolicy: bluemonday.StrictPolicy(),
		basicPolicy:  createBasicPolicy(),
	}
}

// createBasicPolicy creates a policy that allows basic text formatting but removes scripts
func createBasicPolicy() *bluemonday.Policy {
	p := bluemonday.NewPolicy()
	// Allow basic text formatting
	p.AllowElements("b", "i", "em", "strong")
	// Ensure no scripts, iframes, or dangerous elements
	p.AllowAttrs("href").OnElements("a")
	p.RequireNoReferrerOnLinks(true)
	return p
}

// SanitizeUsername removes any characters that are not lowercase letters or numbers
func (s *Sanitizer) SanitizeUsername(username string) (string, error) {
	// Remove any whitespace
	username = strings.TrimSpace(username)

	// Check if username matches expected pattern
	if match, _ := regexp.MatchString("^[a-z0-9]+$", username); !match {
		return "", fmt.Errorf("username must only contain lowercase letters and numbers")
	}

	// Additional length check
	if len(username) < 3 || len(username) > 30 {
		return "", fmt.Errorf("username must be between 3 and 30 characters")
	}

	return username, nil
}

// SanitizeDisplayName removes HTML/script content but allows basic formatting
func (s *Sanitizer) SanitizeDisplayName(displayName string) (string, error) {
	// Remove leading/trailing whitespace
	displayName = strings.TrimSpace(displayName)

	// Check length
	if len(displayName) == 0 {
		return "", fmt.Errorf("display name cannot be empty")
	}
	if len(displayName) > 50 {
		return "", fmt.Errorf("display name cannot exceed 50 characters")
	}

	// Check for potentially dangerous patterns before sanitizing
	if containsSuspiciousPatterns(displayName) {
		return "", fmt.Errorf("display name contains potentially dangerous content")
	}

	// Sanitize HTML/script content
	sanitized := s.strictPolicy.Sanitize(displayName)

	return sanitized, nil
}

// SanitizeDescription removes dangerous HTML while allowing basic formatting
func (s *Sanitizer) SanitizeDescription(description string) (string, error) {
	// Remove leading/trailing whitespace
	description = strings.TrimSpace(description)

	// Check length
	if len(description) > 2000 {
		return "", fmt.Errorf("description cannot exceed 2000 characters")
	}

	// Check for suspicious patterns before sanitizing
	if containsSuspiciousPatterns(description) {
		return "", fmt.Errorf("description contains potentially dangerous content")
	}

	// Sanitize with basic policy (allows some formatting)
	sanitized := s.basicPolicy.Sanitize(description)

	return sanitized, nil
}

// SanitizeMarketTitle sanitizes market question titles
func (s *Sanitizer) SanitizeMarketTitle(title string) (string, error) {
	// Remove leading/trailing whitespace
	title = strings.TrimSpace(title)

	// Check length constraints
	if len(title) == 0 {
		return "", fmt.Errorf("market title cannot be empty")
	}
	if len(title) > 160 {
		return "", fmt.Errorf("market title cannot exceed 160 characters")
	}

	// Check for suspicious patterns before sanitizing
	if containsSuspiciousPatterns(title) {
		return "", fmt.Errorf("market title contains potentially dangerous content")
	}

	// Sanitize HTML/script content
	sanitized := s.strictPolicy.Sanitize(title)

	return sanitized, nil
}

// SanitizePersonalLink validates and sanitizes personal links
func (s *Sanitizer) SanitizePersonalLink(link string) (string, error) {
	// Remove leading/trailing whitespace
	link = strings.TrimSpace(link)

	// Empty links are allowed
	if link == "" {
		return "", nil
	}

	// Check length
	if len(link) > 200 {
		return "", fmt.Errorf("personal link cannot exceed 200 characters")
	}

	// Parse URL to validate format
	parsedURL, err := url.Parse(link)
	if err != nil {
		return "", fmt.Errorf("invalid URL format: %v", err)
	}

	// Ensure scheme is provided and is http/https
	if parsedURL.Scheme == "" {
		// Add https by default
		link = "https://" + link
		parsedURL, err = url.Parse(link)
		if err != nil {
			return "", fmt.Errorf("invalid URL format after adding scheme: %v", err)
		}
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return "", fmt.Errorf("only http and https URLs are allowed")
	}

	// Check for suspicious domains or patterns
	if containsMaliciousDomain(parsedURL.Host) {
		return "", fmt.Errorf("potentially malicious domain detected")
	}

	return parsedURL.String(), nil
}

// SanitizeEmoji validates that the emoji is from an allowed set
func (s *Sanitizer) SanitizeEmoji(emoji string) (string, error) {
	// Remove leading/trailing whitespace
	emoji = strings.TrimSpace(emoji)

	// Check if it's a valid emoji (basic check for unicode emoji ranges)
	if !isValidEmoji(emoji) {
		return "", fmt.Errorf("invalid emoji format")
	}

	// Check length (emojis should be short)
	if len(emoji) > 20 {
		return "", fmt.Errorf("emoji too long")
	}

	return emoji, nil
}

// SanitizePassword validates password strength
func (s *Sanitizer) SanitizePassword(password string) (string, error) {

	// Basic length checks
	password, err := s.CheckPasswordLength(password)
	if err != nil {
		return "", err
	}

	password, err = s.CheckPasswordChars(password)
	if err != nil {
		return "", err
	}

	return password, nil
}

func (s *Sanitizer) CheckPasswordLength(password string) (string, error) {
	const minLength = 8
	const maxLength = 128

	// Check length constraints
	len := len(password)
	if len < minLength || len > maxLength {
		return "", fmt.Errorf("password must be between %d and %d characters long", minLength, maxLength)
	}

	return password, nil
}

// CheckChars checks for the presence of uppercase, lowercase, and digit characters
func (s *Sanitizer) CheckPasswordChars(password string) (string, error) {
	if !(hasUppercase(password) && hasLowercase(password) && hasDigit(password)) {
		return "", fmt.Errorf("password must contain at least one uppercase letter, one lowercase letter, and one digit")
	}

	return password, nil
}

// containsSuspiciousPatterns checks for common XSS and injection patterns
func containsSuspiciousPatterns(input string) bool {
	suspiciousPatterns := []string{
		"javascript:",
		"vbscript:",
		"data:",
		"<script",
		"</script>",
		"<iframe",
		"<object",
		"<embed",
		"<link",
		"<meta",
		"<style",
		"onload=",
		"onerror=",
		"onclick=",
		"onmouseover=",
		"onfocus=",
		"onblur=",
		"onchange=",
		"onsubmit=",
		"eval(",
		"expression(",
		"url(",
		"@import",
		"<!--",
		"-->",
	}

	inputLower := strings.ToLower(input)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(inputLower, pattern) {
			return true
		}
	}

	return false
}

// containsMaliciousDomain checks against known malicious domain patterns
func containsMaliciousDomain(domain string) bool {
	// This is a basic implementation - in production, you'd use a more comprehensive list
	maliciousPatterns := []string{
		"bit.ly", // URL shorteners can be used maliciously
		"tinyurl.com",
		"t.co",
		"localhost", // Prevent internal network access
		"127.0.0.1",
		"0.0.0.0",
		"::1",
		"169.254.", // Link-local addresses
		"10.",      // Private network ranges
		"192.168.",
		"172.16.",
	}

	domainLower := strings.ToLower(domain)
	for _, pattern := range maliciousPatterns {
		if strings.Contains(domainLower, pattern) {
			return true
		}
	}

	return false
}

// isValidEmoji performs basic emoji validation
func isValidEmoji(emoji string) bool {
	if emoji == "" {
		return false
	}

	for _, r := range emoji {
		if isEmojiRune(r) || isASCIIEmojiRune(r) {
			continue
		}
		return false
	}
	return true
}

type runeRange struct {
	start rune
	end   rune
}

// Common emoji ranges (simplified)
var emojiRanges = []runeRange{
	{start: 0x1F600, end: 0x1F64F}, // Emoticons
	{start: 0x1F300, end: 0x1F5FF}, // Misc Symbols
	{start: 0x1F680, end: 0x1F6FF}, // Transport
	{start: 0x2600, end: 0x26FF},   // Misc symbols
	{start: 0x2700, end: 0x27BF},   // Dingbats
	{start: 0xFE00, end: 0xFE0F},   // Variation selectors
	{start: 0x1F900, end: 0x1F9FF}, // Supplemental symbols
	{start: 0x1F1E6, end: 0x1F1FF}, // Regional indicators
}

func isEmojiRune(r rune) bool {
	for _, emojiRange := range emojiRanges {
		if r >= emojiRange.start && r <= emojiRange.end {
			return true
		}
	}
	return false
}

func isASCIIEmojiRune(r rune) bool {
	// Allow basic ASCII characters for simple emojis like :)
	return r >= 32 && r <= 126
}

func hasUppercase(s string) bool {
	for _, char := range s {
		if unicode.IsUpper(char) {
			return true
		}
	}
	return false
}

func hasLowercase(s string) bool {
	for _, char := range s {
		if unicode.IsLower(char) {
			return true
		}
	}
	return false
}

func hasDigit(s string) bool {
	for _, char := range s {
		if unicode.IsDigit(char) {
			return true
		}
	}
	return false
}
