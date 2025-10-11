package http

import (
	"context"
	"encoding/json"
	"net/http"
	"socialpredict/handlers/cms/homepage"
	"socialpredict/middleware"
	"socialpredict/util"
)

type Handler struct {
	svc *homepage.Service
}

func NewHandler(svc *homepage.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) PublicGet(w http.ResponseWriter, r *http.Request) {
	item, err := h.svc.GetHome()
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
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
	db := util.GetDB()
	if err := middleware.ValidateAdminToken(r, db); err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Get username from context/token
	user, httpErr := middleware.ValidateTokenAndGetUser(r, db)
	if httpErr != nil {
		http.Error(w, httpErr.Message, httpErr.StatusCode)
		return
	}

	var in updateReq
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	item, err := h.svc.UpdateHome(homepage.UpdateInput{
		Title:     in.Title,
		Format:    in.Format,
		Markdown:  in.Markdown,
		HTML:      in.HTML,
		Version:   in.Version,
		UpdatedBy: user.Username,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"title":   item.Title,
		"format":  item.Format,
		"html":    item.HTML,
		"version": item.Version,
	})
}

// RequireAdmin middleware wrapper that can be used in routes
func RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db := util.GetDB()
		if err := middleware.ValidateAdminToken(r, db); err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
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
