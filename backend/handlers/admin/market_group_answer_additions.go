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

type marketGroupAnswerAdditionReviewer interface {
	ListAdminMarketGroupAnswerAdditionRows(ctx context.Context, filters dmarkets.AdminAnswerAdditionReviewFilters) (*dmarkets.AdminAnswerAdditionReviewPage, error)
	ApproveMarketGroupAnswerAddition(ctx context.Context, additionID int64, actorUsername string, confirmed bool) (*dmarkets.MarketGroupAnswerAddition, error)
	RejectMarketGroupAnswerAddition(ctx context.Context, additionID int64, actorUsername string, reason string) (*dmarkets.MarketGroupAnswerAddition, error)
}

type marketGroupAnswerAdditionReviewRequest struct {
	Status  string `json:"status"`
	Reason  string `json:"reason"`
	Confirm bool   `json:"confirm"`
}

type marketGroupAnswerAdditionReviewResponse struct {
	ID              int64                      `json:"id"`
	GroupID         int64                      `json:"groupId"`
	MarketID        int64                      `json:"marketId,omitempty"`
	GroupTitle      string                     `json:"groupTitle,omitempty"`
	AnswerLabel     string                     `json:"answerLabel"`
	Status          string                     `json:"status"`
	ProposedBy      string                     `json:"proposedBy"`
	ReviewedBy      string                     `json:"reviewedBy,omitempty"`
	ReviewedAt      *time.Time                 `json:"reviewedAt,omitempty"`
	RejectionReason string                     `json:"rejectionReason,omitempty"`
	AdditionCost    int64                      `json:"additionCost"`
	CreatedAt       time.Time                  `json:"createdAt,omitempty"`
	UpdatedAt       time.Time                  `json:"updatedAt,omitempty"`
	MarketGroup     *marketGroupReviewResponse `json:"marketGroup,omitempty"`
}

type marketGroupAnswerAdditionListResponse struct {
	Additions []marketGroupAnswerAdditionReviewResponse `json:"additions"`
	Total     int                                       `json:"total"`
	Limit     int                                       `json:"limit,omitempty"`
	Offset    int                                       `json:"offset,omitempty"`
}

func ListMarketGroupAnswerAdditionsHandler(svc marketGroupAnswerAdditionReviewer, auth authsvc.Authenticator) http.HandlerFunc {
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
		filters, ok := parseAdminAnswerAdditionFilters(w, r)
		if !ok {
			return
		}
		page, err := svc.ListAdminMarketGroupAnswerAdditionRows(r.Context(), filters)
		if err != nil {
			writeMarketReviewError(w, err)
			return
		}
		if page == nil {
			page = &dmarkets.AdminAnswerAdditionReviewPage{Limit: filters.Limit, Offset: filters.Offset}
		}
		response := marketGroupAnswerAdditionListResponse{
			Additions: marketGroupAnswerAdditionResponsesFromDomain(page.Rows),
			Total:     page.Total,
			Limit:     page.Limit,
			Offset:    page.Offset,
		}
		_ = handlers.WriteResult(w, http.StatusOK, response)
	}
}

func ReviewMarketGroupAnswerAdditionHandler(svc marketGroupAnswerAdditionReviewer, auth authsvc.Authenticator) http.HandlerFunc {
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
		additionID, ok := marketGroupAnswerAdditionIDFromRequest(w, r)
		if !ok {
			return
		}
		var req marketGroupAnswerAdditionReviewRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
			return
		}
		status := dmarkets.NormalizeMarketGroupAnswerAdditionStatus(req.Status)
		var (
			addition *dmarkets.MarketGroupAnswerAddition
			err      error
		)
		switch status {
		case dmarkets.MarketGroupAnswerAdditionStatusApproved:
			addition, err = svc.ApproveMarketGroupAnswerAddition(r.Context(), additionID, admin.Username, req.Confirm)
		case dmarkets.MarketGroupAnswerAdditionStatusRejected:
			addition, err = svc.RejectMarketGroupAnswerAddition(r.Context(), additionID, admin.Username, req.Reason)
		default:
			_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
			return
		}
		if err != nil {
			writeMarketReviewError(w, err)
			return
		}
		_ = handlers.WriteResult(w, http.StatusOK, marketGroupAnswerAdditionReviewResponseFromDomainValue(addition))
	}
}

