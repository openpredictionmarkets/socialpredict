package http

import (
	"encoding/json"
	"net/http"

	"socialpredict/handlers"
	"socialpredict/handlers/authhttp"
	"socialpredict/handlers/cms/socialshare"
	authsvc "socialpredict/internal/service/auth"
	"socialpredict/models"
)

type Handler struct {
	svc  *socialshare.Service
	auth authsvc.Authenticator
}

func NewHandler(svc *socialshare.Service, auth authsvc.Authenticator) *Handler {
	return &Handler{svc: svc, auth: auth}
}

func (h *Handler) PublicGet(w http.ResponseWriter, r *http.Request) {
	item, err := h.svc.GetSettings()
	if err != nil {
		_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
		return
	}
	_ = handlers.WriteResult(w, http.StatusOK, responseFromSettings(item))
}

type updateReq struct {
	SiteName           string `json:"siteName"`
	DefaultDescription string `json:"defaultDescription"`
	DefaultImageURL    string `json:"defaultImageUrl"`
	ImageAlt           string `json:"imageAlt"`
	Version            uint   `json:"version"`
}

func (h *Handler) AdminUpdate(w http.ResponseWriter, r *http.Request) {
	if h.auth == nil {
		_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
		return
	}
	admin, authErr := h.auth.RequireAdmin(r)
	if authErr != nil {
		_ = authhttp.WriteFailure(w, authErr)
		return
	}

	var in updateReq
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
		return
	}

	item, err := h.svc.UpdateSettings(socialshare.UpdateInput{
		SiteName:           in.SiteName,
		DefaultDescription: in.DefaultDescription,
		DefaultImageURL:    in.DefaultImageURL,
		ImageAlt:           in.ImageAlt,
		Version:            in.Version,
		UpdatedBy:          admin.Username,
	})
	if err != nil {
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonValidationFailed)
		return
	}

	_ = handlers.WriteResult(w, http.StatusOK, responseFromSettings(item))
}

func responseFromSettings(item *models.SocialShareSettings) map[string]interface{} {
	return map[string]interface{}{
		"siteName":           item.SiteName,
		"defaultDescription": item.DefaultDescription,
		"defaultImageUrl":    item.DefaultImageURL,
		"imageAlt":           item.ImageAlt,
		"version":            item.Version,
		"updatedAt":          item.UpdatedAt,
	}
}
