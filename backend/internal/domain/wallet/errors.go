package wallet

import "errors"

var (
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrInvalidAmount       = errors.New("amount must be positive")
	ErrAccountNotFound     = errors.New("account not found")
	ErrInvalidTransaction  = errors.New("invalid transaction type")
)
