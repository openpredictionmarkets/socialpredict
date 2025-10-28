package usershandlers

import (
	"errors"
	"net/http"
	"strings"

	dusers "socialpredict/internal/domain/users"
)

func writeProfileError(w http.ResponseWriter, err error, field string) {
	switch {
	case errors.Is(err, dusers.ErrUserNotFound):
		http.Error(w, "User not found", http.StatusNotFound)
	case errors.Is(err, dusers.ErrInvalidUserData):
		http.Error(w, "Invalid user data", http.StatusBadRequest)
	case errors.Is(err, dusers.ErrInvalidCredentials):
		http.Error(w, "Current password is incorrect", http.StatusUnauthorized)
	default:
		message := err.Error()
		if isValidationError(message) {
			http.Error(w, message, http.StatusBadRequest)
			return
		}
		http.Error(w, "Failed to update "+field+": "+message, http.StatusInternalServerError)
	}
}

func isValidationError(message string) bool {
	lower := strings.ToLower(message)
	return strings.Contains(lower, "invalid") ||
		strings.Contains(lower, "exceeds") ||
		strings.Contains(lower, "must") ||
		strings.Contains(lower, "cannot") ||
		strings.Contains(lower, "required")
}
