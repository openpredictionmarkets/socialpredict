package http

import (
	"encoding/json"
	"io"
	"net/http"

	"socialpredict/handlers"
	"socialpredict/handlers/authhttp"
	"socialpredict/handlers/cms/socialshare"
	dusers "socialpredict/internal/domain/users"
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

func (h *Handler) PublicImage(w http.ResponseWriter, r *http.Request) {
	image, err := h.svc.GetImage()
	if err != nil {
		_ = handlers.WriteFailure(w, http.StatusNotFound, handlers.ReasonNotFound)
		return
	}
	// This image is a public share-card asset, so local previews and external
	// crawlers must be able to render it outside the backend origin.
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cross-Origin-Resource-Policy", "cross-origin")
	w.Header().Set("Content-Type", image.ContentType)
	w.Header().Set("Cache-Control", "public, max-age=300")
	w.Header().Set("Content-Length", stringInt64(image.SizeBytes))
	_, _ = w.Write(image.Data)
}

type updateReq struct {
	SiteName           string `json:"siteName"`
	DefaultDescription string `json:"defaultDescription"`
	DefaultImageURL    string `json:"defaultImageUrl"`
	ImageAlt           string `json:"imageAlt"`
	Version            uint   `json:"version"`
}

func (h *Handler) AdminUpdate(w http.ResponseWriter, r *http.Request) {
	admin, ok := h.requireAdmin(w, r)
	if !ok {
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

func (h *Handler) AdminUploadImage(w http.ResponseWriter, r *http.Request) {
	admin, ok := h.requireAdmin(w, r)
	if !ok {
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, socialshare.MaxUploadedImageBytes+1024)
	if err := r.ParseMultipartForm(socialshare.MaxUploadedImageBytes + 1024); err != nil {
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
		return
	}
	file, header, err := r.FormFile("image")
	if err != nil {
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, socialshare.MaxUploadedImageBytes+1))
	if err != nil || int64(len(data)) > socialshare.MaxUploadedImageBytes {
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonValidationFailed)
		return
	}

	item, err := h.svc.UploadImage(socialshare.UploadImageInput{
		FileName:  header.Filename,
		Data:      data,
		ImageAlt:  r.FormValue("imageAlt"),
		UpdatedBy: admin.Username,
	})
	if err != nil {
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonValidationFailed)
		return
	}

	_ = handlers.WriteResult(w, http.StatusOK, responseFromSettings(item))
}

func (h *Handler) requireAdmin(w http.ResponseWriter, r *http.Request) (*dusers.User, bool) {
	if h.auth == nil {
		_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
		return nil, false
	}
	admin, authErr := h.auth.RequireAdmin(r)
	if authErr != nil {
		_ = authhttp.WriteFailure(w, authErr)
		return nil, false
	}
	return admin, true
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

func stringInt64(value int64) string {
	if value == 0 {
		return "0"
	}
	buf := [20]byte{}
	i := len(buf)
	for value > 0 {
		i--
		buf[i] = byte('0' + value%10)
		value /= 10
	}
	return string(buf[i:])
}
