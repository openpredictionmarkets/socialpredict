package betshandlers

import "fmt"

// ErrDustCapExceeded is returned when a sell transaction would generate dust exceeding the configured cap
type ErrDustCapExceeded struct {
	Cap       int64 // Maximum allowed dust per sale
	Requested int64 // Amount of dust that would be generated
}

// Error implements the error interface
func (e ErrDustCapExceeded) Error() string {
	return fmt.Sprintf("dust cap exceeded: would generate %d dust points (cap: %d)", e.Requested, e.Cap)
}

// IsBusinessRuleError identifies this as a business rule violation (HTTP 422)
func (e ErrDustCapExceeded) IsBusinessRuleError() bool {
	return true
}
