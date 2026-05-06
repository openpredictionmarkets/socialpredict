package auth

import (
	"errors"
	"net/http"
)

// ValidateAdminToken uses the provided authenticator to ensure the request is
// made by an admin user. Prefer calling auth.RequireAdmin directly; this helper
// exists for backwards compatibility with legacy call sites.
func ValidateAdminToken(r *http.Request, auth Authenticator) error {
	if auth == nil {
		return errors.New("authenticator is required")
	}

	user, authErr := auth.RequireAdmin(r)
	if authErr != nil {
		return errors.New(authErr.Message)
	}

	// Extra guard: RequireAdmin already checks admin, but ensure status handling.
	if user == nil {
		return errors.New("unauthorized")
	}

	return nil
}
