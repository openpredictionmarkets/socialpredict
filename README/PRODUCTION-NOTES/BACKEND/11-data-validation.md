# Data Validation & Sanitization Implementation Plan

## Overview
Enhance the existing validation and sanitization systems to provide comprehensive input validation, output sanitization, and data integrity checks suitable for production environments.

## Current State Analysis
- Basic input sanitization in `security/sanitizer.go`
- Some validation using `github.com/go-playground/validator/v10`
- Limited validation rules and patterns
- No comprehensive output sanitization
- Missing business rule validation
- No data integrity monitoring

## Implementation Steps

### Step 1: Enhanced Input Validation Framework
**Timeline: 3-4 days**

Create a comprehensive validation framework with custom rules:

```go
// validation/engine.go
package validation

import (
    "reflect"
    "regexp"
    "strings"
    "github.com/go-playground/validator/v10"
)

type ValidationEngine struct {
    validator     *validator.Validate
    customRules   map[string]ValidationRule
    businessRules map[string]BusinessRule
    sanitizer     *Sanitizer
}

type ValidationRule struct {
    Name        string
    Description string
    Validator   func(interface{}) bool
    Message     string
}

type BusinessRule struct {
    Name        string
    Description string
    Validator   func(interface{}, context.Context) error
    Message     string
}

func NewValidationEngine() *ValidationEngine {
    ve := &ValidationEngine{
        validator:     validator.New(),
        customRules:   make(map[string]ValidationRule),
        businessRules: make(map[string]BusinessRule),
        sanitizer:     NewSanitizer(),
    }

    ve.registerCustomValidators()
    ve.registerBusinessRules()

    return ve
}

func (ve *ValidationEngine) registerCustomValidators() {
    // Username validation
    ve.validator.RegisterValidation("username", func(fl validator.FieldLevel) bool {
        username := fl.Field().String()
        // Only alphanumeric, underscore, hyphen, 3-20 chars
        matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]{3,20}$`, username)
        return matched
    })

    // Market title validation
    ve.validator.RegisterValidation("market_title", func(fl validator.FieldLevel) bool {
        title := fl.Field().String()
        // 10-200 characters, no HTML tags
        if len(title) < 10 || len(title) > 200 {
            return false
        }
        // Check for HTML tags
        htmlTag := regexp.MustCompile(`<[^>]*>`)
        return !htmlTag.MatchString(title)
    })

    // Currency amount validation
    ve.validator.RegisterValidation("currency_amount", func(fl validator.FieldLevel) bool {
        amount := fl.Field().Float()
        // Must be positive, max 2 decimal places
        return amount > 0 && amount*100 == float64(int(amount*100))
    })

    // Future date validation
    ve.validator.RegisterValidation("future_date", func(fl validator.FieldLevel) bool {
        date := fl.Field().Interface().(time.Time)
        return date.After(time.Now())
    })
}

// Validation request structure
type ValidationRequest struct {
    Data   interface{}            `json:"data"`
    Rules  map[string]interface{} `json:"rules,omitempty"`
    Context context.Context       `json:"-"`
}

type ValidationResult struct {
    Valid         bool                    `json:"valid"`
    Errors        map[string][]string     `json:"errors,omitempty"`
    Warnings      map[string][]string     `json:"warnings,omitempty"`
    SanitizedData interface{}             `json:"sanitized_data,omitempty"`
    BusinessRules []BusinessRuleResult    `json:"business_rules,omitempty"`
}

