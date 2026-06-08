package adminhandlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"socialpredict/handlers"
	dmarkets "socialpredict/internal/domain/markets"
	authsvc "socialpredict/internal/service/auth"
)

type marketDescriptionAmendmentReviewer interface {
	GetMarketGovernanceSettings(ctx context.Context) (*dmarkets.MarketGovernanceSettings, error)
	UpdateMarketGovernanceSettings(ctx context.Context, update dmarkets.MarketGovernanceSettingsUpdate) (*dmarkets.MarketGovernanceSettings, error)
	ListMarketDescriptionAmendments(ctx context.Context, filters dmarkets.MarketDescriptionAmendmentFilters) ([]dmarkets.MarketDescriptionAmendment, error)
	ReviewMarketDescriptionAmendment(ctx context.Context, amendmentID int64, status string, actorUsername string, reason string) (*dmarkets.MarketDescriptionAmendment, error)
}

type reviewMarketDescriptionAmendmentRequest struct {
	Status string `json:"status"`
	Reason string `json:"reason"`
}

type marketDescriptionAmendmentResponse struct {
	ID                         int64                                `json:"id"`
	MarketID                   int64                                `json:"marketId"`
	MarketTitle                string                               `json:"marketTitle,omitempty"`
	MarketDescription          string                               `json:"marketDescription,omitempty"`
	Version                    int                                  `json:"version"`
	Body                       string                               `json:"body"`
	BodyFormat                 string                               `json:"bodyFormat"`
	Status                     string                               `json:"status"`
	CreatedBy                  string                               `json:"createdBy"`
	CreatedAt                  time.Time                            `json:"createdAt"`
	UpdatedAt                  time.Time                            `json:"updatedAt"`
	ApprovedBy                 string                               `json:"approvedBy,omitempty"`
	ApprovedAt                 *time.Time                           `json:"approvedAt,omitempty"`
	RejectedBy                 string                               `json:"rejectedBy,omitempty"`
	RejectedAt                 *time.Time                           `json:"rejectedAt,omitempty"`
	RejectionReason            string                               `json:"rejectionReason,omitempty"`
	SubmitReason               string                               `json:"submitReason,omitempty"`
	PreviousApprovedAmendments []marketDescriptionAmendmentResponse `json:"previousApprovedAmendments,omitempty"`
}

type marketDescriptionAmendmentListResponse struct {
	Amendments []marketDescriptionAmendmentResponse `json:"amendments"`
	Total      int                                  `json:"total"`
}

type marketGovernanceSettingsResponse struct {
	AutoApproveDescriptionAmendments bool      `json:"autoApproveDescriptionAmendments"`
	AutoApproveMarketProposals       bool      `json:"autoApproveMarketProposals"`
	Version                          uint      `json:"version"`
	UpdatedBy                        string    `json:"updatedBy,omitempty"`
	UpdatedAt                        time.Time `json:"updatedAt"`
}

type updateMarketGovernanceSettingsRequest struct {
	AutoApproveDescriptionAmendments *bool `json:"autoApproveDescriptionAmendments"`
	AutoApproveMarketProposals       *bool `json:"autoApproveMarketProposals"`
	Version                          uint  `json:"version"`
}

func GetMarketGovernanceSettingsHandler(svc marketDescriptionAmendmentReviewer, auth authsvc.Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			_ = handlers.WriteFailure(w, http.StatusMethodNotAllowed, handlers.ReasonMethodNotAllowed)
			return
		}
		if _, ok := requireAdminForMarketReview(w, r, auth); !ok {
			return
		}
		if svc == nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
			return
		}
		settings, err := svc.GetMarketGovernanceSettings(r.Context())
		if err != nil {
			writeMarketReviewError(w, err)
			return
		}
		_ = handlers.WriteResult(w, http.StatusOK, marketGovernanceSettingsResponseFromDomain(settings))
	}
}

func UpdateMarketGovernanceSettingsHandler(svc marketDescriptionAmendmentReviewer, auth authsvc.Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			_ = handlers.WriteFailure(w, http.StatusMethodNotAllowed, handlers.ReasonMethodNotAllowed)
			return
		}
		admin, ok := requireAdminForMarketReview(w, r, auth)
		if !ok {
			return
		}
		if svc == nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
			return
		}
		var req updateMarketGovernanceSettingsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
			return
		}
		settings, err := svc.UpdateMarketGovernanceSettings(r.Context(), dmarkets.MarketGovernanceSettingsUpdate{
			AutoApproveDescriptionAmendments: req.AutoApproveDescriptionAmendments,
			AutoApproveMarketProposals:       req.AutoApproveMarketProposals,
			Version:                          req.Version,
			UpdatedBy:                        admin.Username,
		})
		if err != nil {
			writeMarketReviewError(w, err)
			return
		}
		_ = handlers.WriteResult(w, http.StatusOK, marketGovernanceSettingsResponseFromDomain(settings))
	}
}

