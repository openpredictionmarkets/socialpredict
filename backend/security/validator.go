package security

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator wraps the go-playground validator with custom validation rules
type Validator struct {
	validate *validator.Validate
}

// NewValidator creates a new validator instance with custom rules
func NewValidator() *Validator {
	validate := validator.New()

	// Register custom validation functions
	validate.RegisterValidation("username", validateUsername)
	validate.RegisterValidation("strong_password", validateStrongPassword)
	validate.RegisterValidation("safe_string", validateSafeString)
	validate.RegisterValidation("market_outcome", validateMarketOutcome)
	validate.RegisterValidation("positive_amount", validatePositiveAmount)
	validate.RegisterValidation("market_id", validateMarketID)

	return &Validator{
		validate: validate,
	}
}

// ValidateStruct validates a struct using the configured validator
func (v *Validator) ValidateStruct(s interface{}) error {
	err := v.validate.Struct(s)
	if err != nil {
		return formatValidationErrors(err)
	}
	return nil
}

// formatValidationErrors converts validator errors to user-friendly messages
func formatValidationErrors(err error) error {
	var messages []string

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validationErrors {
			messages = append(messages, getFieldErrorMessage(fieldError))
		}
	}

	return fmt.Errorf("validation failed: %s", strings.Join(messages, "; "))
}

// getFieldErrorMessage returns a user-friendly error message for a field validation error
func getFieldErrorMessage(fe validator.FieldError) string {
	field := strings.ToLower(fe.Field())

	if msg := messageForTag(fe.Tag(), field, fe.Param()); msg != "" {
		return msg
	}

	return fmt.Sprintf("%s is invalid", field)
}

func messageForTag(tag string, field string, param string) string {
	lookup := map[string]string{
		"required":        fmt.Sprintf("%s is required", field),
		"min":             fmt.Sprintf("%s must be at least %s characters", field, param),
		"max":             fmt.Sprintf("%s cannot exceed %s characters", field, param),
		"username":        "username must only contain lowercase letters and numbers",
		"strong_password": "password must be at least 8 characters with uppercase, lowercase, and digit",
		"safe_string":     fmt.Sprintf("%s contains potentially dangerous content", field),
		"market_outcome":  "outcome must be either 'YES' or 'NO'",
		"positive_amount": "amount must be a positive number",
		"market_id":       "invalid market ID format",
	}

	return lookup[tag]
}

// Custom validation functions

// validateUsername checks if the username contains only lowercase letters and numbers
func validateUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()
	if len(username) < 3 || len(username) > 30 {
		return false
	}

	for _, char := range username {
		if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9')) {
			return false
		}
	}
	return true
}

// validateStrongPassword checks password strength requirements
func validateStrongPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	if !isPasswordLengthValid(password) {
		return false
	}

	requirements := passwordRequirements(password)
	return requirements.hasUpper && requirements.hasLower && requirements.hasDigit
}

func isPasswordLengthValid(password string) bool {
	return len(password) >= 8 && len(password) <= 128
}

type passwordFlags struct {
	hasUpper bool
	hasLower bool
	hasDigit bool
}

func passwordRequirements(password string) passwordFlags {
	var flags passwordFlags
	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			flags.hasUpper = true
		case char >= 'a' && char <= 'z':
			flags.hasLower = true
		case char >= '0' && char <= '9':
			flags.hasDigit = true
		}
	}
	return flags
}

// validateSafeString checks for potentially dangerous content
func validateSafeString(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return !containsSuspiciousPatterns(value)
}

// validateMarketOutcome checks if the outcome is valid for prediction markets
func validateMarketOutcome(fl validator.FieldLevel) bool {
	outcome := strings.ToUpper(fl.Field().String())
	return outcome == "YES" || outcome == "NO"
}

// validatePositiveAmount checks if the amount is positive
func validatePositiveAmount(fl validator.FieldLevel) bool {
	field := fl.Field()

	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int() > 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return field.Uint() > 0
	case reflect.Float32, reflect.Float64:
		return field.Float() > 0
	case reflect.String:
		// Try to parse as number
		if value, err := strconv.ParseFloat(field.String(), 64); err == nil {
			return value > 0
		}
		return false
	default:
		return false
	}
}

// validateMarketID checks if the market ID is in a valid format (assuming UUID or positive integer)
func validateMarketID(fl validator.FieldLevel) bool {
	field := fl.Field()

	switch field.Kind() {
	case reflect.String:
		// Check if it's a valid UUID format or positive integer string
		value := field.String()
		if value == "" {
			return false
		}

		// Try parsing as integer first
		if id, err := strconv.ParseInt(value, 10, 64); err == nil {
			return id > 0
		}

		// Could add UUID validation here if using UUIDs
		return len(value) > 0
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int() > 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return field.Uint() > 0
	default:
		return false
	}
}

// ValidationRules contains validation tag strings for common fields
var ValidationRules = struct {
	Username      string
	Password      string
	DisplayName   string
	Description   string
	MarketTitle   string
	PersonalLink  string
	Emoji         string
	MarketOutcome string
	BetAmount     string
	MarketID      string
}{
	Username:      "required,min=3,max=30,username",
	Password:      "required,strong_password",
	DisplayName:   "required,min=1,max=50,safe_string",
	Description:   "max=2000,safe_string",
	MarketTitle:   "required,min=1,max=160,safe_string",
	PersonalLink:  "max=200,url",
	Emoji:         "max=20",
	MarketOutcome: "required,market_outcome",
	BetAmount:     "required,positive_amount",
	MarketID:      "required,market_id",
}