type BusinessRuleResult struct {
    Rule    string `json:"rule"`
    Passed  bool   `json:"passed"`
    Message string `json:"message,omitempty"`
}
```

### Step 2: Request Validation Middleware
**Timeline: 2 days**

Create middleware for automatic request validation:

```go
// middleware/validation.go
func ValidationMiddleware(ve *validation.ValidationEngine) mux.MiddlewareFunc {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Skip validation for GET requests
            if r.Method == "GET" {
                next.ServeHTTP(w, r)
                return
            }

            // Get validation rules for this endpoint
            rules := getValidationRules(r.URL.Path, r.Method)
            if rules == nil {
                next.ServeHTTP(w, r)
                return
            }

            // Parse request body
            var requestData map[string]interface{}
            if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
                writeValidationError(w, "Invalid JSON in request body")
                return
            }

            // Validate request
            result := ve.ValidateRequest(validation.ValidationRequest{
                Data:    requestData,
                Rules:   rules,
                Context: r.Context(),
            })

            if !result.Valid {
                writeValidationError(w, result.Errors)
                return
            }

            // Replace request body with sanitized data
            sanitizedBody, _ := json.Marshal(result.SanitizedData)
            r.Body = ioutil.NopCloser(strings.NewReader(string(sanitizedBody)))

            // Add validation result to context
            ctx := context.WithValue(r.Context(), "validation_result", result)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// Validation rules configuration
func getValidationRules(path, method string) map[string]interface{} {
    rules := map[string]map[string]interface{}{
        "POST:/v1/users": {
            "username": "required,username,min=3,max=20",
            "email":    "required,email",
            "password": "required,min=8,containsany=!@#$%^&*,containsany=0123456789,containsany=ABCDEFGHIJKLMNOPQRSTUVWXYZ",
        },
        "POST:/v1/markets": {
            "title":       "required,market_title",
            "description": "required,min=20,max=1000",
            "end_date":    "required,future_date",
            "category":    "required,oneof=sports politics crypto tech",
        },
        "POST:/v1/bets": {
            "market_id": "required,numeric,min=1",
            "amount":    "required,currency_amount,min=1,max=10000",
            "outcome":   "required,oneof=yes no",
        },
    }

    key := fmt.Sprintf("%s:%s", method, path)
    return rules[key]
}
```

### Step 3: Advanced Input Sanitization
**Timeline: 2-3 days**

Enhance the existing sanitization system:

```go
// sanitization/sanitizer.go
package sanitization

import (
    "html"
    "regexp"
    "strings"
    "unicode"
    "github.com/microcosm-cc/bluemonday"
)

type Sanitizer struct {
    htmlPolicy  *bluemonday.Policy
    sqlPattern  *regexp.Regexp
    xssPatterns []*regexp.Regexp
}

