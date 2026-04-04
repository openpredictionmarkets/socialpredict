package markets

import "errors"

// MarketError exposes message-based sentinel errors behind a reusable type.
type MarketError interface {
	error
	Message() string
}

type staticMarketError struct {
	message string
}

func (e staticMarketError) Error() string   { return e.message }
func (e staticMarketError) Message() string { return e.message }

type marketErrorFactory interface {
	New(message string) MarketError
}

type staticMarketErrorFactory struct{}

func (staticMarketErrorFactory) New(message string) MarketError {
	return staticMarketError{message: message}
}

func newDomainError(message string) error {
	return staticMarketErrorFactory{}.New(message)
}

var (
	// ErrMarketNotFound indicates that the requested market does not exist.
	ErrMarketNotFound = newDomainError("market not found")
	// ErrInvalidQuestionTitle indicates that the market question title is invalid.
	ErrInvalidQuestionTitle = newDomainError("invalid question title")
	// ErrInvalidQuestionLength indicates that the market question title is blank or too long.
	ErrInvalidQuestionLength = newDomainError("question title exceeds maximum length or is blank")
	// ErrInvalidDescriptionLength indicates that the market description is too long.
	ErrInvalidDescriptionLength = newDomainError("question description exceeds maximum length")
	// ErrInvalidLabel indicates that one or more custom labels are invalid.
	ErrInvalidLabel = newDomainError("invalid label")
	// ErrInvalidResolutionTime indicates that the supplied resolution time is invalid.
	ErrInvalidResolutionTime = newDomainError("invalid market resolution time")
	// ErrUserNotFound indicates that the referenced creator user does not exist.
	ErrUserNotFound = newDomainError("creator user not found")
	// ErrInsufficientBalance indicates that the actor does not have enough balance.
	ErrInsufficientBalance = newDomainError("insufficient balance")
	// ErrUnauthorized indicates that the actor is not allowed to perform the action.
	ErrUnauthorized = newDomainError("unauthorized")
	// ErrInvalidInput indicates that one or more request inputs are invalid.
	ErrInvalidInput = newDomainError("invalid input")
	// ErrInvalidState indicates that the market state does not allow the requested action.
	ErrInvalidState = newDomainError("invalid state")
)

// IsMarketNotFound reports whether err represents a missing market.
func IsMarketNotFound(err error) bool {
	return errors.Is(err, ErrMarketNotFound)
}

// IsInvalidInput reports whether err represents invalid caller input.
func IsInvalidInput(err error) bool {
	return errors.Is(err, ErrInvalidInput)
}

// IsUnauthorized reports whether err represents an authorization failure.
func IsUnauthorized(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}

// IsInvalidState reports whether err represents an invalid market state transition.
func IsInvalidState(err error) bool {
	return errors.Is(err, ErrInvalidState)
}
