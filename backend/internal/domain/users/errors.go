package users

import "errors"

var (
	ErrUserNotFound           = errors.New("user not found")
	ErrUserAlreadyExists      = errors.New("user already exists")
	ErrInvalidCredentials     = errors.New("invalid credentials")
	ErrInsufficientBalance    = errors.New("insufficient balance")
	ErrInvalidUserData        = errors.New("invalid user data")
	ErrUnauthorized           = errors.New("unauthorized")
	ErrInvalidTransactionType = errors.New("invalid transaction type")
	ErrInvalidAccountID  = errors.New("account id must be positive")
	ErrInvalidUserID     = errors.New("account user id must exist")
)
