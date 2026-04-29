package marketshandlers

import (
	"errors"
	"net/http"

	"socialpredict/handlers"
	"socialpredict/handlers/authhttp"
	dmarkets "socialpredict/internal/domain/markets"
	authsvc "socialpredict/internal/service/auth"
)

func writeMethodNotAllowed(w http.ResponseWriter) {
	_ = handlers.WriteFailure(w, http.StatusMethodNotAllowed, handlers.ReasonMethodNotAllowed)
}

func writeInvalidRequest(w http.ResponseWriter) {
	_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
}

func writeInternalError(w http.ResponseWriter) {
	_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
}

func writeAuthError(w http.ResponseWriter, authErr *authsvc.AuthError) {
	if authErr == nil {
		writeInternalError(w)
		return
	}

	_ = authhttp.WriteFailure(w, authErr)
}

func writeCreateError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, dmarkets.ErrUserNotFound):
		_ = handlers.WriteFailure(w, http.StatusNotFound, handlers.ReasonUserNotFound)
	case errors.Is(err, dmarkets.ErrInsufficientBalance):
		_ = handlers.WriteFailure(w, http.StatusUnprocessableEntity, handlers.ReasonInsufficientBalance)
	case isValidationError(err):
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonValidationFailed)
	default:
		writeInternalError(w)
	}
}

func writeListError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, dmarkets.ErrInvalidInput):
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonValidationFailed)
	default:
		writeInternalError(w)
	}
}

func writeDetailsError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, dmarkets.ErrMarketNotFound):
		_ = handlers.WriteFailure(w, http.StatusNotFound, handlers.ReasonMarketNotFound)
	case errors.Is(err, dmarkets.ErrInvalidInput):
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonValidationFailed)
	default:
		writeInternalError(w)
	}
}

func writeResolveErrorResponse(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, dmarkets.ErrMarketNotFound):
		_ = handlers.WriteFailure(w, http.StatusNotFound, handlers.ReasonMarketNotFound)
	case errors.Is(err, dmarkets.ErrUserNotFound):
		_ = handlers.WriteFailure(w, http.StatusNotFound, handlers.ReasonUserNotFound)
	case errors.Is(err, dmarkets.ErrUnauthorized):
		_ = handlers.WriteFailure(w, http.StatusForbidden, handlers.ReasonAuthorizationDenied)
	case errors.Is(err, dmarkets.ErrInvalidState):
		_ = handlers.WriteFailure(w, http.StatusConflict, handlers.ReasonMarketClosed)
	case isValidationError(err):
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonValidationFailed)
	default:
		writeInternalError(w)
	}
}

func writeLeaderboardError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, dmarkets.ErrMarketNotFound):
		_ = handlers.WriteFailure(w, http.StatusNotFound, handlers.ReasonMarketNotFound)
	case errors.Is(err, dmarkets.ErrInvalidInput):
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonValidationFailed)
	default:
		writeInternalError(w)
	}
}

func writeProjectionError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, dmarkets.ErrMarketNotFound):
		_ = handlers.WriteFailure(w, http.StatusNotFound, handlers.ReasonMarketNotFound)
	case errors.Is(err, dmarkets.ErrInvalidState):
		_ = handlers.WriteFailure(w, http.StatusConflict, handlers.ReasonMarketClosed)
	case isValidationError(err):
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonValidationFailed)
	default:
		writeInternalError(w)
	}
}

func isValidationError(err error) bool {
	return errors.Is(err, dmarkets.ErrInvalidInput) ||
		errors.Is(err, dmarkets.ErrInvalidQuestionTitle) ||
		errors.Is(err, dmarkets.ErrInvalidQuestionLength) ||
		errors.Is(err, dmarkets.ErrInvalidDescriptionLength) ||
		errors.Is(err, dmarkets.ErrInvalidLabel) ||
		errors.Is(err, dmarkets.ErrInvalidResolutionTime)
}
