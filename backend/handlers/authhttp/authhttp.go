package authhttp

import (
	"net/http"

	"socialpredict/handlers"
	authsvc "socialpredict/internal/service/auth"
)

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
