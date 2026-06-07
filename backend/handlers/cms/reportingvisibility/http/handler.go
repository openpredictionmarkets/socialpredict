package http

import (
	"encoding/json"
	"net/http"
	"time"

	"socialpredict/handlers"
	"socialpredict/handlers/authhttp"
	"socialpredict/handlers/cms/reportingvisibility"
	dusers "socialpredict/internal/domain/users"
	authsvc "socialpredict/internal/service/auth"
	"socialpredict/models"
)

type Handler struct {
	svc  *reportingvisibility.Service
	auth authsvc.Authenticator
}

func NewHandler(svc *reportingvisibility.Service, auth authsvc.Authenticator) *Handler {
	return &Handler{svc: svc, auth: auth}
}

type updateReq struct {
	SystemMetricsPublic     *bool `json:"systemMetricsPublic"`
	GlobalLeaderboardPublic *bool `json:"globalLeaderboardPublic"`
	Version                 uint  `json:"version"`
}

type settingsResponse struct {
	SystemMetricsPublic     bool      `json:"systemMetricsPublic"`
	GlobalLeaderboardPublic bool      `json:"globalLeaderboardPublic"`
	Version                 uint      `json:"version"`
	UpdatedAt               time.Time `json:"updatedAt"`
}

func (h *Handler) PublicGet(w http.ResponseWriter, _ *http.Request) {
	item, err := h.svc.GetSettings()
	if err != nil {
		_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
		return
	}
	_ = handlers.WriteResult(w, http.StatusOK, responseFromSettings(item))
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

	item, err := h.svc.UpdateSettings(reportingvisibility.UpdateInput{
		SystemMetricsPublic:     in.SystemMetricsPublic,
		GlobalLeaderboardPublic: in.GlobalLeaderboardPublic,
		Version:                 in.Version,
		UpdatedBy:               admin.Username,
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

func responseFromSettings(item *models.ReportingVisibilitySettings) settingsResponse {
	if item == nil {
		item = reportingvisibility.DefaultSettings()
	}
	return settingsResponse{
		SystemMetricsPublic:     item.SystemMetricsPublic,
		GlobalLeaderboardPublic: item.GlobalLeaderboardPublic,
		Version:                 item.Version,
		UpdatedAt:               item.UpdatedAt,
	}
}
