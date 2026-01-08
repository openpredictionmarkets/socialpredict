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

	user, httpErr := auth.RequireAdmin(r)
	if httpErr != nil {
		return errors.New(httpErr.Message)
	}

	// Extra guard: RequireAdmin already checks admin, but ensure status handling.
	if user == nil {
		return errors.New("unauthorized")
	}

	return nil
}
