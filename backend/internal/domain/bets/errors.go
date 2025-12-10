package bets

import (
	"errors"
	"fmt"
)

var (
	// ErrInvalidOutcome is returned when the bet outcome is not recognised.
	ErrInvalidOutcome = errors.New("invalid outcome; expected YES or NO")
	// ErrInvalidAmount is returned when the bet amount is not positive.
	ErrInvalidAmount = errors.New("bet amount must be greater than zero")
	// ErrMarketClosed is returned when a bet is attempted on a closed or resolved market.
	ErrMarketClosed = errors.New("market is closed or resolved")
	// ErrInsufficientBalance indicates the user would exceed the maximum allowed debt.
	ErrInsufficientBalance = errors.New("insufficient balance for requested bet")
	// ErrNoPosition indicates the user has no position to sell.
	ErrNoPosition = errors.New("no position found for the given market and outcome")
	// ErrInsufficientShares indicates the user cannot sell the requested credits.
	ErrInsufficientShares = errors.New("not enough shares to satisfy requested sale")
)

// ErrDustCapExceeded is returned when a sell transaction would generate dust above the configured cap.
type ErrDustCapExceeded struct {
	Cap       int64
	Requested int64
}

func (e ErrDustCapExceeded) Error() string {
	return fmt.Sprintf("dust cap exceeded: would generate %d dust points (cap: %d)", e.Requested, e.Cap)
}
