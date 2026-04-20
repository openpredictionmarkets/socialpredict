package http

import (
	"context"
	"encoding/json"
	"net/http"
	"socialpredict/handlers"
	"socialpredict/handlers/cms/homepage"
	authsvc "socialpredict/internal/service/auth"
)

const (
	reasonHomepageAuthUnavailable handlers.FailureReason = "AUTH_SERVICE_UNAVAILABLE"
	reasonHomepageAdminRequired   handlers.FailureReason = "ADMIN_REQUIRED"
	reasonHomepageUpdateFailed    handlers.FailureReason = "HOMEPAGE_UPDATE_FAILED"
	reasonHomepageNotFound        handlers.FailureReason = "HOME_CONTENT_NOT_FOUND"
)

type Handler struct {
	svc  *homepage.Service
	auth authsvc.Authenticator
}

func NewHandler(svc *homepage.Service, auth authsvc.Authenticator) *Handler {
	return &Handler{svc: svc, auth: auth}
}

func (h *Handler) PublicGet(w http.ResponseWriter, r *http.Request) {
	item, err := h.svc.GetHome()
	if err != nil {
		_ = handlers.WriteFailure(w, http.StatusNotFound, reasonHomepageNotFound)
		return
	}

	_ = handlers.WriteResult(w, http.StatusOK, map[string]interface{}{
		"title":     item.Title,
		"format":    item.Format,
		"html":      item.HTML,
		"markdown":  item.Markdown, // optional to expose
		"version":   item.Version,
		"updatedAt": item.UpdatedAt,
	})
}

type updateReq struct {
	Title    string `json:"title"`
	Format   string `json:"format"`   // "markdown" | "html"
	Markdown string `json:"markdown"` // when format=markdown
	HTML     string `json:"html"`     // when format=html
	Version  uint   `json:"version"`
}

func (h *Handler) AdminUpdate(w http.ResponseWriter, r *http.Request) {
	// Validate admin access
	if h.auth == nil {
		_ = handlers.WriteFailure(w, http.StatusInternalServerError, reasonHomepageAuthUnavailable)
		return
	}
	admin, httpErr := h.auth.RequireAdmin(r)
	if httpErr != nil {
		_ = handlers.WriteFailure(w, httpErr.StatusCode, reasonFromAdminAuthError(httpErr))
		return
	}

	var in updateReq
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
		return
	}

	item, err := h.svc.UpdateHome(homepage.UpdateInput{
		Title:     in.Title,
		Format:    in.Format,
		Markdown:  in.Markdown,
		HTML:      in.HTML,
		Version:   in.Version,
		UpdatedBy: admin.Username,
	})
	if err != nil {
		_ = handlers.WriteFailure(w, http.StatusBadRequest, reasonHomepageUpdateFailed)
		return
	}

	_ = handlers.WriteResult(w, http.StatusOK, map[string]interface{}{
		"title":   item.Title,
		"format":  item.Format,
		"html":    item.HTML,
		"version": item.Version,
	})
}

// RequireAdmin middleware wrapper that can be used in routes when an authenticator is available.
func RequireAdmin(auth authsvc.Authenticator, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if auth == nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, reasonHomepageAuthUnavailable)
			return
		}
		if _, httpErr := auth.RequireAdmin(r); httpErr != nil {
			_ = handlers.WriteFailure(w, httpErr.StatusCode, reasonFromAdminAuthError(httpErr))
			return
		}
		next.ServeHTTP(w, r)
	}
}

// UsernameFromContext extracts username from request context (helper function)
func UsernameFromContext(ctx context.Context) string {
	// This is a placeholder - in practice you might store username in context
	// during authentication middleware
	return "admin" // fallback for now
}

func reasonFromAdminAuthError(err *authsvc.HTTPError) handlers.FailureReason {
	if err == nil {
		return handlers.ReasonInternalError
	}

	switch err.Message {
	case "Authorization header is required", "Invalid token":
		return handlers.ReasonInvalidToken
	case "admin privileges required":
		return reasonHomepageAdminRequired
	case "User not found":
		return handlers.ReasonUserNotFound
	default:
		if err.StatusCode >= http.StatusInternalServerError {
			return handlers.ReasonInternalError
		}
		return handlers.ReasonInvalidToken
	}
}
