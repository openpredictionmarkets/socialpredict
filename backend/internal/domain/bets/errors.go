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

func newDomainError(message string) BetError {
	return staticBetErrorFactory{}.New(message)
}

var (
	// ErrInvalidOutcome is returned when the bet outcome is not recognised.
	ErrInvalidOutcome BetError = newDomainError("invalid outcome; expected YES or NO")
	// ErrInvalidAmount is returned when the bet amount is not positive.
	ErrInvalidAmount BetError = newDomainError("bet amount must be greater than zero")
	// ErrMarketClosed is returned when a bet is attempted on a closed or resolved market.
	ErrMarketClosed BetError = newDomainError("market is closed or resolved")
	// ErrInsufficientBalance indicates the user would exceed the maximum allowed debt.
	ErrInsufficientBalance BetError = newDomainError("insufficient balance for requested bet")
	// ErrNoPosition indicates the user has no position to sell.
	ErrNoPosition BetError = newDomainError("no position found for the given market and outcome")
	// ErrInsufficientShares indicates the user cannot sell the requested credits.
	ErrInsufficientShares BetError = newDomainError("not enough shares to satisfy requested sale")
)

// ErrDustCapExceeded is returned when a sell transaction would generate dust above the configured cap.
type ErrDustCapExceeded struct {
	Cap       int64
	Requested int64
}

func (e ErrDustCapExceeded) Error() string {
	return dustCapExceededMessage(e.Cap, e.Requested)
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
	}
}

func dustCapExceededMessage(cap, requested int64) string {
	return fmt.Sprintf("dust cap exceeded: would generate %d dust points (cap: %d)", requested, cap)
}