func ListMarketDescriptionAmendmentsHandler(svc marketDescriptionAmendmentReviewer, auth authsvc.Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			_ = handlers.WriteFailure(w, http.StatusMethodNotAllowed, handlers.ReasonMethodNotAllowed)
			return
		}
		if _, ok := requireAdminForMarketReview(w, r, auth); !ok {
			return
		}
		if svc == nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
			return
		}
		filters, ok := parseAdminAmendmentFilters(w, r)
		if !ok {
			return
		}
		items, err := svc.ListMarketDescriptionAmendments(r.Context(), filters)
		if err != nil {
			writeMarketReviewError(w, err)
			return
		}
		response := marketDescriptionAmendmentListResponse{
			Amendments: make([]marketDescriptionAmendmentResponse, 0, len(items)),
			Total:      len(items),
		}
		for _, item := range items {
			response.Amendments = append(response.Amendments, marketDescriptionAmendmentResponseFromDomain(item))
		}
		_ = handlers.WriteResult(w, http.StatusOK, response)
	}
}

func ReviewMarketDescriptionAmendmentHandler(svc marketDescriptionAmendmentReviewer, auth authsvc.Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			_ = handlers.WriteFailure(w, http.StatusMethodNotAllowed, handlers.ReasonMethodNotAllowed)
			return
		}
		admin, ok := requireAdminForMarketReview(w, r, auth)
		if !ok {
			return
		}
		if svc == nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
			return
		}
		amendmentID, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
		if err != nil || amendmentID <= 0 {
			_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
			return
		}
		var req reviewMarketDescriptionAmendmentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
			return
		}
		amendment, err := svc.ReviewMarketDescriptionAmendment(r.Context(), amendmentID, req.Status, admin.Username, req.Reason)
		if err != nil {
			writeMarketReviewError(w, err)
			return
		}
		_ = handlers.WriteResult(w, http.StatusOK, marketDescriptionAmendmentResponseFromDomain(*amendment))
	}
}

func parseAdminAmendmentFilters(w http.ResponseWriter, r *http.Request) (dmarkets.MarketDescriptionAmendmentFilters, bool) {
	query := r.URL.Query()
	status := dmarkets.NormalizeDescriptionAmendmentStatus(query.Get("status"))
	switch status {
	case dmarkets.DescriptionAmendmentStatusPending, dmarkets.DescriptionAmendmentStatusApproved, dmarkets.DescriptionAmendmentStatusRejected:
	default:
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
		return dmarkets.MarketDescriptionAmendmentFilters{}, false
	}
	marketID := int64(0)
	if raw := strings.TrimSpace(query.Get("marketId")); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || parsed <= 0 {
			_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
			return dmarkets.MarketDescriptionAmendmentFilters{}, false
		}
		marketID = parsed
	}
	return dmarkets.MarketDescriptionAmendmentFilters{
		MarketID: marketID,
		Status:   status,
		Limit:    boundedAdminReviewInt(query.Get("limit"), 50, 1, 200),
		Offset:   boundedAdminReviewInt(query.Get("offset"), 0, 0, 100000),
	}, true
}

func marketGovernanceSettingsResponseFromDomain(settings *dmarkets.MarketGovernanceSettings) marketGovernanceSettingsResponse {
	if settings == nil {
		return marketGovernanceSettingsResponse{Version: 1}
	}
	return marketGovernanceSettingsResponse{
		AutoApproveDescriptionAmendments: settings.AutoApproveDescriptionAmendments,
		AutoApproveMarketProposals:       settings.AutoApproveMarketProposals,
		Version:                          settings.Version,
		UpdatedBy:                        settings.UpdatedBy,
		UpdatedAt:                        settings.UpdatedAt,
	}
}

func marketDescriptionAmendmentResponseFromDomain(item dmarkets.MarketDescriptionAmendment) marketDescriptionAmendmentResponse {
	response := marketDescriptionAmendmentResponse{
		ID:                item.ID,
		MarketID:          item.MarketID,
		MarketTitle:       item.MarketTitle,
		MarketDescription: item.MarketDescription,
		Version:           item.Version,
		Body:              item.Body,
		BodyFormat:        item.BodyFormat,
		Status:            item.Status,
		CreatedBy:         item.CreatedBy,
		CreatedAt:         item.CreatedAt,
		UpdatedAt:         item.UpdatedAt,
		ApprovedBy:        item.ApprovedBy,
		ApprovedAt:        item.ApprovedAt,
		RejectedBy:        item.RejectedBy,
		RejectedAt:        item.RejectedAt,
		RejectionReason:   item.RejectionReason,
		SubmitReason:      item.SubmitReason,
	}
	if len(item.PreviousApprovedAmendments) > 0 {
		response.PreviousApprovedAmendments = make([]marketDescriptionAmendmentResponse, 0, len(item.PreviousApprovedAmendments))
		for _, previous := range item.PreviousApprovedAmendments {
			response.PreviousApprovedAmendments = append(response.PreviousApprovedAmendments, marketDescriptionAmendmentResponseFromDomain(previous))
		}
	}
	return response
}
