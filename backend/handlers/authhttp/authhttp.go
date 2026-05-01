package authhttp

import (
	"net/http"

	"socialpredict/handlers"
	dusers "socialpredict/internal/domain/users"
	authsvc "socialpredict/internal/service/auth"
)

type Failure struct {
	StatusCode int
	Reason     handlers.FailureReason
	Message    string
}

func (f *Failure) Error() string {
	if f == nil {
		return ""
	}
	return f.Message
}

func (f *Failure) Write(w http.ResponseWriter) error {
	if f == nil {
		return handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
	}
	return handlers.WriteFailure(w, f.StatusCode, f.Reason)
}

// CurrentUser resolves the protected-route user and enforces the password
// change gate before mapping auth outcomes to HTTP boundary failures.
func CurrentUser(r *http.Request, svc dusers.ServiceInterface) (*dusers.User, *Failure) {
	user, err := authsvc.ValidateUserAndEnforcePasswordChangeGetUser(r, svc)
	if err != nil {
		return nil, FailureFromAuthError(err)
	}
	return user, nil
}

// TokenUser resolves the authenticated user without the password-change gate.
// It is intentionally used by /v0/changepassword so users with
// mustChangePassword=true can complete the required password change.
func TokenUser(r *http.Request, svc dusers.ServiceInterface) (*dusers.User, *Failure) {
	user, err := authsvc.ValidateTokenAndGetUser(r, svc)
	if err != nil {
		return nil, FailureFromAuthError(err)
	}
	return user, nil
}

func FailureFromAuthError(err *authsvc.AuthError) *Failure {
	return &Failure{
		StatusCode: StatusCode(err),
		Reason:     FailureReason(err),
		Message:    message(err),
	}
}

func StatusCode(err *authsvc.AuthError) int {
	if err == nil {
		return http.StatusInternalServerError
	}

	switch err.Kind {
	case authsvc.ErrorKindMissingToken, authsvc.ErrorKindInvalidToken:
		return http.StatusUnauthorized
	case authsvc.ErrorKindUserNotFound:
		return http.StatusNotFound
	case authsvc.ErrorKindPasswordChangeRequired, authsvc.ErrorKindAdminRequired:
		return http.StatusForbidden
	case authsvc.ErrorKindUserLoadFailed, authsvc.ErrorKindServiceUnavailable:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

func FailureReason(err *authsvc.AuthError) handlers.FailureReason {
	if err == nil {
		return handlers.ReasonInternalError
	}

	switch err.Kind {
	case authsvc.ErrorKindMissingToken, authsvc.ErrorKindInvalidToken:
		return handlers.ReasonInvalidToken
	case authsvc.ErrorKindUserNotFound:
		return handlers.ReasonUserNotFound
	case authsvc.ErrorKindPasswordChangeRequired:
		return handlers.ReasonPasswordChangeRequired
	case authsvc.ErrorKindAdminRequired:
		return handlers.ReasonAuthorizationDenied
	case authsvc.ErrorKindUserLoadFailed, authsvc.ErrorKindServiceUnavailable:
		return handlers.ReasonInternalError
	default:
		return handlers.ReasonInternalError
	}
}

func WriteFailure(w http.ResponseWriter, err *authsvc.AuthError) error {
	return handlers.WriteFailure(w, StatusCode(err), FailureReason(err))
}

func message(err *authsvc.AuthError) string {
	if err == nil || err.Message == "" {
		return "authentication failed"
	}
	return err.Message
}
