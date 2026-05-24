package adminhandlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"socialpredict/handlers"
	"socialpredict/handlers/authhttp"
	dmarkets "socialpredict/internal/domain/markets"
	dusers "socialpredict/internal/domain/users"
	authsvc "socialpredict/internal/service/auth"
)

type marketReviewer interface {
	ApproveProposedMarket(ctx context.Context, marketID int64, actorUsername string, confirmed bool) (*dmarkets.Market, error)
	RejectProposedMarket(ctx context.Context, marketID int64, actorUsername string, reason string) (*dmarkets.Market, error)
}

type approveMarketRequest struct {
	Confirm bool `json:"confirm"`
}

type rejectMarketRequest struct {
	Reason string `json:"reason"`
}

type marketReviewResponse struct {
	ID              int64  `json:"id"`
	Status          string `json:"status"`
	LifecycleStatus string `json:"lifecycleStatus"`
	ApprovedBy      string `json:"approvedBy,omitempty"`
	RejectedBy      string `json:"rejectedBy,omitempty"`
	RejectionReason string `json:"rejectionReason,omitempty"`
}

func ApproveMarketHandler(svc marketReviewer, auth authsvc.Authenticator) http.HandlerFunc {
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
		marketID, ok := marketIDFromRequest(w, r)
		if !ok {
			return
		}
		var req approveMarketRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
			return
		}

		market, err := svc.ApproveProposedMarket(r.Context(), marketID, admin.Username, req.Confirm)
		if err != nil {
			writeMarketReviewError(w, err)
			return
		}
		_ = handlers.WriteResult(w, http.StatusOK, marketReviewResponseFromMarket(market))
	}
}

func RejectMarketHandler(svc marketReviewer, auth authsvc.Authenticator) http.HandlerFunc {
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
		marketID, ok := marketIDFromRequest(w, r)
		if !ok {
			return
		}
		var req rejectMarketRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
			return
		}

		market, err := svc.RejectProposedMarket(r.Context(), marketID, admin.Username, req.Reason)
		if err != nil {
			writeMarketReviewError(w, err)
			return
		}
		_ = handlers.WriteResult(w, http.StatusOK, marketReviewResponseFromMarket(market))
	}
}

func requireAdminForMarketReview(w http.ResponseWriter, r *http.Request, auth authsvc.Authenticator) (*dusers.User, bool) {
	if auth == nil {
		_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
		return nil, false
	}
	admin, authErr := auth.RequireAdmin(r)
	if authErr != nil {
		_ = authhttp.WriteFailure(w, authErr)
		return nil, false
	}
	return admin, true
}

func marketIDFromRequest(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil || id <= 0 {
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
		return 0, false
	}
	return id, true
}

func writeMarketReviewError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, dmarkets.ErrMarketNotFound):
		_ = handlers.WriteFailure(w, http.StatusNotFound, handlers.ReasonMarketNotFound)
	case errors.Is(err, dmarkets.ErrUnauthorized):
		_ = handlers.WriteFailure(w, http.StatusForbidden, handlers.ReasonAuthorizationDenied)
	case errors.Is(err, dmarkets.ErrInvalidState):
		_ = handlers.WriteFailure(w, http.StatusConflict, handlers.ReasonInvalidState)
	case errors.Is(err, dmarkets.ErrInvalidInput):
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonValidationFailed)
	default:
		_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
	}
}

func marketReviewResponseFromMarket(market *dmarkets.Market) marketReviewResponse {
	if market == nil {
		return marketReviewResponse{}
	}
	return marketReviewResponse{
		ID:              market.ID,
		Status:          market.Status,
		LifecycleStatus: market.LifecycleStatus,
		ApprovedBy:      market.ApprovedBy,
		RejectedBy:      market.RejectedBy,
		RejectionReason: market.RejectionReason,
	}
}
