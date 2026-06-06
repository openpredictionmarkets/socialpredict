package adminhandlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"socialpredict/handlers"
	dmarkets "socialpredict/internal/domain/markets"
	authsvc "socialpredict/internal/service/auth"
)

type marketTagManager interface {
	ListMarketTags(ctx context.Context, includeInactive bool) ([]dmarkets.MarketTag, error)
	CreateMarketTag(ctx context.Context, req dmarkets.MarketTagRequest, actorUsername string) (*dmarkets.MarketTag, error)
	UpdateMarketTag(ctx context.Context, slug string, req dmarkets.MarketTagRequest) (*dmarkets.MarketTag, error)
}

type adminMarketTagRequest struct {
	Slug              string `json:"slug"`
	DisplayName       string `json:"displayName"`
	Description       string `json:"description"`
	ColorKey          string `json:"colorKey"`
	SortOrder         int    `json:"sortOrder"`
	IsActive          *bool  `json:"isActive"`
	ConfirmDeactivate bool   `json:"confirmDeactivate"`
}

type marketTagResponse struct {
	ID          int64  `json:"id"`
	Slug        string `json:"slug"`
	DisplayName string `json:"displayName"`
	Description string `json:"description,omitempty"`
	ColorKey    string `json:"colorKey,omitempty"`
	SortOrder   int    `json:"sortOrder"`
	IsActive    bool   `json:"isActive"`
	CreatedBy   string `json:"createdBy,omitempty"`
}

type marketTagsResponse struct {
	Tags  []marketTagResponse `json:"tags"`
	Total int                 `json:"total"`
}

func ListAdminMarketTagsHandler(svc marketTagManager, auth authsvc.Authenticator) http.HandlerFunc {
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

		includeInactive, _ := strconv.ParseBool(r.URL.Query().Get("includeInactive"))
		tags, err := svc.ListMarketTags(r.Context(), includeInactive)
		if err != nil {
			writeMarketTagError(w, err)
			return
		}
		_ = handlers.WriteResult(w, http.StatusOK, marketTagsResponse{
			Tags:  marketTagResponses(tags),
			Total: len(tags),
		})
	}
}

func CreateAdminMarketTagHandler(svc marketTagManager, auth authsvc.Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
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

		var req adminMarketTagRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
			return
		}
		tag, err := svc.CreateMarketTag(r.Context(), marketTagRequestFromAdmin(req), admin.Username)
		if err != nil {
			writeMarketTagError(w, err)
			return
		}
		_ = handlers.WriteResult(w, http.StatusCreated, marketTagResponseFromDomain(*tag))
	}
}

func UpdateAdminMarketTagHandler(svc marketTagManager, auth authsvc.Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
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

		var req adminMarketTagRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
			return
		}
		if req.IsActive != nil && !*req.IsActive && !req.ConfirmDeactivate {
			_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonValidationFailed)
			return
		}

		tag, err := svc.UpdateMarketTag(r.Context(), mux.Vars(r)["slug"], marketTagRequestFromAdmin(req))
		if err != nil {
			writeMarketTagError(w, err)
			return
		}
		_ = handlers.WriteResult(w, http.StatusOK, marketTagResponseFromDomain(*tag))
	}
}

func writeMarketTagError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, dmarkets.ErrInvalidInput):
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonValidationFailed)
	default:
		_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
	}
}

func marketTagRequestFromAdmin(req adminMarketTagRequest) dmarkets.MarketTagRequest {
	return dmarkets.MarketTagRequest{
		Slug:        req.Slug,
		DisplayName: req.DisplayName,
		Description: req.Description,
		ColorKey:    req.ColorKey,
		SortOrder:   req.SortOrder,
		IsActive:    req.IsActive,
	}
}

func marketTagResponses(tags []dmarkets.MarketTag) []marketTagResponse {
	responses := make([]marketTagResponse, 0, len(tags))
	for _, tag := range tags {
		responses = append(responses, marketTagResponseFromDomain(tag))
	}
	return responses
}

func marketTagResponseFromDomain(tag dmarkets.MarketTag) marketTagResponse {
	return marketTagResponse{
		ID:          tag.ID,
		Slug:        tag.Slug,
		DisplayName: tag.DisplayName,
		Description: tag.Description,
		ColorKey:    tag.ColorKey,
		SortOrder:   tag.SortOrder,
		IsActive:    tag.IsActive,
		CreatedBy:   tag.CreatedBy,
	}
}