func NewSanitizer() *Sanitizer {
    s := &Sanitizer{
        htmlPolicy: bluemonday.StrictPolicy(),
        sqlPattern: regexp.MustCompile(`(?i)(union|select|insert|update|delete|drop|exec|script)`),
    }

    // XSS patterns
    s.xssPatterns = []*regexp.Regexp{
        regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`),
        regexp.MustCompile(`(?i)javascript:`),
        regexp.MustCompile(`(?i)on\w+\s*=`),
        regexp.MustCompile(`(?i)data:\s*text/html`),
    }

    return s
}

func (s *Sanitizer) SanitizeString(input string) string {
    if input == "" {
        return input
    }

    // HTML escape
    sanitized := html.EscapeString(input)

    // Remove null bytes
    sanitized = strings.ReplaceAll(sanitized, "\x00", "")

    // Normalize whitespace
    sanitized = normalizeWhitespace(sanitized)

    // Remove potential XSS patterns
    for _, pattern := range s.xssPatterns {
        sanitized = pattern.ReplaceAllString(sanitized, "")
    }

    return sanitized
}

func (s *Sanitizer) SanitizeHTML(input string) string {
    return s.htmlPolicy.Sanitize(input)
}

func (s *Sanitizer) CheckSQLInjection(input string) bool {
    return s.sqlPattern.MatchString(input)
}

func (s *Sanitizer) SanitizeFilename(filename string) string {
    // Remove path separators and dangerous characters
    dangerous := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)
    sanitized := dangerous.ReplaceAllString(filename, "_")

    // Limit length
    if len(sanitized) > 255 {
        sanitized = sanitized[:255]
    }

    return sanitized
}

func normalizeWhitespace(s string) string {
    // Replace multiple whitespace with single space
    space := regexp.MustCompile(`\s+`)
    s = space.ReplaceAllString(s, " ")

    // Trim leading/trailing whitespace
    return strings.TrimSpace(s)
}

// Sanitization rules for different data types
type SanitizationRule struct {
    Field string
    Rules []string
}

func (s *Sanitizer) ApplyRules(data map[string]interface{}, rules []SanitizationRule) map[string]interface{} {
    result := make(map[string]interface{})

    for key, value := range data {
        if strValue, ok := value.(string); ok {
            // Find rules for this field
            for _, rule := range rules {
                if rule.Field == key {
                    result[key] = s.applySanitizationRules(strValue, rule.Rules)
                    break
                }
            }

            // Default sanitization if no specific rules
            if result[key] == nil {
                result[key] = s.SanitizeString(strValue)
            }
        } else {
            result[key] = value
        }
    }

    return result
}

func (s *Sanitizer) applySanitizationRules(input string, rules []string) string {
    result := input

    for _, rule := range rules {
        switch rule {
        case "html_escape":
            result = s.SanitizeString(result)
        case "html_strip":
            result = s.SanitizeHTML(result)
        case "lowercase":
            result = strings.ToLower(result)
        case "uppercase":
            result = strings.ToUpper(result)
        case "trim":
            result = strings.TrimSpace(result)
        case "filename":
            result = s.SanitizeFilename(result)
        }
    }

    return result
}
```

### Step 4: Business Rule Validation
**Timeline: 3 days**

Implement business logic validation:

```go
// validation/business_rules.go
package validation

type BusinessRuleValidator struct {
    db       *gorm.DB
    cache    *cache.Cache
    logger   *logging.Logger
}

func (brv *BusinessRuleValidator) ValidateUserCreation(ctx context.Context, user *models.User) error {
    // Check username uniqueness
    var existingUser models.User
    if err := brv.db.Where("username = ?", user.Username).First(&existingUser).Error; err == nil {
        return errors.New("username already exists")
    }

    // Check email uniqueness
    if err := brv.db.Where("email = ?", user.Email).First(&existingUser).Error; err == nil {
        return errors.New("email already registered")
    }

    // Check username for inappropriate content
    if brv.containsInappropriateContent(user.Username) {
        return errors.New("username contains inappropriate content")
    }

    return nil
}

func (brv *BusinessRuleValidator) ValidateMarketCreation(ctx context.Context, market *models.Market) error {
    // Check if user can create markets
    user := getUserFromContext(ctx)
    if user.Role != "admin" && user.MarketsCreated >= user.MaxMarkets {
        return errors.New("market creation limit reached")
    }

    // Validate end date
    if market.EndDate.Before(time.Now().Add(1 * time.Hour)) {
        return errors.New("market must end at least 1 hour in the future")
    }

    if market.EndDate.After(time.Now().Add(365 * 24 * time.Hour)) {
        return errors.New("market end date cannot be more than 1 year in the future")
    }

    // Check for duplicate markets
    var existingMarket models.Market
    if err := brv.db.Where("title = ? AND creator_id = ?", market.Title, market.CreatorID).First(&existingMarket).Error; err == nil {
        return errors.New("you have already created a market with this title")
    }

    return nil
}

func (brv *BusinessRuleValidator) ValidateBetPlacement(ctx context.Context, bet *models.Bet) error {
    // Check market status
    var market models.Market
    if err := brv.db.First(&market, bet.MarketID).Error; err != nil {
        return errors.New("market not found")
    }

    if market.Status != "active" {
        return errors.New("market is not active")
    }

    if market.EndDate.Before(time.Now()) {
        return errors.New("market has already ended")
    }

    // Check user balance
    user := getUserFromContext(ctx)
    if user.Credits < bet.Amount {
        return errors.New("insufficient credits")
    }

    // Check minimum bet amount
    if bet.Amount < 1 {
        return errors.New("minimum bet amount is 1 credit")
    }

    // Check maximum bet amount
    if bet.Amount > 1000 {
        return errors.New("maximum bet amount is 1000 credits")
    }

    // Check if user is not betting against themselves
    if market.CreatorID == user.ID {
        return errors.New("cannot bet on your own market")
    }

    return nil
}

func (brv *BusinessRuleValidator) containsInappropriateContent(text string) bool {
    // Check against banned words list
    bannedWords := []string{"admin", "moderator", "support", "official"}
    lowerText := strings.ToLower(text)

    for _, word := range bannedWords {
        if strings.Contains(lowerText, word) {
            return true
        }
    }

    return false
}
```

### Step 5: Output Sanitization and Response Security
**Timeline: 2 days**

Implement output sanitization for API responses:

```go
// sanitization/output.go
package sanitization

type OutputSanitizer struct {
    htmlPolicy *bluemonday.Policy
}

func NewOutputSanitizer() *OutputSanitizer {
    // Create policy for user-generated content
    policy := bluemonday.NewPolicy()
    policy.AllowElements("b", "i", "em", "strong", "u")
    policy.AllowElements("p", "br")
    policy.AllowElements("ul", "ol", "li")

    return &OutputSanitizer{
        htmlPolicy: policy,
    }
}

func (os *OutputSanitizer) SanitizeResponse(data interface{}) interface{} {
    return os.sanitizeValue(reflect.ValueOf(data)).Interface()
}

func (os *OutputSanitizer) sanitizeValue(v reflect.Value) reflect.Value {
    switch v.Kind() {
    case reflect.String:
        return reflect.ValueOf(os.sanitizeString(v.String()))
    case reflect.Slice:
        return os.sanitizeSlice(v)
    case reflect.Map:
        return os.sanitizeMap(v)
    case reflect.Struct:
        return os.sanitizeStruct(v)
    case reflect.Ptr:
        if v.IsNil() {
            return v
        }
        elem := os.sanitizeValue(v.Elem())
        ptr := reflect.New(elem.Type())
        ptr.Elem().Set(elem)
        return ptr
    default:
        return v
    }
}

func (os *OutputSanitizer) sanitizeString(s string) string {
    // Remove potential script injection
    scriptPattern := regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
    s = scriptPattern.ReplaceAllString(s, "")

    // Sanitize HTML while allowing basic formatting
    return os.htmlPolicy.Sanitize(s)
}

func (os *OutputSanitizer) sanitizeStruct(v reflect.Value) reflect.Value {
    t := v.Type()
    result := reflect.New(t).Elem()

    for i := 0; i < v.NumField(); i++ {
        field := v.Field(i)
        fieldType := t.Field(i)

        // Check if field should be sanitized
        if shouldSanitizeField(fieldType) {
            sanitized := os.sanitizeValue(field)
            if result.Field(i).CanSet() {
                result.Field(i).Set(sanitized)
            }
        } else {
            if result.Field(i).CanSet() {
                result.Field(i).Set(field)
            }
        }
    }

    return result
}

func shouldSanitizeField(field reflect.StructField) bool {
    // Check struct tags for sanitization rules
    tag := field.Tag.Get("sanitize")
    return tag != "skip" && (field.Type.Kind() == reflect.String || tag == "html")
}
```

### Step 6: Data Integrity Monitoring
**Timeline: 2-3 days**

Implement data integrity checks and monitoring:

```go
// validation/integrity.go
package validation

type IntegrityChecker struct {
    db      *gorm.DB
    logger  *logging.Logger
    metrics *prometheus.CounterVec
}

func (ic *IntegrityChecker) CheckDataIntegrity(ctx context.Context) error {
    checks := []func(context.Context) error{
        ic.checkUserIntegrity,
        ic.checkMarketIntegrity,
        ic.checkBetIntegrity,
        ic.checkFinancialIntegrity,
    }

    for _, check := range checks {
        if err := check(ctx); err != nil {
            ic.logger.WithFields(map[string]interface{}{
                "error": err.Error(),
                "check": "data_integrity",
            }).Error("Data integrity check failed")

            ic.metrics.WithLabelValues("integrity_check_failed").Inc()
            return err
        }
    }

    return nil
}

func (ic *IntegrityChecker) checkUserIntegrity(ctx context.Context) error {
    // Check for negative credits
    var count int64
    ic.db.Model(&models.User{}).Where("credits < 0").Count(&count)
    if count > 0 {
        return fmt.Errorf("found %d users with negative credits", count)
    }

    // Check for duplicate usernames
    var duplicates []string
    ic.db.Raw(`
        SELECT username
        FROM users
        GROUP BY username
        HAVING COUNT(*) > 1
    `).Scan(&duplicates)

    if len(duplicates) > 0 {
        return fmt.Errorf("found duplicate usernames: %v", duplicates)
    }

    return nil
}

func (ic *IntegrityChecker) checkMarketIntegrity(ctx context.Context) error {
    // Check for markets with invalid end dates
    var count int64
    ic.db.Model(&models.Market{}).Where("end_date < created_at").Count(&count)
    if count > 0 {
        return fmt.Errorf("found %d markets with end date before creation date", count)
    }

    // Check for orphaned markets (creator doesn't exist)
    ic.db.Raw(`
        SELECT COUNT(*)
        FROM markets m
        LEFT JOIN users u ON m.creator_id = u.id
        WHERE u.id IS NULL
    `).Scan(&count)

    if count > 0 {
        return fmt.Errorf("found %d orphaned markets", count)
    }

    return nil
}

func (ic *IntegrityChecker) checkBetIntegrity(ctx context.Context) error {
    // Check for bets with zero or negative amounts
    var count int64
    ic.db.Model(&models.Bet{}).Where("amount <= 0").Count(&count)
    if count > 0 {
        return fmt.Errorf("found %d bets with invalid amounts", count)
    }

    // Check bet totals vs user credits
    var users []models.User
    ic.db.Find(&users)

    for _, user := range users {
        var totalBets float64
        ic.db.Model(&models.Bet{}).Where("user_id = ?", user.ID).Select("COALESCE(SUM(amount), 0)").Scan(&totalBets)

        // User's current credits + total bets should equal their initial credits
        // This is a simplified check - real implementation would be more complex
        if user.Credits < 0 {
            return fmt.Errorf("user %d has negative credits: %f", user.ID, user.Credits)
        }
    }

    return nil
}
```

## Directory Structure
```
validation/
├── engine.go              # Main validation engine
├── rules.go               # Validation rule definitions
├── business_rules.go      # Business logic validation
├── middleware.go          # Validation middleware
├── integrity.go           # Data integrity monitoring
└── custom_validators.go   # Custom validation functions

sanitization/
├── sanitizer.go           # Input sanitization
├── output.go              # Output sanitization
├── rules.go               # Sanitization rule definitions
└── policies.go            # HTML sanitization policies

middleware/
├── validation.go          # Request validation middleware
├── sanitization.go        # Sanitization middleware
└── integrity.go           # Data integrity middleware
```

## Configuration
```yaml
validation:
  strict_mode: true
  max_field_length: 10000
  custom_rules:
    username:
      pattern: "^[a-zA-Z0-9_-]{3,20}$"
      message: "Username must be 3-20 characters, alphanumeric, underscore, or hyphen"

sanitization:
  html_policy: "strict"
  allow_html_tags: ["b", "i", "em", "strong", "u", "p", "br"]
  max_input_size: "1MB"

integrity:
  check_interval: "1h"
  alert_on_failure: true
  auto_fix_minor_issues: false
```

## Testing Strategy
- Unit tests for all validation rules
- Integration tests for business rule validation
- Security tests for XSS and SQL injection prevention
- Performance tests for large datasets
- Data integrity monitoring tests

## Benefits
- Comprehensive input validation
- Protection against XSS and injection attacks
- Data consistency and integrity
- Improved user experience with clear error messages
- Automated data quality monitoring
- Compliance with security standards