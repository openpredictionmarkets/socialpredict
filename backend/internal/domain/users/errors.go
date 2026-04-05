package users

// UserError exposes reusable message-based sentinel errors for the users domain.
type UserError interface {
	error
	Message() string
}

type staticUserError struct {
	message string
}

func (e staticUserError) Error() string   { return e.message }
func (e staticUserError) Message() string { return e.message }

type userErrorFactory interface {
	New(message string) UserError
}

type staticUserErrorFactory struct{}

func (staticUserErrorFactory) New(message string) UserError {
	return staticUserError{message: message}
}

var defaultUserErrorFactory userErrorFactory = staticUserErrorFactory{}

func newDomainError(message string) UserError {
	return defaultUserErrorFactory.New(message)
}

var (
	// ErrUserNotFound indicates that the requested user does not exist.
	ErrUserNotFound UserError = newDomainError("user not found")
	// ErrUserAlreadyExists indicates that a create request conflicts with an existing user.
	ErrUserAlreadyExists UserError = newDomainError("user already exists")
	// ErrInvalidCredentials indicates that password verification failed.
	ErrInvalidCredentials UserError = newDomainError("invalid credentials")
	// ErrInsufficientBalance indicates that the user would exceed allowed debt.
	ErrInsufficientBalance UserError = newDomainError("insufficient balance")
	// ErrInvalidUserData indicates that caller-supplied user data is invalid.
	ErrInvalidUserData UserError = newDomainError("invalid user data")
	// ErrUnauthorized indicates that the caller is not allowed to perform the action.
	ErrUnauthorized UserError = newDomainError("unauthorized")
	// ErrInvalidTransactionType indicates that no balance rule exists for the supplied transaction type.
	ErrInvalidTransactionType UserError = newDomainError("invalid transaction type")
)
