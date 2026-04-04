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

func newDomainError(message string) error {
	return staticUserErrorFactory{}.New(message)
}

var (
	// ErrUserNotFound indicates that the requested user does not exist.
	ErrUserNotFound = newDomainError("user not found")
	// ErrUserAlreadyExists indicates that a create request conflicts with an existing user.
	ErrUserAlreadyExists = newDomainError("user already exists")
	// ErrInvalidCredentials indicates that password verification failed.
	ErrInvalidCredentials = newDomainError("invalid credentials")
	// ErrInsufficientBalance indicates that the user would exceed allowed debt.
	ErrInsufficientBalance = newDomainError("insufficient balance")
	// ErrInvalidUserData indicates that caller-supplied user data is invalid.
	ErrInvalidUserData = newDomainError("invalid user data")
	// ErrUnauthorized indicates that the caller is not allowed to perform the action.
	ErrUnauthorized = newDomainError("unauthorized")
	// ErrInvalidTransactionType indicates that no balance rule exists for the supplied transaction type.
	ErrInvalidTransactionType = newDomainError("invalid transaction type")
)
