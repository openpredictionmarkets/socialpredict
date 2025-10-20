package users

import "errors"

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrInvalidUserData     = errors.New("invalid user data")
	ErrUnauthorized        = errors.New("unauthorized")
)
