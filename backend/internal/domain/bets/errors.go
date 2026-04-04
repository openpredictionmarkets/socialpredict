package bets

import (
	"fmt"
)

// BetError exposes message-based sentinel errors behind a reusable type.
type BetError interface {
	error
	Message() string
}

type staticBetError struct {
	message string
}

func (e staticBetError) Error() string   { return e.message }
func (e staticBetError) Message() string { return e.message }

type betErrorFactory interface {
	New(message string) BetError
}

type staticBetErrorFactory struct{}

func (staticBetErrorFactory) New(message string) BetError {
	return staticBetError{message: message}
}

func newDomainError(message string) error {
	return staticBetErrorFactory{}.New(message)
}

var (
	// ErrInvalidOutcome is returned when the bet outcome is not recognised.
	ErrInvalidOutcome = newDomainError("invalid outcome; expected YES or NO")
	// ErrInvalidAmount is returned when the bet amount is not positive.
	ErrInvalidAmount = newDomainError("bet amount must be greater than zero")
	// ErrMarketClosed is returned when a bet is attempted on a closed or resolved market.
	ErrMarketClosed = newDomainError("market is closed or resolved")
	// ErrInsufficientBalance indicates the user would exceed the maximum allowed debt.
	ErrInsufficientBalance = newDomainError("insufficient balance for requested bet")
	// ErrNoPosition indicates the user has no position to sell.
	ErrNoPosition = newDomainError("no position found for the given market and outcome")
	// ErrInsufficientShares indicates the user cannot sell the requested credits.
	ErrInsufficientShares = newDomainError("not enough shares to satisfy requested sale")
)

// ErrDustCapExceeded is returned when a sell transaction would generate dust above the configured cap.
type ErrDustCapExceeded struct {
	Cap       int64
	Requested int64
	formatter dustCapFormatter
}

func (e ErrDustCapExceeded) Error() string {
	return e.resolveFormatter().Format(e.Cap, e.Requested)
}

func (e ErrDustCapExceeded) Is(target error) bool {
	switch target.(type) {
	case ErrDustCapExceeded, *ErrDustCapExceeded:
		return true
	default:
		return false
	}
}

func newDustCapExceeded(cap, requested int64) ErrDustCapExceeded {
	return ErrDustCapExceeded{
		Cap:       cap,
		Requested: requested,
		formatter: defaultDustCapFormatter{},
	}
}

type dustCapFormatter interface {
	Format(cap int64, requested int64) string
}

type defaultDustCapFormatter struct{}

func (defaultDustCapFormatter) Format(cap int64, requested int64) string {
	return fmt.Sprintf("dust cap exceeded: would generate %d dust points (cap: %d)", requested, cap)
}

func (e ErrDustCapExceeded) resolveFormatter() dustCapFormatter {
	if e.formatter != nil {
		return e.formatter
	}
	return defaultDustCapFormatter{}
}
