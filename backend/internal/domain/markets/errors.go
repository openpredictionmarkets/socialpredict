package markets

import "errors"

var (
	ErrMarketNotFound           = errors.New("market not found")
	ErrInvalidQuestionTitle     = errors.New("invalid question title")
	ErrInvalidQuestionLength    = errors.New("question title exceeds maximum length or is blank")
	ErrInvalidDescriptionLength = errors.New("question description exceeds maximum length")
	ErrInvalidLabel             = errors.New("invalid label")
	ErrInvalidResolutionTime    = errors.New("invalid market resolution time")
	ErrUserNotFound             = errors.New("creator user not found")
	ErrInsufficientBalance      = errors.New("insufficient balance")
	ErrUnauthorized             = errors.New("unauthorized")
	ErrInvalidInput             = errors.New("invalid input")
	ErrInvalidState             = errors.New("invalid state")
)
