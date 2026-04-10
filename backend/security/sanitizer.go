package security

import (
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/microcosm-cc/bluemonday"
)

var usernamePattern = regexp.MustCompile("^[a-z0-9]+$")

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
	p.AllowStandardURLs()
	p.RequireParseableURLs(true)
	p.AllowURLSchemes("http", "https")
	p.AllowElements("a", "b", "i", "em", "strong")
	p.AllowAttrs("href").Matching(regexp.MustCompile(`^https?://`)).OnElements("a")
	p.RequireNoReferrerOnLinks(true)
	return p
}

// SanitizeUsername removes any characters that are not lowercase letters or numbers
func (s *Sanitizer) SanitizeUsername(username string) (string, error) {
	username = strings.TrimSpace(username)

	if !usernamePattern.MatchString(username) {
		return "", fmt.Errorf("username must only contain lowercase letters and numbers")
	}

	if utf8.RuneCountInString(username) < 3 || utf8.RuneCountInString(username) > 30 {
		return "", fmt.Errorf("username must be between 3 and 30 characters")
	}

	return username, nil
}

// SanitizeDisplayName removes HTML/script content but allows basic formatting
func (s *Sanitizer) SanitizeDisplayName(displayName string) (string, error) {
	displayName = strings.TrimSpace(displayName)

	if len(displayName) == 0 {
		return "", fmt.Errorf("display name cannot be empty")
	}
	if len(displayName) > 50 {
		return "", fmt.Errorf("display name cannot exceed 50 characters")
	}

	if containsSuspiciousPatterns(displayName) {
		return "", fmt.Errorf("display name contains potentially dangerous content")
	}

	sanitized := strings.TrimSpace(s.strictPolicy.Sanitize(displayName))
	if sanitized == "" {
		return "", fmt.Errorf("display name cannot be empty")
	}

	return sanitized, nil
}

// SanitizeDescription removes dangerous HTML while allowing basic formatting
func (s *Sanitizer) SanitizeDescription(description string) (string, error) {
	description = strings.TrimSpace(description)

	if len(description) > 2000 {
		return "", fmt.Errorf("description cannot exceed 2000 characters")
	}

	if containsSuspiciousPatterns(description) {
		return "", fmt.Errorf("description contains potentially dangerous content")
	}

	sanitized := strings.TrimSpace(s.basicPolicy.Sanitize(description))

	return sanitized, nil
}

// SanitizeMarketTitle sanitizes market question titles
func (s *Sanitizer) SanitizeMarketTitle(title string) (string, error) {
	title = strings.TrimSpace(title)

	if len(title) == 0 {
		return "", fmt.Errorf("market title cannot be empty")
	}
	if len(title) > 160 {
		return "", fmt.Errorf("market title cannot exceed 160 characters")
	}

	if containsSuspiciousPatterns(title) {
		return "", fmt.Errorf("market title contains potentially dangerous content")
	}

	sanitized := strings.TrimSpace(s.strictPolicy.Sanitize(title))
	if sanitized == "" {
		return "", fmt.Errorf("market title cannot be empty")
	}

	return sanitized, nil
}

// SanitizePersonalLink validates and sanitizes personal links
func (s *Sanitizer) SanitizePersonalLink(link string) (string, error) {
	link = strings.TrimSpace(link)
	if link == "" {
		return "", nil
	}

	if err := validatePersonalLinkLength(link); err != nil {
		return "", err
	}

	parsedURL, err := parseURLWithScheme(link)
	if err != nil {
		return "", err
	}

	if err := validateAllowedScheme(parsedURL.Scheme); err != nil {
		return "", err
	}
	if parsedURL.User != nil {
		return "", fmt.Errorf("personal link cannot include credentials")
	}
	if parsedURL.Hostname() == "" {
		return "", fmt.Errorf("personal link must include a valid host")
	}

	if containsMaliciousDomain(parsedURL.Hostname()) {
		return "", fmt.Errorf("potentially malicious domain detected")
	}

	return parsedURL.String(), nil
}

func validatePersonalLinkLength(link string) error {
	if utf8.RuneCountInString(link) > 200 {
		return fmt.Errorf("personal link cannot exceed 200 characters")
	}
	return nil
}

func parseURLWithScheme(link string) (*url.URL, error) {
	if strings.ContainsAny(link, " \t\r\n") {
		return nil, fmt.Errorf("invalid URL format")
	}

	parsedURL, err := url.Parse(link)
	if err != nil {
		return nil, fmt.Errorf("invalid URL format: %v", err)
	}

	if parsedURL.Scheme != "" {
		if parsedURL.Host == "" {
			return nil, fmt.Errorf("invalid URL format")
		}
		return parsedURL, nil
	}

	withScheme := "https://" + link
	parsedURL, err = url.Parse(withScheme)
	if err != nil {
		return nil, fmt.Errorf("invalid URL format after adding scheme: %v", err)
	}
	if parsedURL.Host == "" {
		return nil, fmt.Errorf("invalid URL format")
	}
	return parsedURL, nil
}

func validateAllowedScheme(scheme string) error {
	scheme = strings.ToLower(strings.TrimSpace(scheme))
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("only http and https URLs are allowed")
	}
	return nil
}

// SanitizeEmoji validates that the emoji is from an allowed set
func (s *Sanitizer) SanitizeEmoji(emoji string) (string, error) {
	emoji = strings.TrimSpace(emoji)

	if !isValidEmoji(emoji) {
		return "", fmt.Errorf("invalid emoji format")
	}

	if utf8.RuneCountInString(emoji) > 20 {
		return "", fmt.Errorf("emoji too long")
	}

	return emoji, nil
}

// SanitizePassword validates password strength
func (s *Sanitizer) SanitizePassword(password string) (string, error) {
	if password != strings.TrimSpace(password) {
		return "", fmt.Errorf("password cannot start or end with whitespace")
	}

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

	length := utf8.RuneCountInString(password)
	if length < minLength || length > maxLength {
		return "", fmt.Errorf("password must be between %d and %d characters long", minLength, maxLength)
	}

	return password, nil
}

func (s *Sanitizer) CheckPasswordChars(password string) (string, error) {
	for _, char := range password {
		if unicode.IsSpace(char) || unicode.IsControl(char) {
			return "", fmt.Errorf("password cannot contain whitespace or control characters")
		}
	}

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
	domainLower := strings.ToLower(strings.TrimSpace(domain))
	if domainLower == "" {
		return true
	}

	if ip, err := netip.ParseAddr(domainLower); err == nil {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
			return true
		}
		return false
	}

	blockedDomains := []string{"bit.ly", "tinyurl.com", "t.co", "localhost"}
	for _, blocked := range blockedDomains {
		if domainLower == blocked || strings.HasSuffix(domainLower, "."+blocked) {
			return true
		}
	}

	if host, _, err := net.SplitHostPort(domainLower); err == nil {
		return containsMaliciousDomain(host)
	}

	return false
}

func isValidEmoji(emoji string) bool {
	if emoji == "" {
		return false
	}

	hasUnicodeEmoji := false
	hasASCIIEmoji := false
	for _, r := range emoji {
		switch {
		case isEmojiRune(r):
			hasUnicodeEmoji = true
		case isASCIIEmojiRune(r):
			hasASCIIEmoji = true
			if hasUnicodeEmoji {
				return false
			}
		default:
			return false
		}
	}
	return hasUnicodeEmoji || hasASCIIEmoji
}

type runeRange struct {
	start rune
	end   rune
}

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
	return strings.ContainsRune(":;=8xX()-[]{}\\/|^'\",.<>*oOPpD", r)
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