func parseAdminAnswerAdditionFilters(w http.ResponseWriter, r *http.Request) (dmarkets.AdminAnswerAdditionReviewFilters, bool) {
	query := r.URL.Query()
	status := dmarkets.NormalizeMarketGroupAnswerAdditionStatus(query.Get("status"))
	switch status {
	case dmarkets.MarketGroupAnswerAdditionStatusPending, dmarkets.MarketGroupAnswerAdditionStatusApproved, dmarkets.MarketGroupAnswerAdditionStatusRejected:
	default:
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
		return dmarkets.AdminAnswerAdditionReviewFilters{}, false
	}
	groupID := int64(0)
	if rawGroupID := query.Get("groupId"); rawGroupID != "" {
		parsed, err := strconv.ParseInt(rawGroupID, 10, 64)
		if err != nil || parsed <= 0 {
			_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
			return dmarkets.AdminAnswerAdditionReviewFilters{}, false
		}
		groupID = parsed
	}
	searchQuery := strings.TrimSpace(query.Get("query"))
	if len([]rune(searchQuery)) > 100 {
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
		return dmarkets.AdminAnswerAdditionReviewFilters{}, false
	}
	limit, ok := parseBoundedAdminReviewInt(query.Get("limit"), 50, 1, 200)
	if !ok {
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
		return dmarkets.AdminAnswerAdditionReviewFilters{}, false
	}
	offset, ok := parseBoundedAdminReviewInt(query.Get("offset"), 0, 0, 100000)
	if !ok {
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
		return dmarkets.AdminAnswerAdditionReviewFilters{}, false
	}
	return dmarkets.AdminAnswerAdditionReviewFilters{
		GroupID: groupID,
		Status:  status,
		Query:   searchQuery,
		Limit:   limit,
		Offset:  offset,
	}, true
}

func marketGroupAnswerAdditionIDFromRequest(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil || id <= 0 {
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
		return 0, false
	}
	return id, true
}

func marketGroupAnswerAdditionReviewResponseFromDomainValue(item *dmarkets.MarketGroupAnswerAddition) marketGroupAnswerAdditionReviewResponse {
	if item == nil {
		return marketGroupAnswerAdditionReviewResponse{}
	}
	return marketGroupAnswerAdditionReviewResponseFromDomain(*item)
}

func marketGroupAnswerAdditionResponsesFromDomain(items []dmarkets.MarketGroupAnswerAddition) []marketGroupAnswerAdditionReviewResponse {
	responses := make([]marketGroupAnswerAdditionReviewResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, marketGroupAnswerAdditionReviewResponseFromDomain(item))
	}
	return responses
}

func marketGroupAnswerAdditionReviewResponseFromDomain(item dmarkets.MarketGroupAnswerAddition) marketGroupAnswerAdditionReviewResponse {
	return marketGroupAnswerAdditionReviewResponse{
		ID:              item.ID,
		GroupID:         item.GroupID,
		MarketID:        item.MarketID,
		GroupTitle:      item.GroupTitle,
		AnswerLabel:     item.AnswerLabel,
		Status:          item.Status,
		ProposedBy:      item.ProposedBy,
		ReviewedBy:      item.ReviewedBy,
		ReviewedAt:      item.ReviewedAt,
		RejectionReason: item.RejectionReason,
		AdditionCost:    item.AdditionCost,
		CreatedAt:       item.CreatedAt,
		UpdatedAt:       item.UpdatedAt,
		MarketGroup:     marketGroupReviewResponsePtr(item.MarketGroup),
	}
}

func marketGroupReviewResponsePtr(group *dmarkets.MarketGroup) *marketGroupReviewResponse {
	if group == nil {
		return nil
	}
	response := marketGroupReviewResponseFromGroup(group)
	return &response
}
